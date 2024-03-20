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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
)

// ProvisioningRequest wrapper representation of the ProvisioningRequest
type ProvisioningRequest struct {
	v1Beta1PR           *v1beta1.ProvisioningRequest
	v1Beta1PodTemplates []*apiv1.PodTemplate
}

// PodSet wrapper representation of the PodSet.
type PodSet struct {
	// Count number of pods with given template.
	Count int32
	// PodTemplate template of given pod set.
	PodTemplate apiv1.PodTemplateSpec
}

// NewV1Beta1ProvisioningRequest creates new ProvisioningRequest based on v1beta1 CR.
func NewV1Beta1ProvisioningRequest(v1Beta1PR *v1beta1.ProvisioningRequest, v1Beta1PodTemplates []*apiv1.PodTemplate) *ProvisioningRequest {
	return &ProvisioningRequest{
		v1Beta1PR:           v1Beta1PR,
		v1Beta1PodTemplates: v1Beta1PodTemplates,
	}
}

// Name of the Provisioning Request.
func (pr *ProvisioningRequest) Name() string {
	return pr.v1Beta1PR.Name
}

// Namespace of the Provisioning Request.
func (pr *ProvisioningRequest) Namespace() string {
	return pr.v1Beta1PR.Namespace
}

// CreationTimestamp of the Provisioning Request.
func (pr *ProvisioningRequest) CreationTimestamp() metav1.Time {
	return pr.v1Beta1PR.CreationTimestamp
}

// RuntimeObject returns runtime.Object of the Provisioning Request.
func (pr *ProvisioningRequest) RuntimeObject() runtime.Object {
	return pr.v1Beta1PR
}

// APIVersion returns APIVersion of the Provisioning Request.
func (pr *ProvisioningRequest) APIVersion() string {
	return pr.v1Beta1PR.APIVersion
}

// Kind returns Kind of the Provisioning Request.
func (pr *ProvisioningRequest) Kind() string {
	return pr.v1Beta1PR.Kind

}

// UID returns UID of the Provisioning Request.
func (pr *ProvisioningRequest) UID() types.UID {
	return pr.v1Beta1PR.UID
}

// Conditions of the Provisioning Request.
func (pr *ProvisioningRequest) Conditions() []metav1.Condition {
	return pr.v1Beta1PR.Status.Conditions
}

// SetConditions of the Provisioning Request.
func (pr *ProvisioningRequest) SetConditions(conditions []metav1.Condition) {
	pr.v1Beta1PR.Status.Conditions = conditions
	return
}

// PodSets of the Provisioning Request.
func (pr *ProvisioningRequest) PodSets() ([]PodSet, error) {
	if len(pr.v1Beta1PR.Spec.PodSets) != len(pr.v1Beta1PodTemplates) {
		return nil, errMissingPodTemplates(pr.v1Beta1PR.Spec.PodSets, pr.v1Beta1PodTemplates)
	}
	podSets := make([]PodSet, 0, len(pr.v1Beta1PR.Spec.PodSets))
	for i, podSet := range pr.v1Beta1PR.Spec.PodSets {
		podSets = append(podSets, PodSet{
			Count:       podSet.Count,
			PodTemplate: pr.v1Beta1PodTemplates[i].Template,
		})
	}
	return podSets, nil
}

// V1Beta1 returns v1beta1 object CR, to be used only to pass information to clients.
func (pr *ProvisioningRequest) V1Beta1() *v1beta1.ProvisioningRequest {
	return pr.v1Beta1PR
}

// PodTemplates returns pod templates associated with the Provisioning Request, to be used only to pass information to clients.
func (pr *ProvisioningRequest) PodTemplates() []*apiv1.PodTemplate {
	return pr.v1Beta1PodTemplates
}

// errMissingPodTemplates creates error that is passed when there are missing pod templates.
func errMissingPodTemplates(podSets []v1beta1.PodSet, podTemplates []*apiv1.PodTemplate) error {
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
