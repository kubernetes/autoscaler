//go:build !providerless
// +build !providerless

/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	compute "google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

func newTargetPoolMetricContext(request, region string) *metricContext {
	return newGenericMetricContext("targetpool", request, region, unusedMetricLabel, computeV1Version)
}

// GetTargetPool returns the TargetPool by name.
func (g *Cloud) GetTargetPool(name, region string) (*compute.TargetPool, error) {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	mc := newTargetPoolMetricContext("get", region)
	v, err := g.c.TargetPools().Get(ctx, meta.RegionalKey(name, region))
	return v, mc.Observe(err)
}

// CreateTargetPool creates the passed TargetPool
func (g *Cloud) CreateTargetPool(tp *compute.TargetPool, region string) error {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	mc := newTargetPoolMetricContext("create", region)
	return mc.Observe(g.c.TargetPools().Insert(ctx, meta.RegionalKey(tp.Name, region), tp))
}

// DeleteTargetPool deletes TargetPool by name.
func (g *Cloud) DeleteTargetPool(name, region string) error {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	mc := newTargetPoolMetricContext("delete", region)
	return mc.Observe(g.c.TargetPools().Delete(ctx, meta.RegionalKey(name, region)))
}

// AddInstancesToTargetPool adds instances by link to the TargetPool
func (g *Cloud) AddInstancesToTargetPool(name, region string, instanceRefs []*compute.InstanceReference) error {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	req := &compute.TargetPoolsAddInstanceRequest{
		Instances: instanceRefs,
	}
	mc := newTargetPoolMetricContext("add_instances", region)
	return mc.Observe(g.c.TargetPools().AddInstance(ctx, meta.RegionalKey(name, region), req))
}

// RemoveInstancesFromTargetPool removes instances by link to the TargetPool
func (g *Cloud) RemoveInstancesFromTargetPool(name, region string, instanceRefs []*compute.InstanceReference) error {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	req := &compute.TargetPoolsRemoveInstanceRequest{
		Instances: instanceRefs,
	}
	mc := newTargetPoolMetricContext("remove_instances", region)
	return mc.Observe(g.c.TargetPools().RemoveInstance(ctx, meta.RegionalKey(name, region), req))
}
