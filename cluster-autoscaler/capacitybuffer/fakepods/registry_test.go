/*
Copyright The Kubernetes Authors.

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

package fakepods

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
)

func TestFakePodsRegistry(t *testing.T) {
	buffer1 := &v1beta1.CapacityBuffer{ObjectMeta: metav1.ObjectMeta{Name: "buffer1"}}
	buffer2 := &v1beta1.CapacityBuffer{ObjectMeta: metav1.ObjectMeta{Name: "buffer2"}}
	uid1 := types.UID("uid1")
	uid2 := types.UID("uid2")

	t.Run("NewRegistry with nil", func(t *testing.T) {
		r := NewRegistry(nil)
		assert.NotNil(t, r.fakePodsUIDToBuffer)
		assert.Equal(t, 0, len(r.fakePodsUIDToBuffer))
	})

	t.Run("NewRegistry with existing map", func(t *testing.T) {
		initialMap := map[types.UID]*v1beta1.CapacityBuffer{uid1: buffer1}
		r := NewRegistry(initialMap)
		assert.Equal(t, buffer1, r.GetCapacityBuffer(uid1))
	})

	t.Run("Set and Get", func(t *testing.T) {
		r := NewRegistry(nil)
		r.SetCapacityBuffer(uid1, buffer1)
		r.SetCapacityBuffer(uid2, buffer2)

		assert.Equal(t, buffer1, r.GetCapacityBuffer(uid1))
		assert.Equal(t, buffer2, r.GetCapacityBuffer(uid2))
		assert.Nil(t, r.GetCapacityBuffer("non-existent"))
	})

	t.Run("UnsetCapacityBuffer", func(t *testing.T) {
		r := NewRegistry(nil)
		r.SetCapacityBuffer(uid1, buffer1)

		r.UnsetCapacityBuffer(uid1)

		assert.Nil(t, r.GetCapacityBuffer(uid1))
	})

	t.Run("Clear", func(t *testing.T) {
		r := NewRegistry(nil)
		r.SetCapacityBuffer(uid1, buffer1)
		r.SetCapacityBuffer(uid2, buffer2)

		r.Clear()

		assert.Equal(t, 0, len(r.fakePodsUIDToBuffer))
		assert.Nil(t, r.GetCapacityBuffer(uid1))
		assert.Nil(t, r.GetCapacityBuffer(uid2))
	})
}
