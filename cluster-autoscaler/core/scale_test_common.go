/*
Copyright 2018 The Kubernetes Authors.

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

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

type nodeConfig struct {
	name   string
	cpu    int64
	memory int64
	ready  bool
	group  string
}

type podConfig struct {
	name   string
	cpu    int64
	memory int64
	node   string
}

type groupSizeChange struct {
	groupName  string
	sizeChange int
}

type scaleTestConfig struct {
	nodes                  []nodeConfig
	pods                   []podConfig
	extraPods              []podConfig
	expectedScaleUpOptions []groupSizeChange // we expect that all those options should be included in expansion options passed to expander strategy
	scaleUpOptionToChoose  groupSizeChange   // this will be selected by assertingStrategy.BestOption
	expectedFinalScaleUp   groupSizeChange   // we expect this to be delivered via scale-up event
	expectedScaleDowns     []string
	options                context.AutoscalingOptions
}
