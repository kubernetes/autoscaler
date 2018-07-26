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

package gce

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
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

// builds templates for gce cloud provider
type templateBuilder struct {
	projectId string
}

func (t *templateBuilder) getAcceleratorCount(accelerators []*gce.AcceleratorConfig) int64 {
	count := int64(0)
	for _, accelerator := range accelerators {
		if strings.HasPrefix(accelerator.AcceleratorType, "nvidia-") {
			count += accelerator.AcceleratorCount
		}
	}
	return count
}

func (t *templateBuilder) buildCapacity(machineType string, accelerators []*gce.AcceleratorConfig, zone string, cpu int64, mem int64) (apiv1.ResourceList, error) {
	capacity := apiv1.ResourceList{}
	// TODO: get a real value.
	capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	capacity[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
	capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)

	if accelerators != nil && len(accelerators) > 0 {
		capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(t.getAcceleratorCount(accelerators), resource.DecimalSI)
	}

	return capacity, nil
}

// buildAllocatableFromKubeEnv builds node allocatable based on capacity of the node and
// value of kubeEnv.
// KubeEnv is a multi-line string containing entries in the form of
// <RESOURCE_NAME>:<string>. One of the resources it contains is a list of
// kubelet arguments from which we can extract the resources reserved by
// the kubelet for its operation. Allocated resources are capacity minus reserved.
// If we fail to extract the reserved resources from kubeEnv (e.g it is in a
// wrong format or does not contain kubelet arguments), we return an error.
func (t *templateBuilder) buildAllocatableFromKubeEnv(capacity apiv1.ResourceList, kubeEnv string) (apiv1.ResourceList, error) {
	kubeReserved, err := extractKubeReservedFromKubeEnv(kubeEnv)
	if err != nil {
		return nil, err
	}
	reserved, err := parseKubeReserved(kubeReserved)
	if err != nil {
		return nil, err
	}
	if quantity, found := reserved[apiv1.ResourceMemory]; found {
		reserved[apiv1.ResourceMemory] = *resource.NewQuantity(quantity.Value()+kubeletEvictionHardMemory, resource.BinarySI)
	}
	return t.getAllocatable(capacity, reserved), nil
}

// buildAllocatableFromCapacity builds node allocatable based only on node capacity.
// Calculates reserved as a ratio of capacity. See calculateReserved for more details
func (t *templateBuilder) buildAllocatableFromCapacity(capacity apiv1.ResourceList) apiv1.ResourceList {
	memoryReserved := memoryReservedMB(capacity.Memory().Value() / bytesPerMB)
	cpuReserved := cpuReservedMillicores(capacity.Cpu().MilliValue())
	reserved := apiv1.ResourceList{}
	reserved[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpuReserved, resource.DecimalSI)
	// Duplicating an upstream bug treating MB as MiB (we need to predict the end result accurately).
	memoryReserved = memoryReserved * 1024 * 1024
	memoryReserved += kubeletEvictionHardMemory
	reserved[apiv1.ResourceMemory] = *resource.NewQuantity(memoryReserved, resource.BinarySI)
	return t.getAllocatable(capacity, reserved)
}

func (t *templateBuilder) getAllocatable(capacity, reserved apiv1.ResourceList) apiv1.ResourceList {
	allocatable := apiv1.ResourceList{}
	for key, value := range capacity {
		quantity := *value.Copy()
		if reservedQuantity, found := reserved[key]; found {
			quantity.Sub(reservedQuantity)
		}
		allocatable[key] = quantity
	}
	return allocatable
}

