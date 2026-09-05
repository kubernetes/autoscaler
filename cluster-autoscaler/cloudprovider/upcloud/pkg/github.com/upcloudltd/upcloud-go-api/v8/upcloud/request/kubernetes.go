package request

import (
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
)

const (
	kubernetesClusterBasePath string = "/kubernetes"
)

// GetKubernetesClustersRequest represents a request to list Kubernetes clusters
type GetKubernetesClustersRequest struct{}

func (r *GetKubernetesClustersRequest) RequestURL() string {
	return kubernetesClusterBasePath
}

// Deprecated: KubernetesFilter filter is deprecated. Use QueryFilter instead.
type KubernetesFilter = QueryFilter

// GetKubernetesClustersWithFiltersRequest represents a request to get all clusters
// using labels or label keys as filters.
// Using multiple filters returns only clusters that match all.
type GetKubernetesClustersWithFiltersRequest struct {
	Filters []QueryFilter
}

// RequestURL implements the Request interface.
func (r *GetKubernetesClustersWithFiltersRequest) RequestURL() string {
	if len(r.Filters) == 0 {
		return kubernetesClusterBasePath
	}
	return fmt.Sprintf("%s?%s", kubernetesClusterBasePath, encodeQueryFilters(r.Filters))
}

// GetKubernetesClusterRequest represents a request to get a Kubernetes cluster details
type GetKubernetesClusterRequest struct {
	UUID string
}

func (r *GetKubernetesClusterRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s", kubernetesClusterBasePath, r.UUID)
}

// CreateKubernetesClusterRequest represents a request to create a Kubernetes cluster
type CreateKubernetesClusterRequest struct {
	ControlPlaneIPFilter []string              `json:"control_plane_ip_filter"`
	Labels               []upcloud.Label       `json:"labels,omitempty"`
	Name                 string                `json:"name"`
	Network              string                `json:"network"`
	NetworkCIDR          string                `json:"network_cidr"`
	NodeGroups           []KubernetesNodeGroup `json:"node_groups"`
	Version              string                `json:"version"`
	Zone                 string                `json:"zone"`
	Plan                 string                `json:"plan,omitempty"`
	PrivateNodeGroups    bool                  `json:"private_node_groups"`
	// The default storage encryption strategy for all node groups (optional).
	StorageEncryption upcloud.StorageEncryption `json:"storage_encryption,omitempty"`
}

func (r *CreateKubernetesClusterRequest) RequestURL() string {
	return kubernetesClusterBasePath
}

type ModifyKubernetesCluster struct {
	ControlPlaneIPFilter *[]string        `json:"control_plane_ip_filter,omitempty"`
	Labels               *[]upcloud.Label `json:"labels,omitempty"`
}

type ModifyKubernetesClusterRequest struct {
	ClusterUUID string `json:"-"`
	Cluster     ModifyKubernetesCluster
}

func (r *ModifyKubernetesClusterRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Cluster)
}

func (r *ModifyKubernetesClusterRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s", kubernetesClusterBasePath, r.ClusterUUID)
}

type GetKubernetesClusterAvailableUpgradesRequest struct {
	ClusterUUID string `json:"-"`
}

func (r *GetKubernetesClusterAvailableUpgradesRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/available-upgrades", kubernetesClusterBasePath, r.ClusterUUID)
}

type UpgradeKubernetesClusterRequest struct {
	ClusterUUID string `json:"-"`
	Upgrade     upcloud.KubernetesClusterUpgrade
}

func (r *UpgradeKubernetesClusterRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Upgrade)
}

func (r *UpgradeKubernetesClusterRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/upgrade", kubernetesClusterBasePath, r.ClusterUUID)
}

// DeleteKubernetesClusterRequest represents a request to delete a Kubernetes cluster
type DeleteKubernetesClusterRequest struct {
	UUID string `json:"-"`
}

func (r *DeleteKubernetesClusterRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s", kubernetesClusterBasePath, r.UUID)
}

// WaitForKubernetesClusterStateRequest represents a request to wait for a Kubernetes cluster
// to enter a desired state
type WaitForKubernetesClusterStateRequest struct {
	DesiredState upcloud.KubernetesClusterState `json:"-"`
	UUID         string                         `json:"-"`
}

func (r *WaitForKubernetesClusterStateRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s", kubernetesClusterBasePath, r.UUID)
}

