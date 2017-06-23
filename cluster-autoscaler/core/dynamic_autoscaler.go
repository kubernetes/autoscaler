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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// DynamicAutoscaler is a variant of autoscaler which supports dynamic reconfiguration at runtime
type DynamicAutoscaler struct {
	autoscaler        Autoscaler
	autoscalerBuilder AutoscalerBuilder
	configFetcher     dynamic.ConfigFetcher
}

// NewDynamicAutoscaler builds a DynamicAutoscaler from required parameters
func NewDynamicAutoscaler(autoscalerBuilder AutoscalerBuilder, configFetcher dynamic.ConfigFetcher) (*DynamicAutoscaler, errors.AutoscalerError) {
	autoscaler, err := autoscalerBuilder.Build()
	if err != nil {
		return nil, err
	}
	return &DynamicAutoscaler{
		autoscaler:        autoscaler,
		autoscalerBuilder: autoscalerBuilder,
		configFetcher:     configFetcher,
	}, nil
}

// CleanUp does the work required before all the iterations of a dynamic autoscaler run
func (a *DynamicAutoscaler) CleanUp() {
	a.autoscaler.CleanUp()
}

// CloudProvider returns the cloud provider associated to this autoscaler
func (a *DynamicAutoscaler) CloudProvider() cloudprovider.CloudProvider {
	return a.autoscaler.CloudProvider()
}

// ExitCleanUp cleans-up after autoscaler, so no mess remains after process termination.
func (a *DynamicAutoscaler) ExitCleanUp() {
	a.autoscaler.ExitCleanUp()
}

// RunOnce represents a single iteration of a dynamic autoscaler inside the CA's control-loop
func (a *DynamicAutoscaler) RunOnce(currentTime time.Time) errors.AutoscalerError {
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
		a.autoscaler, err = a.autoscalerBuilder.SetDynamicConfig(*updatedConfig).Build()
		if err != nil {
			return err
		}
		glog.V(4).Infof("Dynamic reconfiguration finished: updatedConfig=%v", updatedConfig)
	}

	return nil
}
