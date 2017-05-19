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
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// DynamicAutoscaler is a variant of autoscaler which supports dynamic reconfiguration at runtime
type DynamicAutoscaler struct {
	autoscaler        Autoscaler
	autoscalerBuilder AutoscalerBuilder
	configFetcher     dynamic.ConfigFetcher
}

// NewDynamicAutoscaler builds a DynamicAutoscaler from required parameters
func NewDynamicAutoscaler(autoscalerBuilder AutoscalerBuilder, configFetcher dynamic.ConfigFetcher) *DynamicAutoscaler {
	return &DynamicAutoscaler{
		autoscaler:        autoscalerBuilder.Build(),
		autoscalerBuilder: autoscalerBuilder,
		configFetcher:     configFetcher,
	}
}

// CleanUp does the work required before all the iterations of a dynamic autoscaler run
func (a *DynamicAutoscaler) CleanUp() {
	a.autoscaler.CleanUp()
}

// ExitCleanUp cleans-up after autoscaler, so no mess remains after process termination.
func (a *DynamicAutoscaler) ExitCleanUp() {
	a.autoscaler.ExitCleanUp()
}

// RunOnce represents a single iteration of a dynamic autoscaler inside the CA's control-loop
func (a *DynamicAutoscaler) RunOnce(currentTime time.Time) *errors.AutoscalerError {
	reconfigureStart := time.Now()
	metrics.UpdateLastTime("reconfigure", reconfigureStart)
	if err := a.Reconfigure(); err != nil {
		glog.Errorf("Failed to reconfigure : %v", err)
	}
	metrics.UpdateDuration("reconfigure", reconfigureStart)
	return a.autoscaler.RunOnce(currentTime)
}

// Reconfigure this dynamic autoscaler if the configmap is updated
func (a *DynamicAutoscaler) Reconfigure() error {
	var updatedConfig *dynamic.Config
	var err error

	if updatedConfig, err = a.configFetcher.FetchConfigIfUpdated(); err != nil {
		return fmt.Errorf("failed to fetch updated config: %v", err)
	}

	if updatedConfig != nil {
		// For safety, any config change should stop and recreate all the stuff running in CA hence recreating all the Autoscaler instance here
		// See https://github.com/kubernetes/contrib/pull/2226#discussion_r94126064
		a.autoscaler = a.autoscalerBuilder.SetDynamicConfig(*updatedConfig).Build()
		glog.V(4).Infof("Dynamic reconfiguration finished: updatedConfig=%v", updatedConfig)
	}

	return nil
}

// AutoscalerBuilder builds an instance of Autoscaler which is the core of CA
type AutoscalerBuilder interface {
	SetDynamicConfig(config dynamic.Config) AutoscalerBuilder
	Build() Autoscaler
}

// AutoscalerBuilderImpl builds new autoscalers from its state including initial `AutoscalingOptions` given at startup and
// `dynamic.Config` read on demand from the configmap
type AutoscalerBuilderImpl struct {
	autoscalingOptions AutoscalingOptions
	dynamicConfig      *dynamic.Config
	kubeClient         kube_client.Interface
	kubeEventRecorder  kube_record.EventRecorder
	predicateChecker   *simulator.PredicateChecker
	listerRegistry     kube_util.ListerRegistry
}

// NewAutoscalerBuilder builds an AutoscalerBuilder from required parameters
func NewAutoscalerBuilder(autoscalingOptions AutoscalingOptions, predicateChecker *simulator.PredicateChecker, kubeClient kube_client.Interface, kubeEventRecorder kube_record.EventRecorder, listerRegistry kube_util.ListerRegistry) *AutoscalerBuilderImpl {
	return &AutoscalerBuilderImpl{
		autoscalingOptions: autoscalingOptions,
		kubeClient:         kubeClient,
		kubeEventRecorder:  kubeEventRecorder,
		predicateChecker:   predicateChecker,
		listerRegistry:     listerRegistry,
	}
}

// SetDynamicConfig sets an instance of dynamic.Config read from a configmap so that
// the new autoscaler built afterwards reflect the latest configuration contained in the configmap
func (b *AutoscalerBuilderImpl) SetDynamicConfig(config dynamic.Config) AutoscalerBuilder {
	b.dynamicConfig = &config
	return b
}

// Build an autoscaler according to the builder's state
func (b *AutoscalerBuilderImpl) Build() Autoscaler {
	options := b.autoscalingOptions
	if b.dynamicConfig != nil {
		c := *(b.dynamicConfig)
		options.NodeGroups = c.NodeGroupSpecStrings()
	}
	return NewStaticAutoscaler(options, b.predicateChecker, b.kubeClient, b.kubeEventRecorder, b.listerRegistry)
}
