/*
Copyright 2016 The Kubernetes Authors.

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

package clusterstate

import (
	"testing"
	"time"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider/test"
	. "k8s.io/contrib/cluster-autoscaler/utils/test"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestOKWithScaleUp(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	setReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	setReadyState(ng2_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        4,
		Time:            now,
		ExpectedAddTime: now.Add(time.Minute),
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy(now))
}

func TestOKOneUnreadyNode(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	setReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	setReadyState(ng2_1, false, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy(now))
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestMissingNodes(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	setReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	setReadyState(ng2_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy(now))
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestToManyUnready(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	setReadyState(ng1_1, false, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	setReadyState(ng2_1, false, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)

	assert.NotNil(t, provider)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.False(t, clusterstate.IsClusterHealthy(now))
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestExpiredScaleUp(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	setReadyState(ng1_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)

	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        4,
		Time:            now.Add(-3 * time.Minute),
		ExpectedAddTime: now.Add(-1 * time.Minute),
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy(now))
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func setReadyState(node *apiv1.Node, ready bool, lastTransition time.Time) {
	if ready {
		node.Status.Conditions = append(node.Status.Conditions,
			apiv1.NodeCondition{
				Type:               apiv1.NodeReady,
				Status:             apiv1.ConditionTrue,
				LastTransitionTime: metav1.Time{Time: lastTransition},
			})
	} else {
		node.Status.Conditions = append(node.Status.Conditions,
			apiv1.NodeCondition{
				Type:               apiv1.NodeReady,
				Status:             apiv1.ConditionFalse,
				LastTransitionTime: metav1.Time{Time: lastTransition},
			})
	}
}

func TestRegisterScaleDown(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)

	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	})

	now := time.Now()

	clusterstate.RegisterScaleDown(&ScaleDownRequest{
		NodeGroupName:      "ng1",
		NodeName:           "ng1-1",
		ExpectedDeleteTime: now.Add(time.Minute),
		Time:               now,
	})
	assert.Equal(t, 1, len(clusterstate.scaleDownRequests))
	clusterstate.cleanUp(now.Add(5 * time.Minute))
	assert.Equal(t, 0, len(clusterstate.scaleDownRequests))
}
