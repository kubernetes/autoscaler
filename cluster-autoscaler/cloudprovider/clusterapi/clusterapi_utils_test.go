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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	uuid1 = "ec21c5fb-a3d5-a45f-887b-6b49aa8fc218"
	uuid2 = "ec23ebb0-bc60-443f-d139-046ec5046283"
)

func TestUtilParseScalingBounds(t *testing.T) {
	for i, tc := range []struct {
		description string
		annotations map[string]string
		error       error
		min         int
		max         int
	}{{
		description: "missing min annotation defaults to 0 and no error",
		annotations: map[string]string{
			nodeGroupMaxSizeAnnotationKey: "0",
		},
	}, {
		description: "missing max annotation defaults to 0 and no error",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
		},
	}, {
		description: "invalid min errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "-1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		error: errInvalidMinAnnotation,
	}, {
		description: "invalid min errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "not-an-int",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		error: errInvalidMinAnnotation,
	}, {
		description: "invalid max errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "-1",
		},
		error: errInvalidMaxAnnotation,
	}, {
		description: "invalid max errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "not-an-int",
		},
		error: errInvalidMaxAnnotation,
	}, {
		description: "negative min errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "-1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		error: errInvalidMinAnnotation,
	}, {
		description: "negative max errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "-1",
		},
		error: errInvalidMaxAnnotation,
	}, {
		description: "max < min errors",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		error: errInvalidMaxAnnotation,
	}, {
		description: "result is: min 0, max 0",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		min: 0,
		max: 0,
	}, {
		description: "result is min 0, max 1",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "1",
		},
		min: 0,
		max: 1,
	}, {
		description: "deprecated min/max annotations still work, result is min 0, max 1",
		annotations: map[string]string{
			deprecatedNodeGroupMinSizeAnnotationKey: "0",
			deprecatedNodeGroupMaxSizeAnnotationKey: "1",
		},
		min: 0,
		max: 1,
	}, {
		description: "deprecated min/max annotations do not take precedence over non-deprecated annotations, result is min 1, max 2",
		annotations: map[string]string{
			deprecatedNodeGroupMinSizeAnnotationKey: "0",
			deprecatedNodeGroupMaxSizeAnnotationKey: "1",
			nodeGroupMinSizeAnnotationKey:           "1",
			nodeGroupMaxSizeAnnotationKey:           "2",
		},
		min: 1,
		max: 2,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			machineSet := unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       machineSetKind,
					"apiVersion": "cluster.x-k8s.io/v1alpha3",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
					},
					"spec":   map[string]interface{}{},
					"status": map[string]interface{}{},
				},
			}
			machineSet.SetAnnotations(tc.annotations)

			min, max, err := parseScalingBounds(machineSet.GetAnnotations())
			if tc.error != nil && err == nil {
				t.Fatalf("test #%d: expected an error", i)
			}

			if tc.error != nil && tc.error != err {
				if !strings.HasPrefix(err.Error(), tc.error.Error()) {
					t.Errorf("expected message to have prefix %q, got %q", tc.error.Error(), err)
				}
			}

			if tc.error == nil {
				if tc.min != min {
					t.Errorf("expected min %d, got %d", tc.min, min)
				}
				if tc.max != max {
					t.Errorf("expected max %d, got %d", tc.max, max)
				}
			}
		})
	}
}

func TestUtilGetOwnerByKindMachineSet(t *testing.T) {
	for _, tc := range []struct {
		description         string
		machineSet          *unstructured.Unstructured
		machineSetOwnerRefs []metav1.OwnerReference
		expectedOwnerRef    *metav1.OwnerReference
	}{{
		description:         "not owned as no owner references",
		machineSet:          &unstructured.Unstructured{},
		machineSetOwnerRefs: []metav1.OwnerReference{},
		expectedOwnerRef:    nil,
	}, {
		description: "not owned as not the same Kind",
		machineSet: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineSetOwnerRefs: []metav1.OwnerReference{
			{
				Kind: "Other",
			},
		},
		expectedOwnerRef: nil,
	}, {
		description: "not owned because no OwnerReference.Name",
		machineSet: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineSetOwnerRefs: []metav1.OwnerReference{
			{
				Kind: machineDeploymentKind,
				UID:  uuid1,
			},
		},
		expectedOwnerRef: nil,
	}, {
		description: "owned as UID values match and same Kind and Name not empty",
		machineSet: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineSetOwnerRefs: []metav1.OwnerReference{
			{
				Kind: machineDeploymentKind,
				Name: "foo",
				UID:  uuid1,
			},
		},
		expectedOwnerRef: &metav1.OwnerReference{
			Kind: machineDeploymentKind,
			Name: "foo",
			UID:  uuid1,
		},
	}} {
		t.Run(tc.description, func(t *testing.T) {
			tc.machineSet.SetOwnerReferences(tc.machineSetOwnerRefs)

			ownerRef := getOwnerForKind(tc.machineSet, machineDeploymentKind)
			if !reflect.DeepEqual(tc.expectedOwnerRef, ownerRef) {
				t.Errorf("expected %v, got %v", tc.expectedOwnerRef, ownerRef)
			}
		})
	}
}

