package upcloud

type (
	KubernetesClusterState       string
	KubernetesNodeGroupState     string
	KubernetesClusterType        string
	KubernetesClusterTaintEffect string
	KubernetesNodeState          string
	KubernetesUpgradeStrategy    string
)

const (
	KubernetesClusterStatePending     KubernetesClusterState = "pending"
	KubernetesClusterStateRunning     KubernetesClusterState = "running"
	KubernetesClusterStateTerminating KubernetesClusterState = "terminating"
	KubernetesClusterStateTerminated  KubernetesClusterState = "terminated"
	KubernetesClusterStateFailed      KubernetesClusterState = "failed"
	KubernetesClusterStateUnknown     KubernetesClusterState = "unknown"

	KubernetesNodeGroupStatePending     KubernetesNodeGroupState = "pending"
	KubernetesNodeGroupStateRunning     KubernetesNodeGroupState = "running"
	KubernetesNodeGroupStateScalingUp   KubernetesNodeGroupState = "scaling-up"
	KubernetesNodeGroupStateScalingDown KubernetesNodeGroupState = "scaling-down"
	KubernetesNodeGroupStateTerminating KubernetesNodeGroupState = "terminating"
	KubernetesNodeGroupStateFailed      KubernetesNodeGroupState = "failed"
	KubernetesNodeGroupStateUnknown     KubernetesNodeGroupState = "unknown"

	KubernetesClusterTaintEffectNoExecute        KubernetesClusterTaintEffect = "NoExecute"
	KubernetesClusterTaintEffectNoSchedule       KubernetesClusterTaintEffect = "NoSchedule"
	KubernetesClusterTaintEffectPreferNoSchedule KubernetesClusterTaintEffect = "PreferNoSchedule"

	KubernetesNodeStateFailed      KubernetesNodeState = "failed"
	KubernetesNodeStatePending     KubernetesNodeState = "pending"
	KubernetesNodeStateRunning     KubernetesNodeState = "running"
	KubernetesNodeStateTerminating KubernetesNodeState = "terminating"
	KubernetesNodeStateUnknown     KubernetesNodeState = "unknown"

	KubernetesStorageTierHDD      StorageTier = StorageTierHDD
	KubernetesStorageTierMaxIOPS  StorageTier = StorageTierMaxIOPS
	KubernetesStorageTierStandard StorageTier = StorageTierStandard

	KubernetesUpgradeStrategyManual        KubernetesUpgradeStrategy = "manual"
	KubernetesUpgradeStrategyRollingUpdate KubernetesUpgradeStrategy = "rolling-update"
)

type KubernetesCluster struct {
	ControlPlaneIPFilter []string               `json:"control_plane_ip_filter,omitempty"`
	Labels               []Label                `json:"labels,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Network              string                 `json:"network,omitempty"`
	NetworkCIDR          string                 `json:"network_cidr,omitempty"`
	NodeGroups           []KubernetesNodeGroup  `json:"node_groups,omitempty"`
	State                KubernetesClusterState `json:"state,omitempty"`
	UUID                 string                 `json:"uuid,omitempty"`
	Version              string                 `json:"version,omitempty"`
	Zone                 string                 `json:"zone,omitempty"`
	Plan                 string                 `json:"plan,omitempty"`
	PrivateNodeGroups    bool                   `json:"private_node_groups,omitempty"`
	// The default storage encryption strategy for all node groups.
	StorageEncryption StorageEncryption `json:"storage_encryption,omitempty"`
}

type KubernetesClusterAvailableUpgrades struct {
	Versions []string `json:"versions"`
}

type KubernetesClusterUpgradeStrategy struct {
	Type KubernetesUpgradeStrategy `json:"type"`
}

type KubernetesClusterUpgrade struct {
	Version  string                            `json:"version"`
	Strategy *KubernetesClusterUpgradeStrategy `json:"strategy,omitempty"`
}

type KubernetesNodeGroup struct {
	AntiAffinity         bool                                `json:"anti_affinity,omitempty"`
	Count                int                                 `json:"count,omitempty"`
	KubeletArgs          []KubernetesKubeletArg              `json:"kubelet_args,omitempty"`
	Labels               []Label                             `json:"labels,omitempty"`
	Name                 string                              `json:"name,omitempty"`
	Plan                 string                              `json:"plan,omitempty"`
	CustomPlan           *KubernetesNodeGroupCustomPlan      `json:"custom_plan,omitempty"`
	CloudNativePlan      *KubernetesNodeGroupCloudNativePlan `json:"cloud_native_plan,omitempty"`
	GPUPlan              *KubernetesNodeGroupGPUPlan         `json:"gpu_plan,omitempty"`
	SSHKeys              []string                            `json:"ssh_keys,omitempty"`
	State                KubernetesNodeGroupState            `json:"state,omitempty"`
	Storage              string                              `json:"storage,omitempty"`
	Taints               []KubernetesTaint                   `json:"taints,omitempty"`
	UtilityNetworkAccess bool                                `json:"utility_network_access,omitempty"`
	// Node group storage encryption strategy.
	StorageEncryption StorageEncryption `json:"storage_encryption,omitempty"`
}

type KubernetesNodeGroupDetails struct {
	KubernetesNodeGroup

	Nodes []KubernetesNode `json:"nodes,omitempty"`
}

type KubernetesNode struct {
	UUID  string              `json:"uuid,omitempty"`
	Name  string              `json:"name,omitempty"`
	State KubernetesNodeState `json:"state,omitempty"`
}

type KubernetesKubeletArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KubernetesTaint struct {
	Effect KubernetesClusterTaintEffect `json:"effect"`
	Key    string                       `json:"key"`
	Value  string                       `json:"value"`
}

type KubernetesPlan struct {
	Name         string `json:"name"`
	ServerNumber int    `json:"server_number"`
	MaxNodes     int    `json:"max_nodes"`
	Deprecated   bool   `json:"deprecated"`
}

type KubernetesVersion struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

// KubernetesNodeGroupCustomPlan represents custom server plan used for each node in the node group.
type KubernetesNodeGroupCustomPlan struct {
	// The number of CPU cores dedicated to individual node group node
	Cores int `json:"cores"`
	// The amount of memory in megabytes to assign to individual node group node
	Memory int `json:"memory"`
	// The size of the storage device in gigabytes.
	StorageSize int `json:"storage_size"`
	// The storage tier is MaxIOPS® or HDD.
	StorageTier StorageTier `json:"storage_tier,omitempty"`
}

// KubernetesNodeGroupCloudNativePlan represents cloud native plan properties used for each node in the node group.
type KubernetesNodeGroupCloudNativePlan struct {
	// The size of the storage device in gigabytes.
	StorageSize int `json:"storage_size"`
	// The storage tier is MaxIOPS, HDD, or Standard.
	StorageTier StorageTier `json:"storage_tier,omitempty"`
}

// KubernetesNodeGroupGPUPlan represents GPU plan properties used for each node in the node group.
type KubernetesNodeGroupGPUPlan struct {
	// The size of the storage device in gigabytes.
	StorageSize int `json:"storage_size"`
	// The storage tier is MaxIOPS, HDD, or Standard.
	StorageTier StorageTier `json:"storage_tier,omitempty"`
}
