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
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// CheckpointWriter persistently stores aggregated historical usage of containers
// controlled by VPA objects. This state can be restored to initialize the model after restart.
type CheckpointWriter interface {
	// StoreCheckpoints writes at least minCheckpoints if there are more checkpoints to write.
	// Checkpoints are written until ctx permits or all checkpoints are written.
	StoreCheckpoints(ctx context.Context, minCheckpoints int, concurrentWorkers int)
}

type checkpointWriter struct {
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	cluster             model.ClusterState
}

// NewCheckpointWriter returns new instance of a CheckpointWriter
func NewCheckpointWriter(cluster model.ClusterState, vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter) CheckpointWriter {
	return &checkpointWriter{
		vpaCheckpointClient: vpaCheckpointClient,
		cluster:             cluster,
	}
}

func isFetchingHistory(vpa *model.Vpa) bool {
	condition, found := vpa.Conditions[vpa_types.FetchingHistory]
	if !found {
		return false
	}
	return condition.Status == v1.ConditionTrue
}

func getVpasToCheckpoint(clusterVpas map[model.VpaID]*model.Vpa) []*model.Vpa {
	vpas := make([]*model.Vpa, 0, len(clusterVpas))
	for _, vpa := range clusterVpas {
		if isFetchingHistory(vpa) {
			klog.V(3).InfoS("VPA is loading history, skipping checkpoints", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName))
			continue
		}
		vpas = append(vpas, vpa)
	}
	sort.Slice(vpas, func(i, j int) bool {
		return vpas[i].CheckpointWritten.Before(vpas[j].CheckpointWritten)
	})
	return vpas
}

func processCheckpointUpdateForVPA(vpa *model.Vpa, writer *checkpointWriter) int {
	checkpointsWritten := 0
	now := time.Now()
	aggregateContainerStateMap := buildAggregateContainerStateMap(vpa, writer.cluster, now)
	for container, aggregatedContainerState := range aggregateContainerStateMap {
		containerCheckpoint, err := aggregatedContainerState.SaveToCheckpoint()
		if err != nil {
			klog.ErrorS(err, "Cannot serialize checkpoint", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName), "container", container)
			continue
		}
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
			klog.ErrorS(err, "Cannot save checkpoint for VPA", "vpa", klog.KRef(vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName), "container", vpaCheckpoint.Spec.ContainerName)
		} else {
			klog.V(3).InfoS("Saved checkpoint for VPA", "vpa", klog.KRef(vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName), "container", vpaCheckpoint.Spec.ContainerName)
			vpa.CheckpointWritten = now
		}
		checkpointsWritten++
	}
	return checkpointsWritten
}

func (writer *checkpointWriter) StoreCheckpoints(ctx context.Context, minCheckpoints int, concurrentWorkers int) {
	vpas := getVpasToCheckpoint(writer.cluster.VPAs())

	// Create a channel to send VPA updates to workers
	vpaCheckpointUpdates := make(chan *model.Vpa, len(vpas))

	// Create a channel to receive the number of checkpoints written
	checkpointCounterChannel := make(chan int, len(vpas))
	defer close(checkpointCounterChannel)

	// Create a separate context for the workers. We don't simply pass the outside context,
	// but want to only cancel the workerCtx if minCheckpoints has been reached already.
	workerCtx, cancelWorkers := context.WithCancel(context.Background())

	go func() {
		for updatedCheckpointsCounter := range checkpointCounterChannel {
			minCheckpoints -= updatedCheckpointsCounter
			select {
			case <-ctx.Done():
				if minCheckpoints <= 0 {
					klog.V(0).InfoS("Failed to store all checkpoints within the configured `checkpoints-timeout`", "err", ctx.Err())
					cancelWorkers()
					return
				}
			default:
			}
		}
	}()

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup
	// Start workers
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vpaToCheckpoint := range vpaCheckpointUpdates {
				select {
				case <-workerCtx.Done():
					return
				default:
				}
				checkpointCounterChannel <- processCheckpointUpdateForVPA(vpaToCheckpoint, writer)
			}
		}()
	}

	// Send VPA Checkpoint updates to the workers
	for _, vpa := range vpas {
		vpaCheckpointUpdates <- vpa
	}

	// Close the channel to signal workers to stop after draining the channel
	close(vpaCheckpointUpdates)

	// Wait for all workers to finish
	wg.Wait()
}

// Build the AggregateContainerState for the purpose of the checkpoint. This is an aggregation of state of all
// containers that belong to pods matched by the VPA.
// Note however that we exclude the most recent memory peak for each container (see below).
func buildAggregateContainerStateMap(vpa *model.Vpa, cluster model.ClusterState, now time.Time) map[string]*model.AggregateContainerState {
	aggregateContainerStateMap := vpa.AggregateStateByContainerName()
	// Note: the memory peak from the current (ongoing) aggregation interval is not included in the
	// checkpoint to avoid having multiple peaks in the same interval after the state is restored from
	// the checkpoint. Therefore we are extracting the current peak from all containers.
	// TODO: Avoid the nested loop over all containers for each VPA.
	for _, pod := range cluster.Pods() {
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
		a.AggregateMemoryPeaks.SubtractSample(model.BytesFromMemoryAmount(container.GetMaxMemoryPeak()), 1.0, container.WindowEnd)
	}
}
