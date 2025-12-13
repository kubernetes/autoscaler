/*
Copyright 2021 The Kubernetes Authors.

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
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"k8s.io/component-base/metrics"
)

func newCaMetricsWithRegistry(registry metrics.KubeRegistry) *caMetrics {
	reg := newCaMetrics()
	reg.registry = registry
	return reg
}

func TestDisabledPerNodeGroupMetrics(t *testing.T) {
	// Use a custom registry for isolation to avoid panics from re-registering metrics.
	reg := metrics.NewKubeRegistry()
	assert.NotNil(t, reg)
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(false)
	assert.False(t, m.nodesGroupMinNodes.IsCreated())
	assert.False(t, m.nodesGroupMaxNodes.IsCreated())
}

func TestEnabledPerNodeGroupMetrics(t *testing.T) {
	// Use a custom registry for isolation
	reg := metrics.NewKubeRegistry()
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(true)
	assert.True(t, m.nodesGroupMinNodes.IsCreated())
	assert.True(t, m.nodesGroupMaxNodes.IsCreated())

	m.UpdateNodeGroupMin("foo", 2)
	m.UpdateNodeGroupMax("foo", 100)

	assert.Equal(t, 2, int(testutil.ToFloat64(m.nodesGroupMinNodes.GaugeVec.WithLabelValues("foo"))))
	assert.Equal(t, 100, int(testutil.ToFloat64(m.nodesGroupMaxNodes.GaugeVec.WithLabelValues("foo"))))
}
