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

package nodeinfosprovider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestTemplateOnlyNodeInfoProviderProcess(t *testing.T) {
	predicateChecker, err := predicatechecker.NewTestPredicateChecker()
	assert.NoError(t, err)

	tni := schedulerframework.NewNodeInfo()
	tn := BuildTestNode("tn", 100, 100)
	tn.SetLabels(map[string]string{apiv1.LabelTopologyZone: "planet-earth"})
	tni.SetNode(tn)

	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, nil, nil,
		map[string]*schedulerframework.NodeInfo{"ng1": tni, "ng2": tni})
	provider1.AddNodeGroup("ng1", 1, 10, 1)
	provider1.AddNodeGroup("ng2", 2, 20, 2)

	ctx := &context.AutoscalingContext{
		PredicateChecker: predicateChecker,
		CloudProvider:    provider1,
	}

	processor := NewTemplateOnlyNodeInfoProvider(nil)
	res, err := processor.Process(ctx, nil, nil, nil, time.Now())

	// nodegroups providing templates
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	assert.Contains(t, res, "ng1")
	assert.Contains(t, res, "ng2")
	assert.Contains(t, res["ng1"].Node().GetLabels(), apiv1.LabelZoneFailureDomain)
	assert.Equal(t, res["ng1"].Node().GetLabels()[apiv1.LabelZoneFailureDomain], "planet-earth")
}
