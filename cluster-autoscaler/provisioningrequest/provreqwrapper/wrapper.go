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
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
)

// ProvisioningRequest wrapper representation of the ProvisioningRequest
type ProvisioningRequest struct {
	*v1.ProvisioningRequest
	PodTemplates []*apiv1.PodTemplate
}

// PodSet wrapper representation of the PodSet.
type PodSet struct {
	// Count number of pods with given template.
	Count int32
	// PodTemplate template of given pod set.
	PodTemplate apiv1.PodTemplateSpec
}

// NewProvisioningRequest creates new ProvisioningRequest based on v1 CR.
func NewProvisioningRequest(pr *v1.ProvisioningRequest, podTemplates []*apiv1.PodTemplate) *ProvisioningRequest {
	return &ProvisioningRequest{
		ProvisioningRequest: pr,
		PodTemplates:        podTemplates,
	}
}

// SetConditions of the Provisioning Request.
func (pr *ProvisioningRequest) SetConditions(conditions []metav1.Condition) {
	pr.Status.Conditions = conditions
	return
}

// PodSets of the Provisioning Request.
func (pr *ProvisioningRequest) PodSets() ([]PodSet, error) {
	if len(pr.Spec.PodSets) != len(pr.PodTemplates) {
		return nil, errMissingPodTemplates(pr.Spec.PodSets, pr.PodTemplates)
	}
	podSets := make([]PodSet, 0, len(pr.Spec.PodSets))
	for i, podSet := range pr.Spec.PodSets {
		podSets = append(podSets, PodSet{
			Count:       podSet.Count,
			PodTemplate: pr.PodTemplates[i].Template,
		})
	}
	return podSets, nil
}

// errMissingPodTemplates creates error that is passed when there are missing pod templates.
func errMissingPodTemplates(podSets []v1.PodSet, podTemplates []*apiv1.PodTemplate) error {
	foundPodTemplates := map[string]struct{}{}
	for _, pt := range podTemplates {
		foundPodTemplates[pt.Name] = struct{}{}
	}
	missingTemplates := make([]string, 0)
	for _, ps := range podSets {
		if _, found := foundPodTemplates[ps.PodTemplateRef.Name]; !found {
			missingTemplates = append(missingTemplates, ps.PodTemplateRef.Name)
		}
	}
	return fmt.Errorf("missing pod templates, %d pod templates were referenced, %d templates were missing: %s", len(podSets), len(missingTemplates), strings.Join(missingTemplates, ","))
}
