/*
Copyright 2021 Oracle and/or its affiliates.

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

package oci

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// IsRetryable returns true if the given error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)

	// Retry on network timeout errors
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}

	// handle oci retryable errors.
	serviceErr, ok := common.IsServiceError(err)
	if !ok {
		return false
	}

	switch serviceErr.GetHTTPStatusCode() {
	case http.StatusTooManyRequests, http.StatusGatewayTimeout,
		http.StatusInternalServerError, http.StatusBadGateway:
		return true
	default:
		return false
	}
}

func newRetryPolicy() *common.RetryPolicy {
	return NewRetryPolicyWithMaxAttempts(uint(8))
}

// NewRetryPolicyWithMaxAttempts returns a RetryPolicy with the specified max retryAttempts
func NewRetryPolicyWithMaxAttempts(retryAttempts uint) *common.RetryPolicy {
	isRetryableOperation := func(r common.OCIOperationResponse) bool {
		return IsRetryable(r.Error)
	}

	nextDuration := func(r common.OCIOperationResponse) time.Duration {
		// you might want wait longer for next retry when your previous one failed
		// this function will return the duration as:
		// 1s, 2s, 4s, 8s, 16s, 32s, 64s etc...
		return time.Duration(math.Pow(float64(2), float64(r.AttemptNumber-1))) * time.Second
	}

	policy := common.NewRetryPolicy(
		retryAttempts, isRetryableOperation, nextDuration,
	)
	return &policy
}

// Missing resource requests on kube-proxy
// Flannel missing priority

func buildCSINodePod() *apiv1.Pod {
	priority := scheduling.SystemCriticalPriority
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("csi-oci-node-%d", rand.Int63()),
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "csi-oci-node",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Image: "iad.ocir.io/oracle/cloud-provider-oci:latest",
				},
			},
			Priority: &priority,
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodRunning,
			Conditions: []apiv1.PodCondition{
				{
					Type:   apiv1.PodReady,
					Status: apiv1.ConditionTrue,
				},
			},
		},
	}
}

func annotateNode(kubeClient kubernetes.Interface, nodeName string, key string, value string) error {

	if nodeName == "" {
		return errors.New("node name is required")
	}
	if kubeClient == nil {
		return errors.New("kubeconfig is required")
	}

	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get node %s %+v", nodeName, err)
		return err
	}

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if v := annotations[key]; v != value {
		node.Annotations[key] = value
		_, err := kubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("failed to annotate node %s %+v", nodeName, err)
			return err
		}
		klog.V(3).Infof("updated annotation %s=%s on node: %s", key, value, nodeName)
	}
	return nil
}

func labelNode(kubeClient kubernetes.Interface, nodeName string, key string, value string) error {

	if nodeName == "" {
		return errors.New("node name is required")
	}
	if kubeClient == nil {
		return errors.New("kubeconfig is required")
	}

	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get node %s %+v", nodeName, err)
		return err
	}

	labels := node.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	if v := labels[key]; v != value {
		node.Labels[key] = value
		_, err := kubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("failed to label node %s %+v", nodeName, err)
			return err
		}
		klog.V(3).Infof("updated label %s=%s on node: %s", key, value, nodeName)
	}
	return nil
}

func setNodeProviderID(kubeClient kubernetes.Interface, nodeName string, value string) error {

	if nodeName == "" {
		return errors.New("node name is required")
	}
	if kubeClient == nil {
		return errors.New("kubeconfig is required")
	}

	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})

	if err != nil {
		klog.Errorf("failed to get node %s %+v", nodeName, err)
		return err
	}

	if node.Spec.ProviderID != value {
		node.Spec.ProviderID = value
		_, err := kubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("failed to update node's provider ID %s %+v", nodeName, err)
			return err
		}
		klog.V(3).Infof("updated provider ID on node: %s", nodeName)
	}
	return nil
}
