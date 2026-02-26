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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	buffersfilter "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// Pods annotation keys and values for fake pods created by capacity buffer pod list processor
const (
	CapacityBufferFakePodAnnotationKey   = "podType"
	CapacityBufferFakePodAnnotationValue = "capacityBufferFakePod"

	// NotReadyForProvisioningReason is the reason used when the buffer is not ready for provisioning.
	NotReadyForProvisioningReason = "NotReadyForProvisioning"
	// BufferIsEmptyReason is the reason used when the buffer has 0 replicas.
	BufferIsEmptyReason = "BufferIsEmpty"
	// FailedToGetPodTemplateReason is the reason used when the pod template cannot be retrieved.
	FailedToGetPodTemplateReason = "FailedToGetPodTemplate"
	// FailedToMakeFakePodsReason is the reason used when creating fake pods fails.
	FailedToMakeFakePodsReason = "FailedToMakeFakePods"
	// FakePodsInjectedReason is the reason used when fake pods are successfully injected.
	FakePodsInjectedReason = "FakePodsInjected"
)

// CapacityBufferPodListProcessor processes the pod lists before scale up
// and adds buffres api virtual pods.
type CapacityBufferPodListProcessor struct {
	client                   *client.CapacityBufferClient
	statusFilter             buffersfilter.Filter
	podTemplateGenFilter     buffersfilter.Filter
	provStrategies           map[string]bool
	buffersRegistry          *capacityBuffersFakePodsRegistry
	forceSafeToEvictFakePods bool
}

// capacityBuffersFakePodsRegistry a struct that keeps the status of capacity buffer
// the fake pods generated for adding buffer event later
type capacityBuffersFakePodsRegistry struct {
	fakePodsUIDToBuffer map[string]*v1beta1.CapacityBuffer
}

// NewCapacityBuffersFakePodsRegistry returns a new pointer to empty capacityBuffersFakePodsRegistry
func NewCapacityBuffersFakePodsRegistry(fakePodsToBuffers map[string]*v1beta1.CapacityBuffer) *capacityBuffersFakePodsRegistry {
	return &capacityBuffersFakePodsRegistry{fakePodsUIDToBuffer: fakePodsToBuffers}
}

// NewDefaultCapacityBuffersFakePodsRegistry returns a new pointer to empty capacityBuffersFakePodsRegistry
func NewDefaultCapacityBuffersFakePodsRegistry() *capacityBuffersFakePodsRegistry {
	return &capacityBuffersFakePodsRegistry{fakePodsUIDToBuffer: map[string]*v1beta1.CapacityBuffer{}}
}

// NewCapacityBufferPodListProcessor creates a new CapacityRequestPodListProcessor.
func NewCapacityBufferPodListProcessor(client *client.CapacityBufferClient, provStrategies []string, buffersRegistry *capacityBuffersFakePodsRegistry, forceSafeToEvictFakePods bool) *CapacityBufferPodListProcessor {
	provStrategiesMap := map[string]bool{}
	for _, ps := range provStrategies {
		provStrategiesMap[ps] = true
	}
	return &CapacityBufferPodListProcessor{
		client: client,
		statusFilter: buffersfilter.NewStatusFilter(map[string]string{
			capacitybuffer.ReadyForProvisioningCondition: string(metav1.ConditionTrue),
			capacitybuffer.ProvisioningCondition:         string(metav1.ConditionTrue),
		}),
		podTemplateGenFilter:     buffersfilter.NewPodTemplateGenerationChangedFilter(client),
		provStrategies:           provStrategiesMap,
		buffersRegistry:          buffersRegistry,
		forceSafeToEvictFakePods: forceSafeToEvictFakePods,
	}
}

// Process updates unschedulablePods by injecting fake pods to match replicas defined in buffers status
func (p *CapacityBufferPodListProcessor) Process(autoscalingCtx *ca_context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	buffers, err := p.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("CapacityBufferPodListProcessor failed to list buffers with error: %v", err.Error())
		return unschedulablePods, nil
	}
	buffers = p.filterBuffersProvStrategy(buffers)
	_, buffers = p.statusFilter.Filter(buffers)
	_, buffers = p.podTemplateGenFilter.Filter(buffers)

	totalFakePods := []*apiv1.Pod{}
	p.clearCapacityBufferRegistry()
	for _, buffer := range buffers {
		fakePods := p.provision(buffer)
		p.updateCapacityBufferRegistry(fakePods, buffer)
		totalFakePods = append(totalFakePods, fakePods...)
	}
	klog.V(2).Infof("Capacity pod processor injecting %v fake pods provisioning %v capacity buffers", len(totalFakePods), len(buffers))
	unschedulablePods = append(unschedulablePods, totalFakePods...)
	return unschedulablePods, nil
}

// CleanUp is called at CA termination
func (p *CapacityBufferPodListProcessor) CleanUp() {
}

func (p *CapacityBufferPodListProcessor) updateCapacityBufferRegistry(fakePods []*apiv1.Pod, buffer *v1beta1.CapacityBuffer) {
	if p.buffersRegistry == nil {
		return
	}
	for _, fakePod := range fakePods {
		p.buffersRegistry.fakePodsUIDToBuffer[string(fakePod.UID)] = buffer
	}
}

