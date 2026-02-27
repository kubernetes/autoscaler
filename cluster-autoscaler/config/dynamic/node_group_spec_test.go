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

package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
)

func TestSpecFromString(t *testing.T) {
	t.Run("invalid string - missing token", func(t *testing.T) {
		specString := "1:50:aks-nodepool-83823-vmss"
		spec, err := SpecFromStringWithLabelsAndTaints(specString, true)
		assert.NoError(t, err)
		assert.Equal(t, *spec, NodeGroupSpec{Name: "aks-nodepool-83823-vmss", MinSize: 1, MaxSize: 50, SupportScaleToZero: true})
	})
	t.Run("valid string", func(t *testing.T) {
		specString := "1:50:Delete:aks-nodepool-83823-vmss"
		spec, err := SpecFromStringWithLabelsAndTaints(specString, true)
		assert.NoError(t, err)
		assert.Equal(t, *spec, NodeGroupSpec{Name: "aks-nodepool-83823-vmss", MinSize: 1, MaxSize: 50, ScaleDownPolicy: deallocate.Delete, SupportScaleToZero: true})
	})
}

func TestSpecFromStringWithLabelsAndTaints(t *testing.T) {
	t.Run("invalid string - missing token", func(t *testing.T) {
		specString := "1:50:aks-nodepool-83823-vmss:{}|"
		spec, err := SpecFromStringWithLabelsAndTaints(specString, true)
		assert.EqualError(t, err, "error while parsing NodeGroupSpec: 1:50:aks-nodepool-83823-vmss:{}|, failed to set scale down policy: aks-nodepool-83823-vmss. Valid values are: Delete, Deallocate")
		assert.Nil(t, spec)
	})
	t.Run("valid string", func(t *testing.T) {
		specString := "1:50:Delete:aks-nodepool-83823-vmss:{}|"
		spec, err := SpecFromStringWithLabelsAndTaints(specString, true)
		assert.NoError(t, err)
		assert.Equal(t, *spec, NodeGroupSpec{Name: "aks-nodepool-83823-vmss", MinSize: 1, MaxSize: 50, ScaleDownPolicy: deallocate.Delete, SupportScaleToZero: true, Labels: map[string]string{}})
	})
}
