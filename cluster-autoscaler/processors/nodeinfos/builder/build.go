/*
Copyright 2020 The Kubernetes Authors.

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

package builder

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfos"
)

var availableNodeInfoProcessors = map[string]BuilderFunc{}

// Build returns a new nodeinfos.NodeInfoProcessor instance with the requested implementation
func Build(opts *core.AutoscalerOptions) (nodeinfos.NodeInfoProcessor, error) {
	if buildFunc, found := availableNodeInfoProcessors[opts.AutoscalingOptions.NodeInfoProcessorName]; found {
		return buildFunc(opts), nil
	}
	return nodeinfos.NewDefaultNodeInfoProcessor(), fmt.Errorf("NodeInfoProcessor %s not found", opts.AutoscalingOptions.NodeInfoProcessorName)
}

// BuilderFunc corresponds to nodeinfos.NodeInfoProcessor
type BuilderFunc func(opts *core.AutoscalerOptions) nodeinfos.NodeInfoProcessor

// Register used to register a nodeinfos.NodeInfoProcessor implementation.
func Register(name string, builderFunc BuilderFunc) {
	availableNodeInfoProcessors[name] = builderFunc
}

// GetAvailableNodeInfoProcessors returns the list of registered NodeInfoProcessor implementation.
func GetAvailableNodeInfoProcessors() []string {
	output := make([]string, 0, len(availableNodeInfoProcessors))
	for key := range availableNodeInfoProcessors {
		output = append(output, key)
	}
	return output
}
