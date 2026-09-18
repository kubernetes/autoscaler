/*
Copyright 2022 The Kubernetes Authors.

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

package utho

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestNormalizeID(t *testing.T) {
	assert.Equal(t, "abc-123", normalizeID("utho://abc-123"))
	assert.Equal(t, "abc-123", normalizeID("abc-123"))
}

func TestToProviderID(t *testing.T) {
	assert.Equal(t, "utho://abc-123", toProviderID("abc-123"))
}

func TestReadyConditions(t *testing.T) {
	conditions := readyConditions()

	assert.Len(t, conditions, 1)
	assert.Equal(t, apiv1.NodeReady, conditions[0].Type)
	assert.Equal(t, apiv1.ConditionTrue, conditions[0].Status)
}

func TestBuildKubeProxy(t *testing.T) {
	pod := buildKubeProxy("pool-1")

	assert.Equal(t, "kube-proxy-pool-1", pod.Name)
	assert.Equal(t, "kube-system", pod.Namespace)
	assert.Equal(t, "kube-proxy", pod.Labels["k8s-app"])
}

func TestJoin(t *testing.T) {
	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, join(map[string]string{"a": "1"}, map[string]string{"b": "2"}))
	assert.Equal(t, map[string]string{"a": "2"}, join(map[string]string{"a": "1"}, map[string]string{"a": "2"}))
	assert.Equal(t, map[string]string{"a": "1"}, join(nil, map[string]string{"a": "1"}))
}
