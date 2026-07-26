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

package routines

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	scopeutil "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/scope"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

var (
	benchmarkContainerStateMap       model.ContainerNameToAggregateStateMap
	benchmarkScopedContainerStateMap map[string]model.ContainerNameToAggregateStateMap
)

func TestGetContainerNameToAggregateStateMapByScopeValue(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.DaemonSetScope, true)
	cluster := model.NewClusterState(time.Minute)
	selector, err := labels.Parse("app=daemon")
	assert.NoError(t, err)

	vpaObj := test.VerticalPodAutoscaler().
		WithName("scoped-vpa").
		WithNamespace("default").
		WithContainer("agent").
		WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
			Kind:       "DaemonSet",
			Name:       "agent-ds",
			APIVersion: "apps/v1",
		}).
		Get()
	vpaObj.Spec.Scope = "node.kubernetes.io/instance-type"
	assert.NoError(t, cluster.AddOrUpdateVpa(vpaObj, selector))

	scopeLabelKey := scopeutil.AggregationLabelKey(string(vpaObj.Spec.Scope))

	podA := model.PodID{Namespace: "default", PodName: "pod-a"}
	cluster.AddOrUpdatePod(podA, labels.Set{
		"app":         "daemon",
		scopeLabelKey: "c3.large",
	}, corev1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(model.ContainerID{PodID: podA, ContainerName: "agent"}, model.Resources{}))

	podB := model.PodID{Namespace: "default", PodName: "pod-b"}
	cluster.AddOrUpdatePod(podB, labels.Set{
		"app":         "daemon",
		scopeLabelKey: "m5.large",
	}, corev1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(model.ContainerID{PodID: podB, ContainerName: "agent"}, model.Resources{}))

	podC := model.PodID{Namespace: "default", PodName: "pod-c"}
	cluster.AddOrUpdatePod(podC, labels.Set{
		"app":         "daemon",
		scopeLabelKey: "",
	}, corev1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(model.ContainerID{PodID: podC, ContainerName: "agent"}, model.Resources{}))

	key := model.VpaID{Namespace: "default", VpaName: "scoped-vpa"}
	vpa := cluster.VPAs()[key]
	assert.NotNil(t, vpa)

	grouped := GetContainerNameToAggregateStateMapByScopeValue(vpa)
	assert.Contains(t, grouped, "c3.large")
	assert.Contains(t, grouped, "m5.large")
	assert.Contains(t, grouped, "")
	assert.Contains(t, grouped["c3.large"], "agent")
	assert.Contains(t, grouped["m5.large"], "agent")
	assert.Contains(t, grouped[""], "agent")

	// Ensure container policy filter still works for grouped output.
	off := vpa_types.ContainerScalingModeOff
	vpa.ResourcePolicy = &vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
			ContainerName: "agent",
			Mode:          &off,
		}},
	}
	grouped = GetContainerNameToAggregateStateMapByScopeValue(vpa)
	assert.Empty(t, grouped["c3.large"])
	assert.Empty(t, grouped["m5.large"])
	assert.Empty(t, grouped[""])
}

func BenchmarkDaemonSetAggregationNoScope1000Pods(b *testing.B) {
	vpa := setupDaemonSetVpaForBenchmark(b, 1000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkContainerStateMap = GetContainerNameToAggregateStateMap(vpa)
	}
}

func BenchmarkDaemonSetAggregationScope1000Groups(b *testing.B) {
	vpa := setupDaemonSetVpaForBenchmark(b, 1000, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkScopedContainerStateMap = GetContainerNameToAggregateStateMapByScopeValue(vpa)
	}
}

func BenchmarkDaemonSetAggregationNoScope5000Pods(b *testing.B) {
	vpa := setupDaemonSetVpaForBenchmark(b, 5000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkContainerStateMap = GetContainerNameToAggregateStateMap(vpa)
	}
}

func BenchmarkDaemonSetAggregationScope5000Groups(b *testing.B) {
	vpa := setupDaemonSetVpaForBenchmark(b, 5000, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkScopedContainerStateMap = GetContainerNameToAggregateStateMapByScopeValue(vpa)
	}
}

func setupDaemonSetVpaForBenchmark(b *testing.B, podCount int, scoped bool) *model.Vpa {
	b.Helper()
	if scoped {
		featuregatetesting.SetFeatureGateDuringTest(b, features.MutableFeatureGate, features.DaemonSetScope, true)
	}
	cluster := model.NewClusterState(time.Minute)
	selector, err := labels.Parse("app=daemon")
	if err != nil {
		b.Fatalf("failed to parse selector: %v", err)
	}

	vpaObj := test.VerticalPodAutoscaler().
		WithName("benchmark-vpa").
		WithNamespace("default").
		WithContainer("agent").
		WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
			Kind:       "DaemonSet",
			Name:       "agent-ds",
			APIVersion: "apps/v1",
		}).
		Get()
	if scoped {
		vpaObj.Spec.Scope = "node.kubernetes.io/instance-type"
	}
	if err := cluster.AddOrUpdateVpa(vpaObj, selector); err != nil {
		b.Fatalf("failed to add vpa: %v", err)
	}

	scopeLabelKey := scopeutil.AggregationLabelKey(string(vpaObj.Spec.Scope))
	for i := range podCount {
		podID := model.PodID{Namespace: "default", PodName: fmt.Sprintf("pod-%d", i)}
		podLabels := labels.Set{"app": "daemon"}
		if scoped {
			podLabels[scopeLabelKey] = fmt.Sprintf("group-%04d", i)
		}
		cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
		if err := cluster.AddOrUpdateContainer(model.ContainerID{PodID: podID, ContainerName: "agent"}, model.Resources{}); err != nil {
			b.Fatalf("failed to add container for %s: %v", podID.PodName, err)
		}
	}

	key := model.VpaID{Namespace: "default", VpaName: "benchmark-vpa"}
	vpa := cluster.VPAs()[key]
	if vpa == nil {
		b.Fatalf("vpa %v was not found", key)
	}
	return vpa
}
