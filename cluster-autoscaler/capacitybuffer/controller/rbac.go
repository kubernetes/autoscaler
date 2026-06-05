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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CapacityBufferRBACUpdater defines an interface for updating RBAC permissions
// or logging warnings/guidelines for target GVKs. This allows cloud providers to provide
// their own implementation for ensuring the controller has necessary permissions.
type CapacityBufferRBACUpdater interface {
	// UpdateRBAC ensures or logs that the controller has permissions to get, list, and watch
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

// UpdateRBAC logs a message that a dynamic watch is being established for the resource.
func (d *DefaultRBACUpdater) UpdateRBAC(mapping *meta.RESTMapping) error {
	gvk := mapping.GroupVersionKind
	gvr := mapping.Resource
	klog.Infof("Establishing dynamic watch for GVK: %v (GVR: %v). Please ensure ClusterAutoscaler has 'get/list/watch' permissions for this resource.", gvk, gvr)
	return nil
}
