/*
Copyright 2017 The Kubernetes Authors.

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

package spot

import (
	"sort"
	"sync"
	"time"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

// ErrEmptySpotPriceHistory implements the error interface
var ErrEmptySpotPriceHistory = errors.New("empty spot price history")

// History represents a set of ordered EC2 spot price items
// It implements the sort.Interface
type History struct {
	items    api.SpotPriceItems
	lastSync time.Time
	maxAge   time.Duration
	mu       sync.RWMutex
}

// Slice returns a copy of the internal spot price item list
func (h *History) Slice() api.SpotPriceItems {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.items[:]
}

// Empty checks whether the history is empty or not
func (h *History) Empty() bool {
	return h.Len() == 0
}

// Len returns the length of the history
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.items.Len()
}

// Housekeep drops items older than maxAge and sorts the history
func (h *History) Housekeep() {
	h.mu.Lock()
	defer h.mu.Unlock()

	c := make(api.SpotPriceItems, 0)

	deadEnd := time.Now().Truncate(h.maxAge)

	for _, item := range h.items {
		if item.Timestamp.Before(deadEnd) {
			continue
		}

		c = append(c, item)
	}

	sort.Sort(c)

	h.items = c
}

// Add adds sorted api.SpotPriceItems and sets the last-sync to current time
func (h *History) Add(items api.SpotPriceItems) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sort.Sort(items)

	h.items = append(h.items, items...)
	h.lastSync = time.Now()
}

// LastItem returns the last item of the history
func (h *History) LastItem() (api.SpotPriceItem, error) {
	if h.Empty() {
		return api.EmptySpotPriceItem, ErrEmptySpotPriceHistory
	}

	idx := h.items.Len() - 1

	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.items[idx], nil
}

// SetLastSync sets last-sync to current time
func (h *History) SetLastSync() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastSync = time.Now()
}

// LastSync returns the time of the last sync
func (h *History) LastSync() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.lastSync
}
