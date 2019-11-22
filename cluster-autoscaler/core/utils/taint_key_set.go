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
	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// TaintKeySet is a set of taint key
type TaintKeySet map[string]bool

var (
	// NodeConditionTaints lists taint keys used as node conditions
	NodeConditionTaints = TaintKeySet{
		v1.TaintNodeNotReady:                    true,
		v1.TaintNodeUnreachable:                 true,
		v1.TaintNodeUnschedulable:               true,
		v1.TaintNodeMemoryPressure:              true,
		v1.TaintNodeDiskPressure:                true,
		v1.TaintNodeNetworkUnavailable:          true,
		v1.TaintNodePIDPressure:                 true,
		schedulerapi.TaintExternalCloudProvider: true,
		schedulerapi.TaintNodeShutdown:          true,
	}
)
