/*
Copyright 2017 The Kubernetes Authors.

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

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"
)

func TestNewAutoscalingContext(t *testing.T) {
	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)

	autoscalingContext, err := NewAutoscalingContext(
		AutoscalingOptions{
			ExpanderName:   expander.RandomExpanderName,
			MaxCoresTotal:  10,
			MinCoresTotal:  1,
			MaxMemoryTotal: 10000000000,
			MinMemoryTotal: 1000000000,
		},
		simulator.NewTestPredicateChecker(),
		fakeClient, fakeRecorder,
		fakeLogRecorder, kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil))
	assert.NoError(t, err)
	assert.NotNil(t, autoscalingContext)
}
