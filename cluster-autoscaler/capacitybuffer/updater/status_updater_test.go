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
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	fakeclientset "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
)

func TestStatusUpdater(t *testing.T) {
	exitingBuffer := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buffer1",
			Namespace: "default",
		},
		Spec: v1.CapacityBufferSpec{},
	}
	notExistingBuffer := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buffer2",
			Namespace: "default",
		},
		Spec: v1.CapacityBufferSpec{},
	}
	fakeClient := fakeclientset.NewSimpleClientset(exitingBuffer)
	tests := []struct {
		name                   string
		buffers                []*v1.CapacityBuffer
		expectedNumberOfCalls  int
		expectedNumberOfErrors int
	}{
		{
			name: "Update one buffer",
			buffers: []*v1.CapacityBuffer{
				exitingBuffer,
			},
			expectedNumberOfCalls:  1,
			expectedNumberOfErrors: 0,
		},
		{
			name: "Update one buffer not existing",
			buffers: []*v1.CapacityBuffer{
				notExistingBuffer,
			},
			expectedNumberOfCalls:  1,
			expectedNumberOfErrors: 1,
		},
		{
			name: "Update multiple buffers",
			buffers: []*v1.CapacityBuffer{
				exitingBuffer,
				notExistingBuffer,
			},
			expectedNumberOfCalls:  2,
			expectedNumberOfErrors: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updateCallsCount := 0
			fakeClient.Fake.PrependReactor("update", "capacitybuffers",
				func(action ctesting.Action) (handled bool, ret runtime.Object, err error) {
					updateCallsCount++
					return false, nil, nil
				},
			)
			buffersUpdater := NewStatusUpdater(fakeClient)
			errors := buffersUpdater.Update(test.buffers)
			assert.Equal(t, test.expectedNumberOfErrors, len(errors))
			assert.Equal(t, test.expectedNumberOfCalls, updateCallsCount)
		})
	}
}
