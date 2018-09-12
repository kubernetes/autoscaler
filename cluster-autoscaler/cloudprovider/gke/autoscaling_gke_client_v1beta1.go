/*
Copyright 2018 The Kubernetes Authors.

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

package gke

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	apiv1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
	gke_api_beta "google.golang.org/api/container/v1beta1"
)

var (
	// This makes me so sad
	taintEffectsMap = map[apiv1.TaintEffect]string{
		apiv1.TaintEffectNoSchedule:       "NO_SCHEDULE",
		apiv1.TaintEffectPreferNoSchedule: "PREFER_NO_SCHEDULE",
		apiv1.TaintEffectNoExecute:        "NO_EXECUTE",
	}
)

type autoscalingGkeClientV1beta1 struct {
	gkeBetaService *gke_api_beta.Service

	clusterPath   string
	nodePoolPath  string
	operationPath string

	operationWaitTimeout  time.Duration
	operationPollInterval time.Duration
}

// NewAutoscalingGkeClientV1beta1 creates a new client for communicating with GKE v1beta1 API.
func NewAutoscalingGkeClientV1beta1(client *http.Client, projectId, location, clusterName string) (*autoscalingGkeClientV1beta1, error) {
	autoscalingGkeClient := &autoscalingGkeClientV1beta1{
		clusterPath:           fmt.Sprintf(clusterPathPrefix, projectId, location, clusterName),
		nodePoolPath:          fmt.Sprintf(nodePoolPathPrefix, projectId, location, clusterName),
		operationPath:         fmt.Sprintf(operationPathPrefix, projectId, location),
		operationWaitTimeout:  defaultOperationWaitTimeout,
		operationPollInterval: defaultOperationPollInterval,
	}

	gkeBetaService, err := gke_api_beta.New(client)
	if err != nil {
		return nil, err
	}
	if *GkeAPIEndpoint != "" {
		gkeBetaService.BasePath = *GkeAPIEndpoint
	}
	autoscalingGkeClient.gkeBetaService = gkeBetaService

	return autoscalingGkeClient, nil
}

func (m *autoscalingGkeClientV1beta1) GetCluster() (Cluster, error) {
	clusterResponse, err := m.gkeBetaService.Projects.Locations.Clusters.Get(m.clusterPath).Do()
	if err != nil {
		return Cluster{}, err
	}
	nodePools := []NodePool{}
	for _, pool := range clusterResponse.NodePools {
		if pool.Autoscaling != nil && pool.Autoscaling.Enabled {
			nodePools = append(nodePools, NodePool{
				Name:              pool.Name,
				InstanceGroupUrls: pool.InstanceGroupUrls,
				Autoscaled:        pool.Autoscaling.Enabled,
				MinNodeCount:      pool.Autoscaling.MinNodeCount,
				MaxNodeCount:      pool.Autoscaling.MaxNodeCount,
				Autoprovisioned:   pool.Autoscaling.Autoprovisioned,
			})
		}
	}
	return Cluster{
		Locations:       clusterResponse.Locations,
		NodePools:       nodePools,
		ResourceLimiter: buildResourceLimiter(clusterResponse),
	}, nil
}

func buildResourceLimiter(cluster *gke_api_beta.Cluster) *cloudprovider.ResourceLimiter {
	if cluster.Autoscaling == nil {
		glog.Warningf("buildResourceLimiter called without autoscaling limits set")
		return nil
	}

	minLimits := make(map[string]int64)
	maxLimits := make(map[string]int64)
	for _, limit := range cluster.Autoscaling.ResourceLimits {
		if _, found := supportedResources[limit.ResourceType]; !found {
			glog.Warningf("Unsupported limit defined %s: %d - %d", limit.ResourceType, limit.Minimum, limit.Maximum)
		}
		minLimits[limit.ResourceType] = limit.Minimum
		maxLimits[limit.ResourceType] = limit.Maximum
	}

	// GKE API provides memory in GB, but ResourceLimiter expects them in bytes
	if _, found := minLimits[cloudprovider.ResourceNameMemory]; found {
		minLimits[cloudprovider.ResourceNameMemory] = minLimits[cloudprovider.ResourceNameMemory] * units.Gigabyte
	}
	if _, found := maxLimits[cloudprovider.ResourceNameMemory]; found {
		maxLimits[cloudprovider.ResourceNameMemory] = maxLimits[cloudprovider.ResourceNameMemory] * units.Gigabyte
	}

	return cloudprovider.NewResourceLimiter(minLimits, maxLimits)
}

func (m *autoscalingGkeClientV1beta1) DeleteNodePool(toBeRemoved string) error {
	deleteOp, err := m.gkeBetaService.Projects.Locations.Clusters.NodePools.Delete(
		fmt.Sprintf(m.nodePoolPath, toBeRemoved)).Do()
	if err != nil {
		return err
	}
	return m.waitForGkeOp(deleteOp)
}

func (m *autoscalingGkeClientV1beta1) CreateNodePool(mig *GkeMig) error {
	// TODO: handle preemptible VMs
	// TODO: handle SSDs

	spec := mig.Spec()

	accelerators := []*gke_api_beta.AcceleratorConfig{}
	if gpuRequest, found := spec.ExtraResources[gpu.ResourceNvidiaGPU]; found {
		gpuType, found := spec.Labels[gpu.GPULabel]
		if !found {
			return fmt.Errorf("failed to create node pool %v with gpu request of unspecified type", mig.NodePoolName())
		}
		gpuConfig := &gke_api_beta.AcceleratorConfig{
			AcceleratorType:  gpuType,
			AcceleratorCount: gpuRequest.Value(),
		}
		accelerators = append(accelerators, gpuConfig)

	}

	taints := []*gke_api_beta.NodeTaint{}
	for _, taint := range spec.Taints {
		if taint.Key == gpu.ResourceNvidiaGPU {
			continue
		}
		effect, found := taintEffectsMap[taint.Effect]
		if !found {
			effect = "EFFECT_UNSPECIFIED"
		}
		taint := &gke_api_beta.NodeTaint{
			Effect: effect,
			Key:    taint.Key,
			Value:  taint.Value,
		}
		taints = append(taints, taint)
	}
	labels := make(map[string]string)
	for k, v := range spec.Labels {
		if k != gpu.GPULabel {
			labels[k] = v
		}
	}

	config := gke_api_beta.NodeConfig{
		MachineType:  spec.MachineType,
		OauthScopes:  defaultOAuthScopes,
		Labels:       labels,
		Accelerators: accelerators,
		Taints:       taints,
	}

	autoscaling := gke_api_beta.NodePoolAutoscaling{
		Enabled:         true,
		MinNodeCount:    napMinNodes,
		MaxNodeCount:    napMaxNodes,
		Autoprovisioned: true,
	}

	createRequest := gke_api_beta.CreateNodePoolRequest{
		NodePool: &gke_api_beta.NodePool{
			Name:             mig.NodePoolName(),
			InitialNodeCount: 0,
			Config:           &config,
			Autoscaling:      &autoscaling,
		},
	}
	createOp, err := m.gkeBetaService.Projects.Locations.Clusters.NodePools.Create(
		m.clusterPath, &createRequest).Do()
	if err != nil {
		return err
	}
	return m.waitForGkeOp(createOp)
}

func (m *autoscalingGkeClientV1beta1) waitForGkeOp(op *gke_api_beta.Operation) error {
	for start := time.Now(); time.Since(start) < m.operationWaitTimeout; time.Sleep(m.operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s", op.TargetLink, op.Name)
		if op, err := m.gkeBetaService.Projects.Locations.Operations.Get(
			fmt.Sprintf(m.operationPath, op.Name)).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s status: %s", op.TargetLink, op.Name, op.Status)
			if op.Status == "DONE" {
				return nil
			}
		} else {
			glog.Warningf("Error while getting operation %s on %s: %v", op.Name, op.TargetLink, err)
		}
	}
	return fmt.Errorf("Timeout while waiting for operation %s on %s to complete.", op.Name, op.TargetLink)
}
