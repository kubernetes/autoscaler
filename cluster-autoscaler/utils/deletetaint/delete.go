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

package deletetaint

import (
	"encoding/json"
	"fmt"
	"time"

	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"

	"github.com/golang/glog"
)

const (
	// ToBeDeletedTaint is a taint used to make the node unschedulable.
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"
)

// MarkToBeDeleted sets a taint that makes the node unschedulable.
func MarkToBeDeleted(node *apiv1.Node, client kube_client.Interface) error {
	// Get the newest version of the node.
	freshNode, err := client.Core().Nodes().Get(node.Name)
	if err != nil || freshNode == nil {
		return fmt.Errorf("failed to get node %v: %v", node.Name, err)
	}

	added, err := addToBeDeletedTaint(freshNode)
	if added == false {
		return err
	}
	_, err = client.Core().Nodes().Update(freshNode)
	if err != nil {
		glog.Warningf("Error while adding taints on node %v: %v", node.Name, err)
		return err
	}
	glog.V(1).Infof("Successfully added toBeDeletedTaint on node %v", node.Name)
	return nil
}

func addToBeDeletedTaint(node *apiv1.Node) (bool, error) {
	taints, err := apiv1.GetTaintsFromNodeAnnotations(node.Annotations)
	if err != nil {
		glog.Warningf("Error while getting Taints for node %v: %v", node.Name, err)
		return false, err
	}
	for _, taint := range taints {
		if taint.Key == ToBeDeletedTaint {
			glog.Infof("ToBeDeletedTaint already present on on node %v", taint, node.Name)
			return false, nil
		}
	}
	taints = append(taints, apiv1.Taint{
		Key:    ToBeDeletedTaint,
		Value:  time.Now().String(),
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taintsJson, err := json.Marshal(taints)
	if err != nil {
		glog.Warningf("Error while adding taints on node %v: %v", node.Name, err)
		return false, err
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[apiv1.TaintsAnnotationKey] = string(taintsJson)
	return true, nil
}

// HasToBeDeletedTaint returns true if ToBeDeleted taint is applied on the node.
func HasToBeDeletedTaint(node *apiv1.Node) bool {
	taints, err := apiv1.GetTaintsFromNodeAnnotations(node.Annotations)
	if err != nil {
		glog.Warningf("Node %v has incorrect taint annotation: %v", err)
		return false
	}
	for _, taint := range taints {
		if taint.Key == ToBeDeletedTaint {
			return true
		}
	}
	return false
}

// CleanToBeDeleted cleans ToBeDeleted taint.
func CleanToBeDeleted(node *apiv1.Node, client kube_client.Interface) (bool, error) {
	taints, err := apiv1.GetTaintsFromNodeAnnotations(node.Annotations)
	if err != nil {
		glog.Warningf("Error while getting Taints for node %v: %v", node.Name, err)
		return false, err
	}

	newTaints := make([]apiv1.Taint, 0)
	for _, taint := range taints {
		if taint.Key == ToBeDeletedTaint {
			glog.V(1).Infof("Releasing taint %+v on node %v", taint, node.Name)
		} else {
			newTaints = append(newTaints, taint)
		}
	}

	if len(newTaints) != len(taints) {
		taintsJson, err := json.Marshal(newTaints)
		if err != nil {
			glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
			return false, err
		}
		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
		}
		node.Annotations[apiv1.TaintsAnnotationKey] = string(taintsJson)
		_, err = client.Core().Nodes().Update(node)
		if err != nil {
			glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
			return false, err
		}
		glog.V(1).Infof("Successfully released toBeDeletedTaint on node %v", node.Name)
		return true, nil
	}
	return false, nil
}