func (p *CapacityBufferPodListProcessor) clearCapacityBufferRegistry() {
	if p.buffersRegistry == nil {
		return
	}
	p.buffersRegistry.fakePodsUIDToBuffer = make(map[string]*v1beta1.CapacityBuffer, 0)
}

func (p *CapacityBufferPodListProcessor) provision(buffer *v1beta1.CapacityBuffer) []*apiv1.Pod {
	if buffer.Status.PodTemplateRef == nil || buffer.Status.Replicas == nil || meta.IsStatusConditionFalse(buffer.Status.Conditions, capacitybuffer.ReadyForProvisioningCondition) {
		changed := common.UpdateBufferStatusToFailedProvisioning(buffer, NotReadyForProvisioningReason, "CapacityBuffer is not ready for provisioning")
		if changed {
			p.updateBufferStatus(buffer)
		}
		return []*apiv1.Pod{}
	}
	if *buffer.Status.Replicas == 0 {
		changed := common.UpdateBufferStatusToFailedProvisioning(buffer, BufferIsEmptyReason, "CapacityBuffer has zero replicas")
		if changed {
			p.updateBufferStatus(buffer)
		}
		return []*apiv1.Pod{}
	}
	podTemplateName := buffer.Status.PodTemplateRef.Name
	replicas := buffer.Status.Replicas
	podTemplate, err := p.client.GetPodTemplate(buffer.Namespace, podTemplateName)
	if err != nil {
		changed := common.UpdateBufferStatusToFailedProvisioning(buffer, FailedToGetPodTemplateReason, fmt.Sprintf("failed to get pod template with error: %v", err.Error()))
		if changed {
			p.updateBufferStatus(buffer)
		}
		return []*apiv1.Pod{}
	}
	fakePods, err := makeFakePods(buffer, &podTemplate.Template, int(*replicas), p.forceSafeToEvictFakePods)
	if err != nil {
		changed := common.UpdateBufferStatusToFailedProvisioning(buffer, FailedToMakeFakePodsReason, fmt.Sprintf("failed to create fake pods with error: %v", err.Error()))
		if changed {
			p.updateBufferStatus(buffer)
		}
		return []*apiv1.Pod{}
	}
	common.UpdateBufferStatusToSuccessfullyProvisioning(buffer, FakePodsInjectedReason)
	p.updateBufferStatus(buffer)
	return fakePods
}

func (p *CapacityBufferPodListProcessor) filterBuffersProvStrategy(buffers []*v1beta1.CapacityBuffer) []*v1beta1.CapacityBuffer {
	var filteredBuffers []*v1beta1.CapacityBuffer
	for _, buffer := range buffers {
		if buffer.Status.ProvisioningStrategy != nil && p.provStrategies[*buffer.Status.ProvisioningStrategy] {
			filteredBuffers = append(filteredBuffers, buffer)
		}
	}
	return filteredBuffers
}

func (p *CapacityBufferPodListProcessor) updateBufferStatus(buffer *v1beta1.CapacityBuffer) {
	_, err := p.client.UpdateCapacityBuffer(buffer)
	if err != nil {
		klog.Errorf("Failed to update buffer status for buffer %v, error: %v", buffer.Name, err.Error())
	}
}

// makeFakePods creates podCount number of copies of the sample pod
func makeFakePods(buffer *v1beta1.CapacityBuffer, samplePodTemplate *apiv1.PodTemplateSpec, podCount int, forceSafeToEvictFakePods bool) ([]*apiv1.Pod, error) {
	var fakePods []*apiv1.Pod
	samplePod := pod.GetPodFromTemplate(samplePodTemplate, buffer.Namespace)
	samplePod.Spec.NodeName = ""
	samplePod = withCapacityBufferFakePodAnnotation(samplePod)
	if forceSafeToEvictFakePods {
		samplePod = withSafeToEvictAnnotation(samplePod)
	}
	for i := 1; i <= podCount; i++ {
		fakePod := samplePod.DeepCopy()
		fakePod.Name = fmt.Sprintf("capacity-buffer-%s-%d", buffer.Name, i)
		fakePod.UID = types.UID(fmt.Sprintf("%s-%d", string(buffer.UID), i))
		fakePods = append(fakePods, fakePod)
	}
	return fakePods, nil
}

func withCapacityBufferFakePodAnnotation(pod *apiv1.Pod) *apiv1.Pod {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}
	pod.Annotations[CapacityBufferFakePodAnnotationKey] = CapacityBufferFakePodAnnotationValue
	return pod
}

func withSafeToEvictAnnotation(pod *apiv1.Pod) *apiv1.Pod {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}
	pod.Annotations[drain.PodSafeToEvictKey] = "true"
	return pod
}

// IsFakeCapacityBuffersPod checks if the pod is a capacity buffer fake pod using pod annotation.
func IsFakeCapacityBuffersPod(pod *apiv1.Pod) bool {
	if pod.Annotations == nil {
		return false
	}
	return pod.Annotations[CapacityBufferFakePodAnnotationKey] == CapacityBufferFakePodAnnotationValue
}
