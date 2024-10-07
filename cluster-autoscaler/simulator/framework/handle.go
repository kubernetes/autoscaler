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
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_config "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	scheduler_plugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerframeworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	schedulermetrics "k8s.io/kubernetes/pkg/scheduler/metrics"
)

type Handle struct {
	Framework        schedulerframework.Framework
	DelegatingLister *DelegatingSchedulerSharedLister
}

func NewHandle(informerFactory informers.SharedInformerFactory, schedConfig *config.KubeSchedulerConfiguration, draEnabled bool) (*Handle, error) {
	if schedConfig == nil {
		var err error
		schedConfig, err = scheduler_config.Default()
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler config: %v", err)
		}
	}

	if len(schedConfig.Profiles) != 1 {
		return nil, fmt.Errorf("unexpected scheduler config: expected one scheduler profile only (found %d profiles)", len(schedConfig.Profiles))
	}
	sharedLister := NewDelegatingSchedulerSharedLister()

	opts := []schedulerframeworkruntime.Option{
		schedulerframeworkruntime.WithInformerFactory(informerFactory),
		schedulerframeworkruntime.WithSnapshotSharedLister(sharedLister),
	}
	if draEnabled {
		opts = append(opts, schedulerframeworkruntime.WithSharedDraManager(sharedLister))
	}

	schedulermetrics.InitMetrics()
	framework, err := schedulerframeworkruntime.NewFramework(
		context.TODO(),
		scheduler_plugins.NewInTreeRegistry(),
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
