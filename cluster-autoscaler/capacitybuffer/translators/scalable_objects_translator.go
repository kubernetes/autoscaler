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
	"fmt"

	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	scalableobject "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators/scalable_objects"
)

// ScalableObjectsTranslator translates buffers processors into pod capacity.
type ScalableObjectsTranslator struct {
	client             *cbclient.CapacityBufferClient
	scaleResolver      *scalableobject.ScaleObjectPodResolver
	supportedResolvers map[string]scalableobject.ScalableObjectTemplateResolver
}

// NewDefaultScalableObjectsTranslator creates an instance of ScalableObjectsTranslator.
func NewDefaultScalableObjectsTranslator(client *cbclient.CapacityBufferClient) *ScalableObjectsTranslator {
	supportedResolvers := map[string]scalableobject.ScalableObjectTemplateResolver{}
	for _, scalableObject := range scalableobject.GetSupportedScalableObjectResolvers(client) {
		supportedResolvers[scalableObject.GetResolverKey()] = scalableObject
	}
	scaleResolver := scalableobject.NewScaleObjectPodResolver(client)

	return &ScalableObjectsTranslator{
		client:             client,
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
		if isScalableObjectBuffer(buffer) {
			createdPodTemplate, podTempErr := t.client.GetPodTemplate(buffer.Namespace, getPodTemplateNameForBuffer(buffer))

			podTemplateSpec, replicasFromScalable, ScaleResolverErr := t.useScaleResolver(buffer)
			if ScaleResolverErr == nil && podTemplateSpec == nil {
				if podTempErr == nil && createdPodTemplate != nil {
					podTemplateSpec = createdPodTemplate.Template.DeepCopy()
				} else {
					ScaleResolverErr = fmt.Errorf("Couldn't resolve buffer %v, Controller requires to observe a running pod for ScalableRef %v", buffer.Name, buffer.Spec.ScalableRef.Name)
				}
			}
			var supportedResolversErr error
			if ScaleResolverErr != nil {
				podTemplateSpec, replicasFromScalable, supportedResolversErr = t.useSupportedResolvers(buffer)
			}

			if supportedResolversErr != nil {
				err := fmt.Errorf("Couldn't resolve buffer %v, error resolving scale object: %v, and error resolving supported objects %v", buffer.Name, ScaleResolverErr.Error(), supportedResolversErr.Error())
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, nil, err)
				errors = append(errors, err)
				continue
			}
			podTemplate := t.getPodTemplateFromSpecs(podTemplateSpec, buffer)
			if podTempErr != nil {
				createdPodTemplate, podTempErr = t.client.CreatePodTemplate(podTemplate)
			} else {
				createdPodTemplate, podTempErr = t.client.UpdatePodTemplate(podTemplate)
			}

			if podTempErr != nil {
				err := fmt.Errorf("Failed to create pod template object for buffer %v with error: %v", buffer.Name, podTempErr.Error())
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, nil, err)
				errors = append(errors, err)
				continue
			}
			numberOfPods := t.getBufferNumberOfPods(buffer, replicasFromScalable)
			if numberOfPods == nil {
				common.SetBufferAsNotReadyForProvisioning(buffer, &apiv1.LocalObjectRef{Name: createdPodTemplate.Name}, &createdPodTemplate.Generation, nil, buffer.Spec.ProvisioningStrategy, fmt.Errorf("Couldn't get number of replicas for buffer %v, replicas and percentage are not defined", buffer.Name))
				continue
			}
			common.SetBufferAsReadyForProvisioning(buffer, &apiv1.LocalObjectRef{Name: createdPodTemplate.Name}, &createdPodTemplate.Generation, numberOfPods, buffer.Spec.ProvisioningStrategy)

		}
	}
	return errors
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

func getPodTemplateNameForBuffer(buffer *apiv1.CapacityBuffer) string {
	return fmt.Sprintf("capacitybuffer-%v-pod-template", buffer.Name)
}

func (t *ScalableObjectsTranslator) getPodTemplateFromSpecs(pts *corev1.PodTemplateSpec, buffer *apiv1.CapacityBuffer) *corev1.PodTemplate {
	return &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPodTemplateNameForBuffer(buffer),
			Namespace: buffer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         common.CapacityBufferApiVersion,
					Kind:               common.CapacityBufferKind,
					Name:               buffer.Name,
					UID:                buffer.UID,
					Controller:         pointerBool(true),
					BlockOwnerDeletion: pointerBool(true),
				},
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      pts.Labels,
				Annotations: pts.Annotations,
			},
			Spec: pts.Spec,
		},
	}
}

func isScalableObjectBuffer(buffer *apiv1.CapacityBuffer) bool {
	return buffer.Spec.ScalableRef != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *ScalableObjectsTranslator) CleanUp() {
}

func pointerBool(b bool) *bool {
	return &b
}
