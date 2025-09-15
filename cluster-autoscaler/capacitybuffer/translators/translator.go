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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
)

// Translator translates the passed buffers to pod template and number of replicas
type Translator interface {
	Translate(buffers []*v1.CapacityBuffer) []error
	CleanUp()
}

// combinedTranslator is a list of Translator
type combinedTranslator struct {
	translators []Translator
}

// NewCombinedTranslator construct combinedTranslator.
func NewCombinedTranslator(Translators []Translator) *combinedTranslator {
	return &combinedTranslator{Translators}
}

// AddTranslator append translator to the list.
func (b *combinedTranslator) AddTranslator(translator Translator) {
	b.translators = append(b.translators, translator)
}

// Translate runs sub-translate sequentially, in case more than one translator acted on same buffer
// last translator overrides the others
func (b *combinedTranslator) Translate(buffers []*v1.CapacityBuffer) []error {
	var errors []error
	for _, translator := range b.translators {
		bufferErrors := translator.Translate(buffers)
		errors = append(errors, bufferErrors...)
	}
	return errors
}

// CleanUp cleans up the translator's internal structures.
func (b *combinedTranslator) CleanUp() {
	for _, translator := range b.translators {
		translator.CleanUp()
	}
}

func setBufferAsReadyForProvisioning(buffer *v1.CapacityBuffer, podTemplateName string, replicas int32) {
	buffer.Status.PodTemplateRef = &v1.LocalObjectRef{
		Name: podTemplateName,
	}
	buffer.Status.Replicas = &replicas
	buffer.Status.PodTemplateGeneration = nil
	readyCondition := metav1.Condition{
		Type:               common.ReadyForProvisioningCondition,
		Status:             common.ConditionTrue,
		Message:            "ready",
		Reason:             "atrtibutesSetSuccessfully",
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	buffer.Status.Conditions = []metav1.Condition{readyCondition}
}

func setBufferAsNotReadyForProvisioning(buffer *v1.CapacityBuffer, errorMessage string) {
	buffer.Status.PodTemplateRef = nil
	buffer.Status.Replicas = nil
	buffer.Status.PodTemplateGeneration = nil
	notReadyCondition := metav1.Condition{
		Type:               common.ReadyForProvisioningCondition,
		Status:             common.ConditionFalse,
		Message:            errorMessage,
		Reason:             "error",
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	buffer.Status.Conditions = []metav1.Condition{notReadyCondition}
}
