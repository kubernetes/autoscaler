/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/estimator"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_record "k8s.io/kubernetes/pkg/client/record"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/golang/glog"
)

// ExpansionOption describes an option to expand the cluster.
type ExpansionOption struct {
	nodeGroup cloudprovider.NodeGroup
	nodeCount int
	debug     string
	pods      []*kube_api.Pod
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occured. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(unschedulablePods []*kube_api.Pod, nodes []*kube_api.Node, cloudProvider cloudprovider.CloudProvider, kubeClient *kube_client.Client,
	predicateChecker *simulator.PredicateChecker, recorder kube_record.EventRecorder, maxNodesTotal int,
	estimatorName string) (bool, error) {

	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return false, nil
	}

	for _, pod := range unschedulablePods {
		glog.V(1).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}

	expansionOptions := make([]ExpansionOption, 0)
	nodeInfos, err := GetNodeInfosForGroups(nodes, cloudProvider, kubeClient)
	if err != nil {
		return false, fmt.Errorf("failed to build node infos for node groups: %v", err)
	}

	podsRemainUnshedulable := make(map[*kube_api.Pod]struct{})
	for _, nodeGroup := range cloudProvider.NodeGroups() {

		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			glog.Errorf("Failed to get node group size: %v", err)
			continue
		}
		if currentSize >= nodeGroup.MaxSize() {
			// skip this node group.
			glog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			continue
		}

		option := ExpansionOption{
			nodeGroup: nodeGroup,
			pods:      make([]*kube_api.Pod, 0),
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			glog.Errorf("No node info for: %s", nodeGroup.Id())
			continue
		}

		for _, pod := range unschedulablePods {
			err = predicateChecker.CheckPredicates(pod, nodeInfo)
			if err == nil {
				option.pods = append(option.pods, pod)
			} else {
				glog.V(2).Infof("Scale-up predicate failed: %v", err)
				podsRemainUnshedulable[pod] = struct{}{}
			}
		}
		if len(option.pods) > 0 {
			if estimatorName == BinpackingEstimatorName {
				binpackingEstimator := estimator.NewBinpackingNodeEstimator(predicateChecker)
				option.nodeCount = binpackingEstimator.Estimate(option.pods, nodeInfo)
			} else if estimatorName == BasicEstimatorName {
				basicEstimator := estimator.NewBasicNodeEstimator()
				for _, pod := range option.pods {
					basicEstimator.Add(pod)
				}
				option.nodeCount, option.debug = basicEstimator.Estimate(nodeInfo.Node())
			} else {
				glog.Fatalf("Unrecognized estimator: %s", estimatorName)
			}
			expansionOptions = append(expansionOptions, option)
		}
	}

	// Pick some expansion option.
	bestOption := BestExpansionOption(expansionOptions)
	if bestOption != nil && bestOption.nodeCount > 0 {
		glog.V(1).Infof("Best option to resize: %s", bestOption.nodeGroup.Id())
		if len(bestOption.debug) > 0 {
			glog.V(1).Info(bestOption.debug)
		}
		glog.V(1).Infof("Estimated %d nodes needed in %s", bestOption.nodeCount, bestOption.nodeGroup.Id())

		currentSize, err := bestOption.nodeGroup.TargetSize()
		if err != nil {
			return false, fmt.Errorf("failed to get node group size: %v", err)
		}
		newSize := currentSize + bestOption.nodeCount
		if newSize >= bestOption.nodeGroup.MaxSize() {
			glog.V(1).Infof("Capping size to MAX (%d)", bestOption.nodeGroup.MaxSize())
			newSize = bestOption.nodeGroup.MaxSize()
		}

		if maxNodesTotal > 0 && len(nodes)+(newSize-currentSize) > maxNodesTotal {
			glog.V(1).Infof("Capping size to max cluster total size (%d)", maxNodesTotal)
			newSize = maxNodesTotal - len(nodes) + currentSize
			if newSize < currentSize {
				return false, fmt.Errorf("max node total count already reached")
			}
		}

		glog.V(0).Infof("Scale-up: setting group %s size to %d", bestOption.nodeGroup.Id(), newSize)

		if err := bestOption.nodeGroup.IncreaseSize(newSize - currentSize); err != nil {
			return false, fmt.Errorf("failed to increase node group size: %v", err)
		}

		for _, pod := range bestOption.pods {
			recorder.Eventf(pod, kube_api.EventTypeNormal, "TriggeredScaleUp",
				"pod triggered scale-up, group: %s, sizes (current/new): %d/%d", bestOption.nodeGroup.Id(), currentSize, newSize)
		}

		return true, nil
	}
	for pod := range podsRemainUnshedulable {
		recorder.Event(pod, kube_api.EventTypeNormal, "NotTriggerScaleUp",
			"pod didn't trigger scale-up (it wouldn't fit if a new node is added)")
	}

	return false, nil
}
