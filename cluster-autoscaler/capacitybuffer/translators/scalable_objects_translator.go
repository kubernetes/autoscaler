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
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
	scalableobject "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators/scalable_objects"
	"k8s.io/klog/v2"
)

// ScalableObjectsTranslator translates buffers processors into pod capacity.
type ScalableObjectsTranslator struct {
	client             *cbclient.CapacityBufferClient
	resolver           fakepods.Resolver
	scaleResolver      *scalableobject.ScaleObjectPodResolver
	supportedResolvers map[string]scalableobject.ScalableObjectTemplateResolver
}

// NewDefaultScalableObjectsTranslator creates an instance of ScalableObjectsTranslator.
func NewDefaultScalableObjectsTranslator(client *cbclient.CapacityBufferClient, resolver fakepods.Resolver) *ScalableObjectsTranslator {
	supportedResolvers := map[string]scalableobject.ScalableObjectTemplateResolver{}
	for _, scalableObject := range scalableobject.GetSupportedScalableObjectResolvers(client) {
		supportedResolvers[scalableObject.GetResolverKey()] = scalableObject
	}
	scaleResolver := scalableobject.NewScaleObjectPodResolver(client)

	return &ScalableObjectsTranslator{
		client:             client,
		resolver:           resolver,
		supportedResolvers: supportedResolvers,
		scaleResolver:      scaleResolver,
	}
}

// GetScalableObjectKey returns a key
func GetScalableObjectKey(apiGroup, kind string) string {
	return fmt.Sprintf("%v-%v", apiGroup, kind)
}

func (t *ScalableObjectsTranslator) useSupportedResolvers(buffer *apiv1.CapacityBuffer) (*corev1.PodTemplateSpec, *int32, error) {
	resolverKey := scalableobject.GetResolverKey(buffer.Spec.ScalableRef.APIGroup, buffer.Spec.ScalableRef.Kind)
	scalableObj, found := t.supportedResolvers[resolverKey]
	if found {
		return scalableObj.GetTemplateAndReplicas(buffer.Namespace, buffer.Spec.ScalableRef.Name)
	}
	err := fmt.Errorf("kind %v/%v is not supported to use pod template", buffer.Spec.ScalableRef.APIGroup, buffer.Spec.ScalableRef.Kind)
	return nil, nil, err
}

func (t *ScalableObjectsTranslator) useScaleResolver(buffer *apiv1.CapacityBuffer) (*corev1.PodTemplateSpec, *int32, error) {
	return t.scaleResolver.GetTemplateAndReplicas(buffer.Namespace, buffer.Spec.ScalableRef.APIGroup, buffer.Spec.ScalableRef.Kind, buffer.Spec.ScalableRef.Name)
}

// Translate translates buffers processors into pod capacity.
func (t *ScalableObjectsTranslator) Translate(buffers []*apiv1.CapacityBuffer) []error {
	errors := []error{}
	for _, buffer := range buffers {
		if !isScalableObjectBuffer(buffer) {
			continue
		}
		if err := t.translateBuffer(context.TODO(), buffer); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (t *ScalableObjectsTranslator) translateBuffer(ctx context.Context, buffer *apiv1.CapacityBuffer) error {
	spec, replicas, err := t.resolveSpecAndReplicas(ctx, buffer)
	if err != nil {
		common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
		return err
	}

	managedTemplate, err := t.ensureManagedPodTemplate(ctx, buffer, spec)
	if err != nil {
		common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
		return err
	}

	numberOfPods := t.getBufferNumberOfPods(buffer, replicas)
	if numberOfPods == nil {
		err := errors.New("couldn't get number of replicas for buffer, replicas and percentage are not defined")
		common.SetBufferAsNotReadyForProvisioning(buffer, &apiv1.LocalObjectRef{Name: managedTemplate.Name}, &managedTemplate.Generation, nil, buffer.Spec.ProvisioningStrategy, err)
		// not returning err here, as it would trigger requeue. Hitting this case means that
		// the buffer is misconfigured and consecutive reconciliations will also fail until
		// the buffer spec is fixed.
		return nil
	}

	common.SetBufferAsReadyForProvisioning(buffer, &apiv1.LocalObjectRef{Name: managedTemplate.Name}, &managedTemplate.Generation, numberOfPods, buffer.Spec.ProvisioningStrategy)
	return nil
}

func (t *ScalableObjectsTranslator) resolveSpecAndReplicas(ctx context.Context, buffer *apiv1.CapacityBuffer) (*corev1.PodTemplateSpec, *int32, error) {
	// 1. Try resolving a live pod.
	// This is the preferred method as it uses the exact spec from a running pod, which is fully defaulted.
	spec, replicas, livePodErr := t.useScaleResolver(buffer)
	// TODO: if fetching the live pod fails, but resolving the scalable object spec
	// succeeds, error from fetching the live pod is hidden from the user. We should
	// record an event here.
	if livePodErr != nil {
		klog.Errorf("capacity buffer scalable objects translator: failed to resolve live pod: %v", livePodErr)
	}
	if spec != nil {
		return spec, replicas, nil
	}

	// 2. Resolve from scalable resource.
	// If live pod is not available, we fetch the spec from the scalable resource and
	// call server dry run on a pod with that spec to trigger pod defaulting and mutation webhooks.
	spec, replicas, resolveErr := t.useSupportedResolvers(buffer)
	if resolveErr != nil {
		if livePodErr != nil {
			return nil, nil, fmt.Errorf("failed to resolve live pod: %w, failed to resolve scalable object: %w", livePodErr, resolveErr)
		}
		return nil, nil, fmt.Errorf("failed to resolve scalable object: %w", resolveErr)
	}

	fakePod, err := t.resolver.Resolve(ctx, buffer.Namespace, spec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve pod spec: %w", err)
	}
	finalSpec := getPodTemplateSpecFromPod(fakePod)
	return finalSpec, replicas, nil
}

func (t *ScalableObjectsTranslator) ensureManagedPodTemplate(ctx context.Context, buffer *apiv1.CapacityBuffer, spec *corev1.PodTemplateSpec) (*corev1.PodTemplate, error) {
	targetPodTemplate := getPodTemplateFromSpec(spec, buffer)

	return t.client.EnsurePodTemplate(ctx, targetPodTemplate)
}

func (t *ScalableObjectsTranslator) getBufferNumberOfPods(buffer *apiv1.CapacityBuffer, scalableReplicas *int32) *int32 {

	var numberOfPodsFromPercentage *int32
	var numberOfPodsFromReplicas *int32

	if buffer.Spec.Percentage != nil {
		if scalableReplicas != nil {
			percentValue := buffer.Spec.Percentage
			numberOfPods := max(0, int32(int32(*percentValue)*(*scalableReplicas)/100.0))
			numberOfPodsFromPercentage = &numberOfPods
		}
	}
	if buffer.Spec.Replicas != nil {
		numberOfPods := max(0, int32(*buffer.Spec.Replicas))
		numberOfPodsFromReplicas = &numberOfPods
	}
	if numberOfPodsFromPercentage != nil && numberOfPodsFromReplicas != nil {
		numberOfPods := min(*numberOfPodsFromPercentage, *numberOfPodsFromReplicas)
		return &numberOfPods
	} else if numberOfPodsFromPercentage != nil {
		return numberOfPodsFromPercentage
	}
	return numberOfPodsFromReplicas
}

func isScalableObjectBuffer(buffer *apiv1.CapacityBuffer) bool {
	return buffer.Spec.ScalableRef != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *ScalableObjectsTranslator) CleanUp() {
}
