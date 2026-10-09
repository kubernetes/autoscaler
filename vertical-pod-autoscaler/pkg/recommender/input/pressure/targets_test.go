/*
Copyright The Kubernetes Authors.

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

package pressure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

type fakeClusterState struct {
	model.ClusterState
	vpas map[model.VpaID]*model.Vpa
	pods map[model.PodID]*model.PodState
}

func (f *fakeClusterState) VPAs() map[model.VpaID]*model.Vpa { return f.vpas }

func (f *fakeClusterState) Pods() map[model.PodID]*model.PodState { return f.pods }

// newFakeClusterState tracks the VPAs and one Running PodState per Pod.
func newFakeClusterState(vpas []*model.Vpa, podIDs ...model.PodID) *fakeClusterState {
	cs := &fakeClusterState{vpas: map[model.VpaID]*model.Vpa{}, pods: map[model.PodID]*model.PodState{}}
	for _, vpa := range vpas {
		cs.vpas[vpa.ID] = vpa
	}
	for _, id := range podIDs {
		cs.pods[id] = &model.PodState{ID: id, Phase: corev1.PodRunning}
	}
	return cs
}

func TestBuildTargetsRequiresManagedMemory(t *testing.T) {
	enabled := vpa_types.PressureDetectionEnabled
	off := vpa_types.ContainerScalingModeOff
	cpuOnly := []corev1.ResourceName{corev1.ResourceCPU}
	withMemory := []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}
	testCases := []struct {
		name   string
		policy vpa_types.ContainerResourcePolicy
		want   int
	}{
		{name: "defaults manage memory", policy: vpa_types.ContainerResourcePolicy{}, want: 1},
		{name: "memory in controlled resources", policy: vpa_types.ContainerResourcePolicy{ControlledResources: &withMemory}, want: 1},
		{name: "scaling mode Off", policy: vpa_types.ContainerResourcePolicy{Mode: &off}, want: 0},
		{name: "memory not controlled", policy: vpa_types.ContainerResourcePolicy{ControlledResources: &cpuOnly}, want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy := tc.policy
			policy.ContainerName = "c"
			policy.PressureDetection = &enabled
			vpa := &model.Vpa{ID: model.VpaID{Namespace: "ns", VpaName: "v"}, PodSelector: labels.Everything(),
				ResourcePolicy: &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{policy}}}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "p", UID: "uid"},
				Spec: corev1.PodSpec{NodeName: "n", Containers: []corev1.Container{{Name: "c",
					Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("512Mi")}}}}},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			}
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
			assert.NoError(t, indexer.Add(pod))
			cs := newFakeClusterState([]*model.Vpa{vpa}, model.PodID{Namespace: "ns", PodName: "p"})

			assert.Len(t, BuildTargets(cs, listersv1.NewPodLister(indexer)), tc.want)
		})
	}
}

func TestBuildTargetsEligibility(t *testing.T) {
	enabled := vpa_types.PressureDetectionEnabled
	appX := labels.SelectorFromSet(labels.Set{"app": "x"})
	testCases := []struct {
		name      string
		vpaNS     string
		selector  labels.Selector
		podLabels map[string]string
		phase     corev1.PodPhase
		limit     string
		want      int
	}{
		{name: "selector matches", vpaNS: "ns", selector: appX, podLabels: map[string]string{"app": "x"}, phase: corev1.PodRunning, limit: "512Mi", want: 1},
		{name: "selector does not match", vpaNS: "ns", selector: appX, podLabels: map[string]string{"app": "y"}, phase: corev1.PodRunning, limit: "512Mi"},
		{name: "VPA in another namespace", vpaNS: "other", selector: labels.Everything(), phase: corev1.PodRunning, limit: "512Mi"},
		{name: "Pod is not running", vpaNS: "ns", selector: labels.Everything(), phase: corev1.PodPending, limit: "512Mi"},
		{name: "container has no memory limit", vpaNS: "ns", selector: labels.Everything(), phase: corev1.PodRunning},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vpa := &model.Vpa{ID: model.VpaID{Namespace: tc.vpaNS, VpaName: "v"}, PodSelector: tc.selector,
				ResourcePolicy: &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{ContainerName: "c", PressureDetection: &enabled}}}}
			container := corev1.Container{Name: "c"}
			if tc.limit != "" {
				container.Resources.Limits = corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(tc.limit)}
			}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "p", UID: "uid", Labels: tc.podLabels},
				Spec:       corev1.PodSpec{NodeName: "n", Containers: []corev1.Container{container}},
			}
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
			assert.NoError(t, indexer.Add(pod))
			podID := model.PodID{Namespace: "ns", PodName: "p"}
			cs := newFakeClusterState([]*model.Vpa{vpa}, podID)
			cs.pods[podID].Phase = tc.phase

			assert.Len(t, BuildTargets(cs, listersv1.NewPodLister(indexer)), tc.want)
		})
	}
}
