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

package exoscale

import (
	"testing"

	"github.com/stretchr/testify/assert"

	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
)

func TestToProviderID(t *testing.T) {
	assert.Equal(t, "exoscale://abc-123", toProviderID("abc-123"))
}

func TestToNodeID(t *testing.T) {
	assert.Equal(t, "abc-123", toNodeID("exoscale://abc-123"))
	assert.Equal(t, "abc-123", toNodeID("abc-123"))
}

func TestToInstanceStatus(t *testing.T) {
	assert.Nil(t, toInstanceStatus(""))
	assert.Equal(t, &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}, toInstanceStatus("starting"))
	assert.Equal(t, &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}, toInstanceStatus("running"))
	assert.Equal(t, &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}, toInstanceStatus("stopping"))

	errStatus := toInstanceStatus("error")
	assert.NotNil(t, errStatus.ErrorInfo)
	assert.Equal(t, cloudprovider.OtherErrorClass, errStatus.ErrorInfo.ErrorClass)
}

func TestToInstance(t *testing.T) {
	id := "abc-123"
	state := "running"
	instance := toInstance(&egoscale.Instance{ID: &id, State: &state})

	assert.Equal(t, "exoscale://abc-123", instance.Id)
	assert.Equal(t, &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}, instance.Status)
}
