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
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			machineSet := MachineSet{
				ObjectMeta: v1.ObjectMeta{
					Annotations: tc.annotations,
				},
			}

			min, max, err := parseScalingBounds(machineSet.Annotations)
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

func TestUtilMachineSetIsOwnedByMachineDeployment(t *testing.T) {
	for _, tc := range []struct {
		description       string
		machineSet        MachineSet
		machineDeployment MachineDeployment
		owned             bool
	}{{
		description:       "not owned as no owner references",
		machineSet:        MachineSet{},
		machineDeployment: MachineDeployment{},
		owned:             false,
	}, {
		description: "not owned as not the same Kind",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "Other",
				}},
			},
		},
		machineDeployment: MachineDeployment{},
		owned:             false,
	}, {
		description: "not owned because no OwnerReference.Name",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineSet",
					UID:  uuid1,
				}},
			},
		},
		machineDeployment: MachineDeployment{
			ObjectMeta: v1.ObjectMeta{
				UID: uuid1,
			},
		},
		owned: false,
	}, {
		description: "not owned as UID values don't match",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineSet",
					Name: "foo",
					UID:  uuid2,
				}},
			},
		},
		machineDeployment: MachineDeployment{
			TypeMeta: v1.TypeMeta{
				Kind: "MachineDeployment",
			},
			ObjectMeta: v1.ObjectMeta{
				UID: uuid1,
			},
		},
		owned: false,
	}, {
		description: "owned as UID values match and same Kind and Name not empty",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineDeployment",
					Name: "foo",
					UID:  uuid1,
				}},
			},
		},
		machineDeployment: MachineDeployment{
			TypeMeta: v1.TypeMeta{
				Kind: "MachineDeployment",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "foo",
				UID:  uuid1,
			},
		},
		owned: true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			owned := machineSetIsOwnedByMachineDeployment(&tc.machineSet, &tc.machineDeployment)
			if tc.owned != owned {
				t.Errorf("expected %t, got %t", tc.owned, owned)
			}
		})
	}
}

func TestUtilMachineIsOwnedByMachineSet(t *testing.T) {
	for _, tc := range []struct {
		description string
		machine     Machine
		machineSet  MachineSet
		owned       bool
	}{{
		description: "not owned as no owner references",
		machine:     Machine{},
		machineSet:  MachineSet{},
		owned:       false,
	}, {
		description: "not owned as not the same Kind",
		machine: Machine{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "Other",
				}},
			},
		},
		machineSet: MachineSet{},
		owned:      false,
	}, {
		description: "not owned because no OwnerReference.Name",
		machine: Machine{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineSet",
					UID:  uuid1,
				}},
			},
		},
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				UID: uuid1,
			},
		},
		owned: false,
	}, {
		description: "not owned as UID values don't match",
		machine: Machine{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineSet",
					Name: "foo",
					UID:  uuid2,
				}},
			},
		},
		machineSet: MachineSet{
			TypeMeta: v1.TypeMeta{
				Kind: "MachineSet",
			},
			ObjectMeta: v1.ObjectMeta{
				UID: uuid1,
			},
		},
		owned: false,
	}, {
		description: "owned as UID values match and same Kind and Name not empty",
		machine: Machine{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineSet",
					Name: "foo",
					UID:  uuid1,
				}},
			},
		},
		machineSet: MachineSet{
			TypeMeta: v1.TypeMeta{
				Kind: "MachineSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "foo",
				UID:  uuid1,
			},
		},
		owned: true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			owned := machineIsOwnedByMachineSet(&tc.machine, &tc.machineSet)
			if tc.owned != owned {
				t.Errorf("expected %t, got %t", tc.owned, owned)
			}
		})
	}
}

func TestUtilMachineSetMachineDeploymentOwnerRef(t *testing.T) {
	for _, tc := range []struct {
		description       string
		machineSet        MachineSet
		machineDeployment MachineDeployment
		owned             bool
	}{{
		description:       "machineset not owned as no owner references",
		machineSet:        MachineSet{},
		machineDeployment: MachineDeployment{},
		owned:             false,
	}, {
		description: "machineset not owned as ownerref not a MachineDeployment",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "Other",
				}},
			},
		},
		machineDeployment: MachineDeployment{},
		owned:             false,
	}, {
		description: "machineset owned as Kind matches and Name not empty",
		machineSet: MachineSet{
			ObjectMeta: v1.ObjectMeta{
				OwnerReferences: []v1.OwnerReference{{
					Kind: "MachineDeployment",
					Name: "foo",
				}},
			},
		},
		machineDeployment: MachineDeployment{
			TypeMeta: v1.TypeMeta{
				Kind: "MachineDeployment",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "foo",
			},
		},
		owned: true,
	}} {
		t.Run(tc.description, func(t *testing.T) {
			owned := machineSetHasMachineDeploymentOwnerRef(&tc.machineSet)
			if tc.owned != owned {
				t.Errorf("expected %t, got %t", tc.owned, owned)
			}
		})
	}
}
