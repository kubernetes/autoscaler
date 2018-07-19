/*
Copyright 2018 The Kubernetes Authors.

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

package nodegroups

import (
	"fmt"
	"testing"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config/static"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"
)

func TestAutoprovisioningNodeGroupManager(t *testing.T) {
	manager := NewAutoprovisioningNodeGroupManager()

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)

	tests := []struct {
		name               string
		createNodeGroupErr error
		wantError          bool
	}{
		{
			name: "create node group",
		},
		{
			name:               "failed to create node group",
			createNodeGroupErr: fmt.Errorf("some error"),
			wantError:          true,
		},
	}
	for _, tc := range tests {
		provider := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil,
			func(string) error { return tc.createNodeGroupErr }, nil, nil, nil)
		context := &context.AutoscalingContext{
			AutoscalingOptions: static.AutoscalingOptions{
				NodeAutoprovisioningEnabled: true,
			},
			CloudProvider: provider,
			LogRecorder:   fakeLogRecorder,
		}

		nodeGroup, err := provider.NewNodeGroup("T1", nil, nil, nil, nil)
		assert.NoError(t, err)
		_, err = manager.CreateNodeGroup(context, nodeGroup)
		if tc.wantError {
			if err == nil {
				glog.Errorf("%s: Got no error, want error", tc.name)
			}
		} else {
			if err != nil {
				glog.Errorf("%s: Unexpected error %v", tc.name, err)
			}
			if len(provider.NodeGroups()) != 1 {
				glog.Errorf("%s: Unexpected number of node groups %d, want 1", tc.name, len(provider.NodeGroups()))
			}
		}
	}
}

func TestRemoveUnneededNodeGroups(t *testing.T) {
	manager := NewAutoprovisioningNodeGroupManager()
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 1000, 1000)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil,
		nil, func(id string) error {
			if id == "ng2" {
				return nil
			}
			return fmt.Errorf("Node group %s shouldn't be deleted", id)
		},
		nil, nil)
	assert.NotNil(t, provider)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddAutoprovisionedNodeGroup("ng2", 0, 10, 0, "mt1")
	provider.AddAutoprovisionedNodeGroup("ng3", 0, 10, 1, "mt1")
	provider.AddAutoprovisionedNodeGroup("ng4", 0, 10, 0, "mt1")
	provider.AddNode("ng3", n1)
	provider.AddNode("ng4", n2)

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	context := &context.AutoscalingContext{
		AutoscalingOptions: static.AutoscalingOptions{
			NodeAutoprovisioningEnabled: true,
		},
		CloudProvider: provider,
		LogRecorder:   fakeLogRecorder,
	}

	assert.NoError(t, manager.RemoveUnneededNodeGroups(context))
}
