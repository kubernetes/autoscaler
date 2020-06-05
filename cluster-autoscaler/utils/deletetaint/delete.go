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
	"context"
	"fmt"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"

	klog "k8s.io/klog/v2"
)

const (
	// ToBeDeletedTaint is a taint used to make the node unschedulable.
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"
	// DeletionCandidateTaint is a taint used to mark unneeded node as preferably unschedulable.
	DeletionCandidateTaint = "DeletionCandidateOfClusterAutoscaler"
)

// Mutable only in unit tests
var (
	maxRetryDeadline      time.Duration = 5 * time.Second
	conflictRetryInterval time.Duration = 750 * time.Millisecond
)

// getKeyShortName converts taint key to short name for logging
func getKeyShortName(key string) string {
	switch key {
	case ToBeDeletedTaint:
		return "ToBeDeletedTaint"
	case DeletionCandidateTaint:
		return "DeletionCandidateTaint"
	default:
		return key
	}
}

// MarkToBeDeleted sets a taint that makes the node unschedulable.
func MarkToBeDeleted(node *apiv1.Node, client kube_client.Interface) error {
	return addTaint(node, client, ToBeDeletedTaint, apiv1.TaintEffectNoSchedule)
}

// MarkDeletionCandidate sets a soft taint that makes the node preferably unschedulable.
func MarkDeletionCandidate(node *apiv1.Node, client kube_client.Interface) error {
	return addTaint(node, client, DeletionCandidateTaint, apiv1.TaintEffectPreferNoSchedule)
}

func addTaint(node *apiv1.Node, client kube_client.Interface, taintKey string, effect apiv1.TaintEffect) error {
	retryDeadline := time.Now().Add(maxRetryDeadline)
	freshNode := node.DeepCopy()
	var err error
	refresh := false
	for {
		if refresh {
			// Get the newest version of the node.
			freshNode, err = client.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
			if err != nil || freshNode == nil {
				klog.Warningf("Error while adding %v taint on node %v: %v", getKeyShortName(taintKey), node.Name, err)
				return fmt.Errorf("failed to get node %v: %v", node.Name, err)
			}
		}

		if !addTaintToSpec(freshNode, taintKey, effect) {
			if !refresh {
				// Make sure we have the latest version before skipping update.
				refresh = true
				continue
			}
			return nil
		}
		_, err = client.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err != nil && errors.IsConflict(err) && time.Now().Before(retryDeadline) {
			refresh = true
			time.Sleep(conflictRetryInterval)
			continue
		}

		if err != nil {
			klog.Warningf("Error while adding %v taint on node %v: %v", getKeyShortName(taintKey), node.Name, err)
			return err
		}
		klog.V(1).Infof("Successfully added %v on node %v", getKeyShortName(taintKey), node.Name)
		return nil
	}
}

func addTaintToSpec(node *apiv1.Node, taintKey string, effect apiv1.TaintEffect) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == taintKey {
			klog.V(2).Infof("%v already present on node %v, taint: %v", taintKey, node.Name, taint)
			return false
		}
	}
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
		Key:    taintKey,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: effect,
	})
	return true
}

// HasToBeDeletedTaint returns true if ToBeDeleted taint is applied on the node.
func HasToBeDeletedTaint(node *apiv1.Node) bool {
	return hasTaint(node, ToBeDeletedTaint)
}

// HasDeletionCandidateTaint returns true if DeletionCandidate taint is applied on the node.
func HasDeletionCandidateTaint(node *apiv1.Node) bool {
	return hasTaint(node, DeletionCandidateTaint)
}

func hasTaint(node *apiv1.Node, taintKey string) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == taintKey {
			return true
		}
	}
	return false
}

// GetToBeDeletedTime returns the date when the node was marked by CA as for delete.
func GetToBeDeletedTime(node *apiv1.Node) (*time.Time, error) {
	return getTaintTime(node, ToBeDeletedTaint)
}

// GetDeletionCandidateTime returns the date when the node was marked by CA as for delete.
func GetDeletionCandidateTime(node *apiv1.Node) (*time.Time, error) {
	return getTaintTime(node, DeletionCandidateTaint)
}

func getTaintTime(node *apiv1.Node, taintKey string) (*time.Time, error) {
	for _, taint := range node.Spec.Taints {
		if taint.Key == taintKey {
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

// CleanToBeDeleted cleans CA's NoSchedule taint from a node.
func CleanToBeDeleted(node *apiv1.Node, client kube_client.Interface) (bool, error) {
	return cleanTaint(node, client, ToBeDeletedTaint)
}

// CleanDeletionCandidate cleans CA's soft NoSchedule taint from a node.
func CleanDeletionCandidate(node *apiv1.Node, client kube_client.Interface) (bool, error) {
	return cleanTaint(node, client, DeletionCandidateTaint)
}

func cleanTaint(node *apiv1.Node, client kube_client.Interface, taintKey string) (bool, error) {
	retryDeadline := time.Now().Add(maxRetryDeadline)
	freshNode := node.DeepCopy()
	var err error
	refresh := false
	for {
		if refresh {
			// Get the newest version of the node.
			freshNode, err = client.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
			if err != nil || freshNode == nil {
				klog.Warningf("Error while adding %v taint on node %v: %v", getKeyShortName(taintKey), node.Name, err)
				return false, fmt.Errorf("failed to get node %v: %v", node.Name, err)
			}
		}
		newTaints := make([]apiv1.Taint, 0)
		for _, taint := range freshNode.Spec.Taints {
			if taint.Key == taintKey {
				klog.V(1).Infof("Releasing taint %+v on node %v", taint, node.Name)
			} else {
				newTaints = append(newTaints, taint)
			}
		}
		if len(newTaints) == len(freshNode.Spec.Taints) {
			if !refresh {
				// Make sure we have the latest version before skipping update.
				refresh = true
				continue
			}
			return false, nil
		}

		freshNode.Spec.Taints = newTaints
		_, err = client.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})

		if err != nil && errors.IsConflict(err) && time.Now().Before(retryDeadline) {
			refresh = true
			time.Sleep(conflictRetryInterval)
			continue
		}

		if err != nil {
			klog.Warningf("Error while releasing %v taint on node %v: %v", getKeyShortName(taintKey), node.Name, err)
			return false, err
		}
		klog.V(1).Infof("Successfully released %v on node %v", getKeyShortName(taintKey), node.Name)
		return true, nil
	}
}

// CleanAllToBeDeleted cleans ToBeDeleted taints from given nodes.
func CleanAllToBeDeleted(nodes []*apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder) {
	cleanAllTaints(nodes, client, recorder, ToBeDeletedTaint)
}

// CleanAllDeletionCandidates cleans DeletionCandidate taints from given nodes.
func CleanAllDeletionCandidates(nodes []*apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder) {
	cleanAllTaints(nodes, client, recorder, DeletionCandidateTaint)
}

func cleanAllTaints(nodes []*apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder, taintKey string) {
	for _, node := range nodes {
		if !hasTaint(node, taintKey) {
			continue
		}
		cleaned, err := cleanTaint(node, client, taintKey)
		if err != nil {
			recorder.Eventf(node, apiv1.EventTypeWarning, "ClusterAutoscalerCleanup",
				"failed to clean %v on node %v: %v", getKeyShortName(taintKey), node.Name, err)
		} else if cleaned {
			recorder.Eventf(node, apiv1.EventTypeNormal, "ClusterAutoscalerCleanup",
				"removed %v taint from node %v", getKeyShortName(taintKey), node.Name)
		}
	}
}
