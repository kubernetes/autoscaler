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

package model

import (
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

// TestConcurrentUpdateRecommendationAndHasRecommendation tests concurrent access
// to recommendation field - one goroutine updating, another checking if exists.
// This reproduces part of the race condition from the production crash.
func TestConcurrentUpdateRecommendationAndHasRecommendation(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	// Add a container state for the recommendation to update
	containerName := "test-container"
	vpa.aggregateContainerStates[aggregateStateKey{
		namespace:     "test",
		containerName: containerName,
	}] = &AggregateContainerState{}

	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Continuously update recommendation
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			rec := test.Recommendation().
				WithContainer(containerName).
				WithTarget("100m", "100Mi").
				Get()
			vpa.UpdateRecommendation(rec)
		}
	}()

	// Goroutine 2: Continuously check if recommendation exists
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = vpa.HasRecommendation()
		}
	}()

	wg.Wait()
}

// TestConcurrentUpdateConditionsAndAsStatus tests concurrent updates to conditions
// while reading the full status. This simulates the actual production scenario.
func TestConcurrentUpdateConditionsAndAsStatus(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())
	vpa.Recommendation = test.Recommendation().WithContainer("test").WithTarget("100m", "100Mi").Get()

	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Continuously update conditions
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			podsMatched := i%2 == 0
			vpa.UpdateConditions(podsMatched)
		}
	}()

	// Goroutine 2: Continuously read status
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = vpa.AsStatus()
		}
	}()

	wg.Wait()
}

// TestConcurrentSetConditionAndConditionActive tests setting conditions
// while checking if they're active.
func TestConcurrentSetConditionAndConditionActive(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: Set NoPodsMatched condition
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			vpa.SetCondition(vpa_types.NoPodsMatched, true, "test", "test message")
		}
	}()

	// Goroutine 2: Set RecommendationProvided condition
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			vpa.SetCondition(vpa_types.RecommendationProvided, i%2 == 0, "test", "test message")
		}
	}()

	// Goroutine 3: Check if conditions are active
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = vpa.ConditionActive(vpa_types.NoPodsMatched)
			_ = vpa.ConditionActive(vpa_types.RecommendationProvided)
		}
	}()

	wg.Wait()
}

// TestConcurrentDeleteConditionAndConditionActive tests deleting conditions
// while checking if they're active.
func TestConcurrentDeleteConditionAndConditionActive(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: Set and delete NoPodsMatched condition
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if i%2 == 0 {
				vpa.SetCondition(vpa_types.NoPodsMatched, true, "test", "test")
			} else {
				vpa.DeleteCondition(vpa_types.NoPodsMatched)
			}
		}
	}()

	// Goroutine 2: Set and delete RecommendationProvided condition
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if i%3 == 0 {
				vpa.SetCondition(vpa_types.RecommendationProvided, true, "test", "test")
			} else {
				vpa.DeleteCondition(vpa_types.RecommendationProvided)
			}
		}
	}()

	// Goroutine 3: Check if conditions are active
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = vpa.ConditionActive(vpa_types.NoPodsMatched)
			_ = vpa.ConditionActive(vpa_types.RecommendationProvided)
		}
	}()

	wg.Wait()
}

