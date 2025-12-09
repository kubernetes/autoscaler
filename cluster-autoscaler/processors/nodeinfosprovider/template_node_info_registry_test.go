/*
Copyright 2025 The Kubernetes Authors.

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
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

type mockTemplateNodeInfoProvider struct {
	nodeInfos map[string]*framework.NodeInfo
}

func (p *mockTemplateNodeInfoProvider) Process(autoscalingCtx *ca_context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError) {
	return p.nodeInfos, nil
}

func (p *mockTemplateNodeInfoProvider) CleanUp() {}

func TestTemplateNodeInfoRegistry(t *testing.T) {
	mockProvider := &mockTemplateNodeInfoProvider{
		nodeInfos: map[string]*framework.NodeInfo{
			"ng1": {},
		},
	}
	registry := NewTemplateNodeInfoRegistry(mockProvider)

	// Test Process
	_, err := registry.Process(nil, nil, nil, taints.TaintConfig{}, time.Now())
	assert.NoError(t, err)

	// Test GetNodeInfo
	info, found := registry.GetNodeInfo("ng1")
	assert.True(t, found)
	assert.NotNil(t, info)

	info, found = registry.GetNodeInfo("ng2")
	assert.False(t, found)
	assert.Nil(t, info)

	// Test GetNodeInfos
	infos := registry.GetNodeInfos()
	assert.Len(t, infos, 1)
	assert.Contains(t, infos, "ng1")

	// Test Update
	mockProvider.nodeInfos = map[string]*framework.NodeInfo{
		"ng1": {},
		"ng2": {},
	}
	registry.Process(nil, nil, nil, taints.TaintConfig{}, time.Now())

	info, found = registry.GetNodeInfo("ng2")
	assert.True(t, found)
	assert.NotNil(t, info)
}
