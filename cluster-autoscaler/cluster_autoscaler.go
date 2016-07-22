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
	"strings"
	"time"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	kube_util "k8s.io/contrib/cluster-autoscaler/utils/kubernetes"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_leaderelection "k8s.io/kubernetes/pkg/client/leaderelection"
	kube_record "k8s.io/kubernetes/pkg/client/record"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	kube_flag "k8s.io/kubernetes/pkg/util/flag"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

// MultiStringFlag is a flag for passing multiple parameters using same flag
type MultiStringFlag []string

// String returns string representation of the node groups.
func (flag *MultiStringFlag) String() string {
	return "[" + strings.Join(*flag, " ") + "]"
}

// Set adds a new configuration.
func (flag *MultiStringFlag) Set(value string) error {
	*flag = append(*flag, value)
	return nil
}

var (
	nodeGroupsFlag          MultiStringFlag
	address                 = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubernetes              = flag.String("kubernetes", "", "Kuberentes master location. Leave blank for default")
	cloudConfig             = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	verifyUnschedulablePods = flag.Bool("verify-unschedulable-pods", true,
		"If enabled CA will ensure that each pod marked by Scheduler as unschedulable actually can't be scheduled on any node."+
			"This prevents from adding unnecessary nodes in situation when CA and Scheduler have different configuration.")
	scaleDownEnabled = flag.Bool("scale-down-enabled", true, "Should CA scale down the cluster")
	scaleDownDelay   = flag.Duration("scale-down-delay", 10*time.Minute,
		"Duration from the last scale up to the time when CA starts to check scale down options")
	scaleDownUnneededTime = flag.Duration("scale-down-unneeded-time", 10*time.Minute,
		"How long the node should be unneeded before it is eligible for scale down")
	scaleDownUtilizationThreshold = flag.Float64("scale-down-utilization-threshold", 0.5,
		"Node utilization level, defined as sum of requested resources divided by capacity, below which a node can be considered for scale down")
	scaleDownTrialInterval = flag.Duration("scale-down-trial-interval", 1*time.Minute,
		"How often scale down possiblity is check")
	scanInterval = flag.Duration("scan-interval", 10*time.Second, "How often cluster is reevaluated for scale up or down")

	cloudProviderFlag = flag.String("cloud-provider", "gce", "Cloud provider type. Allowed values: gce")
)

func createKubeClient() *kube_client.Client {
	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kuberentes url: %v", err)
	}

	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client configuration: %v", err)
	}

	return kube_client.NewOrDie(kubeConfig)
}

func createEventRecorder(kubeClient *kube_client.Client) kube_record.EventRecorder {
	eventBroadcaster := kube_record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(""))
	return eventBroadcaster.NewRecorder(kube_api.EventSource{Component: "cluster-autoscaler"})
}

