/*
Copyright 2024 The Kubernetes Authors.

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

package karpenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
)

func TestSerializeTaints(t *testing.T) {
	assert.Equal(t, "", SerializeTaints(nil))
	assert.Equal(t, "", SerializeTaints([]apiv1.Taint{}))

	singleTaint := []apiv1.Taint{
		{Key: "key1", Value: "val1", Effect: apiv1.TaintEffectNoSchedule},
	}
	assert.Equal(t, "key1=val1:NoSchedule", SerializeTaints(singleTaint))

	// Verify deterministic sorting
	multipleTaints := []apiv1.Taint{
		{Key: "z-key", Value: "valZ", Effect: apiv1.TaintEffectNoExecute},
		{Key: "a-key", Value: "valA", Effect: apiv1.TaintEffectNoSchedule},
	}
	assert.Equal(t, "a-key=valA:NoSchedule,z-key=valZ:NoExecute", SerializeTaints(multipleTaints))
}

func TestInstanceTypeNameFromLabels(t *testing.T) {
	// Stable label precedence
	labels1 := map[string]string{
		apiv1.LabelInstanceTypeStable: "c2-standard-4",
		apiv1.LabelInstanceType:       "legacy-name",
	}
	assert.Equal(t, "c2-standard-4", InstanceTypeNameFromLabels(labels1, "fallback"))

	// Legacy label fallback
	labels2 := map[string]string{
		apiv1.LabelInstanceType: "n1-standard-1",
	}
	assert.Equal(t, "n1-standard-1", InstanceTypeNameFromLabels(labels2, "fallback"))

	// Default fallback
	assert.Equal(t, "fallback", InstanceTypeNameFromLabels(map[string]string{}, "fallback"))
}

func TestConversionResultQueries(t *testing.T) {
	var nilRes *ConversionResult
	assert.Nil(t, nilRes.NodeGroupsFor("pool-1", "it-1"))
	assert.Equal(t, "", nilRes.PoolForNodeGroup("ng-1"))

	ng1 := test.NewTestNodeGroup("ng-1", 10, 1, 1, true, false, "e2-standard-2", nil, nil)
	res := &ConversionResult{
		PoolITToNodeGroups: map[string]map[string][]cloudprovider.NodeGroup{
			"pool-1": {
				"e2-standard-2": {ng1},
			},
		},
		NodeGroupToPool: map[string]string{
			"ng-1": "pool-1",
		},
	}

	assert.Equal(t, []cloudprovider.NodeGroup{ng1}, res.NodeGroupsFor("pool-1", "e2-standard-2"))
	assert.Nil(t, res.NodeGroupsFor("pool-1", "missing-it"))
	assert.Nil(t, res.NodeGroupsFor("missing-pool", "e2-standard-2"))

	assert.Equal(t, "pool-1", res.PoolForNodeGroup("ng-1"))
	assert.Equal(t, "", res.PoolForNodeGroup("missing-ng"))
}
