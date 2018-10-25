package scaling

import (
	"testing"

	"time"

	scalingpolicy "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/apis/scalingpolicy/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/debug"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors/static"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/clock"
)

func TestComputeResources(t *testing.T) {
	grid := []struct {
		Name     string
		Inputs   map[string]float64
		Policy   *scalingpolicy.ScalingPolicySpec
		Expected *v1.PodSpec
	}{
		{
			Name:     "Empty policy",
			Policy:   &scalingpolicy.ScalingPolicySpec{},
			Expected: nil,
		},
		{
			Name: "Trivial no-scale policy",
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Limits: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceCPU,
									Function: scalingpolicy.ResourceScalingFunction{
										Base: resource.MustParse("2000m"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2000m"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Simple linear scale policy",
			Inputs: map[string]float64{
				"pods": 10,
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("100Mi"),
										Slope: resource.MustParse("10Mi"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Policy allows negative coefficients",
			Inputs: map[string]float64{
				"pods": 5,
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("100Mi"),
										Slope: resource.MustParse("-10Mi"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("50Mi"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Mixed units work",
			Inputs: map[string]float64{
				"pods": 5,
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("10Mi"),
										Slope: resource.MustParse("1M"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: *resource.NewScaledQuantity((1024*1024*10)+(1000*1000*5), 0),
							},
						},
					},
				},
			},
		},
		/*{
			Name: "Multiple rules on same resource",
			Inputs: map[string]float64{
				"nodes": 2,
				"pods":  4,
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("100Mi"),
										Slope: resource.MustParse("10Mi"),
									},
								},
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "nodes",
										Slope: resource.MustParse("20Mi"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("180Mi"),
							},
						},
					},
				},
			},
		},
		*/
		{
			Name: "Multiple resources",
			Inputs: map[string]float64{
				"nodes": 2,
				"pods":  4,
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("200Mi"),
										Slope: resource.MustParse("7Mi"),
									},
								},
								{
									Resource: v1.ResourceCPU,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "nodes",
										Base:  resource.MustParse("100m"),
										Slope: resource.MustParse("23m"),
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("228Mi"),
								v1.ResourceCPU:    resource.MustParse("146m"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Segments",
			Inputs: map[string]float64{
				"nodes": 19, // rounds up to 20
				"pods":  26, // rounds up to 30
			},
			Policy: &scalingpolicy.ScalingPolicySpec{
				Containers: []scalingpolicy.ContainerScalingRule{
					{
						Name: "container1",
						Resources: scalingpolicy.ResourceRequirements{
							Requests: []scalingpolicy.ResourceScalingRule{
								{
									Resource: v1.ResourceMemory,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "pods",
										Base:  resource.MustParse("200Mi"),
										Slope: resource.MustParse("7Mi"),
										Segments: []scalingpolicy.ResourceScalingSegment{
											{At: 6, Every: 2},
											{At: 20, Every: 5},
										},
									},
								},
								{
									Resource: v1.ResourceCPU,
									Function: scalingpolicy.ResourceScalingFunction{
										Input: "nodes",
										Base:  resource.MustParse("100m"),
										Slope: resource.MustParse("23m"),
										Segments: []scalingpolicy.ResourceScalingSegment{
											{At: 10, Every: 5},
											{At: 20, Every: 10},
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("410Mi"), // 200Mi + (30 * 7Mi)
								v1.ResourceCPU:    resource.MustParse("560m"),  // 100m + (20 * 23m)
							},
						},
					},
				},
			},
		},
	}

	for _, g := range grid {
		baseTime := time.Now()
		clock := clock.NewFakeClock(baseTime)
		factors := static.NewStaticFactors(clock, g.Inputs)
		snaphot, err := factors.Snapshot()
		if err != nil {
			t.Errorf("snapshot failed: %v", err)
		}
		policy := &scalingpolicy.ScalingPolicy{
			Spec: *g.Policy,
		}
		actual := &v1.PodSpec{}
		for _, c := range g.Policy.Containers {
			actual.Containers = append(actual.Containers, v1.Container{Name: c.Name})
		}
		evaluator := NewScalingPolicyEvaluator(clock, policy)
		evaluator.AddObservation(snaphot)
		changes, err := evaluator.ComputeResources("", actual)
		if err != nil {
			t.Errorf("unexpected error from test\npolicy=%v\nerror=%v", debug.Print(g.Policy), err)
			continue
		}
		if !equality.Semantic.DeepEqual(changes, g.Expected) {
			t.Errorf("test failure\nname=%s\npolicy=%v\n  actual=%v\nexpected=%v", g.Name, debug.Print(g.Policy), debug.Print(changes), debug.Print(g.Expected))
			continue
		}
	}
}

func TestSegments(t *testing.T) {
	fn := &scalingpolicy.ResourceScalingFunction{
		Segments: []scalingpolicy.ResourceScalingSegment{
			{At: 10, Every: 5},
			{At: 20, Every: 10},
		},
	}

	grid := []struct {
		Input    float64
		Expected float64
	}{
		{Input: 1, Expected: 1},
		{Input: 2, Expected: 2},
		{Input: 9, Expected: 9},
		{Input: 10, Expected: 10},
		{Input: 11, Expected: 15},
		{Input: 12, Expected: 15},
		{Input: 15, Expected: 15},
		{Input: 16, Expected: 20},
		{Input: 19, Expected: 20},
		{Input: 20, Expected: 20},
		{Input: 21, Expected: 30},
		{Input: 30, Expected: 30},
		{Input: 31, Expected: 40},
	}

	for _, g := range grid {
		actual := roundInput(fn, g.Input)
		if actual != g.Expected {
			t.Errorf("test failure\fn=%s\n  actual=%v\nexpected=%v", debug.Print(fn), actual, g.Expected)
			continue
		}
	}
}
