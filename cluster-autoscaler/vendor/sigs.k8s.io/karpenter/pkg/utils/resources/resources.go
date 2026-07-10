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

package resources

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resourcehelper "k8s.io/component-helpers/resource"

	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

var Node = v1.ResourceName("nodes")

// RequestsForPods returns the total resources of a variadic list of podspecs.
func RequestsForPods(pods ...*v1.Pod) v1.ResourceList {
	var resources []v1.ResourceList
	for _, pod := range pods {
		resources = append(resources, Ceiling(pod).Requests)
	}
	merged := Merge(resources...)
	merged[v1.ResourcePods] = *resource.NewQuantity(int64(len(pods)), resource.DecimalExponent)
	return merged
}

// LimitsForPods returns the total resources of a variadic list of podspecs
func LimitsForPods(pods ...*v1.Pod) v1.ResourceList {
	var resources []v1.ResourceList
	for _, pod := range pods {
		resources = append(resources, Ceiling(pod).Limits)
	}
	merged := Merge(resources...)
	merged[v1.ResourcePods] = *resource.NewQuantity(int64(len(pods)), resource.DecimalExponent)
	return merged
}

// Merge the resources from the variadic into a single v1.ResourceList
func Merge(resources ...v1.ResourceList) v1.ResourceList {
	if len(resources) == 0 {
		return v1.ResourceList{}
	}
	result := make(v1.ResourceList, len(resources[0]))
	for _, resourceList := range resources {
		for resourceName, quantity := range resourceList {
			current := result[resourceName]
			current.Add(quantity)
			result[resourceName] = current
		}
	}
	return result
}

// MergeInto sums the resources from src into dest, modifying dest. If you need to repeatedly sum
// multiple resource lists, it allocates less to continually sum into an existing list as opposed to
// constructing a new one for each sum like Merge
func MergeInto(dest v1.ResourceList, src v1.ResourceList) v1.ResourceList {
	if dest == nil {
		sz := len(src)
		dest = make(v1.ResourceList, sz)
	}
	for resourceName, quantity := range src {
		current := dest[resourceName]
		current.Add(quantity)
		dest[resourceName] = current
	}
	return dest
}

func Subtract(lhs, rhs v1.ResourceList) v1.ResourceList {
	result := make(v1.ResourceList, len(lhs))
	for k, v := range lhs {
		result[k] = v.DeepCopy()
	}
	for resourceName := range lhs {
		current := lhs[resourceName]
		if rhsValue, ok := rhs[resourceName]; ok {
			current.Sub(rhsValue)
		}
		result[resourceName] = current
	}
	return result
}

// SubtractFrom subtracts the src v1.ResourceList from the dest v1.ResourceList in-place
func SubtractFrom(dest v1.ResourceList, src v1.ResourceList) {
	if dest == nil {
		sz := len(src)
		dest = make(v1.ResourceList, sz)
	}
	for resourceName, quantity := range src {
		current := dest[resourceName]
		current.Sub(quantity)
		dest[resourceName] = current
	}
}

// Ceiling computes the effective resource requirements for a given Pod,
// using the same logic as the scheduler. When InPlacePodVerticalScaling is enabled,
// this returns max(spec.requests, status.allocatedResources) to reflect what the
// kubelet actually holds during active resize transitions.
func Ceiling(pod *v1.Pod) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: resourcehelper.PodRequests(pod, resourcehelper.PodResourcesOptions{UseStatusResources: true}),
		Limits:   resourcehelper.PodLimits(pod, resourcehelper.PodResourcesOptions{}),
	}
}

// RequestsForSpec computes the effective resource requests for a PodSpec using
// standard Kubernetes scheduling semantics (KEP-753 sidecar-aware).
func RequestsForSpec(spec *v1.PodSpec) v1.ResourceList {
	return resourcehelper.PodRequests(&v1.Pod{Spec: *spec}, resourcehelper.PodResourcesOptions{})
}

// MaxResources returns the maximum quantities for a given list of resources.
// Quantity structs in the list returned are only safe to mutate if Quantity.d.Dec == nil
func MaxResources(resources ...v1.ResourceList) v1.ResourceList {
	resourceList := v1.ResourceList{}
	for _, resource := range resources {
		for resourceName, quantity := range resource {
			if value, ok := resourceList[resourceName]; !ok || quantity.Cmp(value) > 0 {
				resourceList[resourceName] = quantity
			}
		}
	}
	return resourceList
}

// MinResources returns the minimum quantities for resources present in all lists (intersection of resources)
// Ex: MinResources({cpu: 1}, {cpu: 2, gpu: 1}) = {cpu: 1}.
// Quantity structs in the list returned are only safe to mutate if Quantity.d.Dec == nil
func MinResources(resources ...v1.ResourceList) v1.ResourceList {
	if len(resources) == 0 {
		return v1.ResourceList{}
	}

	// Start with a copy of the first list's keys, but just copy the Quantity values
	// Safe since we never mutate them
	resourceList := make(v1.ResourceList, len(resources[0]))
	for k, v := range resources[0] {
		resourceList[k] = v
	}

	for _, rl := range resources[1:] {
		for resourceName, quantity := range resourceList {
			if rlQuantity, exists := rl[resourceName]; exists {
				if rlQuantity.Cmp(quantity) < 0 {
					resourceList[resourceName] = rlQuantity
				}
			} else {
				delete(resourceList, resourceName)
			}
		}
	}
	return resourceList
}

// Quantity parses the string value into a *Quantity
func Quantity(value string) *resource.Quantity {
	r := resource.MustParse(value)
	return &r
}

// IsZero implements r.IsZero(). This method is provided to make some code a bit cleaner as the Quantity.IsZero() takes
// a pointer receiver and map index expressions aren't addressable, so it can't be called directly.
func IsZero(r resource.Quantity) bool {
	return r.IsZero()
}

func Cmp(lhs resource.Quantity, rhs resource.Quantity) int {
	return lhs.Cmp(rhs)
}

// Fits returns true if the candidate set of resources is less than or equal to the total set of resources.
func Fits(candidate, total v1.ResourceList) bool {
	// If any of the total resource values are negative then the resource will never fit
	for _, quantity := range total {
		if Cmp(*resource.NewScaledQuantity(0, resource.Kilo), quantity) > 0 {
			return false
		}
	}
	for resourceName, quantity := range candidate {
		if Cmp(quantity, total[resourceName]) > 0 {
			return false
		}
	}
	return true
}

// String returns a string version of the resource list suitable for presenting in a log
func String(list v1.ResourceList) string {
	if len(list) == 0 {
		return "{}"
	}
	return pretty.Concise(list)
}
