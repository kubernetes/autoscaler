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

package nodeinfosprovider

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// TemplateNodeInfoProvider is provides the initial nodeInfos set.
type TemplateNodeInfoProvider interface {
	// Process returns a map of nodeInfos for node groups.
	Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// NewDefaultTemplateNodeInfoProvider returns a default TemplateNodeInfoProvider.
func NewDefaultTemplateNodeInfoProvider(time *time.Duration, forceDaemonSets bool) TemplateNodeInfoProvider {
	return NewMixedTemplateNodeInfoProvider(time, forceDaemonSets)
}
