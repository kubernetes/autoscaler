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

package scheduling

import (
	opmetrics "github.com/awslabs/operatorpkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"sigs.k8s.io/karpenter/pkg/metrics"
)

const (
	ControllerLabel    = "controller"
	schedulingIDLabel  = "scheduling_id"
	schedulerSubsystem = "scheduler"
)

var (
	DurationSeconds = opmetrics.NewPrometheusHistogram(
		crmetrics.Registry,
		prometheus.HistogramOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "scheduling_duration_seconds",
			Help:      "Duration of scheduling simulations used for deprovisioning and provisioning in seconds.",
			Buckets:   metrics.DurationBuckets(),
		},
		[]string{
			ControllerLabel,
		},
	)
	QueueDepth = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "queue_depth",
			Help:      "The number of pods currently waiting to be scheduled.",
		},
		[]string{
			ControllerLabel,
			schedulingIDLabel,
		},
	)
	UnfinishedWorkSeconds = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "unfinished_work_seconds",
			Help:      "How many seconds of work has been done that is in progress and hasn't been observed by scheduling_duration_seconds.",
		},
		[]string{
			ControllerLabel,
			schedulingIDLabel,
		},
	)
	IgnoredPodCount = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "ignored_pods_count",
			Help:      "Number of pods ignored during scheduling by Karpenter",
		},
		[]string{},
	)
	UnschedulablePodsCount = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "unschedulable_pods_count",
			Help:      "The number of unschedulable Pods.",
		},
		[]string{
			ControllerLabel,
		},
	)
	PendingPodsByEffectiveZone = opmetrics.NewPrometheusGauge(
		crmetrics.Registry,
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: schedulerSubsystem,
			Name:      "pending_pods_by_effective_zone_count",
			Help:      "Pending pods dimensioned by effective zone constraint, or the intersection of pod-level zone signals, volume topology (PVC zones), and topology constraints. Values: specific zone name (e.g., 'us-west-2a'), 'flexible' (multiple zones), or 'none' (no valid intersection).",
		},
		[]string{
			ControllerLabel,
			"zone",
		},
	)
)
