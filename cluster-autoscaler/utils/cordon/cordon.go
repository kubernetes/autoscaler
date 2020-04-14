/*
Copyright 2020 The Kubernetes Authors.

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

package cordon

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	// CordonTaint is a taint used to show that the node is unschedulable and cordoned by CA.
	CordonTaint = "CordonedByClusterAutoscaler"
)

var (
	maxRetryDeadlineForCordon   time.Duration = 5 * time.Second
	conflictCordonRetryInterval time.Duration = 750 * time.Millisecond
)

// CordonNode node make it unschedulable
func CordonNode(node *apiv1.Node, client kube_client.Interface) error {
	if node.Spec.Unschedulable {
		if hasNodeCordonTaint(node) {
			klog.V(1).Infof("Node %v already was cordoned by Cluster Autoscaler", node.Name)
			return nil
		}
		if !hasNodeCordonTaint(node) {
			klog.V(1).Infof("Skip cordonning because node %v was not cordoned by Cluster Autoscaler", node.Name)
			return nil
		}
	}
	freshNode := node.DeepCopy()
	retryDeadline := time.Now().Add(maxRetryDeadlineForCordon)
	var err error
	refresh := false
	for {
		if refresh {
			// Get the newest version of the node.
			freshNode, err = client.CoreV1().Nodes().Get(node.Name, metav1.GetOptions{})
		}
		freshNode.Spec.Taints = append(freshNode.Spec.Taints, apiv1.Taint{
			Key:    CordonTaint,
			Value:  "",
			Effect: apiv1.TaintEffectNoSchedule,
		})
		freshNode.Spec.Unschedulable = true
		_, err = client.CoreV1().Nodes().Update(freshNode)
		if err != nil && errors.IsConflict(err) && time.Now().Before(retryDeadline) {
			refresh = true
			time.Sleep(conflictCordonRetryInterval)
			continue
		}

		if err != nil {
			klog.Warningf("Error while cordoning node %v: %v", node.Name, err)
			return nil
		}
		klog.V(1).Infof("Successfully cordoned node %v by Cluster Autoscaler", node.Name)
		return nil
	}

}

// UnCordonNode node make it schedulable
func UnCordonNode(node *apiv1.Node, client kube_client.Interface) error {
	if node.Spec.Unschedulable && hasNodeCordonTaint(node) == false {
		klog.V(1).Infof("Skip uncordonning because node %v was not cordoned by Cluster Autoscaler", node.Name)
		return nil
	}
	freshNode := node.DeepCopy()
	retryDeadline := time.Now().Add(maxRetryDeadlineForCordon)
	var err error
	refresh := false
	for {
		if refresh {
			// Get the newest version of the node.
			freshNode, err = client.CoreV1().Nodes().Get(node.Name, metav1.GetOptions{})
			refresh = false
		}
		newTaints := make([]apiv1.Taint, 0)
		for _, taint := range freshNode.Spec.Taints {
			if taint.Key != CordonTaint {
				newTaints = append(newTaints, taint)
			}
		}

		if len(newTaints) != len(freshNode.Spec.Taints) {
			freshNode.Spec.Taints = newTaints
		} else {
			refresh = true
			continue
		}

		freshNode.Spec.Unschedulable = false
		_, err = client.CoreV1().Nodes().Update(freshNode)
		if err != nil && errors.IsConflict(err) && time.Now().Before(retryDeadline) {
			refresh = true
			time.Sleep(conflictCordonRetryInterval)
			continue
		}

		if err != nil {
			klog.Warningf("Error while uncordoning node %v: %v", node.Name, err)
			return nil
		}
		klog.V(1).Infof("Successfully uncordoned node %v by Cluster Autoscaler", node.Name)
		return nil
	}
}

func hasNodeCordonTaint(node *apiv1.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == CordonTaint {
			return true
		}
	}
	return false
}
