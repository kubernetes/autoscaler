/*
Copyright 2016 The Kubernetes Authors.

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

package gke

import (
	"fmt"
	"math/rand"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	mbPerGB           = 1000
	bytesPerMB        = 1000 * 1000
	millicoresPerCore = 1000
	// Kubelet "evictionHard: {memory.available}" is subtracted from
	// capacity when calculating allocatable (on top of kube-reserved).
	// We don't have a good place to get it from, but it has been hard-coded
	// to 100Mi since at least k8s 1.4.
	kubeletEvictionHardMemory = 100 * 1024 * 1024
)

// GkeTemplateBuilder builds templates for GKE cloud provider.
type GkeTemplateBuilder struct {
	gce.GceTemplateBuilder
	projectId string
}

// BuildNodeFromMigSpec builds node based on MIG's spec.
func (t *GkeTemplateBuilder) BuildNodeFromMigSpec(mig *GkeMig, cpu int64, mem int64) (*apiv1.Node, error) {
	if mig.Spec() == nil {
		return nil, fmt.Errorf("no spec in mig %s", mig.GceRef().Name)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-autoprovisioned-template-%d", mig.GceRef().Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	capacity, err := t.BuildCapacity(mig.Spec().MachineType, nil, mig.GceRef().Zone, cpu, mem)
	if err != nil {
		return nil, err
	}

	if gpuRequest, found := mig.Spec().ExtraResources[gpu.ResourceNvidiaGPU]; found {
		capacity[gpu.ResourceNvidiaGPU] = gpuRequest.DeepCopy()
	}

	node.Status = apiv1.NodeStatus{
		Capacity:    capacity,
		Allocatable: t.BuildAllocatableFromCapacity(capacity),
	}

	labels, err := buildLabelsForAutoprovisionedMig(mig, nodeName)
	if err != nil {
		return nil, err
	}
	node.Labels = labels

	node.Spec.Taints = mig.Spec().Taints

	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildLabelsForAutoprovisionedMig(mig *GkeMig, nodeName string) (map[string]string, error) {
	// GenericLabels
	labels, err := gce.BuildGenericLabels(mig.GceRef(), mig.Spec().MachineType, nodeName)
	if err != nil {
		return nil, err
	}
	for k, v := range mig.Spec().Labels {
		if existingValue, found := labels[k]; found {
			if v != existingValue {
				return map[string]string{}, fmt.Errorf("conflict in labels requested: %s=%s  present: %s=%s",
					k, v, k, existingValue)
			}
		} else {
			labels[k] = v
		}
	}
	return labels, nil
}
