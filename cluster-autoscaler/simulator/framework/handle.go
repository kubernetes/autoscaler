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
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources"
	"k8s.io/client-go/informers"
	schedulerconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedulerconfiglatest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
	schedulerplugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodevolumelimits"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
	schedulerframeworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	schedulermetrics "k8s.io/kubernetes/pkg/scheduler/metrics"
)

var (
	initMetricsOnce sync.Once
)

// NewKarpenterDisabledPluginsSchedulerConfig returns a copy of base KubeSchedulerConfiguration with InterPodAffinity
// and PodTopologySpread plugins disabled.
// WHY: In Karpenter simulation mode, Karpenter's solver natively evaluates and validates all inter-pod affinities
// and topology spread constraints. Disabling these plugins in CA's predicate snapshot framework prevents redundant,
// expensive predicate evaluations during snapshot.SchedulePod, avoiding CPU overhead and false-positive predicate failures.
func NewKarpenterDisabledPluginsSchedulerConfig(base *schedulerconfig.KubeSchedulerConfiguration) (*schedulerconfig.KubeSchedulerConfiguration, error) {
	if base == nil {
		var err error
		base, err = schedulerconfiglatest.Default()
		if err != nil {
			return nil, fmt.Errorf("couldn't create default scheduler config: %v", err)
		}
	}
	cfg := base.DeepCopy()
	if len(cfg.Profiles) > 0 {
		profile := &cfg.Profiles[0]
		if profile.Plugins == nil {
			profile.Plugins = &schedulerconfig.Plugins{}
		}
		disabledPlugins := []schedulerconfig.Plugin{
			{Name: interpodaffinity.Name},
			{Name: podtopologyspread.Name},
		}
		profile.Plugins.PreFilter.Disabled = append(profile.Plugins.PreFilter.Disabled, disabledPlugins...)
		profile.Plugins.Filter.Disabled = append(profile.Plugins.Filter.Disabled, disabledPlugins...)
	}
	return cfg, nil
}

// Handle is meant for interacting with the scheduler framework.
type Handle struct {
	Framework        schedulerimpl.Framework
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
	} else {
		opts = append(opts, schedulerframeworkruntime.WithSharedDRAManager(dynamicresources.NewNoOpDRAManager()))
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
