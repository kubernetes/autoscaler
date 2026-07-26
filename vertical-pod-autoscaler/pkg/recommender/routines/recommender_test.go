/*
Copyright 2025 The Kubernetes Authors.

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
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	scopeutil "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/scope"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type mockPodResourceRecommender struct{}

func (*mockPodResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) logic.RecommendedPodResources {
	return logic.RecommendedPodResources{}
}

type countingPodResourceRecommender struct {
	calls atomic.Int64
}

func (c *countingPodResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) logic.RecommendedPodResources {
	c.calls.Add(1)
	return logic.RecommendedPodResources{}
}

func (c *countingPodResourceRecommender) CallCount() int64 {
	return c.calls.Load()
}

func BenchmarkProcessVPAUpdateDaemonSetNoScope1000Pods(b *testing.B) {
	r, vpaModel, observedVpa := setupProcessVPAUpdateBenchmark(b, 1000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processVPAUpdate(r, vpaModel, observedVpa)
	}
}

func BenchmarkProcessVPAUpdateDaemonSetScope1000Groups(b *testing.B) {
	r, vpaModel, observedVpa := setupProcessVPAUpdateBenchmark(b, 1000, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processVPAUpdate(r, vpaModel, observedVpa)
	}
}

func BenchmarkProcessVPAUpdateDaemonSetNoScope5000Pods(b *testing.B) {
	r, vpaModel, observedVpa := setupProcessVPAUpdateBenchmark(b, 5000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processVPAUpdate(r, vpaModel, observedVpa)
	}
}

func BenchmarkProcessVPAUpdateDaemonSetScope5000Groups(b *testing.B) {
	r, vpaModel, observedVpa := setupProcessVPAUpdateBenchmark(b, 5000, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processVPAUpdate(r, vpaModel, observedVpa)
	}
}

func setupProcessVPAUpdateBenchmark(b *testing.B, podCount int, scoped bool) (*recommender, *model.Vpa, *vpaautoscalingv1.VerticalPodAutoscaler) {
	return setupProcessVPAUpdateScenario(b, podCount, scoped, &mockPodResourceRecommender{})
}

func setupProcessVPAUpdateScenario(tb testing.TB, podCount int, scoped bool, podResourceRecommender logic.PodResourceRecommender) (*recommender, *model.Vpa, *vpaautoscalingv1.VerticalPodAutoscaler) {
	tb.Helper()
	if scoped {
		featuregatetesting.SetFeatureGateDuringTest(tb, features.MutableFeatureGate, features.DaemonSetScope, true)
	}
	clusterState := model.NewClusterState(time.Minute)
	selector, err := labels.Parse("app=daemon")
	if err != nil {
		tb.Fatalf("failed to parse selector: %v", err)
	}

	vpaName := "benchmark-process-vpa-update"
	observedVpa := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithNamespace("default").
		WithContainer("agent").
		WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
			Kind:       "DaemonSet",
			Name:       "agent-ds",
			APIVersion: "apps/v1",
		}).
		Get()
	if scoped {
		observedVpa.Spec.Scope = vpaautoscalingv1.VerticalPodAutoscalerScopeType("node.kubernetes.io/instance-type")
	}

	if err := clusterState.AddOrUpdateVpa(observedVpa, selector); err != nil {
		tb.Fatalf("failed to add vpa: %v", err)
	}
	scopeLabelKey := scopeutil.AggregationLabelKey(string(observedVpa.Spec.Scope))
	for i := range podCount {
		podID := model.PodID{Namespace: "default", PodName: fmt.Sprintf("pod-%d", i)}
		podLabels := labels.Set{"app": "daemon"}
		if scoped {
			podLabels[scopeLabelKey] = fmt.Sprintf("group-%04d", i)
		}
		clusterState.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
		if err := clusterState.AddOrUpdateContainer(model.ContainerID{PodID: podID, ContainerName: "agent"}, model.Resources{}); err != nil {
			tb.Fatalf("failed to add container for %s: %v", podID.PodName, err)
		}
	}

	vpaID := model.VpaID{Namespace: "default", VpaName: vpaName}
	vpaModel := clusterState.VPAs()[vpaID]
	if vpaModel == nil {
		tb.Fatalf("vpa %v was not found", vpaID)
	}

	fakeClient := vpa_fake.NewSimpleClientset(observedVpa).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
	r := &recommender{
		clusterState:                clusterState,
		vpaClient:                   fakeClient,
		podResourceRecommender:      podResourceRecommender,
		recommendationPostProcessor: []RecommendationPostProcessor{},
	}
	return r, vpaModel, observedVpa
}

// TestProcessUpdateVPAsConcurrency tests processVPAUpdate for race conditions when run concurrently
func TestProcessUpdateVPAsConcurrency(t *testing.T) {
	updateWorkerCount := 10

	vpaCount := 1000
	vpas := make(map[model.VpaID]*model.Vpa, vpaCount)
	apiObjectVPAs := make([]*vpaautoscalingv1.VerticalPodAutoscaler, vpaCount)
	fakedClient := make([]runtime.Object, vpaCount)

	for i := range vpaCount {
		vpaName := fmt.Sprintf("test-vpa-%d", i)
		vpaID := model.VpaID{
			Namespace: "default",
			VpaName:   vpaName,
		}
		selector, err := labels.Parse("app=test")
		assert.NoError(t, err, "Failed to parse label selector")
		vpas[vpaID] = model.NewVpa(vpaID, selector, time.Now())

		apiObjectVPAs[i] = test.VerticalPodAutoscaler().
			WithName(vpaName).
			WithNamespace("default").
			WithContainer("test-container").
			Get()

		fakedClient[i] = apiObjectVPAs[i]
	}

	fakeClient := vpa_fake.NewSimpleClientset(fakedClient...).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
	r := &recommender{
		clusterState:                model.NewClusterState(time.Minute),
		vpaClient:                   fakeClient,
		podResourceRecommender:      &mockPodResourceRecommender{},
		recommendationPostProcessor: []RecommendationPostProcessor{},
	}

	labelSelector, err := metav1.ParseToLabelSelector("app=test")
	assert.NoError(t, err, "Failed to parse label selector")
	parsedSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	assert.NoError(t, err, "Failed to convert label selector to selector")

	// Inject into clusterState
	for _, vpa := range apiObjectVPAs {
		err := r.clusterState.AddOrUpdateVpa(vpa, parsedSelector)
		assert.NoError(t, err, "Failed to add or update VPA in cluster state")
	}
	r.clusterState.SetObservedVPAs(apiObjectVPAs)

	// Run processVPAUpdate concurrently for all VPAs
	var wg sync.WaitGroup

	cnt := metrics_recommender.NewObjectCounter()
	defer cnt.Observe()

	// Create a channel to send VPA updates to workers
	vpaUpdates := make(chan *vpaautoscalingv1.VerticalPodAutoscaler, len(apiObjectVPAs))

	var counter atomic.Int64

	// Start workers
	for range updateWorkerCount {
		wg.Go(func() {
			for observedVpa := range vpaUpdates {
				key := model.VpaID{
					Namespace: observedVpa.Namespace,
					VpaName:   observedVpa.Name,
				}

				vpa, found := r.clusterState.VPAs()[key]
				if !found {
					return
				}

				counter.Add(1)

				processVPAUpdate(r, vpa, observedVpa)
				cnt.Add(vpa)
			}
		})
	}

	// Send VPA updates to the workers
	for _, observedVpa := range apiObjectVPAs {
		vpaUpdates <- observedVpa
	}

	close(vpaUpdates)
	wg.Wait()

	assert.Equal(t, int64(vpaCount), counter.Load(), "Not all VPAs were processed")
}

// TestConcurrentAccessToSameVPA tests multiple goroutines updating the same VPA's conditions and recommendations
// simultaneously.
//
// NOTE: This test currently exposes additional race conditions beyond the VPA mutex fix:
// - aggregateContainerStates map is accessed without synchronization
// - RecordRecommendation() reads vpa.Recommendation without holding VPA's mutex
// I don't know that anyone is actually experiencing those, just apparently they can happen.
//
// To run this test and see ONLY the VPA condition/recommendation races that were fixed,
// use the vpa_concurrency_test.go tests in the model package instead.
func TestConcurrentAccessToSameVPA(t *testing.T) {
	t.Skip("This test exposes additional race conditions beyond the VPA mutex fix. " +
		"See vpa_concurrency_test.go for tests specific to the conditions/recommendations mutex.")

	// Create a single VPA that will be accessed by multiple goroutines
	vpaName := "shared-vpa"
	vpaID := model.VpaID{
		Namespace: "default",
		VpaName:   vpaName,
	}

	selector, err := labels.Parse("app=test")
	assert.NoError(t, err, "Failed to parse label selector")

	vpa := model.NewVpa(vpaID, selector, time.Now())

	// Add multiple container states to make the VPA more realistic
	containerNames := []string{"container1", "container2", "container3"}
	for _, containerName := range containerNames {
		vpa.UseAggregationIfMatching(
			mockAggregateStateKey{
				namespace:     "default",
				containerName: containerName,
				labels:        "app=test",
			},
			model.NewAggregateContainerState(),
		)
	}

	apiVpa := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithNamespace("default").
		WithContainer("container1").
		Get()

	fakeClient := vpa_fake.NewSimpleClientset(apiVpa).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954

	r := &recommender{
		clusterState:                model.NewClusterState(time.Minute),
		vpaClient:                   fakeClient,
		podResourceRecommender:      &mockPodResourceRecommender{},
		recommendationPostProcessor: []RecommendationPostProcessor{},
	}

	// Setup cluster state
	labelSelector, err := metav1.ParseToLabelSelector("app=test")
	assert.NoError(t, err, "Failed to parse label selector")
	parsedSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	assert.NoError(t, err, "Failed to convert label selector to selector")

	err = r.clusterState.AddOrUpdateVpa(apiVpa, parsedSelector)
	assert.NoError(t, err, "Failed to add or update VPA in cluster state")
	r.clusterState.SetObservedVPAs([]*vpaautoscalingv1.VerticalPodAutoscaler{apiVpa})

	// Now simulate multiple workers ALL processing the SAME VPA concurrently
	// This is the exact scenario that caused the production crash
	workerCount := 10
	iterations := 100
	var wg sync.WaitGroup

	for w := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for range iterations {
				// Each worker processes the same VPA
				processVPAUpdate(r, vpa, apiVpa)
			}
		}(w)
	}

	wg.Wait()
}

// TestConcurrentVPAMethodAccess tests ONLY the mutex-protected VPA methods
// without involving the full processVPAUpdate path which has additional races.
func TestConcurrentVPAMethodAccess(t *testing.T) {
	// Create a single VPA
	vpaName := "test-vpa"
	vpaID := model.VpaID{
		Namespace: "default",
		VpaName:   vpaName,
	}

	selector, err := labels.Parse("app=test")
	assert.NoError(t, err, "Failed to parse label selector")

	vpa := model.NewVpa(vpaID, selector, time.Now())

	// Add container states (this part is not concurrent)
	containerNames := []string{"container1", "container2", "container3"}
	for _, containerName := range containerNames {
		vpa.UseAggregationIfMatching(
			mockAggregateStateKey{
				namespace:     "default",
				containerName: containerName,
				labels:        "app=test",
			},
			model.NewAggregateContainerState(),
		)
	}

	// Now test concurrent access to the mutex-protected methods
	workerCount := 10
	iterations := 100
	var wg sync.WaitGroup

	for w := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range iterations {
				// Create a recommendation
				rec := test.Recommendation().
					WithContainer(containerNames[i%len(containerNames)]).
					WithTarget("100m", "100Mi").
					Get()

				// Test all the mutex-protected operations
				vpa.UpdateRecommendation(rec)
				vpa.UpdateConditions(i%2 == 0)
				_ = vpa.AsStatus()
				_ = vpa.HasRecommendation()
				_ = vpa.HasMatchedPods()
				_ = vpa.ConditionActive(vpaautoscalingv1.RecommendationProvided)
			}
		}(w)
	}

	wg.Wait()
}

func TestProcessVPAUpdateScopedDaemonSetPublishesGlobalRecommendationAsFallback(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.DaemonSetScope, true)
	vpaName := "scoped-vpa"
	vpaID := model.VpaID{Namespace: "default", VpaName: vpaName}
	selector, err := labels.Parse("app=test")
	assert.NoError(t, err)

	vpaModel := model.NewVpa(vpaID, selector, time.Now())
	observedVpa := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithNamespace("default").
		WithContainer("agent").
		WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
			Kind:       "DaemonSet",
			Name:       "agent",
			APIVersion: "apps/v1",
		}).
		Get()
	observedVpa.Spec.Scope = vpaautoscalingv1.VerticalPodAutoscalerScopeType("node.deckhouse.io/group")

	fakeClient := vpa_fake.NewSimpleClientset(observedVpa).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
	r := &recommender{
		clusterState:                model.NewClusterState(time.Minute),
		vpaClient:                   fakeClient,
		podResourceRecommender:      &mockPodResourceRecommender{},
		recommendationPostProcessor: []RecommendationPostProcessor{},
	}

	processVPAUpdate(r, vpaModel, observedVpa)

	updated, err := fakeClient.VerticalPodAutoscalers("default").Get(context.Background(), vpaName, metav1.GetOptions{})
	assert.NoError(t, err)
	// The global recommendation is always published, including for scoped
	// DaemonSets, so it can serve as a fallback when the DaemonSetScope feature
	// gate is disabled (for example during a rollback).
	assert.NotNil(t, updated.Status.Recommendation)
}

func TestProcessVPAUpdateRegularVPAUsesRecommendationOnly(t *testing.T) {
	vpaName := "regular-vpa"
	vpaID := model.VpaID{Namespace: "default", VpaName: vpaName}
	selector, err := labels.Parse("app=test")
	assert.NoError(t, err)

	vpaModel := model.NewVpa(vpaID, selector, time.Now())
	observedVpa := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithNamespace("default").
		WithContainer("app").
		WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
			Kind:       "Deployment",
			Name:       "app",
			APIVersion: "apps/v1",
		}).
		Get()

	fakeClient := vpa_fake.NewSimpleClientset(observedVpa).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
	r := &recommender{
		clusterState:                model.NewClusterState(time.Minute),
		vpaClient:                   fakeClient,
		podResourceRecommender:      &mockPodResourceRecommender{},
		recommendationPostProcessor: []RecommendationPostProcessor{},
	}

	processVPAUpdate(r, vpaModel, observedVpa)

	updated, err := fakeClient.VerticalPodAutoscalers("default").Get(context.Background(), vpaName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, updated.Status.Recommendation)
	assert.Empty(t, updated.Status.RecommendationGroups)
}

func TestProcessVPAUpdateScopedDaemonSetUsesCacheWhenInputsUnchanged(t *testing.T) {
	counter := &countingPodResourceRecommender{}
	r, vpaModel, observedVpa := setupProcessVPAUpdateScenario(t, 2, true, counter)

	processVPAUpdate(r, vpaModel, observedVpa)
	firstRunCalls := counter.CallCount()
	assert.Equal(t, int64(3), firstRunCalls, "expected 1 summary + 2 grouped recommendation calls")

	processVPAUpdate(r, vpaModel, observedVpa)
	assert.Equal(t, firstRunCalls, counter.CallCount(), "second run should hit scoped cache and avoid recomputation")
}

func TestProcessVPAUpdateScopedDaemonSetInvalidatesCacheOnNewSample(t *testing.T) {
	counter := &countingPodResourceRecommender{}
	r, vpaModel, observedVpa := setupProcessVPAUpdateScenario(t, 2, true, counter)

	processVPAUpdate(r, vpaModel, observedVpa)
	firstRunCalls := counter.CallCount()
	assert.Equal(t, int64(3), firstRunCalls, "expected 1 summary + 2 grouped recommendation calls")

	err := r.clusterState.AddSample(&model.ContainerUsageSampleWithKey{
		ContainerUsageSample: model.ContainerUsageSample{
			MeasureStart: time.Now(),
			Resource:     model.ResourceCPU,
			Usage:        model.CPUAmountFromCores(0.2),
		},
		Container: model.ContainerID{
			PodID: model.PodID{
				Namespace: "default",
				PodName:   "pod-0",
			},
			ContainerName: "agent",
		},
	})
	assert.NoError(t, err)

	processVPAUpdate(r, vpaModel, observedVpa)
	assert.Greater(t, counter.CallCount(), firstRunCalls, "cache should be invalidated after new metrics sample")
}

func TestProcessVPAUpdateRegularVPANotAffectedByScopedCache(t *testing.T) {
	counter := &countingPodResourceRecommender{}
	r, vpaModel, observedVpa := setupProcessVPAUpdateScenario(t, 2, false, counter)

	processVPAUpdate(r, vpaModel, observedVpa)
	processVPAUpdate(r, vpaModel, observedVpa)

	// Regular VPA computes summary recommendation each run and does not use scoped cache.
	assert.Equal(t, int64(2), counter.CallCount())
}

// TestUpdateVPAsRaceCondition tests the UpdateVPAs method for race conditions
// by having multiple workers process overlapping sets of VPAs.
func TestUpdateVPAsRaceCondition(t *testing.T) {
	vpaCount := 20
	apiObjectVPAs := make([]*vpaautoscalingv1.VerticalPodAutoscaler, vpaCount)
	fakedClient := make([]runtime.Object, vpaCount)

	for i := range vpaCount {
		vpaName := fmt.Sprintf("test-vpa-%d", i)
		apiObjectVPAs[i] = test.VerticalPodAutoscaler().
			WithName(vpaName).
			WithNamespace("default").
			WithContainer("test-container").
			Get()
		fakedClient[i] = apiObjectVPAs[i]
	}

	fakeClient := vpa_fake.NewSimpleClientset(fakedClient...).AutoscalingV1() //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
	r := &recommender{
		clusterState:                model.NewClusterState(time.Minute),
		vpaClient:                   fakeClient,
		podResourceRecommender:      &mockPodResourceRecommender{},
		recommendationPostProcessor: []RecommendationPostProcessor{},
		updateWorkerCount:           8, // Similar to production
	}

	labelSelector, err := metav1.ParseToLabelSelector("app=test")
	assert.NoError(t, err, "Failed to parse label selector")
	parsedSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	assert.NoError(t, err, "Failed to convert label selector to selector")

	// Setup VPAs in cluster state
	for _, vpa := range apiObjectVPAs {
		err := r.clusterState.AddOrUpdateVpa(vpa, parsedSelector)
		assert.NoError(t, err, "Failed to add or update VPA in cluster state")
	}
	r.clusterState.SetObservedVPAs(apiObjectVPAs)

	// Run UpdateVPAs multiple times concurrently to increase race detection
	iterations := 10
	var wg sync.WaitGroup

	for range iterations {
		wg.Go(func() {
			r.UpdateVPAs()
		})
	}

	wg.Wait()
}

// mockAggregateStateKey is a simple implementation for testing
type mockAggregateStateKey struct {
	namespace     string
	containerName string
	labels        string
}

func (k mockAggregateStateKey) Namespace() string {
	return k.namespace
}

func (k mockAggregateStateKey) ContainerName() string {
	return k.containerName
}

func (k mockAggregateStateKey) Labels() labels.Labels {
	// Should return empty on error
	labels, _ := labels.ConvertSelectorToLabelsMap(k.labels)
	return labels
}
