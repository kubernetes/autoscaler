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

package core

import (
	"reflect"
	"sort"
	"time"

	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// PollingAutoscaler is a variant of autoscaler which polls the source-of-truth every time RunOnce is invoked
type PollingAutoscaler struct {
	autoscaler        Autoscaler
	autoscalerBuilder AutoscalerBuilder
}

// NewPollingAutoscaler builds a PollingAutoscaler from required parameters
func NewPollingAutoscaler(autoscalerBuilder AutoscalerBuilder) (*PollingAutoscaler, errors.AutoscalerError) {
	autoscaler, err := autoscalerBuilder.Build()
	if err != nil {
		return nil, err
	}

	return &PollingAutoscaler{
		autoscaler:        autoscaler,
		autoscalerBuilder: autoscalerBuilder,
	}, nil
}

// CleanUp does the work required before all the iterations of a polling autoscaler run
func (a *PollingAutoscaler) CleanUp() {
	a.autoscaler.CleanUp()
}

// CloudProvider returns the cloud provider associated to this autoscaler
func (a *PollingAutoscaler) CloudProvider() cloudprovider.CloudProvider {
	return a.autoscaler.CloudProvider()
}

// ExitCleanUp cleans-up after autoscaler, so no mess remains after process termination.
func (a *PollingAutoscaler) ExitCleanUp() {
	a.autoscaler.ExitCleanUp()
}

// RunOnce represents a single iteration of a polling autoscaler inside the CA's control-loop
func (a *PollingAutoscaler) RunOnce(currentTime time.Time) errors.AutoscalerError {
	reconfigureStart := time.Now()
	metrics.UpdateLastTime(metrics.Poll, reconfigureStart)
	if err := a.Poll(); err != nil {
		glog.Errorf("Failed to poll : %v", err)
	}
	metrics.UpdateDurationFromStart(metrics.Poll, reconfigureStart)
	return a.autoscaler.RunOnce(currentTime)
}

// Poll latest data from cloud provider to recreate this autoscaler
func (a *PollingAutoscaler) Poll() error {
	prevAutoscaler := a.autoscaler

	// For safety, any config change should stop and recreate all the stuff running in CA hence recreating all the Autoscaler instance here
	// See https://github.com/kubernetes/contrib/pull/2226#discussion_r94126064
	currentAutoscaler, err := a.autoscalerBuilder.Build()
	if err != nil {
		return err
	}

	// Not to complicate the work, we replace the autoscaler/reset autoscaler state only when the list of target node groups are changed.
	// We consider it changed only when:
	// (1) auto-discovery is enabled and
	// (2) list of node groups matching the criteria(e.g. asg tag for aws provider) changed
	//
	// We should not consider it changed when:
	// *  min/max/target/current size of node group(s) changed
	//
	// See https://github.com/kubernetes/autoscaler/pull/107#issuecomment-307518602 for more context
	prevNodeGroupIds := sortedIdsOfNodeGroups(prevAutoscaler.CloudProvider().NodeGroups())
	currentNodeGroupIds := sortedIdsOfNodeGroups(currentAutoscaler.CloudProvider().NodeGroups())

	if !reflect.DeepEqual(prevNodeGroupIds, currentNodeGroupIds) {
		glog.V(4).Infof("Detected change(s) in node group definitions. Recreating autoscaler...")

		// See https://github.com/kubernetes/autoscaler/issues/252, we need to close any stray resources
		a.autoscaler.CloudProvider().Cleanup()

		// For safety, any config change should stop and recreate all the stuff running in CA hence recreating all the Autoscaler instance here
		// See https://github.com/kubernetes/contrib/pull/2226#discussion_r94126064
		a.autoscaler = currentAutoscaler
	} else {
		// See https://github.com/kubernetes/autoscaler/issues/252, we need to close any stray resources
		currentAutoscaler.CloudProvider().Cleanup()
	}
	glog.V(4).Infof("Poll finished")
	return nil
}

func sortedIdsOfNodeGroups(nodeGroups []cloudprovider.NodeGroup) []string {
	ids := []string{}
	for _, g := range nodeGroups {
		ids = append(ids, g.Id())
	}
	sort.Strings(ids)
	return ids
}
