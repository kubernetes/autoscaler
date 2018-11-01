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
