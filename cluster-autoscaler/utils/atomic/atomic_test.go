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

package atomic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
)

func TestIsAtomicNodeGroup(t *testing.T) {
	n1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "n1"},
		Spec:       apiv1.NodeSpec{ProviderID: "n1"},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	atomicOpts := &config.NodeGroupAutoscalingOptions{
		ZeroOrMaxNodeScaling: true,
	}
	provider.AddNodeGroupWithCustomOptions("ng-atomic", 0, 10, 2, atomicOpts)
	provider.AddNode("ng-atomic", n1)

	n2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "n2"},
		Spec:       apiv1.NodeSpec{ProviderID: "n2"},
	}
	provider.AddNodeGroup("ng-standard", 0, 10, 2)
	provider.AddNode("ng-standard", n2)

	autoscalingCtx := ca_context.AutoscalingContext{
		CloudProvider: provider,
	}

	ng, isAtomic := IsAtomicNodeGroup(&autoscalingCtx, n1)
	assert.True(t, isAtomic)
	assert.Equal(t, "ng-atomic", ng.Id())

	ng2, isAtomic2 := IsAtomicNodeGroup(&autoscalingCtx, n2)
	assert.False(t, isAtomic2)
	assert.Nil(t, ng2)
}

func TestCountRegisteredNodesForGroup(t *testing.T) {
	n1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "n1"},
		Spec:       apiv1.NodeSpec{ProviderID: "n1"},
	}
	n2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "n2",
			Annotations: map[string]string{annotations.NodeUpcomingAnnotation: "true"},
		},
		Spec: apiv1.NodeSpec{ProviderID: "n2"},
	}
	n3 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "n3"},
		Spec:       apiv1.NodeSpec{ProviderID: "n3"},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 0, 10, 3)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)

	ng := provider.GetNodeGroup("ng1")
	allNodes := []*apiv1.Node{n1, n2, n3}
	count, err := CountRegisteredNodesForGroup(ng, allNodes)
	assert.NoError(t, err)
	// n2 has NodeUpcomingAnnotation=true, so registered count should be 2
	assert.Equal(t, 2, count)
}
