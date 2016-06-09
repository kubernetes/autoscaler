/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/contrib/cluster-autoscaler/utils/gce"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_record "k8s.io/kubernetes/pkg/client/record"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	migConfigFlag           config.MigConfigFlag
	address                 = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubernetes              = flag.String("kubernetes", "", "Kuberentes master location. Leave blank for default")
	cloudConfig             = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	verifyUnschedulablePods = flag.Bool("verify-unschedulable-pods", true,
		"If enabled CA will ensure that each pod marked by Scheduler as unschedulable actually can't be scheduled on any node."+
			"This prevents from adding unnecessary nodes in situation when CA and Scheduler have different configuration.")
	scaleDownEnabled = flag.Bool("scale-down-enabled", true, "Should CA scale down the cluster")
	scaleDownDelay   = flag.Duration("scale-down-delay", 10*time.Minute,
		"Duration from the last scale up to the time when CA starts to check scale down options")
	scaleDownUnderutilizedTime = flag.Duration("scale-down-underutilized-time", 10*time.Minute,
		"How long the node should be underutilized before it is eligible for scale down")
	scaleDownUtilizationThreshold = flag.Float64("scale-down-utilization-threshold", 0.5,
		"Node reservation level below which a node can be considered for scale down")
	scaleDownTrialFrequency = flag.Duration("scale-down-trial-frequency", 10*time.Minute,
		"How often scale down possiblity is check")
)

