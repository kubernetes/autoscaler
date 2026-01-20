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

package kamatera

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/kubernetes/fake"
)

func TestInstance_refresh_PoweroffOnScaleDownClearsNodeMetadata(t *testing.T) {
	providerIDPrefix := "rke2://"
	serverName := mockKamateraServerName()
	serverProviderID := formatKamateraProviderID(providerIDPrefix, serverName)

	kubeClient := fake.NewSimpleClientset(&apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: serverName},
		Spec: apiv1.NodeSpec{
			Unschedulable: true,
			Taints: []apiv1.Taint{
				{Key: taints.ToBeDeletedTaint, Value: "123", Effect: apiv1.TaintEffectNoSchedule},
				{Key: taints.DeletionCandidateTaint, Value: "123", Effect: apiv1.TaintEffectPreferNoSchedule},
				{Key: "custom", Value: "x", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
	})

	client := kamateraClientMock{}
	ctx := context.Background()
	client.On("getCommandStatus", ctx, "cmd-poweroff").Return(CommandStatusComplete, nil).Once()

	instance := &Instance{
		Id:                serverProviderID,
		Status:            &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting},
		PowerOn:           false,
		StatusCommandId:   "cmd-poweroff",
		StatusCommandCode: InstanceCommandPoweroff,
	}

	needToDelete := instance.refresh(&client, providerIDPrefix, true, kubeClient, true)
	assert.False(t, needToDelete)
	assert.Nil(t, instance.Status)

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, serverName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.False(t, node.Spec.Unschedulable)
	assert.False(t, taints.HasTaint(node, taints.ToBeDeletedTaint))
	assert.False(t, taints.HasTaint(node, taints.DeletionCandidateTaint))
	assert.True(t, taints.HasTaint(node, "custom"))
}
