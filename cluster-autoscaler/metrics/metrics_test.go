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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
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

func TestUpdateNodesCount(t *testing.T) {
	reg := metrics.NewKubeRegistry()
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(false)

	m.UpdateNodesCount(1, 2, 3, 4, 5, 6)

	assert.Equal(t, 1, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(readyLabel))))
	assert.Equal(t, 2, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(unreadyLabel))))
	assert.Equal(t, 3, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(startingLabel))))
	assert.Equal(t, 4, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(suspendedLabel))))
	assert.Equal(t, 5, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(longUnregisteredLabel))))
	assert.Equal(t, 6, int(testutil.ToFloat64(m.nodesCount.GaugeVec.WithLabelValues(unregisteredLabel))))
}

func TestInitMetricsInitializesFailedNodeCreations(t *testing.T) {
	reg := metrics.NewKubeRegistry()
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(false)

	m.InitMetrics()

	// failed_node_creations_total is only incremented when a scale-up fails, so
	// without initialization it would be absent from the exposition until the
	// first failure. InitMetrics should pre-create a zero-valued series for each
	// FailedScaleUpReason, matching what it already does for failedScaleUpCount.
	assert.Equal(t, 3, testutil.CollectAndCount(m.failedNodeCreationCount.CounterVec))
}

func TestUpdateScaleDownNodeRemovalLatency(t *testing.T) {
	reg := metrics.NewKubeRegistry()
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(false)

	m.UpdateScaleDownNodeRemovalLatency(true, "none", 10*time.Second)
	m.UpdateScaleDownNodeRemovalLatency(false, "BlockedByPod", 20*time.Second)

	var metric1 dto.Metric
	err := m.scaleDownNodeRemovalLatency.HistogramVec.WithLabelValues("true", "none").(prometheus.Histogram).Write(&metric1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), metric1.Histogram.GetSampleCount())
	assert.Equal(t, 10.0, metric1.Histogram.GetSampleSum())

	var metric2 dto.Metric
	err = m.scaleDownNodeRemovalLatency.HistogramVec.WithLabelValues("false", "BlockedByPod").(prometheus.Histogram).Write(&metric2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), metric2.Histogram.GetSampleCount())
	assert.Equal(t, 20.0, metric2.Histogram.GetSampleSum())
}

func TestUpdateNodesCountPerNodeGroup(t *testing.T) {
	reg := metrics.NewKubeRegistry()
	m := newCaMetricsWithRegistry(reg)
	m.RegisterAll(true)

	nodeGroupID := "node-group-id"

	m.UpdateNodesCountPerNodeGroup(1, 2, 3, 4, 5, 6, nodeGroupID)

	assert.Equal(t, 1, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(readyLabel, nodeGroupID))))
	assert.Equal(t, 2, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(unreadyLabel, nodeGroupID))))
	assert.Equal(t, 3, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(startingLabel, nodeGroupID))))
	assert.Equal(t, 4, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(suspendedLabel, nodeGroupID))))
	assert.Equal(t, 5, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(longUnregisteredLabel, nodeGroupID))))
	assert.Equal(t, 6, int(testutil.ToFloat64(m.nodesCountPerNodeGroup.GaugeVec.WithLabelValues(unregisteredLabel, nodeGroupID))))
}
