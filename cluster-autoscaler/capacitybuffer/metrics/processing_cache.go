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

	"k8s.io/apimachinery/pkg/util/sets"
)

// ProcessingCache is a thread safe map for buffers last processing time
type ProcessingCache struct {
	mutex            *sync.RWMutex
	processedBuffers map[string]time.Time
}

// NewProcessingCache creates a new instance of ProcessingCache
func NewProcessingCache() *ProcessingCache {
	return &ProcessingCache{
		mutex:            &sync.RWMutex{},
		processedBuffers: make(map[string]time.Time),
	}
}

// Update updates the underlying map with new entries safely.
func (p *ProcessingCache) Update(newEntries map[string]time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for k, v := range newEntries {
		p.processedBuffers[k] = v
	}
}

// Prune removes entries from the cache that are not present in the provided supportedUIDs list.
func (p *ProcessingCache) Prune(supportedUIDs []string) {
	supported := sets.New(supportedUIDs...)
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for uid := range p.processedBuffers {
		if !supported.Has(uid) {
			delete(p.processedBuffers, uid)
		}
	}
}

// Snapshot creates and returns a copy of the current map.
func (p *ProcessingCache) Snapshot() map[string]time.Time {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	snapshot := make(map[string]time.Time, len(p.processedBuffers))
	for key, value := range p.processedBuffers {
		snapshot[key] = value
	}

	return snapshot
}
