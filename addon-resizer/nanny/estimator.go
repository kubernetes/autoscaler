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

	log "github.com/golang/glog"
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

// ExponentialEstimator estimates resource requirements in a way that prevents
// frequent updates but may end up with larger estimates than actually needed.
type ExponentialEstimator struct {
	// The collection of resources to provide estimates for.
	Resources []Resource
	// The minimum cluster size for which the estimator will provide
	// resource estimates. Must be greater than 1.
	MinClusterSize uint64
	// The multiplier used to compute the next cluster size to provide
	// estimates for. For example, suppose the cluster has 9 nodes,
	// MinClusterSize is 5, and ScaleFactor is 1.5. Then the estimator will
	// provide estimates for FLOOR(5 * 1.5 * 1.5) = 11 nodes since that is
	// the smallest cluster size larger than 9 nodes.
	ScaleFactor float64
}

func (e ExponentialEstimator) scaleWithNodes(numNodes uint64) *corev1.ResourceRequirements {
	n := e.MinClusterSize
	for n < numNodes {
		n = uint64(float64(n)*e.ScaleFactor + eps)
	}

	return calculateResources(n, e.Resources)
}

// Generates and returns a resource value string describing the overhead when
// running on a cluster with the given number of nodes. The per node overhead
// is taken from the Resource instance.
//
// Note this function takes into account resource units allowing it to compute
// resource overhead values even with fractional values. For example, it can
// handle incremental values of 0.5m for cpu resources.
func computeResourceOverheadValueString(numNodes uint64, resource Resource) string {
	perNodeOverhead := resource.ExtraPerNode.String()
	var perNodeValue float64
	var perNodeUnit string
	_, err := fmt.Sscanf(perNodeOverhead, "%f%s", &perNodeValue, &perNodeUnit)
	if err != nil && err.Error() != "EOF" {
		log.Warningf(
			"Failed to parse the per node overhead string '%s'; error=%s",
			perNodeOverhead, err)
		// Default to not specifying any unit to maintain existing
		// behaviour.
		perNodeUnit = ""
	}
	return fmt.Sprintf("%f%s", perNodeValue*float64(numNodes), perNodeUnit)
}

func calculateResources(numNodes uint64, resources []Resource) *corev1.ResourceRequirements {
	limits := make(corev1.ResourceList)
	requests := make(corev1.ResourceList)
	for _, r := range resources {
		overhead := computeResourceOverheadValueString(numNodes, r)
		newRes := r.Base
		newRes.Add(resource.MustParse(overhead))
		limits[r.Name] = newRes
		requests[r.Name] = newRes
	}
	return &corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}
