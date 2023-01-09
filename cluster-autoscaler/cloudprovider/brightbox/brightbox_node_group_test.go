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

package brightbox

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk/mocks"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	//schedulerframework "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	fakeMaxSize                   = 4
	fakeMinSize                   = 1
	fakeNodeGroupDescription      = "1:4"
	fakeDefaultSize               = 3
	fakeNodeGroupID               = "grp-sda44"
	fakeNodeGroupName             = "auto.workers.k8s_fake.cluster.local"
	fakeNodeGroupImageID          = "img-testy"
	fakeNodeGroupServerTypeID     = "typ-zx45f"
	fakeNodeGroupServerTypeHandle = "small"
	fakeNodeGroupZoneID           = "zon-testy"
	fakeNodeGroupMainGroupID      = "grp-y6cai"
	fakeNodeGroupUserData         = "fake userdata"
)

var (
	fakeMapData = map[string]string{
		"min":           strconv.Itoa(fakeMinSize),
		"max":           strconv.Itoa(fakeMaxSize),
		"server_group":  fakeNodeGroupID,
		"default_group": fakeNodeGroupMainGroupID,
		"image":         fakeNodeGroupImageID,
		"type":          fakeNodeGroupServerTypeID,
		"zone":          fakeNodeGroupZoneID,
		"user_data":     fakeNodeGroupUserData,
	}
	fakeMainGroupList           = []string{fakeNodeGroupID, fakeNodeGroupMainGroupID}
	fakeAdditionalGroupList     = []string{"grp-abcde", "grp-testy", "grp-winga"}
	fakeFullGroupList           = append(fakeMainGroupList, fakeAdditionalGroupList...)
	fakeAdditionalGroups        = strings.Join(fakeAdditionalGroupList, ",")
	fakeDefaultAdditionalGroups = fakeNodeGroupMainGroupID + "," + fakeAdditionalGroups
	ErrFake                     = errors.New("fake API Error")
	fakeInstances               = []cloudprovider.Instance{
		{
			Id: "brightbox://srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "brightbox://srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}
	fakeTransitionInstances = []cloudprovider.Instance{
		{
			Id: "brightbox://srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "brightbox://srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
	}
	ErrFakeInstances = []cloudprovider.Instance{
		{
			Id: "brightbox://srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "unavailable",
					ErrorMessage: "unavailable",
				},
			},
		},
		{
			Id: "brightbox://srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "inactive",
					ErrorMessage: "inactive",
				},
			},
		},
	}
)

func TestMaxSize(t *testing.T) {
	assert.Equal(t, makeFakeNodeGroup(t, nil).MaxSize(), fakeMaxSize)
}

func TestMinSize(t *testing.T) {
	assert.Equal(t, makeFakeNodeGroup(t, nil).MinSize(), fakeMinSize)
}

func TestSize(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(t, testclient)
	fakeServerGroup := &fakeGroups()[0]
	t.Run("TargetSize", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(fakeServerGroup, nil).Once()
		size, err := nodeGroup.TargetSize()
		assert.Equal(t, 2, size)
		assert.NoError(t, err)
	})
	t.Run("TargetSizeFail", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(nil, ErrFake).Once()
		size, err := nodeGroup.TargetSize()
		assert.Error(t, err)
		assert.Zero(t, size)
	})
	t.Run("CurrentSize", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(fakeServerGroup, nil).Once()
		size, err := nodeGroup.CurrentSize()
		assert.Equal(t, 2, size)
		assert.NoError(t, err)
	})
	t.Run("CurrentSizeFail", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(nil, ErrFake).Once()
		size, err := nodeGroup.CurrentSize()
		assert.Error(t, err)
		assert.Zero(t, size)
	})
	mockclient.On("ServerGroup", fakeNodeGroupID).
		Return(fakeServerGroup, nil)
	t.Run("DecreaseTargetSizePositive", func(t *testing.T) {
		err := nodeGroup.DecreaseTargetSize(0)
		assert.Error(t, err)
	})
	t.Run("DecreaseTargetSizeFail", func(t *testing.T) {
		err := nodeGroup.DecreaseTargetSize(-1)
		assert.Error(t, err)
	})
	mockclient.AssertExpectations(t)
}

