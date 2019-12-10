/*
Copyright 2018 The Kubernetes Authors.

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

package oom

import (
	"strings"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// OomInfo contains data of the OOM event occurrence
type OomInfo struct {
	Timestamp                 time.Time
	Memory                    resource.Quantity
	Namespace, Pod, Container string
}

// Observer can observe pod resource update and collect OOM events.
type Observer struct {
	ObservedOomsChannel chan OomInfo
}

// NewObserver returns new instance of the Observer.
func NewObserver() Observer {
	return Observer{
		ObservedOomsChannel: make(chan OomInfo),
	}
}

func parseEvictionEvent(event *apiv1.Event) []OomInfo {
	if event.Reason != "Evicted" ||
		event.InvolvedObject.Kind != "Pod" {
		return []OomInfo{}
	}
	extractArray := func(annotationsKey string) []string {
		str, found := event.Annotations[annotationsKey]
		if !found {
			return []string{}
		}
		return strings.Split(str, ",")
	}
	offendingContainers := extractArray("offending_containers")
	offendingContainersUsage := extractArray("offending_containers_usage")
	starvedResource := extractArray("starved_resource")
	if len(offendingContainers) != len(offendingContainersUsage) ||
		len(offendingContainers) != len(starvedResource) {
		return []OomInfo{}
	}

	result := make([]OomInfo, 0, len(offendingContainers))

	for i, container := range offendingContainers {
		if starvedResource[i] != "memory" {
			continue
		}
		memory, err := resource.ParseQuantity(offendingContainersUsage[i])
		if err != nil {
			glog.Errorf("Cannot parse resource quantity in eviction event %v. Error: %v", offendingContainersUsage[i], err)
			continue
		}
		oomInfo := OomInfo{
			Timestamp: event.CreationTimestamp.Time.UTC(),
			Memory:    memory,
			Namespace: event.InvolvedObject.Namespace,
			Pod:       event.InvolvedObject.Name,
			Container: container,
		}
		result = append(result, oomInfo)
	}
	return result
}

// OnEvent inspects k8s eviction events and translates them to OomInfo.
func (o *Observer) OnEvent(event *apiv1.Event) {
	glog.V(1).Infof("OOM Obeserver processing event: %+v", event)
	for _, oomInfo := range parseEvictionEvent(event) {
		info := oomInfo
		go func() {
			o.ObservedOomsChannel <- info
		}()
	}
}

// OnAdd is Noop
func (*Observer) OnAdd(obj interface{}) {}

// OnUpdate inspects if the update contains oom information and
// passess it to the ObservedOomsChannel
func (o *Observer) OnUpdate(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*apiv1.Pod)
	if oldPod == nil || !ok {
		glog.Errorf("OOM observer received invaild oldObj: %v", oldObj)
	}

	newPod, ok := newObj.(*apiv1.Pod)

	if newPod == nil || !ok {
		glog.Errorf("OOM observer received invaild newObj: %v", newObj)
	}

	oldContainersStatusMap := make(map[string]apiv1.ContainerStatus)
	for _, containerStatus := range oldPod.Status.ContainerStatuses {
		oldContainersStatusMap[containerStatus.Name] = containerStatus
	}

	oldContainersSpecMap := make(map[string]apiv1.Container)
	for _, containerSpec := range oldPod.Spec.Containers {
		oldContainersSpecMap[containerSpec.Name] = containerSpec
	}

	for _, containerStatus := range newPod.Status.ContainerStatuses {
		prevContainerStatus, ok := oldContainersStatusMap[containerStatus.Name]
		if ok && containerStatus.RestartCount > prevContainerStatus.RestartCount &&
			containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" {
			if spec, ok := oldContainersSpecMap[containerStatus.Name]; ok {
				oomInfo := OomInfo{
					Namespace: newPod.ObjectMeta.Namespace,
					Pod:       newPod.ObjectMeta.Name,
					Container: containerStatus.Name,
					Memory:    spec.Resources.Requests[apiv1.ResourceMemory],
					Timestamp: containerStatus.LastTerminationState.Terminated.FinishedAt.Time.UTC(),
				}
				go func() {
					o.ObservedOomsChannel <- oomInfo
				}()
			} else {
				glog.Errorf("Shouldn't happen. Missing spec for container %v", containerStatus.Name)
			}

		}
	}

}

// OnDelete is Noop
func (*Observer) OnDelete(obj interface{}) {}
