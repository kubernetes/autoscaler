package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/request"
)

type Kubernetes interface {
	GetKubernetesClusters(ctx context.Context, r *request.GetKubernetesClustersRequest) ([]upcloud.KubernetesCluster, error)
	GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error)
	CreateKubernetesCluster(ctx context.Context, r *request.CreateKubernetesClusterRequest) (*upcloud.KubernetesCluster, error)
	ModifyKubernetesCluster(ctx context.Context, r *request.ModifyKubernetesClusterRequest) (*upcloud.KubernetesCluster, error)
	GetKubernetesClusterAvailableUpgrades(ctx context.Context, r *request.GetKubernetesClusterAvailableUpgradesRequest) (*upcloud.KubernetesClusterAvailableUpgrades, error)
	UpgradeKubernetesCluster(ctx context.Context, r *request.UpgradeKubernetesClusterRequest) (*upcloud.KubernetesClusterUpgrade, error)
	DeleteKubernetesCluster(ctx context.Context, r *request.DeleteKubernetesClusterRequest) error
	GetKubernetesKubeconfig(ctx context.Context, r *request.GetKubernetesKubeconfigRequest) (string, error)
	GetKubernetesVersions(ctx context.Context, r *request.GetKubernetesVersionsRequest) ([]upcloud.KubernetesVersion, error)
	WaitForKubernetesClusterState(ctx context.Context, r *request.WaitForKubernetesClusterStateRequest) (*upcloud.KubernetesCluster, error)
	GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error)
	GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error)
	CreateKubernetesNodeGroup(ctx context.Context, r *request.CreateKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	WaitForKubernetesNodeGroupState(ctx context.Context, r *request.WaitForKubernetesNodeGroupStateRequest) (*upcloud.KubernetesNodeGroup, error)
	DeleteKubernetesNodeGroup(ctx context.Context, r *request.DeleteKubernetesNodeGroupRequest) error
	DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error
	GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error)
}

// GetKubernetesClusters retrieves a list of Kubernetes clusters.
func (s *Service) GetKubernetesClusters(ctx context.Context, r *request.GetKubernetesClustersRequest) ([]upcloud.KubernetesCluster, error) {
	clusters := make([]upcloud.KubernetesCluster, 0)
	return clusters, s.get(ctx, r.RequestURL(), &clusters)
}

// GetKubernetesCluster retrieves details of a Kubernetes cluster.
func (s *Service) GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	kubernetesCluster := upcloud.KubernetesCluster{}
	return &kubernetesCluster, s.get(ctx, r.RequestURL(), &kubernetesCluster)
}

// CreateKubernetesCluster creates a new Kubernetes cluster.
func (s *Service) CreateKubernetesCluster(ctx context.Context, r *request.CreateKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	if r == nil || len(r.Network) == 0 {
		return nil, errors.New("bad request, network is not defined")
	}

	networkDetails, err := s.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{UUID: r.Network})

	if err != nil || networkDetails == nil || len(networkDetails.IPNetworks) == 0 {
		return nil, fmt.Errorf("invalid network: %w", err)
	}

	r.NetworkCIDR = networkDetails.IPNetworks[0].Address

	cluster := upcloud.KubernetesCluster{}

	return &cluster, s.create(ctx, r, &cluster)
}

// ModifyKubernetesCluster modifies an existing Kubernetes cluster.
func (s *Service) ModifyKubernetesCluster(ctx context.Context, r *request.ModifyKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	cluster := upcloud.KubernetesCluster{}
	return &cluster, s.modify(ctx, r, &cluster)
}

// GetKubernetesClusterAvailableUpgrades list available upgrades for the cluster.
func (s *Service) GetKubernetesClusterAvailableUpgrades(ctx context.Context, r *request.GetKubernetesClusterAvailableUpgradesRequest) (*upcloud.KubernetesClusterAvailableUpgrades, error) {
	availableUpgrades := upcloud.KubernetesClusterAvailableUpgrades{}
	return &availableUpgrades, s.get(ctx, r.RequestURL(), &availableUpgrades)
}

// UpgradeKubernetesCluster upgrades an existing Kubernetes cluster.
func (s *Service) UpgradeKubernetesCluster(ctx context.Context, r *request.UpgradeKubernetesClusterRequest) (*upcloud.KubernetesClusterUpgrade, error) {
	upgrade := upcloud.KubernetesClusterUpgrade{}
	return &upgrade, s.create(ctx, r, &upgrade)
}

// DeleteKubernetesCluster deletes an existing Kubernetes cluster.
func (s *Service) DeleteKubernetesCluster(ctx context.Context, r *request.DeleteKubernetesClusterRequest) error {
	return s.delete(ctx, r)
}

