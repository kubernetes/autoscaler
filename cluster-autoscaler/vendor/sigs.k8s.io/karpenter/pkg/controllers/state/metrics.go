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

package state

import (
	opmetrics "github.com/awslabs/operatorpkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"sigs.k8s.io/karpenter/pkg/metrics"
)

const (
	stateSubsystem = "cluster_state"
)

var (
	ClusterStateNodesCount = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: stateSubsystem,
			Name:      "node_count",
			Help:      "Current count of nodes in cluster state",
		},
		[]string{},
	)
	ClusterStateSynced = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: stateSubsystem,
			Name:      "synced",
			Help:      "Returns 1 if cluster state is synced and 0 otherwise. Synced checks that nodeclaims and nodes that are stored in the APIServer have the same representation as Karpenter's cluster state",
		},
		[]string{},
	)
	ClusterStateUnsyncedTimeSeconds = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: stateSubsystem,
			Name:      "unsynced_time_seconds",
			Help:      "The time for which cluster state is not synced",
		},
		[]string{},
	)
	PodSchedulingDecisionSeconds = opmetrics.NewPrometheusHistogram(
		crmetrics.Registry,
		prometheus.HistogramOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.PodSubsystem,
			Name:      "scheduling_decision_duration_seconds",
			Help:      "The time it takes for Karpenter to first try to schedule a pod after it's been seen.",
			Buckets:   metrics.DurationBuckets(),
		},
		[]string{},
	)
)
