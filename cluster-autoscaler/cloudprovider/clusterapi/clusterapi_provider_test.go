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

package clusterapi

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestProviderConstructorProperties(t *testing.T) {
	resourceLimits := cloudprovider.ResourceLimiter{}

	controller, stop := mustCreateTestController(t)
	defer stop()

	provider := newProvider(ProviderName, &resourceLimits, controller)
	if actual := provider.Name(); actual != ProviderName {
		t.Errorf("expected %q, got %q", ProviderName, actual)
	}

	rl, err := provider.GetResourceLimiter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reflect.DeepEqual(rl, resourceLimits) {
		t.Errorf("expected %+v, got %+v", resourceLimits, rl)
	}

	if _, err := provider.Pricing(); err != cloudprovider.ErrNotImplemented {
		t.Errorf("expected an error")
	}

	machineTypes, err := provider.GetAvailableMachineTypes()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(machineTypes) != 0 {
		t.Errorf("expected 0, got %v", len(machineTypes))
	}

	if _, err := provider.NewNodeGroup("foo", nil, nil, nil, nil); err == nil {
		t.Error("expected an error")
	}

	if err := provider.Cleanup(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := provider.Refresh(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	nodegroups := provider.NodeGroups()

	if len(nodegroups) != 0 {
		t.Errorf("expected 0, got %v", len(nodegroups))
	}

	ng, err := provider.NodeGroupForNode(&corev1.Node{
		TypeMeta: v1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "missing-node",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ng != nil {
		t.Fatalf("unexpected nodegroup: %v", ng.Id())
	}

	if got := provider.GPULabel(); got != GPULabel {
		t.Fatalf("expected %q, got %q", GPULabel, got)
	}

	if got := len(provider.GetAvailableGPUTypes()); got != 0 {
		t.Fatalf("expected 0 GPU types, got %d", got)
	}
}
