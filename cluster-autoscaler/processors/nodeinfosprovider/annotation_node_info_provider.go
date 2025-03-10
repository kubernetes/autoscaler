/*
Copyright 2022 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// AnnotationNodeInfoProvider is a wrapper for MixedTemplateNodeInfoProvider.
type AnnotationNodeInfoProvider struct {
	templateNodeInfoProvider TemplateNodeInfoProvider
}

// NewAnnotationNodeInfoProvider returns AnnotationNodeInfoProvider wrapping MixedTemplateNodeInfoProvider.
func NewAnnotationNodeInfoProvider(t *time.Duration, forceDaemonSets bool) *AnnotationNodeInfoProvider {
	return &AnnotationNodeInfoProvider{
		templateNodeInfoProvider: NewMixedTemplateNodeInfoProvider(t, forceDaemonSets),
	}
}

// NewCustomAnnotationNodeInfoProvider returns AnnotationNodeInfoProvider wrapping TemplateNodeInfoProvider.
func NewCustomAnnotationNodeInfoProvider(templateNodeInfoProvider TemplateNodeInfoProvider) *AnnotationNodeInfoProvider {
	return &AnnotationNodeInfoProvider{
		templateNodeInfoProvider: templateNodeInfoProvider,
	}
}

// Process returns the nodeInfos set for this cluster.
func (p *AnnotationNodeInfoProvider) Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError) {
	nodeInfos, err := p.templateNodeInfoProvider.Process(ctx, nodes, daemonsets, taintConfig, currentTime)
	if err != nil {
		return nil, err
	}
	// Add annotations to the NodeInfo to use later in expander.
	nodeGroups := ctx.CloudProvider.NodeGroups()
	for _, ng := range nodeGroups {
		if nodeInfo, ok := nodeInfos[ng.Id()]; ok {
			template, err := ng.TemplateNodeInfo()
			if err != nil {
				continue
			}
			for key, val := range template.Node().Annotations {
				if _, ok := nodeInfo.Node().Annotations[key]; !ok {
					nodeInfo.Node().Annotations[key] = val
				}
			}
		}
	}
	return nodeInfos, nil
}

// CleanUp cleans up processor's internal structures.
func (p *AnnotationNodeInfoProvider) CleanUp() {
}
