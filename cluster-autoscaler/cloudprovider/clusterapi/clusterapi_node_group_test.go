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

package clusterapi

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"testing"
)

func newNodeGroup(t *testing.T) *ClusterapiNodeGroup {
	manager := newTestMachineManager(t)

	return &ClusterapiNodeGroup{
		machineManager: manager,
		machineDeployment: &v1alpha1.MachineDeployment{
			ObjectMeta: v1.ObjectMeta{
				Name: "ngName",
			},
			Spec: v1alpha1.MachineDeploymentSpec{
				Replicas: int32Ptr(5),
			},
		},
		attrs: &MachineDeploymentAttrs{
			maxSize: 10,
			minSize: 0,
		},
	}
}

func TestMaxSize(t *testing.T) {
	ng := newNodeGroup(t)
	assert.Equal(t, 10, ng.MaxSize())
}

func TestMinize(t *testing.T) {
	ng := newNodeGroup(t)
	assert.Equal(t, 0, ng.MinSize())
}

func TestId(t *testing.T) {
	ng := newNodeGroup(t)
	assert.Equal(t, "ngName", ng.Id())
}

func TestDebug(t *testing.T) {
	ng := newNodeGroup(t)
	assert.Equal(t, "ngName (0:10)", ng.Debug())
}

func TestTargetSize(t *testing.T) {
	ng := newNodeGroup(t)
	targetSize, err := ng.TargetSize()

	assert.Nil(t, err)
	assert.Equal(t, 5, targetSize)
}

func TestExists(t *testing.T) {
	ng := newNodeGroup(t)

	assert.True(t, ng.Exist())
}

func TestCreate(t *testing.T) {
	ng := newNodeGroup(t)

	newNg, err := ng.Create()

	assert.Nil(t, newNg)
	assert.EqualError(t, err, "Already exist")
}

func TestDelete(t *testing.T) {
	ng := newNodeGroup(t)

	err := ng.Delete()

	assert.EqualError(t, err, "Not implemented")
}

func TestAutoprovisioned(t *testing.T) {
	ng := newNodeGroup(t)

	assert.False(t, ng.Autoprovisioned())
}

func TestIncreaseWithTooLargeTargetSize(t *testing.T) {
	ng := newNodeGroup(t)

	err := ng.IncreaseSize(100)

	assert.EqualError(t, err, "ClusterapiNodeGroup size increase too large - desired:105 max:10")
}

func TestIncreaseWithNegativeTargetSize(t *testing.T) {
	ng := newNodeGroup(t)

	err := ng.IncreaseSize(-1)

	assert.EqualError(t, err, "ClusterapiNodeGroup size increase size must be positive")
}

func TestDecreaseWithPositiveTargetSize(t *testing.T) {
	ng := newNodeGroup(t)

	err := ng.DecreaseTargetSize(1)

	assert.EqualError(t, err, "ClusterapiNodeGroup size decrease size must be negative")
}
