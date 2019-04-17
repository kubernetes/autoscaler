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

/*
Package nanny implements logic to poll the k8s apiserver for cluster status,
and update a deployment based on that status.
*/
package nanny

import (
	"time"

	log "github.com/golang/glog"
	api "k8s.io/api/core/v1"
)

// checkResource determines whether a specific resource needs to be over-written.
func checkResource(estimatorResult *EstimatorResult, actual api.ResourceList, res api.ResourceName) *api.ResourceList {
	val, ok := actual[res]
	expMinVal, expMinOk := estimatorResult.AcceptableRange.lower[res]
	expMaxVal, expMaxOk := estimatorResult.AcceptableRange.upper[res]
	if ok != expMinOk || ok != expMaxOk {
		// Something changed, but we don't know whether lower or upper bound should be used.
		// It doesn't matter in the long term, so we just pick lower bound arbitrarily here.
		return &estimatorResult.RecommendedRange.lower
	}
	if !ok && !expMinOk && !expMaxOk {
		return nil
	}
	if val.Cmp(expMinVal) == -1 {
		return &estimatorResult.RecommendedRange.lower
	}
	if val.Cmp(expMaxVal) == 1 {
		return &estimatorResult.RecommendedRange.upper
	}
	return nil
}

// shouldOverwriteResources determines if we should over-write the container's
// resource limits. We'll over-write the resource limits if the limited
// resources are different, or if any limit is outside of the accepted range.
// Returns null when no resources should be overridden.
// Otherwise, returns ResourceList that should be used.
func shouldOverwriteResources(estimatorResult *EstimatorResult, limits, reqs api.ResourceList) *api.ResourceRequirements {
	for _, list := range []api.ResourceList{limits, reqs} {
		for _, resourceType := range []api.ResourceName{api.ResourceCPU, api.ResourceMemory, api.ResourceStorage} {
			newReqs := checkResource(estimatorResult, list, resourceType)
			if newReqs != nil {
				log.V(4).Infof("Resource %s is out of bounds.", resourceType)
				return &api.ResourceRequirements{Limits: *newReqs, Requests: *newReqs}
			}
		}
	}
	return nil
}

// KubernetesClient is an object that performs the nanny's requisite interactions with Kubernetes.
type KubernetesClient interface {
	CountNodes() (uint64, error)
	ContainerResources() (*api.ResourceRequirements, error)
	UpdateDeployment(resources *api.ResourceRequirements) error
	Stop()
}

// ResourceEstimator estimates ResourceRequirements for a given criteria. Returned value is a list
// with acceptable values. First element on that list is the recommended one.
type ResourceEstimator interface {
	scaleWithNodes(numNodes uint64) *EstimatorResult
}

// PollAPIServer periodically counts the number of nodes, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
func PollAPIServer(k8s KubernetesClient, est ResourceEstimator, pollPeriod time.Duration) {
	for i := 0; true; i++ {
		if i != 0 {
			// Sleep for the poll period.
			time.Sleep(pollPeriod)
		}

		// Query the apiserver for the number of nodes.
		num, err := k8s.CountNodes()
		if num == 0 {
			log.V(2).Info("No nodes found, probably listers have not synced yet. Skipping current check.")
			continue
		}
		if err != nil {
			log.Error(err)
			continue
		}
		log.V(4).Infof("The number of nodes is %d", num)

		// Query the apiserver for this pod's information.
		resources, err := k8s.ContainerResources()
		if err != nil {
			log.Errorf("Error while querying apiserver for resources: %v", err)
			continue
		}

		// Get the expected resource limits.
		estimation := est.scaleWithNodes(num)

		// If there's a difference, go ahead and set the new values.
		overwrite := shouldOverwriteResources(estimation, resources.Limits, resources.Requests)
		if overwrite == nil {
			log.V(4).Infof("Resources are within the expected limits. Actual: %+v, accepted range: %+v", *resources, estimation.AcceptableRange)
			continue
		}

		log.Infof("Resources are not within the expected limits, updating the deployment. Actual: %+v New: %+v", *resources, *overwrite)
		if err := k8s.UpdateDeployment(overwrite); err != nil {
			log.Error(err)
			continue
		}
	}
}
