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

package actuation

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// UpdateSoftDeletionTaints manages soft taints of unneeded nodes.
func UpdateSoftDeletionTaints(context *context.AutoscalingContext, uneededNodes, neededNodes []*apiv1.Node) (errors []error) {
	defer metrics.UpdateDurationFromStart(metrics.ScaleDownSoftTaintUnneeded, time.Now())
	b := &budgetTracker{
		apiCallBudget: context.AutoscalingOptions.MaxBulkSoftTaintCount,
		timeBudget:    context.AutoscalingOptions.MaxBulkSoftTaintTime,
		startTime:     now(),
	}
	for _, node := range neededNodes {
		if taints.HasToBeDeletedTaint(node) {
			// Do not consider nodes that are scheduled to be deleted
			continue
		}
		if !taints.HasDeletionCandidateTaint(node) {
			continue
		}
		b.processWithinBudget(func() {
			_, err := taints.CleanDeletionCandidate(node, context.ClientSet)
			if err != nil {
				errors = append(errors, err)
				klog.Warningf("Soft taint on %s removal error %v", node.Name, err)
			}
		})
	}
	for _, node := range uneededNodes {
		if taints.HasToBeDeletedTaint(node) {
			// Do not consider nodes that are scheduled to be deleted
			continue
		}
		if taints.HasDeletionCandidateTaint(node) {
			continue
		}
		b.processWithinBudget(func() {
			err := taints.MarkDeletionCandidate(node, context.ClientSet)
			if err != nil {
				errors = append(errors, err)
				klog.Warningf("Soft taint on %s adding error %v", node.Name, err)
			}
		})
	}
	b.reportExceededLimits()
	return
}

// Get current time. Proxy for unit tests.
var now func() time.Time = time.Now

type budgetTracker struct {
	apiCallBudget int
	startTime     time.Time
	timeBudget    time.Duration
	skippedNodes  int
}

func (b *budgetTracker) processWithinBudget(f func()) {
	if b.apiCallBudget <= 0 || now().Sub(b.startTime) >= b.timeBudget {
		b.skippedNodes++
		return
	}
	b.apiCallBudget--
	f()
}

func (b *budgetTracker) reportExceededLimits() {
	if b.skippedNodes > 0 {
		klog.V(4).Infof("Skipped adding/removing soft taints on %v nodes - API call or time limit exceeded", b.skippedNodes)
	}
}