func (t *templateBuilder) buildNodeFromTemplate(mig *Mig, template *gce.InstanceTemplate, cpu int64, mem int64) (*apiv1.Node, error) {

	if template.Properties == nil {
		return nil, fmt.Errorf("instance template %s has no properties", template.Name)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-template-%d", template.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	capacity, err := t.buildCapacity(template.Properties.MachineType, template.Properties.GuestAccelerators, mig.GceRef.Zone, cpu, mem)
	if err != nil {
		return nil, err
	}
	node.Status = apiv1.NodeStatus{
		Capacity: capacity,
	}

	var nodeAllocatable apiv1.ResourceList
	// KubeEnv labels & taints
	if template.Properties.Metadata == nil {
		return nil, fmt.Errorf("instance template %s has no metadata", template.Name)
	}
	for _, item := range template.Properties.Metadata.Items {
		if item.Key == "kube-env" {
			if item.Value == nil {
				return nil, fmt.Errorf("no kube-env content in metadata")
			}
			// Extract labels
			kubeEnvLabels, err := extractLabelsFromKubeEnv(*item.Value)
			if err != nil {
				return nil, err
			}
			node.Labels = cloudprovider.JoinStringMaps(node.Labels, kubeEnvLabels)
			// Extract taints
			kubeEnvTaints, err := extractTaintsFromKubeEnv(*item.Value)
			if err != nil {
				return nil, err
			}
			node.Spec.Taints = append(node.Spec.Taints, kubeEnvTaints...)

			if allocatable, err := t.buildAllocatableFromKubeEnv(node.Status.Capacity, *item.Value); err == nil {
				nodeAllocatable = allocatable
			}
		}
	}
	if nodeAllocatable == nil {
		glog.Warningf("could not extract kube-reserved from kubeEnv for mig %q, setting allocatable to capacity.", mig.Name)
		node.Status.Allocatable = node.Status.Capacity
	} else {
		node.Status.Allocatable = nodeAllocatable
	}
	// GenericLabels
	labels, err := buildGenericLabels(mig.GceRef, template.Properties.MachineType, nodeName)
	if err != nil {
		return nil, err
	}
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, labels)

	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func (t *templateBuilder) buildNodeFromAutoprovisioningSpec(mig *Mig, cpu int64, mem int64) (*apiv1.Node, error) {
	if mig.spec == nil {
		return nil, fmt.Errorf("no spec in mig %s", mig.Name)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-autoprovisioned-template-%d", mig.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	capacity, err := t.buildCapacity(mig.spec.machineType, nil, mig.GceRef.Zone, cpu, mem)
	if err != nil {
		return nil, err
	}

	if gpuRequest, found := mig.spec.extraResources[gpu.ResourceNvidiaGPU]; found {
		capacity[gpu.ResourceNvidiaGPU] = gpuRequest.DeepCopy()
	}

	node.Status = apiv1.NodeStatus{
		Capacity:    capacity,
		Allocatable: t.buildAllocatableFromCapacity(capacity),
	}

	labels, err := buildLabelsForAutoprovisionedMig(mig, nodeName)
	if err != nil {
		return nil, err
	}
	node.Labels = labels

	node.Spec.Taints = mig.spec.taints

	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildLabelsForAutoprovisionedMig(mig *Mig, nodeName string) (map[string]string, error) {
	// GenericLabels
	labels, err := buildGenericLabels(mig.GceRef, mig.spec.machineType, nodeName)
	if err != nil {
		return nil, err
	}
	if mig.spec.labels != nil {
		for k, v := range mig.spec.labels {
			if existingValue, found := labels[k]; found {
				if v != existingValue {
					return map[string]string{}, fmt.Errorf("conflict in labels requested: %s=%s  present: %s=%s",
						k, v, k, existingValue)
				}
			} else {
				labels[k] = v
			}
		}
	}
	return labels, nil
}

func buildGenericLabels(ref GceRef, machineType string, nodeName string) (map[string]string, error) {
	result := make(map[string]string)

	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[kubeletapis.LabelInstanceType] = machineType
	ix := strings.LastIndex(ref.Zone, "-")
	if ix == -1 {
		return nil, fmt.Errorf("unexpected zone: %s", ref.Zone)
	}
	result[kubeletapis.LabelZoneRegion] = ref.Zone[:ix]
	result[kubeletapis.LabelZoneFailureDomain] = ref.Zone
	result[kubeletapis.LabelHostname] = nodeName
	return result, nil
}

func parseKubeReserved(kubeReserved string) (apiv1.ResourceList, error) {
	resourcesMap, err := parseKeyValueListToMap(kubeReserved)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kube-reserved from kube-env: %q", err)
	}
	reservedResources := apiv1.ResourceList{}
	for name, quantity := range resourcesMap {
		switch apiv1.ResourceName(name) {
		case apiv1.ResourceCPU, apiv1.ResourceMemory, apiv1.ResourceEphemeralStorage:
			if q, err := resource.ParseQuantity(quantity); err == nil && q.Sign() >= 0 {
				reservedResources[apiv1.ResourceName(name)] = q
			}
		default:
			glog.Warningf("ignoring resource from kube-reserved: %q", name)
		}
	}
	return reservedResources, nil
}

func extractLabelsFromKubeEnv(kubeEnv string) (map[string]string, error) {
	// In v1.10+, labels are only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	labels, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "node_labels")
	if err != nil {
		glog.Errorf("node_labels not found via AUTOSCALER_ENV_VARS due to error, will try NODE_LABELS: %v", err)
		labels, err = extractFromKubeEnv(kubeEnv, "NODE_LABELS")
		if err != nil {
			return nil, err
		}
	}
	return parseKeyValueListToMap(labels)
}

func extractTaintsFromKubeEnv(kubeEnv string) ([]apiv1.Taint, error) {
	// In v1.10+, taints are only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	taints, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "node_taints")
	if err != nil {
		glog.Errorf("node_taints not found via AUTOSCALER_ENV_VARS due to error, will try NODE_TAINTS: %v", err)
		taints, err = extractFromKubeEnv(kubeEnv, "NODE_TAINTS")
		if err != nil {
			return nil, err
		}
	}
	taintMap, err := parseKeyValueListToMap(taints)
	if err != nil {
		return nil, err
	}
	return buildTaints(taintMap)
}

func extractKubeReservedFromKubeEnv(kubeEnv string) (string, error) {
	// In v1.10+, kube-reserved is only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	kubeReserved, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "kube_reserved")
	if err != nil {
		glog.Errorf("kube_reserved not found via AUTOSCALER_ENV_VARS due to error, will try kube-reserved in KUBELET_TEST_ARGS: %v", err)
		kubeletArgs, err := extractFromKubeEnv(kubeEnv, "KUBELET_TEST_ARGS")
		if err != nil {
			return "", err
		}
		resourcesRegexp := regexp.MustCompile(`--kube-reserved=([^ ]+)`)

		matches := resourcesRegexp.FindStringSubmatch(kubeletArgs)
		if len(matches) > 1 {
			return matches[1], nil
		}
		return "", fmt.Errorf("kube-reserved not in kubelet args in kube-env: %q", kubeletArgs)
	}
	return kubeReserved, nil
}

func extractAutoscalerVarFromKubeEnv(kubeEnv, name string) (string, error) {
	const autoscalerVars = "AUTOSCALER_ENV_VARS"
	autoscalerVals, err := extractFromKubeEnv(kubeEnv, autoscalerVars)
	if err != nil {
		return "", err
	}
	for _, val := range strings.Split(autoscalerVals, ";") {
		val = strings.Trim(val, " ")
		items := strings.SplitN(val, "=", 2)
		if len(items) != 2 {
			return "", fmt.Errorf("malformed autoscaler var: %s", val)
		}
		if strings.Trim(items[0], " ") == name {
			return strings.Trim(items[1], " \"'"), nil
		}
	}
	return "", fmt.Errorf("var %s not found in %s: %v", name, autoscalerVars, autoscalerVals)
}

func extractFromKubeEnv(kubeEnv, resource string) (string, error) {
	kubeEnvMap := make(map[string]string)
	err := yaml.Unmarshal([]byte(kubeEnv), &kubeEnvMap)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling kubeEnv: %v", err)
	}
	return kubeEnvMap[resource], nil
}

func parseKeyValueListToMap(kvList string) (map[string]string, error) {
	result := make(map[string]string)
	if len(kvList) == 0 {
		return result, nil
	}
	for _, keyValue := range strings.Split(kvList, ",") {
		kvItems := strings.SplitN(keyValue, "=", 2)
		if len(kvItems) != 2 {
			return nil, fmt.Errorf("error while parsing key-value list, val: %s", keyValue)
		}
		result[kvItems[0]] = kvItems[1]
	}
	return result, nil
}

func buildTaints(kubeEnvTaints map[string]string) ([]apiv1.Taint, error) {
	taints := make([]apiv1.Taint, 0)
	for key, value := range kubeEnvTaints {
		values := strings.SplitN(value, ":", 2)
		if len(values) != 2 {
			return nil, fmt.Errorf("error while parsing node taint value and effect: %s", value)
		}
		taints = append(taints, apiv1.Taint{
			Key:    key,
			Value:  values[0],
			Effect: apiv1.TaintEffect(values[1]),
		})
	}
	return taints, nil
}

type allocatableBracket struct {
	threshold            int64
	marginalReservedRate float64
}

func memoryReservedMB(memoryCapacityMB int64) int64 {
	if memoryCapacityMB <= 1*mbPerGB {
		// do not set any memory reserved for nodes with less than 1 Gb of capacity
		return 0
	}
	return calculateReserved(memoryCapacityMB, []allocatableBracket{
		{
			threshold:            0,
			marginalReservedRate: 0.25,
		},
		{
			threshold:            4 * mbPerGB,
			marginalReservedRate: 0.2,
		},
		{
			threshold:            8 * mbPerGB,
			marginalReservedRate: 0.1,
		},
		{
			threshold:            16 * mbPerGB,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            128 * mbPerGB,
			marginalReservedRate: 0.02,
		},
	})
}

func cpuReservedMillicores(cpuCapacityMillicores int64) int64 {
	return calculateReserved(cpuCapacityMillicores, []allocatableBracket{
		{
			threshold:            0,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            1 * millicoresPerCore,
			marginalReservedRate: 0.01,
		},
		{
			threshold:            2 * millicoresPerCore,
			marginalReservedRate: 0.005,
		},
		{
			threshold:            4 * millicoresPerCore,
			marginalReservedRate: 0.0025,
		},
	})
}

// calculateReserved calculates reserved using capacity and a series of
// brackets as follows:  the marginalReservedRate applies to all capacity
// greater than the bracket, but less than the next bracket.  For example, if
// the first bracket is threshold: 0, rate:0.1, and the second bracket has
// threshold: 100, rate: 0.4, a capacity of 100 results in a reserved of
// 100*0.1 = 10, but a capacity of 200 results in a reserved of
// 10 + (200-100)*.4 = 50.  Using brackets with marginal rates ensures that as
// capacity increases, reserved always increases, and never decreases.
func calculateReserved(capacity int64, brackets []allocatableBracket) int64 {
	var reserved float64
	for i, bracket := range brackets {
		c := capacity
		if i < len(brackets)-1 && brackets[i+1].threshold < capacity {
			c = brackets[i+1].threshold
		}
		additionalReserved := float64(c-bracket.threshold) * bracket.marginalReservedRate
		if additionalReserved > 0 {
			reserved += additionalReserved
		}
	}
	return int64(reserved)
}
