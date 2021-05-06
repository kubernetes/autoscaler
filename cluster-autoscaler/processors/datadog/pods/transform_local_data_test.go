/*
Copyright 2021 The Kubernetes Authors.

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

package pods

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/common"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	testRemoteClass    = "remote-data"
	testLocalClass     = "local-data"
	testNamespace      = "foons"
	testEmptyResources = corev1.ResourceList{}
	testLdResources    = corev1.ResourceList{
		common.DatadogLocalDataResource: common.DatadogLocalDataQuantity.DeepCopy(),
	}
)

func TestTransformLocalDataProcess(t *testing.T) {
	tests := []struct {
		name     string
		pods     []*corev1.Pod
		pvcs     []*corev1.PersistentVolumeClaim
		expected []*corev1.Pod
	}{
		{
			"No modification on remote volumes",
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1")},
			[]*corev1.PersistentVolumeClaim{buildPVC("pvc-1", testRemoteClass)},
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1")},
		},

		{
			"Cope with pod not having volumes",
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources)},
			[]*corev1.PersistentVolumeClaim{},
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources)},
		},

		{
			"local-data volumes are removed, and custom resources added",
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1")},
			[]*corev1.PersistentVolumeClaim{buildPVC("pvc-1", testLocalClass)},
			[]*corev1.Pod{buildPod("pod1", testLdResources, testLdResources)},
		},

		{
			"mixed local-data and remote volumes don't cause confusion",
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1", "pvc-2", "pvc-3")},
			[]*corev1.PersistentVolumeClaim{
				buildPVC("pvc-1", testRemoteClass),
				buildPVC("pvc-2", testLocalClass),
				buildPVC("pvc-3", testRemoteClass),
			},
			[]*corev1.Pod{buildPod("pod1", testLdResources, testLdResources, "pvc-1", "pvc-3")},
		},

		{
			"volumes using missing pvcs are conserved",
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1")},
			[]*corev1.PersistentVolumeClaim{},
			[]*corev1.Pod{buildPod("pod1", testEmptyResources, testEmptyResources, "pvc-1")},
		},

		{
			"empty pod list don't crash",
			[]*corev1.Pod{},
			[]*corev1.PersistentVolumeClaim{},
			[]*corev1.Pod{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvcLister, err := newTestPVCLister(tt.pvcs)
			assert.NoError(t, err)

			tld := transformLocalData{
				pvcLister: pvcLister,
			}
			actual, err := tld.Process(&context.AutoscalingContext{}, tt.pods)
			assert.NoError(t, err)
			assert.True(t, apiequality.Semantic.DeepEqual(tt.expected, actual))
		})
	}

}

func newTestPVCLister(pvcs []*corev1.PersistentVolumeClaim) (v1lister.PersistentVolumeClaimLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, pvc := range pvcs {
		err := store.Add(pvc)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1lister.NewPersistentVolumeClaimLister(store), nil
}

func buildPod(name string, requests, limits corev1.ResourceList, claimNames ...string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{},
			Containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: requests,
						Limits:   limits,
					},
				},
			},
		},
	}

	for _, name := range claimNames {
		pod.Spec.Volumes = append(pod.Spec.Volumes,
			corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: name,
					},
				},
			})
	}

	return pod
}

func buildPVC(name string, storageClassName string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
		},
	}
}
