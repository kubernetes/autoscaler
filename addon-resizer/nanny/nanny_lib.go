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
	"encoding/json"
	"time"

	log "github.com/golang/glog"
	api "k8s.io/api/core/v1"
)

type operation int

const (
	unknown   operation = iota
	scaleDown operation = iota
	scaleUp   operation = iota
)

type updateResult int

const (
	noChange  updateResult = iota
	postpone  updateResult = iota
	overwrite updateResult = iota
)

// checkResource determines whether a specific resource needs to be over-written.
func checkResource(estimatorResult *EstimatorResult, actual api.ResourceList, res api.ResourceName) (*api.ResourceList, operation) {
	val, ok := actual[res]
	expMinVal, expMinOk := estimatorResult.AcceptableRange.lower[res]
	expMaxVal, expMaxOk := estimatorResult.AcceptableRange.upper[res]
	if ok != expMinOk || ok != expMaxOk {
		// Something changed, but we don't know whether lower or upper bound should be used.
		// It doesn't matter in the long term, so we just pick lower bound arbitrarily here.
		return &estimatorResult.RecommendedRange.lower, unknown
	}
	if !ok && !expMinOk && !expMaxOk {
		return nil, unknown
	}
	if val.Cmp(expMinVal) == -1 {
		return &estimatorResult.RecommendedRange.lower, scaleUp
	}
	if val.Cmp(expMaxVal) == 1 {
		return &estimatorResult.RecommendedRange.upper, scaleDown
	}
	return nil, unknown
}

// shouldOverwriteResources determines if we should over-write the container's resource limits and determines type of the operation.
// We'll over-write the resource limits if the limited resources are different, or if any limit is outside of the accepted range.
// Returns null ResourceRequirements when no resources should be overridden.
// Returned operation type (scale up/down or unknown) is calculated based on values taken from the first resource which requires overwrite.
func shouldOverwriteResources(estimatorResult *EstimatorResult, limits, reqs api.ResourceList) (*api.ResourceRequirements, operation) {
	for _, list := range []api.ResourceList{limits, reqs} {
		for _, resourceType := range []api.ResourceName{api.ResourceCPU, api.ResourceMemory, api.ResourceStorage} {
			newReqs, op := checkResource(estimatorResult, list, resourceType)
			if newReqs != nil {
				log.V(4).Infof("Resource %s is out of bounds.", resourceType)
				return &api.ResourceRequirements{Limits: *newReqs, Requests: *newReqs}, op
			}
		}
	}
	return nil, unknown
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
func PollAPIServer(k8s KubernetesClient, est ResourceEstimator, pollPeriod, scaleDownDelay, scaleUpDelay time.Duration) {
	lastChange := time.Now()
	lastResult := noChange

	for i := 0; true; i++ {
		if i != 0 {
			// Sleep for the poll period.
			time.Sleep(pollPeriod)
		}

		if lastResult = updateResources(k8s, est, time.Now(), lastChange, scaleDownDelay, scaleUpDelay, lastResult); lastResult == overwrite {
			lastChange = time.Now()
		}
	}
}

// updateResources counts the number of nodes, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
// It returns overwrite if deployment has been updated, postpone if the change
// could not be applied due to scale up/down delay and noChange if the estimated
// expected ResourceRequirements are in line with the actual ResourceRequirements.
func updateResources(k8s KubernetesClient, est ResourceEstimator, now, lastChange time.Time, scaleDownDelay, scaleUpDelay time.Duration, prevResult updateResult) updateResult {

	// Query the apiserver for the number of nodes.
	num, err := k8s.CountNodes()
	if num == 0 {
		log.V(2).Info("No nodes found, probably listers have not synced yet. Skipping current check.")
		return noChange
	}
	if err != nil {
		log.Error(err)
		return noChange
	}
	log.V(4).Infof("The number of nodes is %d", num)

	// Query the apiserver for this pod's information.
	resources, err := k8s.ContainerResources()
	if err != nil {
		log.Errorf("Error while querying apiserver for resources: %v", err)
		return noChange
	}

	// Get the expected resource limits.
	estimation := est.scaleWithNodes(num)

	// If there's a difference, go ahead and set the new values.
	overwriteResReq, op := shouldOverwriteResources(estimation, resources.Limits, resources.Requests)
	if overwriteResReq == nil {
		log.V(4).Infof("Resources are within the expected limits. Actual: %+v, accepted range: %+v", jsonOrValue(*resources), jsonOrValue(estimation.AcceptableRange))
		return noChange
	}

	if (op == scaleDown && now.Before(lastChange.Add(scaleDownDelay))) ||
		(op == scaleUp && now.Before(lastChange.Add(scaleUpDelay))) {
		log.Infof("Resources are not within the expected limits, Actual: %+v, accepted range: %+v. Skipping resource update because of scale up/down delay", jsonOrValue(*resources), jsonOrValue(estimation.AcceptableRange))
		return postpone
	}

	log.Infof("Resources are not within the expected limits, updating the deployment. Actual: %+v New: %+v", *resources, jsonOrValue(*overwriteResReq))
	if err := k8s.UpdateDeployment(overwriteResReq); err != nil {
		log.Error(err)
		return noChange
	}
	lastChange = time.Now()
	return overwrite
}

func jsonOrValue(val interface{}) interface{} {
	bytes, err := json.Marshal(val)
	if err != nil {
		return val
	}
	return string(bytes)
}
