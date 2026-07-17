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

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
)

func TestSetUnremovableNodesInfo(t *testing.T) {
	n1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "n1",
		},
	}
	n2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "n2",
		},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	// n2 is not added to any node group

	unremovableNodes := []*simulator.UnremovableNode{
		{
			Node:   n1,
			Reason: simulator.NotUnderutilized,
		},
		{
			Node:   n2,
			Reason: simulator.NotUnderutilized,
		},
	}

	nodeUtilizationMap := map[string]utilization.Info{
		"n1": {Utilization: 0.1},
		"n2": {Utilization: 0.1},
	}

	s := &ScaleDownStatus{}
	s.SetUnremovableNodesInfo(unremovableNodes, nodeUtilizationMap, provider)

	assert.Equal(t, 1, len(s.UnremovableNodes))
	assert.Equal(t, "n1", s.UnremovableNodes[0].Node.Name)
}
