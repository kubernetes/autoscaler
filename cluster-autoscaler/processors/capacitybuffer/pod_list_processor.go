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

package capacitybufferpodlister

import (
	"fmt"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	api_v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	client "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	buffersfilter "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// Pods annotation keys and values for fake pods created by capacity buffer pod list processor
const (
	FakeCapacityBufferPodAnnotationKey   = "podType"
	FakeCapacityBufferPodAnnotationValue = "capacityBufferFakePod"
)

// CapacityBufferPodListProcessor processes the pod lists before scale up
// and adds buffres api virtual pods.
type CapacityBufferPodListProcessor struct {
	client               *client.CapacityBufferClient
	statusFilter         buffersfilter.Filter
	podTemplateGenFilter buffersfilter.Filter
	provStrategies       map[string]bool
}

// NewCapacityBufferPodListProcessor creates a new CapacityRequestPodListProcessor.
func NewCapacityBufferPodListProcessor(client *client.CapacityBufferClient, provStrategies []string) *CapacityBufferPodListProcessor {
	provStrategiesMap := map[string]bool{}
	for _, ps := range provStrategies {
		provStrategiesMap[ps] = true
	}
	return &CapacityBufferPodListProcessor{
		client: client,
		statusFilter: buffersfilter.NewStatusFilter(map[string]string{
			common.ReadyForProvisioningCondition: common.ConditionTrue,
			common.ProvisioningCondition:         common.ConditionTrue,
		}),
		podTemplateGenFilter: buffersfilter.NewPodTemplateGenerationChangedFilter(client),
		provStrategies:       provStrategiesMap,
	}
}

// Process updates unschedulablePods by injecting fake pods to match replicas defined in buffers status
func (p *CapacityBufferPodListProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	buffers, err := p.client.ListCapacityBuffers()
	if err != nil {
		klog.Errorf("CapacityBufferPodListProcessor failed to list buffers with error: %v", err.Error())
		return unschedulablePods, nil
	}
	buffers = p.filterBuffersProvStrategy(buffers)
	_, buffers = p.statusFilter.Filter(buffers)
	_, buffers = p.podTemplateGenFilter.Filter(buffers)

	totalFakePods := []*apiv1.Pod{}
	for _, buffer := range buffers {
		fakePods := p.provision(buffer)
		totalFakePods = append(totalFakePods, fakePods...)
	}
	klog.V(2).Infof("Capacity pod processor injecting %v fake pods provisioning %v capacity buffers", len(totalFakePods), len(buffers))
	unschedulablePods = append(unschedulablePods, totalFakePods...)
	return unschedulablePods, nil
}

// CleanUp is called at CA termination
func (p *CapacityBufferPodListProcessor) CleanUp() {
}

func (p *CapacityBufferPodListProcessor) provision(buffer *api_v1.CapacityBuffer) []*apiv1.Pod {
	if buffer.Status.PodTemplateRef == nil || buffer.Status.Replicas == nil {
		return []*apiv1.Pod{}
	}
	podTemplateName := buffer.Status.PodTemplateRef.Name
	replicas := buffer.Status.Replicas
	podTemplate, err := p.client.GetPodTemplate(buffer.Namespace, podTemplateName)
	if err != nil {
		common.UpdateBufferStatusToFailedProvisioing(buffer, "FailedToGetPodTemplate", fmt.Sprintf("failed to get pod template with error: %v", err.Error()))
		p.updateBufferStatus(buffer)
		return []*apiv1.Pod{}
	}
	fakePods, err := makeFakePods(buffer.Name, &podTemplate.Template, int(*replicas))
	if err != nil {
		common.UpdateBufferStatusToFailedProvisioing(buffer, "FailedToMakeFakePods", fmt.Sprintf("failed to create fake pods with error: %v", err.Error()))
		p.updateBufferStatus(buffer)
		return []*apiv1.Pod{}
	}
	common.UpdateBufferStatusToSuccessfullyProvisioing(buffer, "FakePodsInjected")
	p.updateBufferStatus(buffer)
	return fakePods
}

func (p *CapacityBufferPodListProcessor) filterBuffersProvStrategy(buffers []*api_v1.CapacityBuffer) []*api_v1.CapacityBuffer {
	var filteredBuffers []*api_v1.CapacityBuffer
	for _, buffer := range buffers {

		if buffer.Status.ProvisioningStrategy != nil && p.provStrategies[*buffer.Status.ProvisioningStrategy] {
			filteredBuffers = append(filteredBuffers, buffer)
		}
	}
	return filteredBuffers
}

func (p *CapacityBufferPodListProcessor) updateBufferStatus(buffer *api_v1.CapacityBuffer) {
	_, err := p.client.UpdateCapacityBuffer(buffer)
	if err != nil {
		klog.Errorf("Failed to update buffer status for buffer %v, error: %v", buffer.Name, err.Error())
	}
}

// makeFakePods creates podCount number of copies of the sample pod
func makeFakePods(bufferName string, samplePodTemplate *apiv1.PodTemplateSpec, podCount int) ([]*apiv1.Pod, error) {
	var fakePods []*apiv1.Pod
	samplePod := getPodFromTemplate(samplePodTemplate)
	samplePod = withCapacityBufferFakePodAnnotation(samplePod)
	for i := 1; i <= podCount; i++ {
		newPod := samplePod.DeepCopy()
		newPod.Name = fmt.Sprintf("capacity-buffer-%s-%d", bufferName, i)
		newPod.UID = types.UID(fmt.Sprintf("%s-%d", string(bufferName), i))
		newPod.Spec.NodeName = ""
		fakePods = append(fakePods, newPod)
	}
	return fakePods, nil
}

func withCapacityBufferFakePodAnnotation(pod *apiv1.Pod) *apiv1.Pod {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}
	pod.Annotations[FakeCapacityBufferPodAnnotationKey] = FakeCapacityBufferPodAnnotationValue
	return pod
}

func isFakeCapacityBuffersPod(pod *apiv1.Pod) bool {
	if pod.Annotations == nil {
		return false
	}
	return pod.Annotations[FakeCapacityBufferPodAnnotationKey] == FakeCapacityBufferPodAnnotationValue
}

func getPodFromTemplate(template *apiv1.PodTemplateSpec) *apiv1.Pod {
	desiredLabels := getPodsLabelSet(template)
	desiredFinalizers := getPodsFinalizers(template)
	desiredAnnotations := getPodsAnnotationSet(template)

	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       desiredLabels,
			Namespace:    template.Namespace,
			Annotations:  desiredAnnotations,
			GenerateName: uuid.NewString(),
			Finalizers:   desiredFinalizers,
		},
	}

	pod.Spec = template.Spec
	return pod
}

func getPodsLabelSet(template *apiv1.PodTemplateSpec) labels.Set {
	desiredLabels := make(labels.Set)
	for k, v := range template.Labels {
		desiredLabels[k] = v
	}
	return desiredLabels
}

func getPodsFinalizers(template *apiv1.PodTemplateSpec) []string {
	desiredFinalizers := make([]string, len(template.Finalizers))
	copy(desiredFinalizers, template.Finalizers)
	return desiredFinalizers
}

func getPodsAnnotationSet(template *apiv1.PodTemplateSpec) labels.Set {
	desiredAnnotations := make(labels.Set)
	for k, v := range template.Annotations {
		desiredAnnotations[k] = v
	}
	return desiredAnnotations
}
