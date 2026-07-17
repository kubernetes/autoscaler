/*
Copyright 2025 The Kubernetes Authors.

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

package updater

import (
	"testing"

	ctesting "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	fakeclientset "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
)

func TestStatusUpdater(t *testing.T) {
	exitingBuffer := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buffer1",
			Namespace: "default",
			UID:       types.UID("uid1"),
		},
		Spec: v1.CapacityBufferSpec{},
	}
	notExistingBuffer := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buffer2",
			Namespace: "default",
			UID:       types.UID("uid2"),
		},
		Spec: v1.CapacityBufferSpec{},
	}
	fakeClient := fakeclientset.NewSimpleClientset(exitingBuffer)
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClient(fakeClient, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name               string
		buffers            []*v1.CapacityBuffer
		wantNumberOfCalls  int
		wantNumberOfErrors int
		wantUpdatedCount   int
	}{
		{
			name: "Update one buffer",
			buffers: []*v1.CapacityBuffer{
				exitingBuffer,
			},
			wantNumberOfCalls:  1,
			wantNumberOfErrors: 0,
			wantUpdatedCount:   1,
		},
		{
			name: "Update one buffer not existing",
			buffers: []*v1.CapacityBuffer{
				notExistingBuffer,
			},
			wantNumberOfCalls:  1,
			wantNumberOfErrors: 1,
			wantUpdatedCount:   0,
		},
		{
			name: "Update multiple buffers",
			buffers: []*v1.CapacityBuffer{
				exitingBuffer,
				notExistingBuffer,
			},
			wantNumberOfCalls:  2,
			wantNumberOfErrors: 1,
			wantUpdatedCount:   1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateCallsCount := 0
			fakeClient.Fake.PrependReactor("update", "capacitybuffers",
				func(action ctesting.Action) (handled bool, ret runtime.Object, err error) {
					updateCallsCount++
					return false, nil, nil
				},
			)
			buffersUpdater := NewStatusUpdater(fakeCapacityBuffersClient)
			updatedBuffers, errors := buffersUpdater.Update(tc.buffers)
			assert.Equal(t, tc.wantNumberOfErrors, len(errors))
			assert.Equal(t, tc.wantNumberOfCalls, updateCallsCount)
			assert.Equal(t, tc.wantUpdatedCount, len(updatedBuffers))
		})
	}
}
