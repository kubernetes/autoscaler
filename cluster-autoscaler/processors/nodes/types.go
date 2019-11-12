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

package nodes

import (
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// ScaleDownNodeProcessor contains methods to get harbor and scale down candidate nodes
type ScaleDownNodeProcessor interface {
	// GetPodDestinationCandidates returns nodes that potentially could act as destinations for pods
	// that would become unscheduled after a scale down.
	GetPodDestinationCandidates(*context.AutoscalingContext, []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError)
	// GetScaleDownCandidates returns nodes that potentially could be scaled down.
	GetScaleDownCandidates(*context.AutoscalingContext, []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError)
	// CleanUp is called at CA termination
	CleanUp()
}