// In order to meet interface criteria for LeaderElectionConfig we need to
// take stop channell as an argument. However, since we are committing a suicide
// after loosing mastership we can safely ignore it.
func run(_ <-chan struct{}) {
	kubeClient := createKubeClient()

	predicateChecker, err := simulator.NewPredicateChecker(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create predicate checker: %v", err)
	}
	unschedulablePodLister := kube_util.NewUnschedulablePodLister(kubeClient, kube_api.NamespaceAll)
	scheduledPodLister := kube_util.NewScheduledPodLister(kubeClient)
	nodeLister := kube_util.NewNodeLister(kubeClient)

	lastScaleUpTime := time.Now()
	lastScaleDownFailedTrial := time.Now()
	unneededNodes := make(map[string]time.Time)
	podLocationHints := make(map[string]string)
	usageTracker := simulator.NewUsageTracker()

	recorder := createEventRecorder(kubeClient)

	var cloudProvider cloudprovider.CloudProvider

	if *cloudProviderFlag == "gce" {
		// GCE Manager
		var gceManager *gce.GceManager
		var gceError error
		if *cloudConfig != "" {
			config, fileErr := os.Open(*cloudConfig)
			if fileErr != nil {
				glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", *cloudConfig, err)
			}
			defer config.Close()
			gceManager, gceError = gce.CreateGceManager(config)
		} else {
			gceManager, gceError = gce.CreateGceManager(nil)
		}
		if gceError != nil {
			glog.Fatalf("Failed to create GCE Manager: %v", err)
		}
		cloudProvider, err = gce.BuildGceCloudProvider(gceManager, nodeGroupsFlag)
		if err != nil {
			glog.Fatalf("Failed to create GCE cloud provider: %v", err)
		}
	}

	for {
		select {
		case <-time.After(*scanInterval):
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

				if err := CheckGroupsAndNodes(nodes, cloudProvider); err != nil {
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
						glog.V(2).Info("Schedulable pods present")
						schedulablePodsPresent = true
					}
					unschedulablePodsToHelp = newUnschedulablePodsToHelp
				}

				if len(unschedulablePodsToHelp) == 0 {
					glog.V(1).Info("No unschedulable pods")
				} else {
					scaleUpStart := time.Now()
					updateLastTime("scaleup")
					scaledUp, err := ScaleUp(unschedulablePodsToHelp, nodes, cloudProvider, kubeClient, predicateChecker, recorder)

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
					unneededStart := time.Now()

					// In dry run only utilization is updated
					calculateUnneededOnly := lastScaleUpTime.Add(*scaleDownDelay).After(time.Now()) ||
						lastScaleDownFailedTrial.Add(*scaleDownTrialInterval).After(time.Now()) ||
						schedulablePodsPresent

					glog.V(4).Infof("Scale down status: unneededOnly=%v lastScaleUpTime=%s "+
						"lastScaleDownFailedTrail=%s schedulablePodsPresent=%v", calculateUnneededOnly,
						lastScaleUpTime, lastScaleDownFailedTrial, schedulablePodsPresent)

					updateLastTime("findUnneeded")
					glog.V(4).Infof("Calculating unneded nodes")

					usageTracker.CleanUp(time.Now().Add(-(*scaleDownUnneededTime)))
					unneededNodes, podLocationHints = FindUnneededNodes(
						nodes,
						unneededNodes,
						*scaleDownUtilizationThreshold,
						allScheduled,
						predicateChecker,
						podLocationHints,
						usageTracker, time.Now())

					updateDuration("findUnneeded", unneededStart)

					for key, val := range unneededNodes {
						if glog.V(4) {
							glog.V(4).Infof("%s is unneeded since %s duration %s", key, val.String(), time.Now().Sub(val).String())
						}
					}

					if !calculateUnneededOnly {
						glog.V(4).Infof("Starting scale down")

						scaleDownStart := time.Now()
						updateLastTime("scaledown")

						result, err := ScaleDown(
							nodes,
							unneededNodes,
							*scaleDownUnneededTime,
							allScheduled,
							cloudProvider,
							kubeClient,
							predicateChecker,
							podLocationHints,
							usageTracker)

						updateDuration("scaledown", scaleDownStart)

						// TODO: revisit result handling
						if err != nil {
							glog.Errorf("Failed to scale down: %v", err)
						} else {
							if result == ScaleDownError || result == ScaleDownNoNodeDeleted {
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

func main() {
	leaderElection := kube_leaderelection.DefaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = true

	kube_leaderelection.BindFlags(&leaderElection, pflag.CommandLine)
	flag.Var(&nodeGroupsFlag, "nodes", "sets min,max size and other configuration data for a node group in a format accepted by cloud provider."+
		"Can be used multiple times. Format: <min>:<max>:<other...>")
	kube_flag.InitFlags()

	glog.Infof("Cluster Autoscaler %s", ClusterAutoscalerVersion)

	go func() {
		http.Handle("/metrics", prometheus.Handler())
		err := http.ListenAndServe(*address, nil)
		glog.Fatalf("Failed to start metrics: %v", err)
	}()

	if !leaderElection.LeaderElect {
		run(nil)
	} else {
		id, err := os.Hostname()
		if err != nil {
			glog.Fatalf("Unable to get hostname: %v", err)
		}

		kubeClient := createKubeClient()
		kube_leaderelection.RunOrDie(kube_leaderelection.LeaderElectionConfig{
			EndpointsMeta: kube_api.ObjectMeta{
				Namespace: "kube-system",
				Name:      "cluster-autoscaler",
			},
			Client:        kubeClient,
			Identity:      id,
			EventRecorder: createEventRecorder(kubeClient),
			LeaseDuration: leaderElection.LeaseDuration.Duration,
			RenewDeadline: leaderElection.RenewDeadline.Duration,
			RetryPeriod:   leaderElection.RetryPeriod.Duration,
			Callbacks: kube_leaderelection.LeaderCallbacks{
				OnStartedLeading: run,
				OnStoppedLeading: func() {
					glog.Fatalf("lost master")
				},
			},
		})
	}
}

func updateDuration(label string, start time.Time) {
	duration.WithLabelValues(label).Observe(durationToMicro(start))
	lastDuration.WithLabelValues(label).Set(durationToMicro(start))
}

func updateLastTime(label string) {
	lastTimestamp.WithLabelValues(label).Set(float64(time.Now().Unix()))
}
