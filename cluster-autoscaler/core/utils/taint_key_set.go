/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// TaintKeySet is a set of taint key
type TaintKeySet map[string]bool

var (
	// NodeConditionTaints lists taint keys used as node conditions
	NodeConditionTaints = TaintKeySet{
		schedulerapi.TaintNodeNotReady:           true,
		schedulerapi.TaintNodeUnreachable:        true,
		schedulerapi.TaintNodeUnschedulable:      true,
		schedulerapi.TaintNodeMemoryPressure:     true,
		schedulerapi.TaintNodeDiskPressure:       true,
		schedulerapi.TaintNodeNetworkUnavailable: true,
		schedulerapi.TaintNodePIDPressure:        true,
		schedulerapi.TaintExternalCloudProvider:  true,
		schedulerapi.TaintNodeShutdown:           true,
	}
)
