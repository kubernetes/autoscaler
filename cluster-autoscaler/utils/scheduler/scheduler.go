/*
Copyright 2017 The Kubernetes Authors.

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

package scheduler

import (
	"fmt"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	scheduler_config "k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_scheme "k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	scheduler_validation "k8s.io/kubernetes/pkg/scheduler/apis/config/validation"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	schedulerConfigDecodeErr   = "couldn't decode scheduler config"
	schedulerConfigLoadErr     = "couldn't load scheduler config"
	schedulerConfigTypeCastErr = "couldn't assert type as KubeSchedulerConfiguration"
	schedulerConfigInvalidErr  = "invalid KubeSchedulerConfiguration"
)

// CreateNodeNameToInfoMap obtains a list of pods and pivots that list into a map where the keys are node names
// and the values are the aggregated information for that node. Pods waiting lower priority pods preemption
// (pod.Status.NominatedNodeName is set) are also added to list of pods for a node.
func CreateNodeNameToInfoMap(pods []*apiv1.Pod, nodes []*apiv1.Node) map[string]*framework.NodeInfo {
	nodeNameToNodeInfo := make(map[string]*framework.NodeInfo)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			nodeName = pod.Status.NominatedNodeName
		}
		if _, ok := nodeNameToNodeInfo[nodeName]; !ok {
			nodeNameToNodeInfo[nodeName] = framework.NewNodeInfo(nil, nil)
		}
		nodeNameToNodeInfo[nodeName].AddPod(&framework.PodInfo{Pod: pod})
	}

	for _, node := range nodes {
		if _, ok := nodeNameToNodeInfo[node.Name]; !ok {
			nodeNameToNodeInfo[node.Name] = framework.NewNodeInfo(nil, nil)
		}
		nodeNameToNodeInfo[node.Name].SetNode(node)
	}

	// Some pods may be out of sync with node lists. Removing incomplete node infos.
	keysToRemove := make([]string, 0)
	for key, nodeInfo := range nodeNameToNodeInfo {
		if nodeInfo.Node() == nil {
			keysToRemove = append(keysToRemove, key)
		}
	}
	for _, key := range keysToRemove {
		delete(nodeNameToNodeInfo, key)
	}

	return nodeNameToNodeInfo
}

func isHugePageResourceName(name apiv1.ResourceName) bool {
	return strings.HasPrefix(string(name), apiv1.ResourceHugePagesPrefix)
}

// ResourceToResourceList returns a resource list of the resource.
func ResourceToResourceList(r *schedulerframework.Resource) apiv1.ResourceList {
	result := apiv1.ResourceList{
		apiv1.ResourceCPU:              *resource.NewMilliQuantity(r.MilliCPU, resource.DecimalSI),
		apiv1.ResourceMemory:           *resource.NewQuantity(r.Memory, resource.BinarySI),
		apiv1.ResourcePods:             *resource.NewQuantity(int64(r.AllowedPodNumber), resource.BinarySI),
		apiv1.ResourceEphemeralStorage: *resource.NewQuantity(r.EphemeralStorage, resource.BinarySI),
	}
	for rName, rQuant := range r.ScalarResources {
		if isHugePageResourceName(rName) {
			result[rName] = *resource.NewQuantity(rQuant, resource.BinarySI)
		} else {
			result[rName] = *resource.NewQuantity(rQuant, resource.DecimalSI)
		}
	}
	return result
}

// ConfigFromPath loads scheduler config from a path.
// TODO(vadasambar): replace code to parse scheduler config with upstream function
// once https://github.com/kubernetes/kubernetes/pull/119057 is merged
func ConfigFromPath(path string) (*scheduler_config.KubeSchedulerConfiguration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", schedulerConfigLoadErr, err)
	}

	obj, gvk, err := scheduler_scheme.Codecs.UniversalDecoder().Decode(data, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", schedulerConfigDecodeErr, err)
	}

	cfgObj, ok := obj.(*scheduler_config.KubeSchedulerConfiguration)
	if !ok {
		return nil, fmt.Errorf("%s, gvk: %s", schedulerConfigTypeCastErr, gvk)
	}

	// this needs to be set explicitly because config's api version is empty after decoding
	// check kubernetes/cmd/kube-scheduler/app/options/configfile.go for more info
	cfgObj.TypeMeta.APIVersion = gvk.GroupVersion().String()

	if err := scheduler_validation.ValidateKubeSchedulerConfiguration(cfgObj); err != nil {
		return nil, fmt.Errorf("%s: %v", schedulerConfigInvalidErr, err)
	}

	return cfgObj, nil
}

// GetBypassedSchedulersMap returns a map of scheduler names that should be bypassed as keys, and values are set to true
// Also sets "" (empty string) to true if default scheduler is bypassed
func GetBypassedSchedulersMap(bypassedSchedulers []string) map[string]bool {
	bypassedSchedulersMap := make(map[string]bool, len(bypassedSchedulers))
	for _, scheduler := range bypassedSchedulers {
		bypassedSchedulersMap[scheduler] = true
	}
	if canBypass := bypassedSchedulersMap[apiv1.DefaultSchedulerName]; canBypass {
		bypassedSchedulersMap[""] = true
	}
	return bypassedSchedulersMap
}
