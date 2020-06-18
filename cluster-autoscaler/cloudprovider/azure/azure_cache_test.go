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

package azure

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	provider := newTestProvider(t)
	ss := newTestScaleSet(provider.azureManager, "ss")

	ac, err := newAsgCache()
	assert.NoError(t, err)
	ac.registeredAsgs = []cloudprovider.NodeGroup{ss}

	isSuccess := ac.Register(ss)
	assert.False(t, isSuccess)

	ss1 := newTestScaleSet(provider.azureManager, "ss")
	ss1.minSize = 2
	isSuccess = ac.Register(ss1)
	assert.True(t, isSuccess)
}

func TestUnRegister(t *testing.T) {
	provider := newTestProvider(t)
	ss := newTestScaleSet(provider.azureManager, "ss")
	ss1 := newTestScaleSet(provider.azureManager, "ss1")

	ac, err := newAsgCache()
	assert.NoError(t, err)
	ac.registeredAsgs = []cloudprovider.NodeGroup{ss, ss1}

	isSuccess := ac.Unregister(ss)
	assert.True(t, isSuccess)
	assert.Equal(t, 1, len(ac.registeredAsgs))
}

func TestFindForInstance(t *testing.T) {
	ac, err := newAsgCache()
	assert.NoError(t, err)

	inst := azureRef{Name: "/subscriptions/sub/resourceGroups/rg/providers/foo"}
	ac.notInRegisteredAsg = make(map[azureRef]bool)
	ac.notInRegisteredAsg[inst] = true
	nodeGroup, err := ac.FindForInstance(&inst, vmTypeVMSS)
	assert.Nil(t, nodeGroup)
	assert.NoError(t, err)

	ac.notInRegisteredAsg[inst] = false
	nodeGroup, err = ac.FindForInstance(&inst, vmTypeStandard)
	assert.Nil(t, nodeGroup)
	assert.NoError(t, err)
	assert.True(t, ac.notInRegisteredAsg[inst])
}
