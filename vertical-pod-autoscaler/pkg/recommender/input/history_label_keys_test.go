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

package input

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestSanitizeLabelKey(t *testing.T) {
	testCases := map[string]string{
		"app":                    "app",
		"app_name":               "app_name",
		"app.kubernetes.io/name": "app_kubernetes_io_name",
		"example.com/team":       "example_com_team",
		"my-app":                 "my_app",
		"1st-tier":               "1st_tier",
		"appName":                "app_name",
		"APP":                    "app",
	}
	for key, expected := range testCases {
		t.Run(key, func(t *testing.T) {
			assert.Equal(t, expected, sanitizeLabelKey(key))
		})
	}
}

func TestRestoreLabelKeys(t *testing.T) {
	newVpa := func(namespace, name string, selector labels.Selector) *model.Vpa {
		return model.NewVpa(model.VpaID{Namespace: namespace, VpaName: name}, selector, time.Time{})
	}
	matchLabels := func(key string) labels.Selector {
		return labels.SelectorFromSet(labels.Set{key: "foo"})
	}
	matchExpression := func(key string) labels.Selector {
		selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{{Key: key, Operator: metav1.LabelSelectorOpIn, Values: []string{"foo"}}},
		})
		assert.NoError(t, err)
		return selector
	}

	testCases := []struct {
		name      string
		vpas      []*model.Vpa
		namespace string
		labels    map[string]string
		expected  map[string]string
	}{
		{
			name:      "restores key with dots and slash",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", matchLabels("app.kubernetes.io/name"))},
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo", "pod_template_hash": "abc"},
			expected:  map[string]string{"app.kubernetes.io/name": "foo", "pod_template_hash": "abc"},
		},
		{
			name:      "keeps both keys when the pod already has the selector key",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", matchLabels("app.kubernetes.io/name"))},
			namespace: "ns1",
			labels:    map[string]string{"app.kubernetes.io/name": "foo", "app_kubernetes_io_name": "bar"},
			expected:  map[string]string{"app.kubernetes.io/name": "foo", "app_kubernetes_io_name": "bar"},
		},
		{
			name:      "restores key with dash",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", matchLabels("my-app"))},
			namespace: "ns1",
			labels:    map[string]string{"my_app": "foo"},
			expected:  map[string]string{"my-app": "foo"},
		},
		{
			name:      "keeps key that is already valid",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", matchLabels("app"))},
			namespace: "ns1",
			labels:    map[string]string{"app": "foo"},
			expected:  map[string]string{"app": "foo"},
		},
		{
			name:      "restores key from matchExpressions",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", matchExpression("example.com/team"))},
			namespace: "ns1",
			labels:    map[string]string{"example_com_team": "foo"},
			expected:  map[string]string{"example.com/team": "foo"},
		},
		{
			name:      "keeps labels when no VPA exists",
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo"},
			expected:  map[string]string{"app_kubernetes_io_name": "foo"},
		},
		{
			name:      "keeps labels when the VPA is in another namespace",
			vpas:      []*model.Vpa{newVpa("ns2", "vpa", matchLabels("app.kubernetes.io/name"))},
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo"},
			expected:  map[string]string{"app_kubernetes_io_name": "foo"},
		},
		{
			name: "picks the first key in sorted order on a collision",
			vpas: []*model.Vpa{
				newVpa("ns1", "vpa", matchLabels("app.kubernetes.io/name")),
				newVpa("ns1", "other", matchLabels("app-kubernetes-io/name")),
			},
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo"},
			expected:  map[string]string{"app-kubernetes-io/name": "foo"},
		},
		{
			name: "keeps key that is itself a selector key on a collision",
			vpas: []*model.Vpa{
				newVpa("ns1", "vpa", matchLabels("app.kubernetes.io/name")),
				newVpa("ns1", "other", matchLabels("app_kubernetes_io_name")),
			},
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo"},
			expected:  map[string]string{"app_kubernetes_io_name": "foo"},
		},
		{
			name:      "ignores VPA without selector",
			vpas:      []*model.Vpa{newVpa("ns1", "vpa", nil)},
			namespace: "ns1",
			labels:    map[string]string{"app_kubernetes_io_name": "foo"},
			expected:  map[string]string{"app_kubernetes_io_name": "foo"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vpas := make(map[model.VpaID]*model.Vpa, len(tc.vpas))
			for _, vpa := range tc.vpas {
				vpas[vpa.ID] = vpa
			}
			selectorKeys := selectorLabelKeysByNamespace(vpas)
			assert.Equal(t, tc.expected, restoreLabelKeys(tc.labels, selectorKeys[tc.namespace]))
		})
	}
}
