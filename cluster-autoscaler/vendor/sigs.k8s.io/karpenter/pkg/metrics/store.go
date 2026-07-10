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

	opmetrics "github.com/awslabs/operatorpkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
)

// Store is a mapping from a key to a list of Metrics
// Each time Update() is called for a key on Store, the metric store ensures that all metrics are "refreshed"
// for all currently tracked metrics assigned to the key. This means that any metric that contains the same labels
// as a previous metric will be updated through the standard prometheus.Gauge metric Set() call while any metric with
// different labels than the recently fired metrics will be removed from the prometheus client response and the Store
type Store struct {
	sync.Mutex
	store map[string][]*StoreMetric
}

func NewStore() *Store {
	return &Store{store: map[string][]*StoreMetric{}}
}

// StoreMetric is a single state metric associated with a metrics.Gauge
type StoreMetric struct {
	opmetrics.GaugeMetric
	Value  float64
	Labels prometheus.Labels
}

// update is an internal non-thread-safe method for updating metrics given a key in the Store
func (s *Store) update(key string, metrics []*StoreMetric) {
	for _, metric := range metrics {
		metric.Set(metric.Value, metric.Labels)
	}
	// Cleanup old metrics if the old metric family has metrics that weren't updated by this round of metrics
	if oldMetrics, ok := s.store[key]; ok {
		for _, oldMetric := range oldMetrics {
			if _, ok = lo.Find(metrics, func(m *StoreMetric) bool {
				return oldMetric.GaugeMetric == m.GaugeMetric && equality.Semantic.DeepEqual(oldMetric.Labels, m.Labels)
			}); !ok {
				oldMetric.Delete(oldMetric.Labels)
			}
		}
	}
	s.store[key] = metrics
}

// Update calls the update() method internally
func (s *Store) Update(key string, metrics []*StoreMetric) {
	s.Lock()
	defer s.Unlock()

	s.update(key, metrics)
}

// ReplaceAll replaces all metrics in the store with the new metrics passes into the ReplaceAll function. This calls
// the update method as normal for any keys that match existing keys while removing any keys that existed in the old
// store but don't exist in the new store.
func (s *Store) ReplaceAll(newStore map[string][]*StoreMetric) {
	s.Lock()
	defer s.Unlock()

	for k, v := range newStore {
		s.update(k, v)
	}
	for k := range s.store {
		if _, ok := newStore[k]; !ok {
			s.delete(k)
		}
	}
}

// delete is an internal non-thread-safe method for deleting metrics given a key in the Store
func (s *Store) delete(key string) {
	if metrics, ok := s.store[key]; ok {
		for _, metric := range metrics {
			metric.Delete(metric.Labels)
		}
		delete(s.store, key)
	}
}

// Delete calls the delete() method internally
func (s *Store) Delete(key string) {
	s.Lock()
	defer s.Unlock()

	s.delete(key)
}