// TestConcurrentRecommendationAndConditionUpdates simulates the exact production
// scenario where multiple worker goroutines update the same VPA's recommendations
// and conditions while other code reads them.
func TestConcurrentRecommendationAndConditionUpdates(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	// Add container states
	containerNames := []string{"container1", "container2", "container3"}
	for _, name := range containerNames {
		vpa.aggregateContainerStates[aggregateStateKey{
			namespace:     "test",
			containerName: name,
		}] = &AggregateContainerState{}
	}

	workerCount := 10
	iterations := 100
	var wg sync.WaitGroup

	// Simulate multiple workers updating VPA (like in UpdateVPAs)
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Update recommendation
				cpu := resource.NewMilliQuantity(int64(100+workerID*10+i), resource.DecimalSI)
				mem := resource.NewQuantity(int64(100+workerID*10+i)*1024*1024, resource.BinarySI)

				rec := &vpa_types.RecommendedPodResources{
					ContainerRecommendations: []vpa_types.RecommendedContainerResources{
						{
							ContainerName: containerNames[i%len(containerNames)],
							Target: corev1.ResourceList{
								corev1.ResourceCPU:    *cpu,
								corev1.ResourceMemory: *mem,
							},
							UncappedTarget: corev1.ResourceList{
								corev1.ResourceCPU:    *cpu,
								corev1.ResourceMemory: *mem,
							},
						},
					},
				}
				vpa.UpdateRecommendation(rec)

				// Update conditions
				vpa.UpdateConditions(i%2 == 0)

				// Read status (like when writing back to API)
				_ = vpa.AsStatus()

				// Check conditions
				_ = vpa.ConditionActive(vpa_types.RecommendationProvided)
				_ = vpa.HasRecommendation()
			}
		}(w)
	}

	// Additional goroutines that just read (simulating other parts of the system)
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations*workerCount; i++ {
				_ = vpa.AsStatus()
				_ = vpa.HasRecommendation()
				_ = vpa.HasMatchedPods()
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentUpdateRecommendationAndAsStatus specifically tests the race
// between updating recommendation and reading status, which was happening
// in the production crash.
func TestConcurrentUpdateRecommendationAndAsStatus(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	containerName := "test-container"
	vpa.aggregateContainerStates[aggregateStateKey{
		namespace:     "test",
		containerName: containerName,
	}] = &AggregateContainerState{}

	iterations := 1000
	workerCount := 8 // Similar to production worker count
	var wg sync.WaitGroup

	// Multiple workers updating recommendation and reading status
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				rec := test.Recommendation().
					WithContainer(containerName).
					WithTarget("100m", "100Mi").
					Get()

				vpa.UpdateRecommendation(rec)
				vpa.UpdateConditions(true)

				// This is what happens in processVPAUpdate - we read the status
				// to write it back via UpdateVpaStatusIfNeeded
				_ = vpa.AsStatus()
			}
		}(w)
	}

	wg.Wait()
}

// TestConcurrentConditionMapAccess tests direct concurrent access patterns
// to the conditions map through the thread-safe methods.
func TestConcurrentConditionMapAccess(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	iterations := 1000
	var wg sync.WaitGroup

	conditionTypes := []vpa_types.VerticalPodAutoscalerConditionType{
		vpa_types.RecommendationProvided,
		vpa_types.NoPodsMatched,
		vpa_types.LowConfidence,
		vpa_types.FetchingHistory,
	}

	// Multiple goroutines setting different conditions
	for _, condType := range conditionTypes {
		wg.Add(1)
		go func(ct vpa_types.VerticalPodAutoscalerConditionType) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				vpa.SetCondition(ct, i%2 == 0, "test", "test message")
			}
		}(condType)
	}

	// Goroutines reading conditions
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				for _, ct := range conditionTypes {
					_ = vpa.ConditionActive(ct)
				}
			}
		}()
	}

	// Goroutines calling AsStatus (which reads all conditions)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = vpa.AsStatus()
			}
		}()
	}

	wg.Wait()
}

// TestHighContentionScenario creates a very high contention scenario with
// many goroutines all accessing the VPA simultaneously.
func TestHighContentionScenario(t *testing.T) {
	vpa := NewVpa(VpaID{Namespace: "test", VpaName: "test-vpa"}, labels.Nothing(), time.Now())

	// Setup multiple containers
	for i := 0; i < 5; i++ {
		containerName := "container" + string(rune('0'+i))
		vpa.aggregateContainerStates[aggregateStateKey{
			namespace:     "test",
			containerName: containerName,
		}] = &AggregateContainerState{
			lastRecommendation: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		}
	}

	goroutineCount := 50
	iterations := 100
	var wg sync.WaitGroup

	for g := 0; g < goroutineCount; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				switch id % 7 {
				case 0:
					// Update recommendation
					rec := test.Recommendation().
						WithContainer("container0").
						WithTarget("100m", "100Mi").
						Get()
					vpa.UpdateRecommendation(rec)
				case 1:
					// Update conditions
					vpa.UpdateConditions(i%2 == 0)
				case 2:
					// Read status
					_ = vpa.AsStatus()
				case 3:
					// Check if has recommendation
					_ = vpa.HasRecommendation()
				case 4:
					// Check if has matched pods
					_ = vpa.HasMatchedPods()
				case 5:
					// Set specific condition
					vpa.SetCondition(vpa_types.RecommendationProvided, true, "test", "test")
				case 6:
					// Check condition active
					_ = vpa.ConditionActive(vpa_types.RecommendationProvided)
				default:
					// linter requires default case
				}
			}
		}(g)
	}

	wg.Wait()
}
