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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
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

func TestScaleFromZeroEnabled(t *testing.T) {
	for _, tc := range []struct {
		description string
		enabled     bool
		annotations map[string]string
	}{{
		description: "nil annotations",
		enabled:     false,
	}, {
		description: "empty annotations",
		annotations: map[string]string{},
		enabled:     false,
	}, {
		description: "non-matching annotation",
		annotations: map[string]string{"foo": "bar"},
		enabled:     false,
	}, {
		description: "matching key, incomplete annotations",
		annotations: map[string]string{
			"foo":  "bar",
			cpuKey: "1",
		},
		enabled: false,
	}, {
		description: "matching key, complete annotations",
		annotations: map[string]string{
			"foo":     "bar",
			cpuKey:    "1",
			memoryKey: "2Mi",
		},
		enabled: true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			got := scaleFromZeroAnnotationsEnabled(tc.annotations)
			if tc.enabled != got {
				t.Errorf("expected %t, got %t", tc.enabled, got)
			}
		})
	}
}

func TestParseCPUCapacity(t *testing.T) {
	for _, tc := range []struct {
		description      string
		annotations      map[string]string
		expectedQuantity resource.Quantity
		expectedError    bool
	}{{
		description:      "nil annotations",
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "empty annotations",
		annotations:      map[string]string{},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "bad quantity",
		annotations:      map[string]string{cpuKey: "not-a-quantity"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}, {
		description:      "valid quantity with units",
		annotations:      map[string]string{cpuKey: "123m"},
		expectedError:    false,
		expectedQuantity: resource.MustParse("123m"),
	}, {
		description:      "valid quantity without units",
		annotations:      map[string]string{cpuKey: "1"},
		expectedError:    false,
		expectedQuantity: resource.MustParse("1"),
	}, {
		description:      "valid fractional quantity without units",
		annotations:      map[string]string{cpuKey: "0.1"},
		expectedError:    false,
		expectedQuantity: resource.MustParse("0.1"),
	}} {
		t.Run(tc.description, func(t *testing.T) {
			got, err := parseCPUCapacity(tc.annotations)
			if tc.expectedError && err == nil {
				t.Fatal("expected an error")
			}
			if tc.expectedQuantity.Cmp(got) != 0 {
				t.Errorf("expected %v, got %v", tc.expectedQuantity.String(), got.String())
			}
		})
	}
}

func TestParseMemoryCapacity(t *testing.T) {
	for _, tc := range []struct {
		description      string
		annotations      map[string]string
		expectedQuantity resource.Quantity
		expectedError    bool
	}{{
		description:      "nil annotations",
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "empty annotations",
		annotations:      map[string]string{},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "bad quantity",
		annotations:      map[string]string{memoryKey: "not-a-quantity"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}, {
		description:      "quantity as with no unit type",
		annotations:      map[string]string{memoryKey: "1024"},
		expectedQuantity: *resource.NewQuantity(1024, resource.DecimalSI),
		expectedError:    false,
	}, {
		description:      "quantity with unit type (Mi)",
		annotations:      map[string]string{memoryKey: "456Mi"},
		expectedError:    false,
		expectedQuantity: *resource.NewQuantity(456*units.MiB, resource.DecimalSI),
	}, {
		description:      "quantity with unit type (Gi)",
		annotations:      map[string]string{memoryKey: "8Gi"},
		expectedError:    false,
		expectedQuantity: *resource.NewQuantity(8*units.GiB, resource.DecimalSI),
	}} {
		t.Run(tc.description, func(t *testing.T) {
			got, err := parseMemoryCapacity(tc.annotations)
			if tc.expectedError && err == nil {
				t.Fatal("expected an error")
			}
			if tc.expectedQuantity.Cmp(got) != 0 {
				t.Errorf("expected %v, got %v", tc.expectedQuantity.String(), got.String())
			}
		})
	}
}

func TestParseGPUCapacity(t *testing.T) {
	for _, tc := range []struct {
		description      string
		annotations      map[string]string
		expectedQuantity resource.Quantity
		expectedError    bool
	}{{
		description:      "nil annotations",
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "empty annotations",
		annotations:      map[string]string{},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "bad quantity",
		annotations:      map[string]string{gpuCountKey: "not-a-quantity"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}, {
		description:      "valid quantity",
		annotations:      map[string]string{gpuCountKey: "13"},
		expectedError:    false,
		expectedQuantity: resource.MustParse("13"),
	}, {
		description:      "valid quantity, bad unit type",
		annotations:      map[string]string{gpuCountKey: "13Mi"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			got, err := parseGPUCount(tc.annotations)
			if tc.expectedError && err == nil {
				t.Fatal("expected an error")
			}
			if tc.expectedQuantity.Cmp(got) != 0 {
				t.Errorf("expected %v, got %v", tc.expectedQuantity.String(), got.String())
			}
		})
	}
}

func TestParseMaxPodsCapacity(t *testing.T) {
	for _, tc := range []struct {
		description      string
		annotations      map[string]string
		expectedQuantity resource.Quantity
		expectedError    bool
	}{{
		description:      "nil annotations",
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "empty annotations",
		annotations:      map[string]string{},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    false,
	}, {
		description:      "bad quantity",
		annotations:      map[string]string{maxPodsKey: "not-a-quantity"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}, {
		description:      "valid quantity",
		annotations:      map[string]string{maxPodsKey: "13"},
		expectedError:    false,
		expectedQuantity: resource.MustParse("13"),
	}, {
		description:      "valid quantity, bad unit type",
		annotations:      map[string]string{maxPodsKey: "13Mi"},
		expectedQuantity: zeroQuantity.DeepCopy(),
		expectedError:    true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			got, err := parseMaxPodsCapacity(tc.annotations)
			if tc.expectedError && err == nil {
				t.Fatal("expected an error")
			}
			if tc.expectedQuantity.Cmp(got) != 0 {
				t.Errorf("expected %v, got %v", tc.expectedQuantity.String(), got.String())
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

func Test_getKeyHelpers(t *testing.T) {
	for _, tc := range []struct {
		name     string
		expected string
		testfunc func() string
	}{
		{
			name:     "default group, min size annotation key",
			expected: fmt.Sprintf("%s/cluster-api-autoscaler-node-group-min-size", defaultCAPIGroup),
			testfunc: getNodeGroupMinSizeAnnotationKey,
		},
		{
			name:     "default group, max size annotation key",
			expected: fmt.Sprintf("%s/cluster-api-autoscaler-node-group-max-size", defaultCAPIGroup),
			testfunc: getNodeGroupMaxSizeAnnotationKey,
		},
		{
			name:     "default group, machine delete annotation key",
			expected: fmt.Sprintf("%s/delete-machine", defaultCAPIGroup),
			testfunc: getMachineDeleteAnnotationKey,
		},
		{
			name:     "default group, machine annotation key",
			expected: fmt.Sprintf("%s/machine", defaultCAPIGroup),
			testfunc: getMachineAnnotationKey,
		},
		{
			name:     "default group, cluster name label",
			expected: fmt.Sprintf("%s/cluster-name", defaultCAPIGroup),
			testfunc: getClusterNameLabel,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			observed := tc.testfunc()
			if observed != tc.expected {
				t.Errorf("%s, mismatch, expected=%s, observed=%s", tc.name, observed, tc.expected)
			}
		})
	}

	testgroup := "test.k8s.io"
	t.Setenv(CAPIGroupEnvVar, testgroup)

	for _, tc := range []struct {
		name     string
		expected string
		testfunc func() string
	}{
		{
			name:     "test group, min size annotation key",
			expected: fmt.Sprintf("%s/cluster-api-autoscaler-node-group-min-size", testgroup),
			testfunc: getNodeGroupMinSizeAnnotationKey,
		},
		{
			name:     "test group, max size annotation key",
			expected: fmt.Sprintf("%s/cluster-api-autoscaler-node-group-max-size", testgroup),
			testfunc: getNodeGroupMaxSizeAnnotationKey,
		},
		{
			name:     "test group, machine delete annotation key",
			expected: fmt.Sprintf("%s/delete-machine", testgroup),
			testfunc: getMachineDeleteAnnotationKey,
		},
		{
			name:     "test group, mark machine for delete annotation key",
			expected: fmt.Sprintf("%s/machine-marked-for-delete", testgroup),
			testfunc: getMarkMachineForDeleteAnnotationKey,
		},
		{
			name:     "test group, machine annotation key",
			expected: fmt.Sprintf("%s/machine", testgroup),
			testfunc: getMachineAnnotationKey,
		},
		{
			name:     "test group, cluster name label",
			expected: fmt.Sprintf("%s/cluster-name", testgroup),
			testfunc: getClusterNameLabel,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			observed := tc.testfunc()
			if observed != tc.expected {
				t.Errorf("%s, mismatch, expected=%s, observed=%s", tc.name, observed, tc.expected)
			}
		})
	}
}