func TestUtilGetOwnerByKindMachine(t *testing.T) {
	for _, tc := range []struct {
		description      string
		machine          *unstructured.Unstructured
		machineOwnerRefs []metav1.OwnerReference
		expectedOwnerRef *metav1.OwnerReference
	}{{
		description:      "not owned as no owner references",
		machine:          &unstructured.Unstructured{},
		machineOwnerRefs: []metav1.OwnerReference{},
		expectedOwnerRef: nil,
	}, {
		description: "not owned as not the same Kind",
		machine: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineOwnerRefs: []metav1.OwnerReference{
			{
				Kind: "Other",
				Name: "foo",
				UID:  uuid1,
			},
		},
		expectedOwnerRef: nil,
	}, {
		description: "not owned because no OwnerReference.Name",
		machine: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineOwnerRefs: []metav1.OwnerReference{
			{
				Kind: machineSetKind,
				UID:  uuid1,
			},
		},
		expectedOwnerRef: nil,
	}, {
		description: "owned as UID values match and same Kind and Name not empty",
		machine: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineOwnerRefs: []metav1.OwnerReference{
			{
				Kind: machineSetKind,
				Name: "foo",
				UID:  uuid2,
			},
		},
		expectedOwnerRef: &metav1.OwnerReference{
			Kind: machineSetKind,
			Name: "foo",
			UID:  uuid2,
		},
	}} {
		t.Run(tc.description, func(t *testing.T) {
			tc.machine.SetOwnerReferences(tc.machineOwnerRefs)

			ownerRef := getOwnerForKind(tc.machine, machineSetKind)
			if !reflect.DeepEqual(tc.expectedOwnerRef, ownerRef) {
				t.Errorf("expected %v, got %v", tc.expectedOwnerRef, ownerRef)
			}
		})
	}
}

func TestUtilMachineSetHasMachineDeploymentOwnerRef(t *testing.T) {
	for _, tc := range []struct {
		description         string
		machineSet          *unstructured.Unstructured
		machineSetOwnerRefs []metav1.OwnerReference
		owned               bool
	}{{
		description:         "machineset not owned as no owner references",
		machineSet:          &unstructured.Unstructured{},
		machineSetOwnerRefs: []metav1.OwnerReference{},
		owned:               false,
	}, {
		description: "machineset not owned as ownerref not a MachineDeployment",
		machineSet: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineSetOwnerRefs: []metav1.OwnerReference{
			{
				Kind: "Other",
			},
		},
		owned: false,
	}, {
		description: "machineset owned as Kind matches and Name not empty",
		machineSet: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
		machineSetOwnerRefs: []metav1.OwnerReference{
			{
				Kind: machineDeploymentKind,
				Name: "foo",
			},
		},
		owned: true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			tc.machineSet.SetOwnerReferences(tc.machineSetOwnerRefs)
			owned := machineSetHasMachineDeploymentOwnerRef(tc.machineSet)
			if tc.owned != owned {
				t.Errorf("expected %t, got %t", tc.owned, owned)
			}
		})
	}
}

func TestUtilNormalizedProviderID(t *testing.T) {
	for _, tc := range []struct {
		description string
		providerID  string
		expectedID  normalizedProviderID
	}{{
		description: "nil string yields empty string",
		providerID:  "",
		expectedID:  "",
	}, {
		description: "empty string",
		providerID:  "",
		expectedID:  "",
	}, {
		description: "id without / characters",
		providerID:  "i-12345678",
		expectedID:  "i-12345678",
	}, {
		description: "id with / characters",
		providerID:  "aws:////i-12345678",
		expectedID:  "i-12345678",
	}} {
		t.Run(tc.description, func(t *testing.T) {
			actualID := normalizedProviderString(tc.providerID)
			if actualID != tc.expectedID {
				t.Errorf("expected %v, got %v", tc.expectedID, actualID)
			}
		})
	}
}

func Test_clusterNameFromResource(t *testing.T) {
	for _, tc := range []struct {
		name     string
		resource *unstructured.Unstructured
		want     string
	}{{
		name: "cluster name not set, v1alpha1 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "",
	}, {
		name: "cluster name not set, v1alpha1 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "",
	}, {
		name: "cluster name set in MachineSet labels, v1alpha1 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
					"labels": map[string]interface{}{
						deprecatedClusterNameLabel: "bar",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in MachineDeployment, v1alpha1 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
					"labels": map[string]interface{}{
						deprecatedClusterNameLabel: "bar",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in Machine template labels, v1alpha1 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								deprecatedClusterNameLabel: "bar",
							},
						},
					},
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in Machine template, v1alpha1 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.k8s.io/v1alpha1",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								deprecatedClusterNameLabel: "bar",
							},
						},
					},
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name not set, v1alpha2 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha2",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "",
	}, {
		name: "cluster name not set, v1alpha2 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha2",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "",
	}, {
		name: "cluster name set in MachineSet labels, v1alpha2 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha2",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
					"labels": map[string]interface{}{
						clusterNameLabel: "bar",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in MachineDeployment, v1alpha2 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha2",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
					"labels": map[string]interface{}{
						clusterNameLabel: "bar",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in spec, v1alpha3 MachineSet",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"clusterName": "bar",
					"replicas":    int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}, {
		name: "cluster name set in spec, v1alpha3 MachineDeployment",
		resource: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineDeploymentKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      "foo",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"clusterName": "bar",
					"replicas":    int64(1),
				},
				"status": map[string]interface{}{},
			},
		},
		want: "bar",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			if got := clusterNameFromResource(tc.resource); got != tc.want {
				t.Errorf("clusterNameFromResource() = %v, want %v", got, tc.want)
			}
		})
	}
}
