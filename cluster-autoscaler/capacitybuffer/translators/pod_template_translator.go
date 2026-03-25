/*
Copyright 2025 The Kubernetes Authors.

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

package translator

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
)

// podTemplateBufferTranslator translates podTemplateRef buffers specs to fill their status.
type podTemplateBufferTranslator struct {
	client   *cbclient.CapacityBufferClient
	resolver fakepods.Resolver
}

// NewPodTemplateBufferTranslator creates an instance of podTemplateBufferTranslator.
func NewPodTemplateBufferTranslator(client *cbclient.CapacityBufferClient, resolver fakepods.Resolver) *podTemplateBufferTranslator {
	return &podTemplateBufferTranslator{
		client:   client,
		resolver: resolver,
	}
}

// Translate translates buffers podTemplateRef specs to fill their status.
func (t *podTemplateBufferTranslator) Translate(buffers []*v1.CapacityBuffer) []error {
	var errs []error
	var podTemplateRef *v1.LocalObjectRef
	for _, buffer := range buffers {
		if !isPodTemplateBasedBuffer(buffer) {
			continue
		}
		podTemplateRef = buffer.Spec.PodTemplateRef
		sourcePodTemplate, err := t.client.GetPodTemplate(buffer.Namespace, podTemplateRef.Name)
		if err != nil {
			common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
			errs = append(errs, err)
			continue
		}

		managedPodTemplate, err := t.ensureManagedPodTemplate(context.TODO(), buffer, sourcePodTemplate)
		if err != nil {
			errs = append(errs, err)
			common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
			continue
		}

		numberOfPods, err := getBufferNumberOfPods(buffer, managedPodTemplate.Template, nil)
		if err != nil {
			conditionErr := fmt.Errorf("couldn't get number of replicas for buffer: %w", err)
			common.SetBufferAsNotReadyForProvisioning(buffer, &v1.LocalObjectRef{Name: managedPodTemplate.Name}, &managedPodTemplate.Generation, nil, buffer.Spec.ProvisioningStrategy, conditionErr)
			// not recording an error here, as it would trigger requeue. Hitting this case means that
			// the buffer is misconfigured and consecutive reconciliations will also fail until
			// the buffer spec or a related object (pod template, scalable object) is fixed.
			continue
		}
		common.SetBufferAsReadyForProvisioning(buffer, &v1.LocalObjectRef{Name: managedPodTemplate.Name}, &managedPodTemplate.Generation, &numberOfPods, buffer.Spec.ProvisioningStrategy)
	}
	return errs
}

func (t *podTemplateBufferTranslator) ensureManagedPodTemplate(ctx context.Context, buffer *v1.CapacityBuffer, sourcePodTemplate *corev1.PodTemplate) (*corev1.PodTemplate, error) {
	fakePod, err := t.resolver.Resolve(ctx, buffer.Namespace, &sourcePodTemplate.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to create fake pod: %v", err)
	}

	newSpec := getPodTemplateSpecFromPod(fakePod)
	targetPodTemplate := getPodTemplateFromSpec(newSpec, buffer)
	managedPodTemplate, err := t.client.EnsurePodTemplate(ctx, targetPodTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update managed pod template: %v", err)
	}

	return managedPodTemplate, nil
}

func isPodTemplateBasedBuffer(buffer *v1.CapacityBuffer) bool {
	return buffer.Spec.PodTemplateRef != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *podTemplateBufferTranslator) CleanUp() {
}
