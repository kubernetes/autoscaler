/*
Copyright The Kubernetes Authors.

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

package fakepods

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
)

// Registry tracks the relationship between fake pods (created for capacity buffers)
// and their originating CapacityBuffer objects.
type Registry struct {
	fakePodsUIDToBuffer map[types.UID]*v1beta1.CapacityBuffer
	mutex               sync.RWMutex
}

// NewRegistry returns a new instance of Registry.
// If fakePodsToBuffers is nil, it initializes an empty registry.
func NewRegistry(fakePodsToBuffers map[types.UID]*v1beta1.CapacityBuffer) *Registry {
	if fakePodsToBuffers == nil {
		fakePodsToBuffers = make(map[types.UID]*v1beta1.CapacityBuffer)
	}
	return &Registry{fakePodsUIDToBuffer: fakePodsToBuffers}
}

// GetCapacityBuffer returns the CapacityBuffer associated with the given fake pod UID.
// Returns nil if no such mapping exists.
func (r *Registry) GetCapacityBuffer(fakePodUID types.UID) *v1beta1.CapacityBuffer {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.fakePodsUIDToBuffer[fakePodUID]
}

// SetCapacityBuffer registers a mapping between a fake pod's UID and the CapacityBuffer it was created from.
func (r *Registry) SetCapacityBuffer(fakePodUID types.UID, buffer *v1beta1.CapacityBuffer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.fakePodsUIDToBuffer[fakePodUID] = buffer
}

// UnsetCapacityBuffer removes the mapping for the specified fake pod UID.
func (r *Registry) UnsetCapacityBuffer(fakePodUID types.UID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.fakePodsUIDToBuffer, fakePodUID)
}

// Clear removes all mappings from the registry, effectively resetting it.
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	clear(r.fakePodsUIDToBuffer)
}
