/*
Copyright 2016 The Kubernetes Authors.

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

package nanny

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	eps = float64(0.01)
)

// Resource defines the name of a resource, the quantity, and the marginal value.
type Resource struct {
	Base, ExtraPerNode resource.Quantity
	Name               corev1.ResourceName
}

// LinearEstimator estimates the amount of resources as r = base + extra*nodes.
type LinearEstimator struct {
	Resources []Resource
}

func (e LinearEstimator) scaleWithNodes(numNodes uint64) *corev1.ResourceRequirements {
	return calculateResources(numNodes, e.Resources)
}

// ExponentialEstimator estimates the amount of resources in the way that
// prevents from frequent updates but may end up with larger resource usage
// than actually needed (though no more than ScaleFactor).
type ExponentialEstimator struct {
	Resources   []Resource
	ScaleFactor float64
}

func (e ExponentialEstimator) scaleWithNodes(numNodes uint64) *corev1.ResourceRequirements {
	n := uint64(16)
	for n < numNodes {
		n = uint64(float64(n)*e.ScaleFactor + eps)
	}
	return calculateResources(n, e.Resources)
}

func calculateResources(numNodes uint64, resources []Resource) *corev1.ResourceRequirements {
	limits := make(corev1.ResourceList)
	requests := make(corev1.ResourceList)
	for _, r := range resources {
		// Since we want to enable passing values smaller than e.g. 1 millicore per node,
		// we need to have some more hacky solution here than operating on MilliValues.
		perNodeString := r.ExtraPerNode.String()
		var perNode float64
		read, _ := fmt.Sscanf(perNodeString, "%f", &perNode)
		overhead := resource.MustParse(fmt.Sprintf("%f%s", perNode*float64(numNodes), perNodeString[read:]))

		newRes := r.Base
		newRes.Add(overhead)

		limits[r.Name] = newRes
		requests[r.Name] = newRes
	}
	return &corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}
