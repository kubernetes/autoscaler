/*
Copyright 2023 The Kubernetes Authors.

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

package upcloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/request"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/config/dynamic"
)

type upCloudService interface {
	GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error)
	GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error)
	GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error)
	ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error
	GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error)
}

// manager manages node group cache
type manager struct {
	clusterID      uuid.UUID
	svc            upCloudService
	nodeGroups     []*upCloudNodeGroup
	nodeGroupSpecs map[string]dynamic.NodeGroupSpec

	maxNodesTotal int

	mu sync.Mutex
}

// refresh updates manager's node group cache
func (m *manager) refresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, timeoutGetRequest)
	defer cancel()
	groups := make([]*upCloudNodeGroup, 0)
	upcloudNodeGroups, err := m.svc.GetKubernetesNodeGroups(ctx, &request.GetKubernetesNodeGroupsRequest{
		ClusterUUID: m.clusterID.String(),
	})
	if err != nil {
		return err
	}
	for _, g := range upcloudNodeGroups {
		nodes, err := nodeGroupNodes(ctx, m.svc, m.clusterID, g.Name)
		if err != nil {
			klog.ErrorS(err, "failed to get node group nodes")
			return err
		}
		group := upCloudNodeGroup{
			clusterID: m.clusterID,
			name:      g.Name,
			size:      g.Count,
			minSize:   nodeGroupMinSize,
			maxSize:   m.maxNodesTotal,
			svc:       m.svc,
			nodes:     nodes,
			mu:        sync.Mutex{},
		}
		if spec, ok := m.nodeGroupSpecs[group.name]; ok && spec.Name == group.name {
			group.minSize = spec.MinSize
			group.maxSize = spec.MaxSize
		}
		klog.V(logInfo).Infof("caching cluster %s node group %s size=%d minSize=%d maxSize=%d nodes=%d",
			m.clusterID.String(), group.name, group.size, group.minSize, group.maxSize, len(nodes))
		groups = append(groups, &group)
	}
	m.nodeGroups = groups
	klog.V(logInfo).Infof("refreshed node groups (%d)", len(m.nodeGroups))
	return nil
}

func newManager(ctx context.Context, svc upCloudService, cfg upCloudConfig, opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions) (*manager, error) {
	clusterUUID, err := uuid.Parse(cfg.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster ID %s is not valid UUID %w", envUpCloudClusterID, err)
	}

	maxNodesTotal, err := clusterMaxNodes(ctx, svc, clusterUUID, opts.MaxNodesTotal)
	if err != nil {
		return nil, err
	}
	nodeGroupSpecs, err := nodeGroupSpecsFromDiscoveryOptions(&do, nodeGroupMinSize == 0, maxNodesTotal)
	if err != nil {
		return nil, err
	}

	return &manager{
		clusterID:      clusterUUID,
		maxNodesTotal:  maxNodesTotal,
		svc:            svc,
		nodeGroups:     make([]*upCloudNodeGroup, 0),
		nodeGroupSpecs: nodeGroupSpecs,
		mu:             sync.Mutex{},
	}, nil
}

func nodeGroupSpecsFromDiscoveryOptions(do *cloudprovider.NodeGroupDiscoveryOptions, supportScaleToZero bool, maxNodesTotal int) (map[string]dynamic.NodeGroupSpec, error) {
	specs := make(map[string]dynamic.NodeGroupSpec)
	if do == nil || len(do.NodeGroupSpecs) == 0 {
		return specs, nil
	}
	for _, spec := range do.NodeGroupSpecs {
		s, err := dynamic.SpecFromString(spec, supportScaleToZero)
		if err != nil {
			return specs, fmt.Errorf("failed to parse node group spec, format should be `<minSize>:<maxSize>:<nodeGroupName>`: %v", err)
		}
		if s.MaxSize > maxNodesTotal {
			return specs, fmt.Errorf("failed to validate node group spec, max size %d is greater than cluster plan maximum %d`", s.MaxSize, maxNodesTotal)
		}
		specs[s.Name] = *s
	}
	return specs, nil
}

func clusterMaxNodes(ctx context.Context, svc upCloudService, clusterID uuid.UUID, requestedMaxNodesTotal int) (int, error) {
	cluster, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{
		UUID: clusterID.String(),
	})
	if err != nil {
		var p *upcloud.Problem
		if !errors.As(err, &p) {
			return requestedMaxNodesTotal, err
		}
		if p.Status == http.StatusForbidden {
			return requestedMaxNodesTotal, fmt.Errorf("unable to get cluster %s info, permission denied", clusterID.String())
		}
		if p.Status == http.StatusNotFound {
			return requestedMaxNodesTotal, fmt.Errorf("cluster %s not found", clusterID.String())
		}
		return requestedMaxNodesTotal, err
	}

	plan, err := clusterPlanByName(ctx, svc, cluster.Plan)
	if err != nil {
		return requestedMaxNodesTotal, err
	}

	if requestedMaxNodesTotal == 0 {
		return plan.MaxNodes, nil
	}
	if requestedMaxNodesTotal > plan.MaxNodes {
		return requestedMaxNodesTotal, fmt.Errorf("MaxNodesTotal %d is greater than maximum allowed %d with selected %s cluster plan", requestedMaxNodesTotal, plan.MaxNodes, cluster.Plan)
	}
	return requestedMaxNodesTotal, nil
}

func clusterPlanByName(ctx context.Context, svc upCloudService, name string) (upcloud.KubernetesPlan, error) {
	plans, err := svc.GetKubernetesPlans(ctx, &request.GetKubernetesPlansRequest{})
	if err != nil {
		return upcloud.KubernetesPlan{}, err
	}
	for i := range plans {
		if strings.EqualFold(plans[i].Name, name) {
			return plans[i], nil
		}
	}
	return upcloud.KubernetesPlan{}, fmt.Errorf("can't get cluster plan by name '%s'", name)
}

func nodeGroupNodes(ctx context.Context, svc upCloudService, clusterID uuid.UUID, name string) ([]cloudprovider.Instance, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutGetRequest)
	defer cancel()
	instances := make([]cloudprovider.Instance, 0)
	klog.V(logInfo).Infof("fetching node group %s/%s details", clusterID.String(), name)
	ng, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: clusterID.String(),
		Name:        name,
	})
	if err != nil {
		return instances, err
	}
	for i := range ng.Nodes {
		node := ng.Nodes[i]
		instances = append(instances, cloudprovider.Instance{
			Id:     fmt.Sprintf("upcloud:////%s", node.UUID),
			Status: nodeStateToInstanceStatus(node.State),
		})
	}
	return instances, err
}

func nodeStateToInstanceStatus(nodeState upcloud.KubernetesNodeState) *cloudprovider.InstanceStatus {
	var s cloudprovider.InstanceState
	var e *cloudprovider.InstanceErrorInfo
	switch nodeState {
	case upcloud.KubernetesNodeStateRunning:
		s = cloudprovider.InstanceRunning
	case upcloud.KubernetesNodeStateTerminating:
		s = cloudprovider.InstanceDeleting
	case upcloud.KubernetesNodeStatePending:
		s = cloudprovider.InstanceCreating
	default:
		e = &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  string(nodeState),
		}
	}
	return &cloudprovider.InstanceStatus{
		State:     s,
		ErrorInfo: e,
	}
}