func TestIncreaseSize(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(t, testclient)
	t.Run("Creating details set properly", func(t *testing.T) {
		assert.Equal(t, fakeNodeGroupID, nodeGroup.id)
		assert.Equal(t, fakeNodeGroupName, *nodeGroup.serverOptions.Name)
		assert.Equal(t, fakeNodeGroupServerTypeID, nodeGroup.serverOptions.ServerType)
		assert.Equal(t, fakeNodeGroupImageID, nodeGroup.serverOptions.Image)
		assert.Equal(t, fakeNodeGroupZoneID, nodeGroup.serverOptions.Zone)
		assert.ElementsMatch(t, []string{fakeNodeGroupMainGroupID, fakeNodeGroupID}, nodeGroup.serverOptions.ServerGroups)
		assert.Equal(t, fakeNodeGroupUserData, *nodeGroup.serverOptions.UserData)
	})
	t.Run("Require positive delta", func(t *testing.T) {
		err := nodeGroup.IncreaseSize(0)
		assert.Error(t, err)
	})
	fakeServerGroup := &fakeGroups()[0]
	t.Run("Don't exceed max size", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(fakeServerGroup, nil).Once()
		err := nodeGroup.IncreaseSize(4)
		assert.Error(t, err)
	})
	t.Run("Fail to create one new server", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(fakeServerGroup, nil).Once()
		mockclient.On("CreateServer", mock.Anything).
			Return(nil, ErrFake).Once()
		err := nodeGroup.IncreaseSize(1)
		assert.Error(t, err)
	})
	t.Run("Create one new server", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(fakeServerGroup, nil).Once()
		mockclient.On("CreateServer", mock.Anything).
			Return(nil, nil).Once()
		mockclient.On("ServerGroup", fakeNodeGroupID).
			Return(&fakeServerGroupsPlusOne()[0], nil).Once()
		err := nodeGroup.IncreaseSize(1)
		assert.NoError(t, err)
	})
}

func TestDeleteNodes(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(t, testclient)
	fakeServerGroup := &fakeGroups()[0]
	mockclient.On("ServerGroup", fakeNodeGroupID).
		Return(fakeServerGroup, nil).
		On("Server", fakeServer).
		Return(fakeServertesty(), nil)
	t.Run("Empty Nodes", func(t *testing.T) {
		err := nodeGroup.DeleteNodes(nil)
		assert.NoError(t, err)
	})
	t.Run("Foreign Node", func(t *testing.T) {
		err := nodeGroup.DeleteNodes([]*v1.Node{makeNode(fakeServer)})
		assert.Error(t, err)
	})
	t.Run("Delete Node", func(t *testing.T) {
		mockclient.On("Server", "srv-rp897").
			Return(fakeServerrp897(), nil).Once().
			On("Server", "srv-rp897").
			Return(deletedFakeServer(fakeServerrp897()), nil).
			Once().
			On("DestroyServer", "srv-rp897").
			Return(nil).Once()
		err := nodeGroup.DeleteNodes([]*v1.Node{makeNode("srv-rp897")})
		assert.NoError(t, err)
	})
	t.Run("Delete All Nodes", func(t *testing.T) {
		truncateServers := mocks.ServerListReducer(fakeServerGroup)
		mockclient.On("Server", "srv-rp897").
			Return(fakeServerrp897(), nil).Once().
			On("Server", "srv-rp897").
			Return(deletedFakeServer(fakeServerrp897()), nil).
			Once().
			On("DestroyServer", "srv-rp897").
			Return(nil).Once().Run(truncateServers)
		err := nodeGroup.DeleteNodes([]*v1.Node{
			makeNode("srv-rp897"),
			makeNode("srv-lv426"),
		})
		assert.Error(t, err)
	})

}

func TestExist(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(t, testclient)
	fakeServerGroup := &fakeGroups()[0]
	t.Run("Find Group", func(t *testing.T) {
		mockclient.On("ServerGroup", nodeGroup.Id()).
			Return(fakeServerGroup, nil).Once()
		assert.True(t, nodeGroup.Exist())
	})
	t.Run("Fail to Find Group", func(t *testing.T) {
		mockclient.On("ServerGroup", nodeGroup.Id()).
			Return(nil, serverNotFoundError(nodeGroup.Id()))
		assert.False(t, nodeGroup.Exist())
	})
	mockclient.AssertExpectations(t)
}

func TestNodes(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(t, testclient)
	fakeServerGroup := &fakeGroups()[0]
	mockclient.On("ServerGroup", fakeNodeGroupID).
		Return(fakeServerGroup, nil)
	t.Run("Both Active", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "active"
		fakeServerGroup.Servers[1].Status = "active"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeInstances, nodes)
	})
	t.Run("Creating and Deleting", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "creating"
		fakeServerGroup.Servers[1].Status = "deleting"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeTransitionInstances, nodes)
	})
	t.Run("Inactive and Unavailable", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "inactive"
		fakeServerGroup.Servers[1].Status = "unavailable"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, ErrFakeInstances, nodes)
	})
}

func TestTemplateNodeInfo(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	mockclient.On("ServerType", fakeNodeGroupServerTypeID).
		Return(fakeServerTypezx45f(), nil)
	obj, err := makeFakeNodeGroup(t, testclient).TemplateNodeInfo()
	require.NoError(t, err)
	assert.Equal(t, fakeResource(), obj.Allocatable)
}

