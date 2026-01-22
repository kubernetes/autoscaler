/*
Copyright 2024 The Kubernetes Authors.

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

package framework

import (
	"context"
	"fmt"
	"k8s.io/client-go/informers"
	schedulerconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedulerconfiglatest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	schedulerplugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodevolumelimits"
	schedulerframeworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	schedulermetrics "k8s.io/kubernetes/pkg/scheduler/metrics"
	"sync"
)

var (
	initMetricsOnce sync.Once
)

// Handle is meant for interacting with the scheduler framework.
type Handle struct {
	Framework        schedulerframework.Framework
	DelegatingLister *DelegatingSchedulerSharedLister
}

// NewHandle builds a framework Handle based on the provided informers and scheduler config.
func NewHandle(ctx context.Context, informerFactory informers.SharedInformerFactory, schedConfig *schedulerconfig.KubeSchedulerConfiguration, draEnabled bool, csiEnabled bool) (*Handle, error) {
	if schedConfig == nil {
		var err error
		schedConfig, err = schedulerconfiglatest.Default()
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler config: %v", err)
		}
	}
	if len(schedConfig.Profiles) != 1 {
		return nil, fmt.Errorf("unexpected scheduler config: expected one scheduler profile only (found %d profiles)", len(schedConfig.Profiles))
	}

	sharedLister := NewDelegatingSchedulerSharedLister()
	sharedCSIManager := nodevolumelimits.NewCSIManager(informerFactory.Storage().V1().CSINodes().Lister())
	opts := []schedulerframeworkruntime.Option{
		schedulerframeworkruntime.WithInformerFactory(informerFactory),
		schedulerframeworkruntime.WithSnapshotSharedLister(sharedLister),
		schedulerframeworkruntime.WithSharedCSIManager(sharedCSIManager),
	}
	if draEnabled {
		opts = append(opts, schedulerframeworkruntime.WithSharedDRAManager(sharedLister))
	}
	// TODO: We should always use sharedLister once this CSINode aware changes in CAS are
	// enabled by default.
	if csiEnabled {
		opts = append(opts, schedulerframeworkruntime.WithSharedCSIManager(sharedLister))
	} else {
		sharedCSIManager := nodevolumelimits.NewCSIManager(informerFactory.Storage().V1().CSINodes().Lister())
		opts = append(opts, schedulerframeworkruntime.WithSharedCSIManager(sharedCSIManager))
	}
	initMetricsOnce.Do(func() {
		schedulermetrics.InitMetrics()
	})
	framework, err := schedulerframeworkruntime.NewFramework(
		ctx,
		schedulerplugins.NewInTreeRegistry(),
		&schedConfig.Profiles[0],
		opts...,
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't create scheduler framework; %v", err)
	}

	return &Handle{
		Framework:        framework,
		DelegatingLister: sharedLister,
	}, nil
}
