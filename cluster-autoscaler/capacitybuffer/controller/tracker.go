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
	"reflect"
	"sync"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// DirtyNamespaceTracker tracks namespaces that need reconciliation based on events.
type DirtyNamespaceTracker struct {
	client          *cbclient.CapacityBufferClient
	dirtyNamespaces map[string]bool
	lock            sync.Mutex
}

// NewDirtyNamespaceTracker creates a new DirtyNamespaceTracker.
func NewDirtyNamespaceTracker(client *cbclient.CapacityBufferClient) *DirtyNamespaceTracker {
	tracker := &DirtyNamespaceTracker{
		client:          client,
		dirtyNamespaces: make(map[string]bool),
	}
	tracker.startInformers()
	return tracker
}

func (t *DirtyNamespaceTracker) startInformers() {
	// CapacityBuffer Informer
	t.client.GetBufferInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Add event for Buffer")
			t.markDirty(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Filter out Status updates to prevent double reconciliation
			oldBuf := oldObj.(*v1.CapacityBuffer)
			newBuf := newObj.(*v1.CapacityBuffer)
			if !reflect.DeepEqual(oldBuf.Spec, newBuf.Spec) || !reflect.DeepEqual(oldBuf.Labels, newBuf.Labels) || !reflect.DeepEqual(oldBuf.Annotations, newBuf.Annotations) {
				klog.V(4).Infof("DirtyNamespaceTracker: Update event for Buffer %s (Spec changed)", newBuf.Name)
				t.markDirty(newObj)
			} else {
				klog.V(4).Infof("DirtyNamespaceTracker: Skipping status update for Buffer %s", newBuf.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Delete event for Buffer")
			t.markDirty(obj)
		},
	})

	// ResourceQuota Informer
	t.client.GetResourceQuotaInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Add event for Quota")
			t.markDirty(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// We react to Spec AND Status changes for Quotas (used resources change)
			klog.V(4).Infof("DirtyNamespaceTracker: Update event for Quota")
			t.markDirty(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Delete event for Quota")
			t.markDirty(obj)
		},
	})

	// PodTemplate Informer
	t.client.GetPodTemplateInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Add event for PodTemplate")
			t.markDirty(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Update event for PodTemplate")
			t.markDirty(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(4).Infof("DirtyNamespaceTracker: Delete event for PodTemplate")
			t.markDirty(obj)
		},
	})
}

func (t *DirtyNamespaceTracker) markDirty(obj interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	var ns string
	if object, ok := obj.(interface{ GetNamespace() string }); ok {
		ns = object.GetNamespace()
	} else if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		if object, ok := tombstone.Obj.(interface{ GetNamespace() string }); ok {
			ns = object.GetNamespace()
		}
	}

	if ns != "" {
		klog.V(4).Infof("Marking namespace dirty: %s", ns)
		t.dirtyNamespaces[ns] = true
	}
}

// GetAndClearDirtyNamespaces returns the list of dirty namespaces and clears the tracker.
func (t *DirtyNamespaceTracker) GetAndClearDirtyNamespaces() []string {
	t.lock.Lock()
	defer t.lock.Unlock()

	namespaces := make([]string, 0, len(t.dirtyNamespaces))
	for ns := range t.dirtyNamespaces {
		namespaces = append(namespaces, ns)
	}
	t.dirtyNamespaces = make(map[string]bool)
	return namespaces
}
