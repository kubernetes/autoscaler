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

package metrics

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
)

// ReconciliationCache is a thread safe map for buffers last reconciliation time
type ReconciliationCache struct {
	reconciledBuffers map[types.UID]time.Time
	mutex             sync.RWMutex
}

// NewReconciliationCache creates a new instance of ReconciliationCache
func NewReconciliationCache() *ReconciliationCache {
	return &ReconciliationCache{
		reconciledBuffers: make(map[types.UID]time.Time),
	}
}

// Update updates the underlying map with the provided buffers and time.
func (r *ReconciliationCache) Update(buffers []*v1beta1.CapacityBuffer, t time.Time) {
	if len(buffers) == 0 {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, buffer := range buffers {
		r.reconciledBuffers[buffer.UID] = t
	}
}

// Prune removes entries from the cache that are not present in the provided buffers list.
func (r *ReconciliationCache) Prune(buffers []*v1beta1.CapacityBuffer) {
	existingBuffers := sets.New[types.UID]()
	for _, buffer := range buffers {
		existingBuffers.Insert(buffer.UID)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for uid := range r.reconciledBuffers {
		if !existingBuffers.Has(uid) {
			delete(r.reconciledBuffers, uid)
		}
	}
}

// Snapshot creates and returns a copy of the current map.
func (r *ReconciliationCache) Snapshot() map[types.UID]time.Time {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	snapshot := make(map[types.UID]time.Time, len(r.reconciledBuffers))
	for key, value := range r.reconciledBuffers {
		snapshot[key] = value
	}

	return snapshot
}
