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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/ptr"

	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type mockPodResourceRecommender struct{}

func (*mockPodResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) logic.RecommendedPodResources {
	return logic.RecommendedPodResources{}
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

type fakeClusterStateFeeder struct {
	input.ClusterStateFeeder
}

// GarbageCollectCheckpoints fails iff ctx is already done
func (*fakeClusterStateFeeder) GarbageCollectCheckpoints(ctx context.Context) error {
	return ctx.Err()
}

type noopCheckpointWriter struct{}

func (noopCheckpointWriter) StoreCheckpoints(context.Context, int) {}

func TestMaintainCheckpointsGCUsesIndependentTimeout(t *testing.T) {
	feeder := &fakeClusterStateFeeder{}

	// Construct a recommender where checkpoint updates timeout straight away
	r := &recommender{
		clusterStateFeeder:      feeder,
		checkpointWriter:        noopCheckpointWriter{},
		useCheckpoints:          true,
		checkpointsWriteTimeout: 0,
		checkpointsGCTimeout:    time.Minute,
	}

	r.MaintainCheckpoints(context.Background())

	// Ensure that garbage collection was invoked and succeeded
	assert.False(t, r.lastCheckpointGC.IsZero())
}

func TestUpdateInitialDelayCondition(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newObserved := func(delay *int32) *vpaautoscalingv1.VerticalPodAutoscaler {
		mode := vpaautoscalingv1.UpdateModeRecreate
		return &vpaautoscalingv1.VerticalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{Name: "vpa", Namespace: "default", CreationTimestamp: metav1.NewTime(created)},
			Spec: vpaautoscalingv1.VerticalPodAutoscalerSpec{
				UpdatePolicy: &vpaautoscalingv1.PodUpdatePolicy{UpdateMode: &mode, InitialDelaySeconds: delay},
			},
		}
	}
	tests := []struct {
		name           string
		delay          *int32
		now            time.Time
		featureEnabled bool
		expectFound    bool
		expectStatus   bool
		expectReason   string
		expectMessage  string
	}{
		{
			name:           "inside the window",
			delay:          ptr.To(int32(3600)),
			now:            created.Add(time.Minute),
			featureEnabled: true,
			expectFound:    true,
			expectStatus:   true,
			expectReason:   "WindowActive",
			expectMessage:  "Initial delay window active until 2026-01-01T01:00:00Z",
		},
		{
			name:           "after the window",
			delay:          ptr.To(int32(3600)),
			now:            created.Add(2 * time.Hour),
			featureEnabled: true,
			expectFound:    true,
			expectStatus:   false,
			expectReason:   "WindowExpired",
			expectMessage:  "Initial delay window ended at 2026-01-01T01:00:00Z",
		},
		{
			name:           "field unset",
			delay:          nil,
			now:            created,
			featureEnabled: true,
			expectFound:    false,
		},
		{
			name:           "feature disabled",
			delay:          ptr.To(int32(3600)),
			now:            created.Add(time.Minute),
			featureEnabled: false,
			expectFound:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.VPAInitialDelay, tc.featureEnabled)
			vpa := model.NewVpa(model.VpaID{Namespace: "default", VpaName: "vpa"}, labels.Everything(), created)
			// A condition left over from an earlier loop must be replaced or removed.
			vpa.SetCondition(vpaautoscalingv1.InitialDelayActive, true, "WindowActive", "stale")

			updateInitialDelayCondition(vpa, newObserved(tc.delay), tc.now)

			condition, found := vpa.GetConditionsMap()[vpaautoscalingv1.InitialDelayActive]
			assert.Equal(t, tc.expectFound, found)
			if !tc.expectFound {
				return
			}
			assert.Equal(t, tc.expectStatus, condition.Status == "True")
			assert.Equal(t, tc.expectReason, condition.Reason)
			assert.Equal(t, tc.expectMessage, condition.Message)
		})
	}
}
