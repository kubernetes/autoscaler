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
	"net/url"
	"path"
	"regexp"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/golang/glog"
)

// builds templates for gce cloud provider
type templateBuilder struct {
	service   *gce.Service
	zone      string
	projectId string
}

func (t *templateBuilder) getMigTemplate(mig *Mig) (*gce.InstanceTemplate, error) {
	igm, err := t.service.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return nil, err
	}
	templateUrl, err := url.Parse(igm.InstanceTemplate)
	if err != nil {
		return nil, err
	}
	_, templateName := path.Split(templateUrl.EscapedPath())
	instanceTemplate, err := t.service.InstanceTemplates.Get(mig.Project, templateName).Do()
	if err != nil {
		return nil, err
	}
	return instanceTemplate, nil
}

func (t *templateBuilder) getCpuAndMemoryForMachineType(machineType string) (cpu int64, mem int64, err error) {
	if strings.HasPrefix(machineType, "custom-") {
		return parseCustomMachineType(machineType)
	}
	machine, geterr := t.service.MachineTypes.Get(t.projectId, t.zone, machineType).Do()
	if geterr != nil {
		return 0, 0, geterr
	}
	return machine.GuestCpus, machine.MemoryMb * 1024 * 1024, nil
}

func (t *templateBuilder) buildCapacity(machineType string) (apiv1.ResourceList, error) {
	capacity := apiv1.ResourceList{}
	// TODO: get a real value.
	// TODO: handle GPU
	capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)

	cpu, mem, err := t.getCpuAndMemoryForMachineType(machineType)
	if err != nil {
		return apiv1.ResourceList{}, err
	}
	capacity[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
	capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	return capacity, nil
}

// buildAllocatable builds node allocatable based on capacity of the node and
// value of kubeEnv.
// KubeEnv is a multi line string containing entries in the form of
// <RESOURCE_NAME>:<string>. One of the resources it contains is a list of
// kubelet arguments from which we can extract the resources reseved by
// the kubelet for its operation. Allocated resources are capacity - reserved.
// If we fail to extract the reserved resources from kubeEnv (e.g it is in a
// wrong format or does not contain kubelet arguments), we return an error.
func (t *templateBuilder) buildAllocatable(capacity apiv1.ResourceList, kubeEnv string) (apiv1.ResourceList, error) {
	kubeReserved, err := extractKubeReservedFromKubeEnv(kubeEnv)
	if err != nil {
		return nil, err
	}
	reserved, err := parseKubeReserved(kubeReserved)
	if err != nil {
		return nil, err
	}
	allocatable := apiv1.ResourceList{}
	for key, value := range capacity {
		quantity := *value.Copy()
		if reservedQuantity, found := reserved[key]; found {
			quantity.Sub(reservedQuantity)
		}
		allocatable[key] = quantity
	}
	return allocatable, nil
}

func (t *templateBuilder) buildNodeFromTemplate(mig *Mig, template *gce.InstanceTemplate) (*apiv1.Node, error) {

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
	capacity, err := t.buildCapacity(template.Properties.MachineType)
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

			if allocatable, err := t.buildAllocatable(node.Status.Capacity, *item.Value); err == nil {
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

func (t *templateBuilder) buildNodeFromAutoprovisioningSpec(mig *Mig) (*apiv1.Node, error) {

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
	capacity, err := t.buildCapacity(mig.spec.machineType)
	if err != nil {
		return nil, err
	}
	node.Status = apiv1.NodeStatus{
		Capacity: capacity,
	}
	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	labels, err := buildLablesForAutoprovisionedMig(mig, nodeName)
	if err != nil {
		return nil, err
	}
	node.Labels = labels
	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildLablesForAutoprovisionedMig(mig *Mig, nodeName string) (map[string]string, error) {
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

func parseCustomMachineType(machineType string) (cpu, mem int64, err error) {
	// example custom-2-2816
	var count int
	count, err = fmt.Sscanf(machineType, "custom-%d-%d", &cpu, &mem)
	if err != nil {
		return
	}
	if count != 2 {
		return 0, 0, fmt.Errorf("failed to parse all params in %s", machineType)
	}
	// Mb to bytes
	mem = mem * 1024 * 1024
	return
}

func parseKubeReserved(kubeReserved string) (apiv1.ResourceList, error) {
	resourcesMap, err := parseKeyValueListToMap([]string{kubeReserved})
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
	labels, err := extractFromKubeEnv(kubeEnv, "NODE_LABELS")
	if err != nil {
		return nil, err
	}
	return parseKeyValueListToMap(labels)
}

func extractTaintsFromKubeEnv(kubeEnv string) ([]apiv1.Taint, error) {
	taints, err := extractFromKubeEnv(kubeEnv, "NODE_TAINTS")
	if err != nil {
		return nil, err
	}
	taintMap, err := parseKeyValueListToMap(taints)
	if err != nil {
		return nil, err
	}
	return buildTaints(taintMap)
}

func extractKubeReservedFromKubeEnv(kubeEnv string) (string, error) {
	kubeletArgs, err := extractFromKubeEnv(kubeEnv, "KUBELET_TEST_ARGS")
	if err != nil {
		return "", err
	}
	resourcesRegexp := regexp.MustCompile(`--kube-reserved=([^ ]+)`)

	for _, value := range kubeletArgs {
		matches := resourcesRegexp.FindStringSubmatch(value)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("kube-reserved not in kubelet args in kube-env: %q", strings.Join(kubeletArgs, " "))
}

func extractFromKubeEnv(kubeEnv, resource string) ([]string, error) {
	result := make([]string, 0)

	for line, env := range strings.Split(kubeEnv, "\n") {
		env = strings.Trim(env, " ")
		if len(env) == 0 {
			continue
		}
		items := strings.SplitN(env, ":", 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("wrong content in kube-env at line: %d", line)
		}
		key := strings.Trim(items[0], " ")
		value := strings.Trim(items[1], " \"'")
		if key == resource {
			result = append(result, value)
		}
	}
	return result, nil
}

func parseKeyValueListToMap(values []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, value := range values {
		for _, val := range strings.Split(value, ",") {
			valItems := strings.SplitN(val, "=", 2)
			if len(valItems) != 2 {
				return nil, fmt.Errorf("error while parsing kube env value: %s", val)
			}
			result[valItems[0]] = valItems[1]
		}
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
