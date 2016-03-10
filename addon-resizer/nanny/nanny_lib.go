/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	api "k8s.io/kubernetes/pkg/api/v1"
	inf "speter.net/go/exp/math/dec/inf"
)

// checkResource determines whether a specific resource needs to be over-written.
func checkResource(threshold int64, actual, expected api.ResourceList, res api.ResourceName) bool {
	val, ok := actual[res]
	expVal, expOk := expected[res]
	if ok != expOk {
		return true
	}
	if !ok && !expOk {
		return false
	}
	q := new(inf.Dec).QuoRound(val.Amount, expVal.Amount, 2, inf.RoundDown)
	lower := inf.NewDec(100-threshold, 2)
	upper := inf.NewDec(100+threshold, 2)
	if q.Cmp(lower) == -1 || q.Cmp(upper) == 1 {
		return true
	}
	return false
}

// shouldOverwriteResources determines if we should over-write the container's
// resource limits. We'll over-write the resource limits if the limited
// resources are different, or if any limit is violated by a threshold.
func shouldOverwriteResources(threshold int64, limits, reqs, expLimits, expReqs api.ResourceList) bool {
	return checkResource(threshold, limits, expLimits, api.ResourceCPU) ||
		checkResource(threshold, limits, expLimits, api.ResourceMemory) ||
		checkResource(threshold, limits, expLimits, api.ResourceStorage) ||
		checkResource(threshold, reqs, expReqs, api.ResourceCPU) ||
		checkResource(threshold, reqs, expReqs, api.ResourceMemory) ||
		checkResource(threshold, reqs, expReqs, api.ResourceStorage)
}

type KubernetesClient interface {
	CountNodes() (uint64, error)
	ContainerResources() (*api.ResourceRequirements, error)
	UpdateDeployment(resources *api.ResourceRequirements) error
}

type ResourceEstimator interface {
	scaleWithNodes(numNodes uint64) *api.ResourceRequirements
}

// PollApiServer periodically counts the number of nodes, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
func PollApiServer(k8s KubernetesClient, est ResourceEstimator, contName string, pollPeriod time.Duration, threshold uint64) {
	for i := 0; true; i++ {
		if i != 0 {
			// Sleep for the poll period.
			time.Sleep(pollPeriod)
		}

		// Query the apiserver for the number of nodes.
		num, err := k8s.CountNodes()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("The number of nodes is %d", num)

		// Query the apiserver for this pod's information.
		resources, err := k8s.ContainerResources()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("The container resources are %v", resources)

		// Get the expected resource limits.
		expResources := est.scaleWithNodes(num)
		log.Infof("The expected resources are %v", expResources)

		// If there's a difference, go ahead and set the new values.
		if !shouldOverwriteResources(int64(threshold), resources.Limits, resources.Requests, expResources.Limits, expResources.Requests) {
			log.Infof("Resources are within the expected limits.")
			continue
		}
		log.Infof("Resources are not within the expected limits: updating the deployment.")
		if err := k8s.UpdateDeployment(expResources); err != nil {
			log.Error(err)
			continue
		}
	}
}
