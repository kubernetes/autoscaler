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

package utils

import (
	"errors"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"

	"github.com/stretchr/testify/assert"
)

type testInfo struct {
	client       *fake.Clientset
	configMap    *apiv1.ConfigMap
	namespace    string
	getError     error
	getCalled    bool
	updateCalled bool
	createCalled bool
	t            *testing.T
}

func setUpTest(t *testing.T) *testInfo {
	namespace := "kube-system"
	result := testInfo{
		client: &fake.Clientset{},
		configMap: &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      StatusConfigMapName,
			},
			Data: map[string]string{},
		},
		namespace:    namespace,
		getCalled:    false,
		updateCalled: false,
		createCalled: false,
		t:            t,
	}
	result.client.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		get := action.(core.GetAction)
		assert.Equal(result.t, namespace, get.GetNamespace())
		assert.Equal(result.t, StatusConfigMapName, get.GetName())
		result.getCalled = true
		if result.getError != nil {
			return true, nil, result.getError
		}
		return true, result.configMap, nil
	})
	result.client.Fake.AddReactor("update", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		assert.Equal(result.t, namespace, update.GetNamespace())
		result.updateCalled = true
		return true, result.configMap, nil
	})
	result.client.Fake.AddReactor("create", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		create := action.(core.CreateAction)
		assert.Equal(result.t, namespace, create.GetNamespace())
		configMap := create.GetObject().(*apiv1.ConfigMap)
		assert.Equal(result.t, StatusConfigMapName, configMap.ObjectMeta.Name)
		result.createCalled = true
		return true, configMap, nil
	})
	return &result
}

func TestWriteStatusConfigMapExisting(t *testing.T) {
	ti := setUpTest(t)
	result, err := WriteStatusConfigMap(ti.client, ti.namespace, "TEST_MSG", nil)
	assert.Equal(t, ti.configMap, result)
	assert.Contains(t, result.Data["status"], "TEST_MSG")
	assert.Contains(t, result.ObjectMeta.Annotations, ConfigMapLastUpdatedKey)
	assert.Nil(t, err)
	assert.True(t, ti.getCalled)
	assert.True(t, ti.updateCalled)
	assert.False(t, ti.createCalled)
}

func TestWriteStatusConfigMapCreate(t *testing.T) {
	ti := setUpTest(t)
	ti.getError = kube_errors.NewNotFound(apiv1.Resource("configmap"), "nope, not found")
	result, err := WriteStatusConfigMap(ti.client, ti.namespace, "TEST_MSG", nil)
	assert.Contains(t, result.Data["status"], "TEST_MSG")
	assert.Contains(t, result.ObjectMeta.Annotations, ConfigMapLastUpdatedKey)
	assert.Nil(t, err)
	assert.True(t, ti.getCalled)
	assert.False(t, ti.updateCalled)
	assert.True(t, ti.createCalled)
}

func TestWriteStatusConfigMapError(t *testing.T) {
	ti := setUpTest(t)
	ti.getError = errors.New("stuff bad")
	result, err := WriteStatusConfigMap(ti.client, ti.namespace, "TEST_MSG", nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "stuff bad")
	assert.Nil(t, result)
	assert.True(t, ti.getCalled)
	assert.False(t, ti.updateCalled)
	assert.False(t, ti.createCalled)
}
