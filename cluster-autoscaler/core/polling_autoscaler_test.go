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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestRunOnce(t *testing.T) {
	currentTime := time.Now()

	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)

	initialCloudProvider := testprovider.NewTestCloudProvider(nil, nil)
	initialCloudProvider.AddNodeGroup("ng1", 1, 10, 2)
	initialCloudProvider.AddNode("ng1", n1)

	newCloudProvider := testprovider.NewTestCloudProvider(nil, nil)
	newCloudProvider.AddNodeGroup("ng1", 1, 10, 2)
	newCloudProvider.AddNode("ng1", n1)
	newCloudProvider.AddNodeGroup("ng2", 1, 10, 2)
	newCloudProvider.AddNode("ng2", n2)

	initialAutoscaler := &AutoscalerMock{}
	initialAutoscaler.On("CloudProvider").Return(initialCloudProvider).Once()

	newAutoscaler := &AutoscalerMock{}
	newAutoscaler.On("RunOnce", currentTime).Once()
	newAutoscaler.On("CloudProvider").Return(newCloudProvider).Once()

	builder := &AutoscalerBuilderMock{}
	builder.On("Build").Return(initialAutoscaler).Once()
	builder.On("Build").Return(newAutoscaler).Once()

	a, _ := NewPollingAutoscaler(builder)
	a.RunOnce(currentTime)

	initialAutoscaler.AssertNotCalled(t, "RunOnce", mock.AnythingOfType("time.Time"))
	newAutoscaler.AssertExpectations(t)
	builder.AssertExpectations(t)
}