func main() {
	flag.Var(&migConfigFlag, "nodes", "sets min,max size and url of a MIG to be controlled by Cluster Autoscaler. "+
		"Can be used multiple times. Format: <min>:<max>:<migurl>")
	flag.Parse()

	go func() {
		http.Handle("/metrics", prometheus.Handler())
		err := http.ListenAndServe(*address, nil)
		glog.Fatalf("Failed to start metrics: %v", err)
	}()

	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kuberentes url: %v", err)
	}

	// Configuration
	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client configuration: %v", err)
	}
	migConfigs := make([]*config.MigConfig, 0, len(migConfigFlag))
	for i := range migConfigFlag {
		migConfigs = append(migConfigs, &migConfigFlag[i])
	}

	// GCE Manager
	var gceManager *gce.GceManager
	if *cloudConfig != "" {
		config, err := os.Open(*cloudConfig)
		if err != nil {
			glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", *cloudConfig, err)
		}
		defer config.Close()
		gceManager, err = gce.CreateGceManager(migConfigs, config)
	} else {
		gceManager, err = gce.CreateGceManager(migConfigs, nil)
	}
	if err != nil {
		glog.Fatalf("Failed to create GCE Manager: %v", err)
	}

	kubeClient := kube_client.NewOrDie(kubeConfig)

	predicateChecker, err := simulator.NewPredicateChecker(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create predicate checker: %v", err)
	}
	unschedulablePodLister := NewUnschedulablePodLister(kubeClient)
	scheduledPodLister := NewScheduledPodLister(kubeClient)
	nodeLister := NewNodeLister(kubeClient)

	lastScaleUpTime := time.Now()
	lastScaleDownFailedTrial := time.Now()
	underutilizedNodes := make(map[string]time.Time)

	eventBroadcaster := kube_record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(""))
	recorder := eventBroadcaster.NewRecorder(kube_api.EventSource{Component: "cluster-autoscaler"})

	for {
		select {
		case <-time.After(time.Minute):
			{
				loopStart := time.Now()
				updateLastTime("main")

				nodes, err := nodeLister.List()
				if err != nil {
					glog.Errorf("Failed to list nodes: %v", err)
					continue
				}
				if len(nodes) == 0 {
					glog.Errorf("No nodes in the cluster")
					continue
				}

				if err := CheckMigsAndNodes(nodes, gceManager); err != nil {
					glog.Warningf("Cluster is not ready for autoscaling: %v", err)
					continue
				}

				allUnschedulablePods, err := unschedulablePodLister.List()
				if err != nil {
					glog.Errorf("Failed to list unscheduled pods: %v", err)
					continue
				}

				allScheduled, err := scheduledPodLister.List()
				if err != nil {
					glog.Errorf("Failed to list scheduled pods: %v", err)
					continue
				}

				// We need to reset all pods that have been marked as unschedulable not after
				// the newest node became available for the scheduler.
				allNodesAvailableTime := GetAllNodesAvailableTime(nodes)
				podsToReset, unschedulablePodsToHelp := SlicePodsByPodScheduledTime(allUnschedulablePods, allNodesAvailableTime)
				ResetPodScheduledCondition(kubeClient, podsToReset)

				// We need to check whether pods marked as unschedulable are actually unschedulable.
				// This should prevent from adding unnecessary nodes. Example of such situation:
				// - CA and Scheduler has slightly different configuration
				// - Scheduler can't schedule a pod and marks it as unschedulable
				// - CA added a node which should help the pod
				// - Scheduler doesn't schedule the pod on the new node
				//   because according to it logic it doesn't fit there
				// - CA see the pod is still unschedulable, so it adds another node to help it
				//
				// With the check enabled the last point won't happen because CA will ignore a pod
				// which is supposed to schedule on an existing node.
				//
				// Without below check cluster might be unnecessary scaled up to the max allowed size
				// in the describe situation.
				schedulablePodsPresent := false
				if *verifyUnschedulablePods {
					newUnschedulablePodsToHelp := FilterOutSchedulable(unschedulablePodsToHelp, nodes, allScheduled, predicateChecker)

					if len(newUnschedulablePodsToHelp) != len(unschedulablePodsToHelp) {
						schedulablePodsPresent = true
					}
					unschedulablePodsToHelp = newUnschedulablePodsToHelp
				}

				if len(unschedulablePodsToHelp) == 0 {
					glog.V(1).Info("No unschedulable pods")
				} else {
					scaleUpStart := time.Now()
					updateLastTime("scaleup")
					scaledUp, err := ScaleUp(unschedulablePodsToHelp, nodes, migConfigs, gceManager, kubeClient, predicateChecker, recorder)

					updateDuration("scaleup", scaleUpStart)

					if err != nil {
						glog.Errorf("Failed to scale up: %v", err)
						continue
					} else {
						if scaledUp {
							lastScaleUpTime = time.Now()
							// No scale down in this iteration.
							continue
						}
					}
				}

				if *scaleDownEnabled {
					utilizationStart := time.Now()

					// In dry run only utilization is updated
					calculateUtilizationOnly := lastScaleUpTime.Add(*scaleDownDelay).After(time.Now()) ||
						lastScaleDownFailedTrial.Add(*scaleDownTrialFrequency).After(time.Now()) ||
						schedulablePodsPresent

					updateLastTime("utilization")

					underutilizedNodes = CalculateUnderutilizedNodes(
						nodes,
						underutilizedNodes,
						*scaleDownUtilizationThreshold,
						allScheduled,
						predicateChecker)

					updateDuration("utilization", utilizationStart)

					if !calculateUtilizationOnly {
						scaleDownStart := time.Now()
						updateLastTime("scaledown")

						result, err := ScaleDown(
							nodes,
							underutilizedNodes,
							*scaleDownUnderutilizedTime,
							allScheduled,
							gceManager, kubeClient, predicateChecker)

						updateDuration("scaledown", scaleDownStart)

						if err != nil {
							glog.Errorf("Failed to scale down: %v", err)
						} else {
							if result == ScaleDownNodeDeleted {
								// Clean the utilization map to be super sure that the simulated
								// deletions are made in the new context.
								underutilizedNodes = make(map[string]time.Time, len(underutilizedNodes))
							} else {
								lastScaleDownFailedTrial = time.Now()
							}
						}
					}
				}
				updateDuration("main", loopStart)
			}
		}
	}
}

func updateDuration(label string, start time.Time) {
	duration.WithLabelValues(label).Observe(durationToMicro(start))
	lastDuration.WithLabelValues(label).Set(durationToMicro(start))
}

func updateLastTime(label string) {
	lastTimestamp.WithLabelValues(label).Set(float64(time.Now().Unix()))
}
