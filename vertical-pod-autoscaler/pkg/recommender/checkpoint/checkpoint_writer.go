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

package checkpoint

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// CheckpointVersion should be incremented every time the format or semantics of the checkpoint
// is changed in a way that is not compatible with the previous version.
const CheckpointVersion string = "1"

// CheckpointWriter persistently stores aggregated historical usage of containers
// controlled by VPA objects. This state can be restored to initialize the model after restart.
type CheckpointWriter interface {
	StoreCheckpoints(now time.Time)
}

type checkpointWriter struct {
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	cluster             *model.ClusterState
}

// NewCheckpointWriter returns new instance of a CheckpointWriter
func NewCheckpointWriter(cluster *model.ClusterState, vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter) CheckpointWriter {
	return &checkpointWriter{
		vpaCheckpointClient: vpaCheckpointClient,
		cluster:             cluster,
	}
}

func (writer *checkpointWriter) StoreCheckpoints(now time.Time) {
	for _, vpa := range writer.cluster.Vpas {
		aggregateContainerStateMap := buildAggregateContainerStateMap(vpa, writer.cluster, now)
		for container, aggregatedContainerState := range aggregateContainerStateMap {
			containerCheckpoint, err := aggregatedContainerState.SaveToCheckpoint()
			if err != nil {
				glog.Errorf("Cannot serialize checkpoint for vpa %v container %v. Reason: %+v", vpa.ID.VpaName, container, err)
				continue
			}
			containerCheckpoint.Version = CheckpointVersion
			checkpointName := fmt.Sprintf("%s-%s", vpa.ID.VpaName, container)
			vpaCheckpoint := vpa_types.VerticalPodAutoscalerCheckpoint{
				ObjectMeta: metav1.ObjectMeta{Name: checkpointName},
				Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
					ContainerName: container,
					VPAObjectName: vpa.ID.VpaName,
				},
				Status: *containerCheckpoint,
			}
			err = api_util.CreateOrUpdateVpaCheckpoint(writer.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(vpa.ID.Namespace), &vpaCheckpoint)
			if err != nil {
				glog.Errorf("Cannot save VPA %s/%s checkpoint for %s. Reason: %+v",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName, err)
			} else {
				glog.V(3).Infof("Saved VPA %s/%s checkpoint for %s",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName)
			}
		}
	}
}

// Build the AggregateContainerState for the purpose of the checkpoint. This is an aggregation of state of all
// containers that belong to pods matched by the VPA.
// Note however that we exclude the most recent memory peak for each container (see below).
func buildAggregateContainerStateMap(vpa *model.Vpa, cluster *model.ClusterState, now time.Time) map[string]*model.AggregateContainerState {
	aggregateContainerStateMap := vpa.AggregateStateByContainerName()
	// Note: the memory peak from the current (ongoing) aggregation interval is not included in the
	// checkpoint to avoid having multiple peaks in the same interval after the state is restored from
	// the checkpoint. Therefore we are extracting the current peak from all containers.
	// TODO: Avoid the nested loop over all containers for each VPA.
	for _, pod := range cluster.Pods {
		for containerName, container := range pod.Containers {
			aggregateKey := cluster.MakeAggregateStateKey(pod, containerName)
			if vpa.UsesAggregation(aggregateKey) {
				if aggregateContainerState, exists := aggregateContainerStateMap[containerName]; exists {
					subtractCurrentContainerMemoryPeak(aggregateContainerState, container, now)
				}
			}
		}
	}
	return aggregateContainerStateMap
}

func subtractCurrentContainerMemoryPeak(a *model.AggregateContainerState, container *model.ContainerState, now time.Time) {
	if now.Before(container.WindowEnd) {
		a.AggregateMemoryPeaks.SubtractSample(model.BytesFromMemoryAmount(container.MemoryPeak), 1.0, container.WindowEnd)
	}
}
