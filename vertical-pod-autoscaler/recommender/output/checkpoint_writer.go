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

package output

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

// CheckpointWriter is implements how histogram checkpoints are stored.
type CheckpointWriter interface {
	StoreCheckpoints()
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

func (writer *checkpointWriter) StoreCheckpoints() {
	for _, vpa := range writer.cluster.Vpas {
		aggregateContainerStateMap := model.BuildAggregateContainerStateMap(vpa, model.MergeForRecommendation, time.Now())
		for container, aggregatedContainerState := range aggregateContainerStateMap {
			containerCheckpoint, err := aggregatedContainerState.SaveToCheckpoint()
			if err != nil {
				glog.Errorf("Cannot serialize checkpoint for vpa %v container %v. Reason: %+v", vpa.ID.VpaName, container, err)
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
				glog.Errorf("Cannot save VPA %s/%s checkpoint for %s. Reason: %+v",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName, err)
			} else {
				glog.V(3).Infof("Saved VPA %s/%s checkpoint for %s",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName)
			}
		}
	}
}
