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
	"net/url"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/utils/gce"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/golang/glog"
)

var (
	migConfigFlag config.MigConfigFlag
	kubernetes    = flag.String("kubernetes", "", "Kuberentes master location. Leave blank for default")
)

func main() {
	flag.Var(&migConfigFlag, "nodes", "sets min,max size and url of a MIG to be controlled by Cluster Autoscaler. "+
		"Can be used multiple times. Format: <min>:<max>:<migurl>")
	flag.Parse()

	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kuberentes url: %v", err)
	}
	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client configuration: %v", err)
	}

	kubeClient := kube_client.NewOrDie(kubeConfig)
	unscheduledPodLister := NewUnscheduledPodLister(kubeClient)
	nodeLister := NewNodeLister(kubeClient)

	migConfigs := make([]*config.MigConfig, 0, len(migConfigFlag))
	gceManager, err := gce.CreateGceManager(migConfigs)
	if err != nil {
		glog.Fatalf("Failed to create GCE Manager %v", err)
	}

	for {
		select {
		case <-time.After(time.Minute):
			{
				pods, err := unscheduledPodLister.List()
				if err != nil {
					glog.Errorf("Failed to list pods: %v", err)
					continue
				}
				if len(pods) == 0 {
					glog.V(1).Info("No unscheduled pods")
					continue
				}

				for _, pod := range pods {
					glog.V(1).Infof("Pod %s/%s is not scheduled", pod.Namespace, pod.Name)
				}

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

				// Checks if scheduler tried to schedule the pods after thew newest node was added.
				newestNode := GetNewestNode(nodes)
				if newestNode == nil {
					glog.Errorf("No newest node")
					continue
				}
				oldestSchedulingTrial := GetOldestFailedSchedulingTrail(pods)
				if oldestSchedulingTrial == nil {
					glog.Errorf("No oldest unschedueled trial: %v", err)
					continue
				}

				// TODO: Find better way to check if all pods were checked after the newest node
				// was added.
				if newestNode.CreationTimestamp.After(oldestSchedulingTrial.Add(-1 * time.Minute)) {
					// Lets give scheduler another chance.
					glog.V(1).Infof("One of the pods have not been tried after adding %s", newestNode.Name)
					continue
				}
			}
		}
	}
}
