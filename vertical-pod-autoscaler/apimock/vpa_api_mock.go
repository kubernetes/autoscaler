/*
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

// Package apimock contains temporary definitions of Vertical Pod Autoscaler related objects - to be replaced with real implementation
// Definitions based on VPA design doc : https://github.com/kubernetes/community/pull/338
package apimock

import (
	"math/rand"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// VerticalPodAutoscaler Represents Vertical Pod Autoscaler configuration - to be replaced by real implementation
type VerticalPodAutoscaler struct {
	// Specification
	Spec Spec
	// Current state of VPA
	Status Status
}

// Spec holds Vertical Pod Autoscaler configuration
type Spec struct {
	// Defines which pods this Autoscaler should watch and update
	Target Target
	// Policy for pod updates
	UpdatePolicy UpdatePolicy
	// Policy for container resources updates
	ResourcesPolicy ResourcesPolicy
}

// Status holds current Vertical Pod Autoscaler state
type Status struct {
	// Recommended resources allocation
	Recommendation *Recommendation
}

// Target Specifies pods to be managed by Vertical Pod Autoscaler
type Target struct {
	// label query over pods. The string will be in the same format as the query-param syntax.
	// More info about label selectors: http://kubernetes.io/docs/user-guide/labels#label-selectors
	// +optional
	Selector string `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`
}

// UpdatePolicy defines policy for Pod updates
type UpdatePolicy struct {
	// Mode for update policy
	Mode Mode
}

// Mode of update policy
type Mode struct {
}

// ResourcesPolicy represents Resources allocation policy
type ResourcesPolicy struct {
	// Resources allocation policy for containers
	Containers []ContainerPolicy
}

// Recommendation represents resources allocation recommended by Vertical Pod Autoscaler
type Recommendation struct {
	// Recommended resources allocation for containers
	Containers []ContainerRecommendation
}

// ContainerPolicy hold resources allocation policy for single container
type ContainerPolicy struct {
	// Name of the container
	Name string
	// Memory allocation policy
	ResourcePolicy map[apiv1.ResourceName]Policy
}

// Policy holds resource allocation policy
type Policy struct {
	// Minimal resource quantity
	Min resource.Quantity
	// Maximal resource quantity
	Max resource.Quantity
}

// ContainerRecommendation holds resource allocation recommendation for container
type ContainerRecommendation struct {
	// Name of the container
	Name string
	// Resources allocation recommended
	Resources map[apiv1.ResourceName]resource.Quantity
}

// VerticalPodAutoscalerLister provides list of all configured Vertical Pod Autoscalers
type VerticalPodAutoscalerLister interface {
	// List returns all configured Vertical Pod Autoscalers
	List() (ret []*VerticalPodAutoscaler, err error)
}

type vpaLister struct{}

// List Mock implementation of Vertical Pod Autoscaler Lister - to be replaced with real implementation
func (lister *vpaLister) List() (ret []*VerticalPodAutoscaler, err error) {
	return []*VerticalPodAutoscaler{fake()}, nil
}

// NewVpaLister returns mock VerticalPodAutoscalerLister - to be replaced with real implementation
func NewVpaLister(_ interface{}) VerticalPodAutoscalerLister {
	return &vpaLister{}
}

// RecommenderAPI defines api returning Vertical Pod Autoscaler recommendations for pods
type RecommenderAPI interface {
	// GetRecommendation returns Vertical Pod Autoscaler recommendation for given pod
	GetRecommendation(spec *apiv1.PodSpec) (*Recommendation, error)
}

type recommenderAPIImpl struct {
}

// NewRecommenderAPI return mock RecommenderAPI - to be replaced with real implementation
func NewRecommenderAPI() RecommenderAPI {
	return &recommenderAPIImpl{}
}

// GetRecommendation Returns random recommendation of resources increase / decrease by 0 - 100 %
// To be replaced with real implementation of recommender request
func (rec *recommenderAPIImpl) GetRecommendation(spec *apiv1.PodSpec) (*Recommendation, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]ContainerRecommendation, len(spec.Containers))
	for i, podContainer := range spec.Containers {

		// scale memory and cpu by random factor [0, 2]

		diffFactor := 2 * rand.Float64()
		memory := podContainer.Resources.Requests.Memory().DeepCopy()
		memory.Set(int64(float64(memory.Value()) * diffFactor))

		diffFactor = 2 * rand.Float64()
		cpu := podContainer.Resources.Requests.Cpu().DeepCopy()
		cpu.Set(int64(float64(cpu.Value()) * diffFactor))

		result[i] = ContainerRecommendation{
			Name: podContainer.Name,
			Resources: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceCPU: cpu, apiv1.ResourceMemory: memory}}
	}
	return &Recommendation{Containers: result}, nil
}

func fake() *VerticalPodAutoscaler {
	minCpu, _ := resource.ParseQuantity("1")
	maxCpu, _ := resource.ParseQuantity("4")
	minMem, _ := resource.ParseQuantity("10M")
	maxMem, _ := resource.ParseQuantity("5G")
	return &VerticalPodAutoscaler{
		Spec: Spec{
			Target:       Target{Selector: "app = redis"},
			UpdatePolicy: UpdatePolicy{Mode: Mode{}},
			ResourcesPolicy: ResourcesPolicy{Containers: []ContainerPolicy{{
				Name: "slave",
				ResourcePolicy: map[apiv1.ResourceName]Policy{
					apiv1.ResourceCPU:    {minCpu, maxCpu},
					apiv1.ResourceMemory: {minMem, maxMem}},
			}}}},
	}
}
