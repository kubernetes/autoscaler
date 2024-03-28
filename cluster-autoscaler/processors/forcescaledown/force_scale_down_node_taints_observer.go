/*
Copyright 2024 The Kubernetes Authors.

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

package forcescaledown

import (
	"context"
	"fmt"
	"reflect"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// ForceScaleDownNodeTaintsObserver is an observer to handle force-scale-down taints.
type ForceScaleDownNodeTaintsObserver struct {
	kubeClient     kube_client.Interface
	listerRegistry kube_util.ListerRegistry
}

// NewForceScaleDownNodeTaintsObserver returns a constructed ForceScaleDownNodeTaintsObserver struct.
func NewForceScaleDownNodeTaintsObserver(kubeClient kube_client.Interface, informerFactory informers.SharedInformerFactory) *ForceScaleDownNodeTaintsObserver {
	listerRegistry := kube_util.NewListerRegistryWithDefaultListers(informerFactory)
	return &ForceScaleDownNodeTaintsObserver{kubeClient, listerRegistry}
}

// Refresh starts to backfill empty TimeAdded for all force-scale-down taints.
func (p *ForceScaleDownNodeTaintsObserver) Refresh() {
	if err := p.backfillTaintTimeAddedIfEmtpy(context.Background()); err != nil {
		klog.Warningf("Failed to backfill TimeAdded for force-scale-down taints: %v", err)
	}
}

func (p *ForceScaleDownNodeTaintsObserver) backfillTaintTimeAddedIfEmtpy(ctx context.Context) error {
	nodes, err := p.listerRegistry.AllNodeLister().List()
	if err != nil {
		return fmt.Errorf("failed to list all nodes: %w", err)
	}
	taintUpdater := func(node *apiv1.Node) {
		for index := range node.Spec.Taints {
			if node.Spec.Taints[index].Key != taints.ForceScaleDownTaint {
				continue
			}
			if node.Spec.Taints[index].TimeAdded != nil && !node.Spec.Taints[index].TimeAdded.IsZero() {
				continue
			}
			node.Spec.Taints[index].TimeAdded = &metav1.Time{Time: time.Now()}
		}
	}
	nodesToUpdate := []*apiv1.Node{}
	for _, node := range nodes {
		copiedNode := node.DeepCopy()
		taintUpdater(copiedNode)
		if !reflect.DeepEqual(node, copiedNode) {
			nodesToUpdate = append(nodesToUpdate, node)
		}
	}
	for _, node := range nodesToUpdate {
		klog.V(2).Infof("Backfilling TimeAdded for force-scale-down node taint %s: %+v", node.Name, node.Spec.Taints)
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			node, err := p.kubeClient.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get node %s: %w", node, err)
			}
			taintUpdater(node)
			if _, err := p.kubeClient.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("failed to update node %s: %w", node.Name, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update %s node taint after retries: %w", node.Name, err)
		}
	}
	return nil
}