func TestNodeGroupErrors(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	emptyMapData := map[string]string{}
	obj, err := makeNodeGroupFromAPIDetails(
		fakeNodeGroupName,
		emptyMapData,
		fakeMinSize,
		fakeMaxSize,
		testclient,
	)
	assert.Equal(t, cloudprovider.ErrIllegalConfiguration, err)
	assert.Nil(t, obj)
}

func TestMultipleGroups(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	t.Run("server group and multi default group", func(t *testing.T) {
		multigroupData := map[string]string{
			"min":           strconv.Itoa(fakeMinSize),
			"max":           strconv.Itoa(fakeMaxSize),
			"image":         fakeNodeGroupImageID,
			"type":          fakeNodeGroupServerTypeID,
			"zone":          fakeNodeGroupZoneID,
			"user_data":     fakeNodeGroupUserData,
			"server_group":  fakeNodeGroupID,
			"default_group": fakeDefaultAdditionalGroups,
		}
		obj, err := makeFakeNodeGroupFromMap(
			testclient,
			multigroupData,
		)
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeFullGroupList, obj.serverOptions.ServerGroups)
	})
	t.Run("server group and multi additional group", func(t *testing.T) {
		multigroupData := map[string]string{
			"min":               strconv.Itoa(fakeMinSize),
			"max":               strconv.Itoa(fakeMaxSize),
			"image":             fakeNodeGroupImageID,
			"type":              fakeNodeGroupServerTypeID,
			"zone":              fakeNodeGroupZoneID,
			"user_data":         fakeNodeGroupUserData,
			"server_group":      fakeNodeGroupID,
			"additional_groups": fakeDefaultAdditionalGroups,
		}
		obj, err := makeFakeNodeGroupFromMap(
			testclient,
			multigroupData,
		)
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeFullGroupList, obj.serverOptions.ServerGroups)
	})
	t.Run("server group, default group and multi additional group", func(t *testing.T) {
		multigroupData := map[string]string{
			"min":               strconv.Itoa(fakeMinSize),
			"max":               strconv.Itoa(fakeMaxSize),
			"image":             fakeNodeGroupImageID,
			"type":              fakeNodeGroupServerTypeID,
			"zone":              fakeNodeGroupZoneID,
			"user_data":         fakeNodeGroupUserData,
			"server_group":      fakeNodeGroupID,
			"default_group":     fakeNodeGroupMainGroupID,
			"additional_groups": fakeAdditionalGroups,
		}
		obj, err := makeFakeNodeGroupFromMap(
			testclient,
			multigroupData,
		)
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeFullGroupList, obj.serverOptions.ServerGroups)
	})
	t.Run("server group, default group and multi additional group with duplicates", func(t *testing.T) {
		multigroupData := map[string]string{
			"min":               strconv.Itoa(fakeMinSize),
			"max":               strconv.Itoa(fakeMaxSize),
			"image":             fakeNodeGroupImageID,
			"type":              fakeNodeGroupServerTypeID,
			"zone":              fakeNodeGroupZoneID,
			"user_data":         fakeNodeGroupUserData,
			"server_group":      fakeNodeGroupID,
			"default_group":     fakeNodeGroupMainGroupID,
			"additional_groups": fakeDefaultAdditionalGroups,
		}
		obj, err := makeFakeNodeGroupFromMap(
			testclient,
			multigroupData,
		)
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeFullGroupList, obj.serverOptions.ServerGroups)
	})
}

func TestCreate(t *testing.T) {
	obj, err := makeFakeNodeGroup(t, nil).Create()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.Nil(t, obj)
}

func TestDelete(t *testing.T) {
	assert.Equal(t, cloudprovider.ErrNotImplemented, makeFakeNodeGroup(t, nil).Delete())
}

func TestAutoprovisioned(t *testing.T) {
	assert.False(t, makeFakeNodeGroup(t, nil).Autoprovisioned())
}

func fakeResource() *schedulerframework.Resource {
	return &schedulerframework.Resource{
		MilliCPU:         2000,
		Memory:           1979711488,
		EphemeralStorage: 80530636800,
		AllowedPodNumber: 110,
	}
}

func makeFakeNodeGroup(t *testing.T, brightboxCloudClient *k8ssdk.Cloud) *brightboxNodeGroup {
	obj, err := makeFakeNodeGroupFromMap(
		brightboxCloudClient,
		fakeMapData,
	)
	require.NoError(t, err)
	return obj
}

func makeFakeNodeGroupFromMap(brightboxCloudClient *k8ssdk.Cloud, mapData map[string]string) (*brightboxNodeGroup, error) {
	return makeNodeGroupFromAPIDetails(
		fakeNodeGroupName,
		mapData,
		fakeMinSize,
		fakeMaxSize,
		brightboxCloudClient,
	)
}
