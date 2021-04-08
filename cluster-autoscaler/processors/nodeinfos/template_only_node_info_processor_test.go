/*
Copyright 2021 The Kubernetes Authors.

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

package nodeinfos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestTemplateOnlyNodeInfoProcessorProcess(t *testing.T) {
	predicateChecker, err := simulator.NewTestPredicateChecker()
	assert.NoError(t, err)

	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(BuildTestNode("tn", 100, 100))

	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, nil, nil,
		map[string]*schedulerframework.NodeInfo{"ng1": tni, "ng2": tni})
	provider1.AddNodeGroup("ng1", 1, 10, 1)
	provider1.AddNodeGroup("ng2", 2, 20, 2)

	ctx := &context.AutoscalingContext{
		CloudProvider:    provider1,
		PredicateChecker: predicateChecker,
	}

	processor := NewTemplateOnlyNodeInfoProcessor()
	res, err := processor.Process(ctx, nil, nil, nil)

	// nodegroups providing templates
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	assert.Contains(t, res, "ng1")
	assert.Contains(t, res, "ng2")

	// nodegroup not providing templates
	provider1.AddNodeGroup("ng3", 0, 1000, 0)
	_, err = processor.Process(ctx, nil, nil, nil)
	assert.Error(t, err)
}
