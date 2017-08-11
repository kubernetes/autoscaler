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
	"fmt"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"

	"github.com/golang/glog"
)

const (
	// ToBeDeletedTaint is a taint used to make the node unschedulable.
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"
)

// MarkToBeDeleted sets a taint that makes the node unschedulable.
func MarkToBeDeleted(node *apiv1.Node, client kube_client.Interface) error {
	// Get the newest version of the node.
	freshNode, err := client.Core().Nodes().Get(node.Name, metav1.GetOptions{})
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
	for _, taint := range node.Spec.Taints {
		if taint.Key == ToBeDeletedTaint {
			glog.Infof("ToBeDeletedTaint already present on on node %v", taint, node.Name)
			return false, nil
		}
	}
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
		Key:    ToBeDeletedTaint,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: apiv1.TaintEffectNoSchedule,
	})
	return true, nil
}

// HasToBeDeletedTaint returns true if ToBeDeleted taint is applied on the node.
func HasToBeDeletedTaint(node *apiv1.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == ToBeDeletedTaint {
			return true
		}
	}
	return false
}

// GetToBeDeletedTime returns the date when the node was marked by CA as for delete.
func GetToBeDeletedTime(node *apiv1.Node) (*time.Time, error) {
	for _, taint := range node.Spec.Taints {
		if taint.Key == ToBeDeletedTaint {
			resultTimestamp, err := strconv.ParseInt(taint.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			result := time.Unix(resultTimestamp, 0)
			return &result, nil
		}
	}
	return nil, nil
}

// CleanToBeDeleted cleans ToBeDeleted taint.
func CleanToBeDeleted(node *apiv1.Node, client kube_client.Interface) (bool, error) {
	freshNode, err := client.Core().Nodes().Get(node.Name, metav1.GetOptions{})
	if err != nil || freshNode == nil {
		return false, fmt.Errorf("failed to get node %v: %v", node.Name, err)
	}
	newTaints := make([]apiv1.Taint, 0)
	for _, taint := range freshNode.Spec.Taints {
		if taint.Key == ToBeDeletedTaint {
			glog.V(1).Infof("Releasing taint %+v on node %v", taint, node.Name)
		} else {
			newTaints = append(newTaints, taint)
		}
	}

	if len(newTaints) != len(freshNode.Spec.Taints) {
		freshNode.Spec.Taints = newTaints
		_, err := client.Core().Nodes().Update(freshNode)
		if err != nil {
			glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
			return false, err
		}
		glog.V(1).Infof("Successfully released toBeDeletedTaint on node %v", node.Name)
		return true, nil
	}
	return false, nil
}
