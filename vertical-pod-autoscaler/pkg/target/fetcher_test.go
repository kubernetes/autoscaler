/*
Copyright 2026 The Kubernetes Authors.

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

package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestGetLabelSelectorCronJobNilOrEmptyTemplateLabels(t *testing.T) {
	t.Parallel()
	for name, labelsMap := range map[string]map[string]string{
		"nil labels":   nil,
		"empty labels": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cj := &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "cj"},
				Spec: batchv1.CronJobSpec{
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{Labels: labelsMap},
							},
						},
					},
				},
			}
			sel, err := getLabelSelectorFromCronJob(cj)
			require.NoError(t, err)
			assert.True(t, sel.Matches(labels.Set{}), "pod with no labels should match")
			assert.True(t, sel.Matches(labels.Set{"batch.kubernetes.io/job-name": "cj-123"}), "pod with controller labels should match")
		})
	}
}

func TestGetLabelSelectorCronJobNonEmptyTemplateLabels(t *testing.T) {
	t.Parallel()
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "cj"},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "longrunning"}},
					},
				},
			},
		},
	}
	sel, err := getLabelSelectorFromCronJob(cj)
	require.NoError(t, err)
	assert.True(t, sel.Matches(labels.Set{"app": "longrunning"}))
	assert.False(t, sel.Matches(labels.Set{"app": "other"}))
}
