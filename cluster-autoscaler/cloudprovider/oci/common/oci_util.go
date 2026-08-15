/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
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

// NewRetryPolicy returns an exponential backoff retry policy
func NewRetryPolicy() *common.RetryPolicy {
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

// AnnotateNode adds an annotation to a new based on the key/value
func AnnotateNode(kubeClient kubernetes.Interface, nodeName string, key string, value string) error {

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

// LabelNode adds a label to a new based on the key/value
func LabelNode(kubeClient kubernetes.Interface, nodeName string, key string, value string) error {

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

// SetNodeProviderID sets the provider id value on the node object
func SetNodeProviderID(kubeClient kubernetes.Interface, nodeName string, value string) error {

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

// GetPoolType returns the resource type of the specified group i.e. (instance pool or node pool) or an error if it cannot be determined
func GetPoolType(group string) (string, error) {
	ocidParts := strings.Split(group, ".")
	if len(ocidParts) >= 2 {
		// we just need to populate the map key
		return ocidParts[1], nil
	}
	return "", fmt.Errorf("unsupported ocid value. Could not determine ocid type of %s", group)
}

// GetAllPoolTypes returns the resource type of the specified groups i.e. (instance pool or node pool) or an error if it cannot be determined
func GetAllPoolTypes(groups []string) (string, error) {
	// set of all the ocid types
	// we only support instance pools or node pools
	ocidTypes := make(map[string]interface{})
	for _, group := range groups {
		ocidParts := strings.Split(group, ".")
		if len(ocidParts) >= 2 {
			// we just need to populate the map key
			ocidTypes[ocidParts[1]] = nil
		} else {
			return "", fmt.Errorf("unsupported ocid value. Could not determine ocid type of %s", group)
		}
	}

	if len(ocidTypes) > 1 {
		return "", fmt.Errorf("found multiple ocid types, but cluster autoscaler currently only supports either instance pools OR node pools: %v", ocidTypes)
	}

	var ocidType string
	// ocidTypes set should only have a single entry in it
	for key := range ocidTypes {
		ocidType = key
		break
	}
	return ocidType, nil
}

// HasNodeGroupTags checks if nodepoolTags is provided
func HasNodeGroupTags(nodeGroupAutoDiscoveryList []string) (bool, bool, error) {
	instancePoolTagsFound := false
	nodePoolTagsFound := false
	for _, arg := range nodeGroupAutoDiscoveryList {
		if strings.Contains(arg, "nodepoolTags") {
			nodePoolTagsFound = true
		}
		if strings.Contains(arg, "instancepoolTags") {
			instancePoolTagsFound = true
		}
	}
	if instancePoolTagsFound == true && nodePoolTagsFound == true {
		return instancePoolTagsFound, nodePoolTagsFound, fmt.Errorf("can not use both instancepoolTags and nodepoolTags in node-group-auto-discovery")
	}
	if len(nodeGroupAutoDiscoveryList) > 0 && instancePoolTagsFound == false && nodePoolTagsFound == false {
		return instancePoolTagsFound, nodePoolTagsFound, fmt.Errorf("either instancepoolTags or nodepoolTags should be provided in node-group-auto-discovery")
	}
	return instancePoolTagsFound, nodePoolTagsFound, nil
}
