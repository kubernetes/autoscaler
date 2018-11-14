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
	corefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"
	"testing"
)

func TestGetMachineDeploymentAttrs(t *testing.T) {
	attrs := GetMachineDeploymentAttrs(&v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				MinSizeAnnotation: "1",
				MaxSizeAnnotation: "10",
			},
		},
	})

	assert.Equal(t, 1, attrs.minSize)
	assert.Equal(t, 10, attrs.maxSize)
}

func TestGetMachineDeploymentAttrsMissingMinSize(t *testing.T) {
	attrs := GetMachineDeploymentAttrs(&v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				MaxSizeAnnotation: "10",
			},
		},
	})

	assert.Nil(t, attrs)
}

func TestGetMachineDeploymentAttrsMissingMaxSize(t *testing.T) {
	attrs := GetMachineDeploymentAttrs(&v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				MinSizeAnnotation: "1",
			},
		},
	})

	assert.Nil(t, attrs)
}

func TestGetMachineDeploymentAttrsInvalidMinSize(t *testing.T) {
	attrs := GetMachineDeploymentAttrs(&v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				MinSizeAnnotation: "invalid",
				MaxSizeAnnotation: "10",
			},
		},
	})

	assert.Nil(t, attrs)
}

func TestGetMachineDeploymentAttrsInvalidMaxSize(t *testing.T) {
	attrs := GetMachineDeploymentAttrs(&v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				MinSizeAnnotation: "1",
				MaxSizeAnnotation: "invalid",
			},
		},
	})

	assert.Nil(t, attrs)
}

func TestDeploymentsAndNodes(t *testing.T) {
	md1 := buildTestMachineDeployment("md1", 1, 0, 10)
	md2 := buildTestMachineDeployment("md2", 2, 0, 10)
	md3 := buildTestMachineDeployment("md3", 2, 0, -1)

	ms1 := buildTestMachineSet(md2, "ms1", 2)

	n1 := buildTestNode("n1")
	m1 := buildTestMachine(ms1, "m1", n1)

	n2 := buildTestNode("n2")
	m2 := buildTestMachine(ms1, "m2", n2)

	m3 := buildTestMachine(ms1, "m3", nil)

	ms2 := buildTestMachineSet(md3, "ms2", 2)

	n4 := buildTestNode("n4")
	m4 := buildTestMachine(ms2, "m4", n4)

	coreApiClient := corefake.NewSimpleClientset(n1, n2, n4)
	clusterApiClient := clusterfake.NewSimpleClientset(m1, m2, m3, m4, ms1, ms2, md1, md2, md3)

	mm := NewMachineManagerFromApiStubs(coreApiClient, clusterApiClient)
	err := mm.Refresh()
	if !assert.Nil(t, err) {
		return
	}

	depls := mm.AllDeployments()
	assert.Len(t, depls, 2)
	assert.Contains(t, depls, md1)
	assert.Contains(t, depls, md2)

	assert.Equal(t, md2, mm.DeploymentForNode(n1))
	assert.Equal(t, md2, mm.DeploymentForNode(n2))
	assert.NotEqual(t, md1, mm.DeploymentForNode(n2))

	assert.Nil(t, mm.NodesForDeployment(md1))
	assert.Nil(t, mm.NodesForDeployment(md3))
	nodes := mm.NodesForDeployment(md2)
	assert.Len(t, nodes, 2)
	assert.Contains(t, nodes, n1)
	assert.Contains(t, nodes, n2)

	assert.Nil(t, mm.SetDeploymentSize(md2, 5))
	assert.Equal(t, int32(5), *md2.Spec.Replicas)
}
