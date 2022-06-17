package routines

import (
	"testing"
	"time"

	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"

	"github.com/stretchr/testify/assert"
)

func TestSortedRecommendation(t *testing.T) {
	cases := []struct {
		name         string
		resources    logic.RecommendedPodResources
		expectedLast []string
	}{
		{
			name: "All recommendations sorted",
			resources: logic.RecommendedPodResources{
				"a-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"b-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"c-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"d-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
			},
			expectedLast: []string{
				"a-container",
				"b-container",
				"c-container",
				"d-container",
			},
		},
		{
			name: "All recommendations unsorted",
			resources: logic.RecommendedPodResources{
				"b-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"a-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"d-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
				"c-container": logic.RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1000)}},
			},
			expectedLast: []string{
				"a-container",
				"b-container",
				"c-container",
				"d-container",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			namespace := "test-namespace"
			vpa := model.NewVpa(model.VpaID{Namespace: namespace, VpaName: "my-vpa"}, labels.Nothing(), time.Unix(0, 0))
			vpa.UpdateRecommendation(getCappedRecommendation(vpa.ID, tc.resources, nil))
			// Check that the slice is in the correct order.
			for i := range vpa.Recommendation.ContainerRecommendations {
				assert.Equal(t, tc.expectedLast[i], vpa.Recommendation.ContainerRecommendations[i].ContainerName)
			}
		})
	}
}