// WaitForKubernetesNodeGroupStateRequest represents a request to wait for a Kubernetes node group
// to enter a desired state
type WaitForKubernetesNodeGroupStateRequest struct {
	DesiredState upcloud.KubernetesNodeGroupState `json:"-"`
	ClusterUUID  string                           `json:"-"`
	Name         string                           `json:"-"`
}

func (r *WaitForKubernetesNodeGroupStateRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s", kubernetesClusterBasePath, r.ClusterUUID, r.Name)
}

// GetKubernetesKubeconfigRequest represents a request to get kubeconfig for a Kubernetes cluster
type GetKubernetesKubeconfigRequest struct {
	UUID string `json:"-"`
}

func (r *GetKubernetesKubeconfigRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/kubeconfig", kubernetesClusterBasePath, r.UUID)
}

// GetKubernetesVersionsRequest represents a request to list available Kubernetes cluster versions
type GetKubernetesVersionsRequest struct{}

func (r *GetKubernetesVersionsRequest) RequestURL() string {
	return fmt.Sprintf("%s/versions", kubernetesClusterBasePath)
}

type GetKubernetesNodeGroupsRequest struct {
	ClusterUUID string
}

func (r *GetKubernetesNodeGroupsRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups", kubernetesClusterBasePath, r.ClusterUUID)
}

type GetKubernetesNodeGroupRequest struct {
	ClusterUUID string
	Name        string
}

func (r *GetKubernetesNodeGroupRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s", kubernetesClusterBasePath, r.ClusterUUID, r.Name)
}

type KubernetesNodeGroup struct {
	Count                int                            `json:"count,omitempty"`
	Labels               []upcloud.Label                `json:"labels,omitempty"`
	Name                 string                         `json:"name,omitempty"`
	Plan                 string                         `json:"plan,omitempty"`
	SSHKeys              []string                       `json:"ssh_keys,omitempty"`
	Storage              string                         `json:"storage,omitempty"`
	KubeletArgs          []upcloud.KubernetesKubeletArg `json:"kubelet_args,omitempty"`
	Taints               []upcloud.KubernetesTaint      `json:"taints,omitempty"`
	AntiAffinity         bool                           `json:"anti_affinity,omitempty"`
	UtilityNetworkAccess *bool                          `json:"utility_network_access,omitempty"`
	// Node group custom plan properties. Required when plan is set as "custom".
	CustomPlan *upcloud.KubernetesNodeGroupCustomPlan `json:"custom_plan,omitempty"`
	// Node group cloud native plan properties. Optional to customise a cloud native plan.
	CloudNativePlan *upcloud.KubernetesNodeGroupCloudNativePlan `json:"cloud_native_plan,omitempty"`
	// Node group GPU plan properties. Optional to customise a GPU plan.
	GPUPlan *upcloud.KubernetesNodeGroupGPUPlan `json:"gpu_plan,omitempty"`
	// node group storage encryption strategy (optional).
	StorageEncryption upcloud.StorageEncryption `json:"storage_encryption,omitempty"`
}

type CreateKubernetesNodeGroupRequest struct {
	ClusterUUID string `json:"-"`
	NodeGroup   KubernetesNodeGroup
}

func (r *CreateKubernetesNodeGroupRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.NodeGroup)
}

func (r *CreateKubernetesNodeGroupRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups", kubernetesClusterBasePath, r.ClusterUUID)
}

type ModifyKubernetesNodeGroup struct {
	Count int `json:"count,omitempty"`
}

type ModifyKubernetesNodeGroupRequest struct {
	ClusterUUID string `json:"-"`
	Name        string `json:"-"`
	NodeGroup   ModifyKubernetesNodeGroup
}

func (r *ModifyKubernetesNodeGroupRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.NodeGroup)
}

func (r *ModifyKubernetesNodeGroupRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s", kubernetesClusterBasePath, r.ClusterUUID, r.Name)
}

type DeleteKubernetesNodeGroupRequest struct {
	ClusterUUID string
	Name        string
}

func (r *DeleteKubernetesNodeGroupRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s", kubernetesClusterBasePath, r.ClusterUUID, r.Name)
}

type GetKubernetesPlansRequest struct{}

func (r *GetKubernetesPlansRequest) RequestURL() string {
	return fmt.Sprintf("%s/plans", kubernetesClusterBasePath)
}

type DeleteKubernetesNodeGroupNodeRequest struct {
	ClusterUUID string
	Name        string
	NodeName    string
}

func (r *DeleteKubernetesNodeGroupNodeRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s/%s", kubernetesClusterBasePath, r.ClusterUUID, r.Name, r.NodeName)
}
