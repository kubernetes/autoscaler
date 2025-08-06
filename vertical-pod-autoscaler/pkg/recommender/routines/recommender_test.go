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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type mockPodResourceRecommender struct{}

func (m *mockPodResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) logic.RecommendedPodResources {
	return logic.RecommendedPodResources{}
}

// TestProcessUpdateVPAsConcurrency tests processVPAUpdate for race conditions when run concurrently
func TestProcessUpdateVPAsConcurrency(t *testing.T) {
	updateWorkerCount := 10

	vpaCount := 1000
	vpas := make(map[model.VpaID]*model.Vpa, vpaCount)
	apiObjectVPAs := make([]*v1.VerticalPodAutoscaler, vpaCount)
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

	fakeClient := vpa_fake.NewSimpleClientset(fakedClient...).AutoscalingV1()
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
	vpaUpdates := make(chan *v1.VerticalPodAutoscaler, len(apiObjectVPAs))

	var counter int64

	// Start workers
	for range updateWorkerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for observedVpa := range vpaUpdates {
				key := model.VpaID{
					Namespace: observedVpa.Namespace,
					VpaName:   observedVpa.Name,
				}

				vpa, found := r.clusterState.VPAs()[key]
				if !found {
					return
				}

				atomic.AddInt64(&counter, 1)

				processVPAUpdate(r, vpa, observedVpa)
				cnt.Add(vpa)
			}
		}()
	}

	// Send VPA updates to the workers
	for _, observedVpa := range apiObjectVPAs {
		vpaUpdates <- observedVpa
	}

	close(vpaUpdates)
	wg.Wait()

	assert.Equal(t, int64(vpaCount), atomic.LoadInt64(&counter), "Not all VPAs were processed")
}
