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
	opmetrics "github.com/awslabs/operatorpkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	NodeSubsystem      = "nodes"
	NodeClaimSubsystem = "nodeclaims"
	NodePoolSubsystem  = "nodepools"
	PodSubsystem       = "pods"
)

var (
	NodeClaimsCreatedTotal = opmetrics.NewPrometheusCounter(
		crmetrics.Registry,
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: NodeClaimSubsystem,
			Name:      "created_total",
			Help:      "Number of nodeclaims created in total by Karpenter. Labeled by reason the nodeclaim was created, the owning nodepool, and if min values was relaxed for this nodeclaim.",
		},
		[]string{
			ReasonLabel,
			NodePoolLabel,
			MinValuesRelaxedLabel,
		},
	)
	NodeClaimsTerminatedTotal = opmetrics.NewPrometheusCounter(
		crmetrics.Registry,
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: NodeClaimSubsystem,
			Name:      "terminated_total",
			Help:      "Number of nodeclaims terminated in total by Karpenter. Labeled by the owning nodepool, capacity type, and zone.",
		},
		[]string{
			NodePoolLabel,
			CapacityTypeLabel,
			ZoneLabel,
		},
	)
	NodeClaimsDisruptedTotal = opmetrics.NewPrometheusCounter(
		crmetrics.Registry,
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: NodeClaimSubsystem,
			Name:      "disrupted_total",
			Help:      "Number of nodeclaims disrupted in total by Karpenter. Labeled by reason the nodeclaim was disrupted and the owning nodepool.",
		},
		[]string{
			ReasonLabel,
			NodePoolLabel,
			CapacityTypeLabel,
		},
	)
	NodesCreatedTotal = opmetrics.NewPrometheusCounter(
		crmetrics.Registry,
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: NodeSubsystem,
			Name:      "created_total",
			Help:      "Number of nodes created in total by Karpenter. Labeled by owning nodepool and zone.",
		},
		[]string{
			NodePoolLabel,
			ZoneLabel,
		},
	)
	NodesTerminatedTotal = opmetrics.NewPrometheusCounter(
		crmetrics.Registry,
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: NodeSubsystem,
			Name:      "terminated_total",
			Help:      "Number of nodes terminated in total by Karpenter. Labeled by owning nodepool and zone.",
		},
		[]string{
			NodePoolLabel,
			ZoneLabel,
		},
	)
)
