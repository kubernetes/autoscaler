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

package exoscale

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale"
)

func testSetupNodeGroup(url string) (*NodeGroup, error) {
	os.Setenv("EXOSCALE_API_KEY", "KEY")
	os.Setenv("EXOSCALE_API_SECRET", "SECRET")
	os.Setenv("EXOSCALE_API_ENDPOINT", url)

	manager, err := newManager()
	if err != nil {
		return nil, err
	}

	nodeGroup := &NodeGroup{
		id:      testMockInstancePool1ID,
		manager: manager,
		instancePool: &egoscale.InstancePool{
			ID:     egoscale.MustParseUUID(testMockInstancePool1ID),
			Size:   1,
			ZoneID: egoscale.MustParseUUID(testMockGetZoneID),
			VirtualMachines: []egoscale.VirtualMachine{
				{
					ID:    egoscale.MustParseUUID(testMockInstance1ID),
					State: string(egoscale.VirtualMachineRunning),
				},
			},
		},
	}

	return nodeGroup, nil
}

func TestNodeGroup_MaxSize(t *testing.T) {
	ts := newTestServer(
		testHTTPResponse{200, testMockResourceLimit},
	)

	nodeGroup, err := testSetupNodeGroup(ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	assert.Equal(t, testMockResourceLimitMax, nodeGroup.MaxSize())
}

func TestNodeGroup_MinSize(t *testing.T) {
	nodeGroup, err := testSetupNodeGroup("url")
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	assert.Equal(t, 1, nodeGroup.MinSize())
}

func TestNodeGroup_TargetSize(t *testing.T) {
	nodeGroup, err := testSetupNodeGroup("url")
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	target, err := nodeGroup.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, nodeGroup.instancePool.Size, target)
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ts := newTestServer(
		testHTTPResponse{200, testMockResourceLimit},
		testHTTPResponse{200, testMockBooleanResponse("scaleinstancepoolresponse")},
		testHTTPResponse{200, testMockInstancePool1},
	)

	nodeGroup, err := testSetupNodeGroup(ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	err = nodeGroup.IncreaseSize(2)
	assert.NoError(t, err)
}

func TestNodeGroup_IncreaseSizeFailure(t *testing.T) {
	ts := newTestServer(
		testHTTPResponse{200, testMockResourceLimit},
		testHTTPResponse{200, testMockBooleanResponse("scaleinstancepoolresponse")},
		testHTTPResponse{200, testMockInstancePool1},
	)

	nodeGroup, err := testSetupNodeGroup(ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	err = nodeGroup.IncreaseSize(testMockResourceLimitMax + 1)
	assert.Error(t, err)
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	ts := newTestServer(
		testHTTPResponse{200, testMockInstancePool1},
		testHTTPResponse{200, testMockBooleanResponse("evictinstancepoolmembersresponse")},
		testHTTPResponse{200, testMockInstancePool1},
	)

	nodeGroup, err := testSetupNodeGroup(ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance1ID),
		},
	}

	err = nodeGroup.DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
}

func TestNodeGroup_Id(t *testing.T) {
	nodeGroup, err := testSetupNodeGroup("url")
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	id := nodeGroup.Id()
	assert.Equal(t, testMockInstancePool1ID, id)
}

func TestNodeGroup_Nodes(t *testing.T) {
	nodeGroup, err := testSetupNodeGroup("url")
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	instances, err := nodeGroup.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, testMockInstance1ID, toNodeID(instances[0].Id))
	assert.Equal(t, cloudprovider.InstanceRunning, instances[0].Status.State)
}

func TestNodeGroup_Exist(t *testing.T) {
	nodeGroup, err := testSetupNodeGroup("url")
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	exist := nodeGroup.Exist()
	assert.True(t, exist)
}
