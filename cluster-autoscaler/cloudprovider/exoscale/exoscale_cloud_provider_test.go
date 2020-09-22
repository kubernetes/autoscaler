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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func testSetupCloudProvider(url string) (*exoscaleCloudProvider, error) {
	os.Setenv("EXOSCALE_API_KEY", "KEY")
	os.Setenv("EXOSCALE_API_SECRET", "SECRET")
	os.Setenv("EXOSCALE_API_ENDPOINT", url)

	manager, err := newManager()
	if err != nil {
		return nil, err
	}

	provider, err := newExoscaleCloudProvider(manager, &cloudprovider.ResourceLimiter{})
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func TestExoscaleCloudProvider_Name(t *testing.T) {
	provider, err := testSetupCloudProvider("url")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "exoscale", provider.Name())
}

func TestExoscaleCloudProvider_NodeGroupForNode(t *testing.T) {
	url := testMockAPICloudProviderTest()
	assert.NotEmpty(t, url)

	provider, err := testSetupCloudProvider(url)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance1ID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testMockGetZoneName,
			},
		},
	}

	nodeGroup, err := provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	assert.Equal(t, testMockInstancePool1ID, nodeGroup.Id())

	// Testing a second time with a node belonging to a different
	// node group.
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance2ID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testMockGetZoneName,
			},
		},
	}

	nodeGroup, err = provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	assert.Equal(t, testMockInstancePool2ID, nodeGroup.Id())
}

func TestExoscaleCloudProvider_NodeGroupForNodeWithoutZone(t *testing.T) {
	url := testMockAPICloudProviderTest()
	assert.NotEmpty(t, url)

	provider, err := testSetupCloudProvider(url)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance1ID),
		},
	}

	nodeGroup, err := provider.NodeGroupForNode(node)
	assert.Error(t, err)
	assert.Nil(t, nodeGroup)
}

func TestExoscaleCloudProvider_NodeGroups(t *testing.T) {
	url := testMockAPICloudProviderTest()
	assert.NotEmpty(t, url)

	provider, err := testSetupCloudProvider(url)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance1ID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testMockGetZoneName,
			},
		},
	}

	// Referencing a second node group to test if the cloud provider
	// manager cache is successfully updated.
	nodeGroup, err := provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testMockInstance2ID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testMockGetZoneName,
			},
		},
	}

	nodeGroup, err = provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)

	nodeGroups := provider.NodeGroups()
	assert.Len(t, nodeGroups, 2)
	assert.Equal(t, testMockInstancePool1ID, nodeGroups[0].Id())
	assert.Equal(t, testMockInstancePool2ID, nodeGroups[1].Id())
}
