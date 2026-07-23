/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

const (
	// DelayDeletionAnnotationPrefix is the prefix of annotation marking node as it needs to wait
	// for other K8s components before deleting node.
	DelayDeletionAnnotationPrefix = "delay-deletion.cluster-autoscaler.kubernetes.io/"
)

// WaitForDelayDeletion waits until the provided node has no annotations beginning with DelayDeletionAnnotationPrefix,
// or until the provided timeout is reached - whichever comes first.
func WaitForDelayDeletion(ctx context.Context, node *apiv1.Node, nodeLister kubernetes.NodeLister, timeout time.Duration) errors.AutoscalerError {
	logger := klog.FromContext(ctx)
	if timeout != 0 && hasDelayDeletionAnnotation(node) {
		logger.V(1).Info("Wait for removing annotations on node", "DelayDeletionAnnotationPrefix", DelayDeletionAnnotationPrefix, "node", node.Name)
		err := wait.Poll(5*time.Second, timeout, func() (bool, error) {
			logger.V(5).Info("Waiting for removing annotations on node", "DelayDeletionAnnotationPrefix", DelayDeletionAnnotationPrefix, "node", node.Name)
			freshNode, err := nodeLister.Get(node.Name)
			if err != nil || freshNode == nil {
				return false, fmt.Errorf("failed to get node %v: %v", node.Name, err)
			}
			return !hasDelayDeletionAnnotation(freshNode), nil
		})
		if err != nil && err != wait.ErrWaitTimeout {
			return errors.ToAutoscalerError(errors.ApiCallError, err)
		}
		if err == wait.ErrWaitTimeout {
			logger.Info("Delay node deletion timed out for node , delay deletion annotation wasn't removed within , this might slow down scale down.", "node", node.Name, "timeout", timeout)
		} else {
			logger.V(2).Info("Annotation removed from node", "DelayDeletionAnnotationPrefix", DelayDeletionAnnotationPrefix, "node", node.Name)
		}
	}
	return nil
}

func hasDelayDeletionAnnotation(node *apiv1.Node) bool {
	for annotation := range node.Annotations {
		if strings.HasPrefix(annotation, DelayDeletionAnnotationPrefix) {
			return true
		}
	}
	return false
}