// WaitForKubernetesClusterState blocks execution until the specified Kubernetes cluster has entered the
// specified state. If the state changes favorably, cluster details are returned. The method will give up
// after the specified timeout
func (s *Service) WaitForKubernetesClusterState(ctx context.Context, r *request.WaitForKubernetesClusterStateRequest) (*upcloud.KubernetesCluster, error) {
	return retry(ctx, func(i int, c context.Context) (*upcloud.KubernetesCluster, error) {
		details, err := s.GetKubernetesCluster(c, &request.GetKubernetesClusterRequest{
			UUID: r.UUID,
		})
		if err != nil {
			// Ignore first two 404 responses to avoid errors caused by possible false NOT_FOUND responses right after cluster has been created.
			var ucErr *upcloud.Problem
			if errors.As(err, &ucErr) && ucErr.Status == http.StatusNotFound && i < 3 {
				log.Printf("ERROR: %+v", err)
			} else {
				return nil, err
			}
		}

		if details.State == r.DesiredState {
			return details, nil
		}
		return nil, nil
	}, nil)
}

// WaitForKubernetesNodeGroupState blocks execution until the specified Kubernetes node group has entered the
// specified state. If the state changes favorably, node group is returned. The method will give up
// after the specified timeout
func (s *Service) WaitForKubernetesNodeGroupState(ctx context.Context, r *request.WaitForKubernetesNodeGroupStateRequest) (*upcloud.KubernetesNodeGroup, error) {
	return retry(ctx, func(i int, c context.Context) (*upcloud.KubernetesNodeGroup, error) {
		ng, err := s.GetKubernetesNodeGroup(c, &request.GetKubernetesNodeGroupRequest{
			ClusterUUID: r.ClusterUUID,
			Name:        r.Name,
		})
		if err != nil {
			// Ignore first two 404 responses to avoid errors caused by possible false NOT_FOUND responses right after cluster has been created.
			var ucErr *upcloud.Problem
			if errors.As(err, &ucErr) && ucErr.Status == http.StatusNotFound && i < 3 {
				log.Printf("ERROR: %+v", err)
			} else {
				return nil, err
			}
		}

		if ng.State == r.DesiredState {
			return &ng.KubernetesNodeGroup, nil
		}
		return nil, nil
	}, nil)
}

// GetKubernetesKubeconfig retrieves kubeconfig of a Kubernetes cluster.
func (s *Service) GetKubernetesKubeconfig(ctx context.Context, r *request.GetKubernetesKubeconfigRequest) (string, error) {
	data := struct {
		Kubeconfig string `json:"kubeconfig"`
	}{}

	_, err := s.WaitForKubernetesClusterState(ctx, &request.WaitForKubernetesClusterStateRequest{
		DesiredState: upcloud.KubernetesClusterStateRunning,
		UUID:         r.UUID,
	})
	if err != nil {
		return "", err
	}

	err = s.get(ctx, r.RequestURL(), &data)
	return data.Kubeconfig, err
}

// GetKubernetesVersions retrieves a list of Kubernetes cluster versions.
func (s *Service) GetKubernetesVersions(ctx context.Context, r *request.GetKubernetesVersionsRequest) ([]upcloud.KubernetesVersion, error) {
	versions := make([]upcloud.KubernetesVersion, 0)
	return versions, s.get(ctx, r.RequestURL(), &versions)
}

// GetKubernetesNodeGroups retrieves a list of Kubernetes cluster node groups.
func (s *Service) GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error) {
	ng := make([]upcloud.KubernetesNodeGroup, 0)
	return ng, s.get(ctx, r.RequestURL(), &ng)
}

// GetKubernetesNodeGroup retrieves details of a node group.
func (s *Service) GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error) {
	ng := upcloud.KubernetesNodeGroupDetails{}
	return &ng, s.get(ctx, r.RequestURL(), &ng)
}

// CreateKubernetesNodeGroup creates a new node group.
func (s *Service) CreateKubernetesNodeGroup(ctx context.Context, r *request.CreateKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	ng := upcloud.KubernetesNodeGroup{}
	return &ng, s.create(ctx, r, &ng)
}

// ModifyKubernetesNodeGroup modifies an existing node group.
func (s *Service) ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	ng := upcloud.KubernetesNodeGroup{}
	return &ng, s.modify(ctx, r, &ng)
}

// DeleteKubernetesNodeGroup deletes an existing node group.
func (s *Service) DeleteKubernetesNodeGroup(ctx context.Context, r *request.DeleteKubernetesNodeGroupRequest) error {
	return s.delete(ctx, r)
}

// DeleteKubernetesNodeGroupNode deletes an existing node from the node group.
func (s *Service) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	return s.delete(ctx, r)
}

// GetKubernetesPlans retrieves a list of Kubernetes plans.
func (s *Service) GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error) {
	plans := make([]upcloud.KubernetesPlan, 0)
	return plans, s.get(ctx, r.RequestURL(), &plans)
}
