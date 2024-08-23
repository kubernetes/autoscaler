/*
Copyright 2023 The Kubernetes Authors.

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

package provreqwrapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
)

func TestProvisioningRequestWrapper(t *testing.T) {
	creationTimestamp := metav1.NewTime(time.Date(2023, 11, 12, 13, 14, 15, 0, time.UTC))
	conditions := []metav1.Condition{
		{
			LastTransitionTime: metav1.NewTime(time.Date(2022, 11, 12, 13, 14, 15, 0, time.UTC)),
			Message:            "Message",
			ObservedGeneration: 1,
			Reason:             "Reason",
			Status:             "Status",
			Type:               "ConditionType",
		},
	}
	podSets := []PodSet{
		{
			Count: 1,
			PodTemplate: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
			},
		},
	}

	podTemplates := []*apiv1.PodTemplate{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "name-pod-template-beta",
				Namespace:         "namespace-beta",
				CreationTimestamp: creationTimestamp,
			},
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
			},
		},
	}
	v1PR := &v1.ProvisioningRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "beta-api",
			Kind:       "beta-kind",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "name-beta",
			Namespace:         "namespace-beta",
			CreationTimestamp: creationTimestamp,
			UID:               types.UID("beta-uid"),
		},
		Spec: v1.ProvisioningRequestSpec{
			ProvisioningClassName: "queued-provisioning.gke.io",
			PodSets: []v1.PodSet{
				{
					Count: 1,
					PodTemplateRef: v1.Reference{
						Name: podTemplates[0].Name,
					},
				},
			},
		},
		Status: v1.ProvisioningRequestStatus{
			Conditions:               conditions,
			ProvisioningClassDetails: map[string]v1.Detail{},
		},
	}

	wrappedBetaPR := NewProvisioningRequest(v1PR, podTemplates)

	// Check Name, Namespace and Creation accessors
	assert.Equal(t, "name-beta", wrappedBetaPR.Name)
	assert.Equal(t, "namespace-beta", wrappedBetaPR.Namespace)
	assert.Equal(t, creationTimestamp, wrappedBetaPR.CreationTimestamp)

	// Check APIVersion, Kind and UID accessors
	assert.Equal(t, "beta-api", wrappedBetaPR.APIVersion)
	assert.Equal(t, "beta-kind", wrappedBetaPR.Kind)
	assert.Equal(t, types.UID("beta-uid"), wrappedBetaPR.UID)

	// Check the initial conditions
	assert.Equal(t, conditions, wrappedBetaPR.Status.Conditions)

	// Clear conditions and check the values
	wrappedBetaPR.SetConditions(nil)
	assert.Nil(t, wrappedBetaPR.Status.Conditions)

	// Set conditions and check the values
	wrappedBetaPR.SetConditions(conditions)
	assert.Equal(t, conditions, wrappedBetaPR.Status.Conditions)

	// Check the PodSets
	betaPodSets, betaErr := wrappedBetaPR.PodSets()
	assert.Nil(t, betaErr)
	assert.Equal(t, podSets, betaPodSets)

	// Check the type accessors.
	assert.Equal(t, v1PR, wrappedBetaPR.ProvisioningRequest)
	assert.Equal(t, podTemplates, wrappedBetaPR.PodTemplates)

	// Check case where the Provisioning Request is missing Pod Templates.
	wrappedBetaPRMissingPodTemplates := NewProvisioningRequest(v1PR, nil)
	podSets, err := wrappedBetaPRMissingPodTemplates.PodSets()
	assert.Nil(t, podSets)
	assert.EqualError(t, err, "missing pod templates, 1 pod templates were referenced, 1 templates were missing: name-pod-template-beta")
}
