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

package spotinst

import (
	"context"
	"testing"
	"time"

	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/azure"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/gce"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type groupServiceMock struct {
	mock.Mock
	providerAWS *awsServiceMock
}

func (s *groupServiceMock) CloudProviderAWS() aws.Service {
	return s.providerAWS
}

func (s *groupServiceMock) CloudProviderGCE() gce.Service {
	return nil // not implemented
}

func (s *groupServiceMock) CloudProviderAzure() azure.Service {
	return nil // not implemented
}

type awsServiceMock struct {
	mock.Mock
}

func (s *awsServiceMock) List(ctx context.Context, input *aws.ListGroupsInput) (*aws.ListGroupsOutput, error) {
	return nil, nil
}

func (s *awsServiceMock) Create(ctx context.Context, input *aws.CreateGroupInput) (*aws.CreateGroupOutput, error) {
	return nil, nil
}

func (s *awsServiceMock) Read(ctx context.Context, input *aws.ReadGroupInput) (*aws.ReadGroupOutput, error) {
	out := &aws.ReadGroupOutput{
		Group: &aws.Group{
			Capacity: &aws.Capacity{
				Target: spotinst.Int(2),
			},
		},
	}
	return out, nil
}

func (s *awsServiceMock) Update(ctx context.Context, input *aws.UpdateGroupInput) (*aws.UpdateGroupOutput, error) {
	args := s.Called(ctx, input)
	return args.Get(0).(*aws.UpdateGroupOutput), nil
}

func (s *awsServiceMock) Delete(ctx context.Context, input *aws.DeleteGroupInput) (*aws.DeleteGroupOutput, error) {
	return nil, nil
}

func (s *awsServiceMock) Status(ctx context.Context, input *aws.StatusGroupInput) (*aws.StatusGroupOutput, error) {
	out := &aws.StatusGroupOutput{
		Instances: []*aws.Instance{
			{
				ID: spotinst.String("test-instance-id"),
			},
			{
				ID: spotinst.String("second-test-instance-id"),
			},
		},
	}
	return out, nil
}

func (s *awsServiceMock) Detach(ctx context.Context, input *aws.DetachGroupInput) (*aws.DetachGroupOutput, error) {
	args := s.Called(ctx, input)
	return args.Get(0).(*aws.DetachGroupOutput), nil
}

func (s *awsServiceMock) Roll(ctx context.Context, input *aws.RollGroupInput) (*aws.RollGroupOutput, error) {
	return nil, nil
}

func testCloudManager(t *testing.T) *CloudManager {
	return &CloudManager{
		groupService: &groupServiceMock{
			providerAWS: new(awsServiceMock),
		},
		groups:          make([]*Group, 0),
		cache:           make(map[string]*Group),
		interruptCh:     make(chan struct{}),
		refreshInterval: time.Minute,
	}
}

func testCloudProvider(t *testing.T, m *CloudManager) *CloudProvider {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	cloud, err := NewCloudProvider(m, resourceLimiter)
	assert.NoError(t, err)
	return cloud
}

func TestNewCloudProvider(t *testing.T) {
	testCloudProvider(t, testCloudManager(t))
}

func TestAddNodeGroup(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("bad spec")
	assert.Error(t, err)
	assert.Equal(t, len(provider.manager.groups), 0)

	err = provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)
}

func TestName(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	assert.Equal(t, provider.Name(), "spotinst")
}

func TestNodeGroups(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	assert.Equal(t, len(provider.NodeGroups()), 0)
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	group, err := provider.NodeGroupForNode(node)

	assert.NoError(t, err)
	assert.Equal(t, group.Id(), "sig-test")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	// test node in cluster that is not in a group managed by cluster autoscaler
	nodeNotInGroup := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-not-in-group",
		},
	}

	group, err = provider.NodeGroupForNode(nodeNotInGroup)
	assert.NoError(t, err)
	assert.Nil(t, group)
}

func TestExtractInstanceId(t *testing.T) {
	_, err := extractInstanceId("bad spec")
	assert.Error(t, err)

	instanceID, err := extractInstanceId("aws:///us-east-1a/i-260942b3")
	assert.NoError(t, err)
	assert.Equal(t, instanceID, "i-260942b3")
}

func TestMaxSize(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)
	assert.Equal(t, provider.manager.groups[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)
	assert.Equal(t, provider.manager.groups[0].MinSize(), 1)
}

func TestTargetSize(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	targetSize, err := provider.manager.groups[0].TargetSize()
	assert.Equal(t, targetSize, 2)
	assert.NoError(t, err)
}

func TestIncreaseSize(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)

	cloud := provider.manager.groupService.CloudProviderAWS().(*awsServiceMock)
	cloud.On("Update", context.Background(), &aws.UpdateGroupInput{
		Group: &aws.Group{
			ID: spotinst.String(provider.manager.groups[0].Id()),
			Capacity: &aws.Capacity{
				Target:  spotinst.Int(3),
				Minimum: spotinst.Int(provider.manager.groups[0].minSize),
				Maximum: spotinst.Int(provider.manager.groups[0].maxSize),
			},
		},
	}).Return(&aws.UpdateGroupOutput{})

	err = provider.manager.groups[0].IncreaseSize(1)
	assert.NoError(t, err)
	cloud.AssertExpectations(t)
}

func TestBelongs(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/invalid-instance-id",
		},
	}
	_, err = provider.manager.groups[0].Belongs(invalidNode)
	assert.Error(t, err)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	belongs, err := provider.manager.groups[0].Belongs(validNode)
	assert.Equal(t, belongs, true)
	assert.NoError(t, err)
}

func TestDeleteNodes(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)

	cloud := provider.manager.groupService.CloudProviderAWS().(*awsServiceMock)
	cloud.On("Detach", context.Background(), &aws.DetachGroupInput{
		GroupID:                       spotinst.String(provider.manager.groups[0].Id()),
		InstanceIDs:                   []string{"test-instance-id"},
		ShouldDecrementTargetCapacity: spotinst.Bool(true),
		ShouldTerminateInstances:      spotinst.Bool(true),
	}).Return(&aws.DetachGroupOutput{})

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}

	err = provider.manager.groups[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	cloud.AssertExpectations(t)
}

func TestId(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.manager.addNodeGroup("1:5:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.manager.groups), 1)
	assert.Equal(t, provider.manager.groups[0].Id(), "sig-test")
}

func TestDebug(t *testing.T) {
	grp := Group{
		manager: testCloudManager(t),
		minSize: 5,
		maxSize: 55,
	}
	grp.groupID = "sig-test"
	assert.Equal(t, grp.Debug(), "sig-test (5:55)")
}

func TestBuildGroup(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))

	_, err := provider.manager.buildGroupFromSpec("a")
	assert.Error(t, err)
	_, err = provider.manager.buildGroupFromSpec("a:b:c")
	assert.Error(t, err)
	_, err = provider.manager.buildGroupFromSpec("1:")
	assert.Error(t, err)
	_, err = provider.manager.buildGroupFromSpec("1:2:")
	assert.Error(t, err)

	grp, err := provider.manager.buildGroupFromSpec("111:222:sig-test")
	assert.NoError(t, err)
	assert.Equal(t, 111, grp.MinSize())
	assert.Equal(t, 222, grp.MaxSize())
	assert.Equal(t, "sig-test", grp.Id())
}

func TestGetResourceLimiter(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	_, err := provider.GetResourceLimiter()
	assert.NoError(t, err)
}

func TestCleanup(t *testing.T) {
	provider := testCloudProvider(t, testCloudManager(t))
	err := provider.Cleanup()
	assert.NoError(t, err)
}
