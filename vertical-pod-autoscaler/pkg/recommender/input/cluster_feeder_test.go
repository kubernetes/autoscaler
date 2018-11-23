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

package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// PodID contains information needed to identify a Pod within a cluster.
type PodID struct {
	// Namespaces where the Pod is defined.
	Namespace string
	// PodName is the name of the pod unique within a namespace.
	PodName string
}

type ContainerID struct {
	PodID
	// ContainerName is the name of the container, unique within a pod.
	ContainerName string
}

type vpaCRD struct {
	Namespace string
	VpaName   string
}

var (
	podID = PodID{Namespace: "namespace-1", PodName: "Pod1"}

	testContainerID = ContainerID{testPodID, "container-1"}	
)
func TestLoadPods(t *testing.T) {

	podPhase  := "running"
	podLabels  := "labelKey"
	pods := make(map[model.PodID]*spec.BasicPodSpec)

	for _, pod := range pods {
		assert.NoError(t, feeder.clusterState.AddOrUpdatePod(pod.ID, pod.PodLabels, pod.Phase))
		for _, container := range pod.Containers {
			assert.NoError(t, feeder.clusterState.AddOrUpdateContainer(container.ID, container.Request))
		}
	}

}

func TestLoadVPAs(t *testing.T) {

	vpaCRD.namespace = "default"
	vpaCRD.VpaName = "vpa-1"
	vpaKeys := make(map[model.VpaID]bool)
	for _, vpaCRD := range vpaCRDs {
		vpaID := model.VpaID{
			Namespace: vpaCRD.Namespace,
			VpaName:   vpaCRD.Name}
		assert.NoError(t, feeder.clusterState.AddOrUpdateVpa(vpaCRD))
		vpaKeys[vpaID] = true
	}
	if _, exists := vpaKeys[vpaID]
		assert.NoError(t, feeder.clusterState.DeleteVpa(vpaID))
	}
}
