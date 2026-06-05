/*
Copyright 2025 The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// DefaultAggregationLabel is the label used to aggregate RBAC rules to the main cluster-autoscaler role.
	DefaultAggregationLabel = "rbac.authorization.k8s.io/aggregate-to-cluster-autoscaler"
)

// CapacityBufferRBACUpdater defines an interface for dynamically updating RBAC permissions
// for target GVKs. This allows cloud providers to provide their own implementation
// for ensuring the controller has necessary permissions.
type CapacityBufferRBACUpdater interface {
	// UpdateRBAC ensures that the controller has permissions to get, list, and watch
	// the specified resource.
	UpdateRBAC(mapping *meta.RESTMapping) error
}

// DefaultRBACUpdater is a default implementation of CapacityBufferRBACUpdater.
type DefaultRBACUpdater struct {
	kubeClient kubernetes.Interface
}

// NewDefaultRBACUpdater returns a new DefaultRBACUpdater.
func NewDefaultRBACUpdater(kubeClient kubernetes.Interface) *DefaultRBACUpdater {
	return &DefaultRBACUpdater{
		kubeClient: kubeClient,
	}
}

// UpdateRBAC logs a guide for the user on how to manually provide necessary RBAC permissions.
func (d *DefaultRBACUpdater) UpdateRBAC(mapping *meta.RESTMapping) error {
	gvk := mapping.GroupVersionKind
	gvr := mapping.Resource

	roleName := fmt.Sprintf("cluster-autoscaler-dynamic-%s-%s",
		strings.ReplaceAll(strings.ToLower(gvr.Group), ".", "-"),
		strings.ReplaceAll(strings.ToLower(gvr.Resource), ".", "-"))

	// Truncate if too long
	if len(roleName) > 253 {
		roleName = roleName[:253]
	}

	guide := fmt.Sprintf(`
[RBAC GUIDE] To enable event-driven reconciliation for %v, please create the following ClusterRole and aggregate it to the cluster-autoscaler:

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
  labels:
    %s: "true"
rules:
- apiGroups: ["%s"]
  resources: ["%s", "%s/scale"]
  verbs: ["get", "list", "watch"]
`, gvk, roleName, DefaultAggregationLabel, gvr.Group, gvr.Resource, gvr.Resource)

	klog.Info(guide)
	return nil
}
