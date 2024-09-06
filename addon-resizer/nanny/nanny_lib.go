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

	inf "gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/addon-resizer/healthcheck"
	"k8s.io/klog/v2"
)

type operation int

const (
	unknown               operation = iota
	scaleDown             operation = iota
	scaleUp               operation = iota
	ContainerProportional           = "container-proportional"
	NodeProportional                = "node-proportional"
)

type updateResult int

const (
	noChange  updateResult = iota
	postpone  updateResult = iota
	overwrite updateResult = iota
)

// checkResource determines whether a specific resource needs to be over-written and determines type of the operation.
func checkResource(threshold int64, actual, expected corev1.ResourceList, res corev1.ResourceName) (bool, operation) {
	val, ok := actual[res]
	expVal, expOk := expected[res]
	if ok != expOk {
		return true, unknown
	}
	if !ok && !expOk {
		return false, unknown
	}
	q := new(inf.Dec).QuoRound(val.AsDec(), expVal.AsDec(), 2, inf.RoundDown)
	lower := inf.NewDec(100-threshold, 2)
	upper := inf.NewDec(100+threshold, 2)
	if q.Cmp(lower) == -1 {
		return true, scaleUp
	}
	if q.Cmp(upper) == 1 {
		return true, scaleDown
	}
	return false, unknown
}

// shouldOverwriteResources determines if we should over-write the container's resource limits and determines type of the operation.
// We'll over-write the resource limits if the limited resources are different, or if any limit is violated by a threshold.
// Returned operation type (scale up/down or unknown) is calculated based on values taken from the first resource which requires overwrite.
func shouldOverwriteResources(threshold int64, limits, reqs, expLimits, expReqs corev1.ResourceList) (bool, operation) {
	for _, resourceType := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceStorage} {
		overwrite, op := checkResource(threshold, limits, expLimits, resourceType)
		if overwrite {
			return true, op
		}
		overwrite, op = checkResource(threshold, reqs, expReqs, resourceType)
		if overwrite {
			return true, op
		}
	}
	return false, unknown
}

// KubernetesClient is an object that performs the nanny's requisite interactions with Kubernetes.
type KubernetesClient interface {
	CountNodes() (uint64, error)
	CountContainers() (uint64, error)
	ContainerResources() (*corev1.ResourceRequirements, error)
	UpdateDeployment(resources *corev1.ResourceRequirements) error
}

// ResourceEstimator estimates ResourceRequirements for a given criteria.
type ResourceEstimator interface {
	scale(clusterSize uint64) *corev1.ResourceRequirements
	updatedResourceEstimator(resources []Resource) ResourceEstimator
}

type NannyConfigUpdater interface {
	CurrentResources() ([]Resource, error)
}

// PollAPIServer periodically counts the size of the cluster, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
func PollAPIServer(k8s KubernetesClient, est ResourceEstimator, hc *healthcheck.HealthCheck, pollPeriod, scaleDownDelay, scaleUpDelay time.Duration, threshold uint64, scalingMode string, runOnControlPlane bool, nannyConfigUpdater NannyConfigUpdater) {
	hc.StartMonitoring()
	lastChange := time.Now()
	lastResult := noChange

	for i := 0; true; i++ {
		if i != 0 {
			// Sleep for the poll period.
			time.Sleep(pollPeriod)
		}

		if runOnControlPlane {
			resources, err := nannyConfigUpdater.CurrentResources()
			if err != nil {
				klog.Fatal(err)
			}
			est = est.updatedResourceEstimator(resources)
		}

		if lastResult = updateResources(k8s, est, time.Now(), lastChange, scaleDownDelay, scaleUpDelay, threshold, lastResult, scalingMode); lastResult == overwrite {
			lastChange = time.Now()
		}
		hc.UpdateLastActivity()
	}
}

// updateResources counts the cluster size, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
// It returns overwrite if deployment has been updated, postpone if the change
// could not be applied due to scale up/down delay and noChange if the estimated
// expected ResourceRequirements are in line with the actual ResourceRequirements.
func updateResources(k8s KubernetesClient, est ResourceEstimator, now, lastChange time.Time, scaleDownDelay, scaleUpDelay time.Duration, threshold uint64, prevResult updateResult, scalingMode string) updateResult {
	// Query the apiserver for the cluster size.
	var num uint64
	var err error
	if scalingMode == ContainerProportional {
		num, err = k8s.CountContainers()
	} else {
		num, err = k8s.CountNodes()
	}

	if err != nil {
		klog.Error(err)
		ObserveOutcome("error_getting_count")
		return noChange
	}
	klog.V(4).Infof("The cluster size is %d", num)

	// Query the apiserver for this pod's information.
	resources, err := k8s.ContainerResources()
	if err != nil {
		klog.Errorf("Error while querying apiserver for resources: %v", err)
		ObserveOutcome("error_getting_resources")
		return noChange
	}

	// Get the expected resource limits.
	expResources := est.scale(num)

	// If there's a difference, go ahead and set the new values.
	overwriteRes, op := shouldOverwriteResources(int64(threshold), resources.Limits, resources.Requests, expResources.Limits, expResources.Requests)
	if !overwriteRes {
		if klog.V(4).Enabled() {
			klog.V(4).Infof("Resources are within the expected limits. Actual: %+v Expected: %+v", jsonOrValue(*resources), jsonOrValue(*expResources))
		}
		ObserveOutcome("resources_within_limits")
		return noChange
	}

	if (op == scaleDown && now.Before(lastChange.Add(scaleDownDelay))) ||
		(op == scaleUp && now.Before(lastChange.Add(scaleUpDelay))) {
		if prevResult != postpone {
			klog.Infof("Resources are not within the expected limits, Actual: %+v, accepted range: %+v. Skipping resource update because of scale up/down delay", jsonOrValue(*resources), jsonOrValue(*expResources))
		}
		ObserveOutcome("scale_up_down_delay")
		return postpone
	}

	klog.Infof("Resources are not within the expected limits, updating the deployment. Actual: %+v Expected: %+v. Resource will be updated.", jsonOrValue(*resources), jsonOrValue(*expResources))
	if err := k8s.UpdateDeployment(expResources); err != nil {
		klog.Error(err)
		ObserveOutcome("error_updating_deployment")
		return noChange
	}
	ObserveOutcome("updated_resources")
	return overwrite
}

func jsonOrValue(val interface{}) interface{} {
	bytes, err := json.Marshal(val)
	if err != nil {
		return val
	}
	return string(bytes)
}
