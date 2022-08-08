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

package v20180525

import (
	"encoding/json"

	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type AcquireClusterAdminRoleRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *AcquireClusterAdminRoleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AcquireClusterAdminRoleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "AcquireClusterAdminRoleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type AcquireClusterAdminRoleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AcquireClusterAdminRoleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AcquireClusterAdminRoleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AddClusterCIDRRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 增加的ClusterCIDR
	ClusterCIDRs []*string `json:"ClusterCIDRs,omitempty" name:"ClusterCIDRs"`

	// 是否忽略ClusterCIDR与VPC路由表的冲突
	IgnoreClusterCIDRConflict *bool `json:"IgnoreClusterCIDRConflict,omitempty" name:"IgnoreClusterCIDRConflict"`
}

func (r *AddClusterCIDRRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddClusterCIDRRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterCIDRs")
	delete(f, "IgnoreClusterCIDRConflict")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "AddClusterCIDRRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type AddClusterCIDRResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AddClusterCIDRResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddClusterCIDRResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AddExistedInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 实例列表，不支持竞价实例
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 实例额外需要设置参数信息(默认值)
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 增强服务。通过该参数可以指定是否开启云安全、云监控等服务。若不指定该参数，则默认开启云监控、云安全服务。
	EnhancedService *EnhancedService `json:"EnhancedService,omitempty" name:"EnhancedService"`

	// 节点登录信息（目前仅支持使用Password或者单个KeyIds）
	LoginSettings *LoginSettings `json:"LoginSettings,omitempty" name:"LoginSettings"`

	// 重装系统时，可以指定修改实例的HostName(集群为HostName模式时，此参数必传，规则名称除不支持大写字符外与[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口HostName一致)
	HostName *string `json:"HostName,omitempty" name:"HostName"`

	// 实例所属安全组。该参数可以通过调用 DescribeSecurityGroups 的返回值中的sgId字段来获取。若不指定该参数，则绑定默认安全组。（目前仅支持设置单个sgId）
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`

	// 节点池选项
	NodePool *NodePoolOption `json:"NodePool,omitempty" name:"NodePool"`

	// 校验规则相关选项，可配置跳过某些校验规则。目前支持GlobalRouteCIDRCheck（跳过GlobalRouter的相关校验），VpcCniCIDRCheck（跳过VpcCni相关校验）
	SkipValidateOptions []*string `json:"SkipValidateOptions,omitempty" name:"SkipValidateOptions"`

	// 参数InstanceAdvancedSettingsOverride数组用于定制化地配置各台instance，与InstanceIds顺序对应。当传入InstanceAdvancedSettingsOverrides数组时，将覆盖默认参数InstanceAdvancedSettings；当没有传入参数InstanceAdvancedSettingsOverrides时，InstanceAdvancedSettings参数对每台instance生效。
	//
	// 参数InstanceAdvancedSettingsOverride数组的长度应与InstanceIds数组一致；当长度大于InstanceIds数组长度时将报错；当长度小于InstanceIds数组时，没有对应配置的instace将使用默认配置。
	InstanceAdvancedSettingsOverrides []*InstanceAdvancedSettings `json:"InstanceAdvancedSettingsOverrides,omitempty" name:"InstanceAdvancedSettingsOverrides"`
}

func (r *AddExistedInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddExistedInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "InstanceIds")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "EnhancedService")
	delete(f, "LoginSettings")
	delete(f, "HostName")
	delete(f, "SecurityGroupIds")
	delete(f, "NodePool")
	delete(f, "SkipValidateOptions")
	delete(f, "InstanceAdvancedSettingsOverrides")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "AddExistedInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type AddExistedInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 失败的节点ID
		// 注意：此字段可能返回 null，表示取不到有效值。
		FailedInstanceIds []*string `json:"FailedInstanceIds,omitempty" name:"FailedInstanceIds"`

		// 成功的节点ID
		// 注意：此字段可能返回 null，表示取不到有效值。
		SuccInstanceIds []*string `json:"SuccInstanceIds,omitempty" name:"SuccInstanceIds"`

		// 超时未返回出来节点的ID(可能失败，也可能成功)
		// 注意：此字段可能返回 null，表示取不到有效值。
		TimeoutInstanceIds []*string `json:"TimeoutInstanceIds,omitempty" name:"TimeoutInstanceIds"`

		// 失败的节点的失败原因
		// 注意：此字段可能返回 null，表示取不到有效值。
		FailedReasons []*string `json:"FailedReasons,omitempty" name:"FailedReasons"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AddExistedInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddExistedInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AddNodeToNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点id
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`
}

func (r *AddNodeToNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddNodeToNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "InstanceIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "AddNodeToNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type AddNodeToNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AddNodeToNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddNodeToNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AddVpcCniSubnetsRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 为集群容器网络增加的子网列表
	SubnetIds []*string `json:"SubnetIds,omitempty" name:"SubnetIds"`

	// 集群所属的VPC的ID
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`
}

func (r *AddVpcCniSubnetsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddVpcCniSubnetsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "SubnetIds")
	delete(f, "VpcId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "AddVpcCniSubnetsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type AddVpcCniSubnetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AddVpcCniSubnetsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *AddVpcCniSubnetsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AppChart struct {

	// chart名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// chart的标签
	// 注意：此字段可能返回 null，表示取不到有效值。
	Label *string `json:"Label,omitempty" name:"Label"`

	// chart的版本
	LatestVersion *string `json:"LatestVersion,omitempty" name:"LatestVersion"`
}

type AutoScalingGroupRange struct {

	// 伸缩组最小实例数
	MinSize *int64 `json:"MinSize,omitempty" name:"MinSize"`

	// 伸缩组最大实例数
	MaxSize *int64 `json:"MaxSize,omitempty" name:"MaxSize"`
}

type AutoscalingAdded struct {

	// 正在加入中的节点数量
	Joining *int64 `json:"Joining,omitempty" name:"Joining"`

	// 初始化中的节点数量
	Initializing *int64 `json:"Initializing,omitempty" name:"Initializing"`

	// 正常的节点数量
	Normal *int64 `json:"Normal,omitempty" name:"Normal"`

	// 节点总数
	Total *int64 `json:"Total,omitempty" name:"Total"`
}

type Capabilities struct {

	// 启用安全能力项列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Add []*string `json:"Add,omitempty" name:"Add"`

	// 禁用安全能力向列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Drop []*string `json:"Drop,omitempty" name:"Drop"`
}

type CbsVolume struct {

	// cbs volume 数据卷名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 腾讯云cbs盘Id
	CbsDiskId *string `json:"CbsDiskId,omitempty" name:"CbsDiskId"`
}

type CheckInstancesUpgradeAbleRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点列表，空为全部节点
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 升级类型
	UpgradeType *string `json:"UpgradeType,omitempty" name:"UpgradeType"`

	// 分页Offset
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`

	// 分页Limit
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`

	// 过滤
	Filter []*Filter `json:"Filter,omitempty" name:"Filter"`
}

func (r *CheckInstancesUpgradeAbleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CheckInstancesUpgradeAbleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "InstanceIds")
	delete(f, "UpgradeType")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "Filter")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CheckInstancesUpgradeAbleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CheckInstancesUpgradeAbleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群master当前小版本
		ClusterVersion *string `json:"ClusterVersion,omitempty" name:"ClusterVersion"`

		// 集群master对应的大版本目前最新小版本
		LatestVersion *string `json:"LatestVersion,omitempty" name:"LatestVersion"`

		// 可升级节点列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		UpgradeAbleInstances []*UpgradeAbleInstancesItem `json:"UpgradeAbleInstances,omitempty" name:"UpgradeAbleInstances"`

		// 总数
		// 注意：此字段可能返回 null，表示取不到有效值。
		Total *int64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CheckInstancesUpgradeAbleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CheckInstancesUpgradeAbleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type Cluster struct {

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 集群描述
	ClusterDescription *string `json:"ClusterDescription,omitempty" name:"ClusterDescription"`

	// 集群版本（默认值为1.10.5）
	ClusterVersion *string `json:"ClusterVersion,omitempty" name:"ClusterVersion"`

	// 集群系统。centos7.2x86_64 或者 ubuntu16.04.1 LTSx86_64，默认取值为ubuntu16.04.1 LTSx86_64
	ClusterOs *string `json:"ClusterOs,omitempty" name:"ClusterOs"`

	// 集群类型，托管集群：MANAGED_CLUSTER，独立集群：INDEPENDENT_CLUSTER。
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`

	// 集群网络相关参数
	ClusterNetworkSettings *ClusterNetworkSettings `json:"ClusterNetworkSettings,omitempty" name:"ClusterNetworkSettings"`

	// 集群当前node数量
	ClusterNodeNum *uint64 `json:"ClusterNodeNum,omitempty" name:"ClusterNodeNum"`

	// 集群所属的项目ID
	ProjectId *uint64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 标签描述列表。
	// 注意：此字段可能返回 null，表示取不到有效值。
	TagSpecification []*TagSpecification `json:"TagSpecification,omitempty" name:"TagSpecification"`

	// 集群状态 (Running 运行中  Creating 创建中 Abnormal 异常  )
	ClusterStatus *string `json:"ClusterStatus,omitempty" name:"ClusterStatus"`

	// 集群属性(包括集群不同属性的MAP，属性字段包括NodeNameType (lan-ip模式和hostname 模式，默认无lan-ip模式))
	// 注意：此字段可能返回 null，表示取不到有效值。
	Property *string `json:"Property,omitempty" name:"Property"`

	// 集群当前master数量
	ClusterMaterNodeNum *uint64 `json:"ClusterMaterNodeNum,omitempty" name:"ClusterMaterNodeNum"`

	// 集群使用镜像id
	// 注意：此字段可能返回 null，表示取不到有效值。
	ImageId *string `json:"ImageId,omitempty" name:"ImageId"`

	// OsCustomizeType 系统定制类型
	// 注意：此字段可能返回 null，表示取不到有效值。
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 集群运行环境docker或container
	// 注意：此字段可能返回 null，表示取不到有效值。
	ContainerRuntime *string `json:"ContainerRuntime,omitempty" name:"ContainerRuntime"`

	// 创建时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// 删除保护开关
	// 注意：此字段可能返回 null，表示取不到有效值。
	DeletionProtection *bool `json:"DeletionProtection,omitempty" name:"DeletionProtection"`

	// 集群是否开启第三方节点支持
	// 注意：此字段可能返回 null，表示取不到有效值。
	EnableExternalNode *bool `json:"EnableExternalNode,omitempty" name:"EnableExternalNode"`
}

type ClusterAdvancedSettings struct {

	// 是否启用IPVS
	IPVS *bool `json:"IPVS,omitempty" name:"IPVS"`

	// 是否启用集群节点自动扩缩容(创建集群流程不支持开启此功能)
	AsEnabled *bool `json:"AsEnabled,omitempty" name:"AsEnabled"`

	// 集群使用的runtime类型，包括"docker"和"containerd"两种类型，默认为"docker"
	ContainerRuntime *string `json:"ContainerRuntime,omitempty" name:"ContainerRuntime"`

	// 集群中节点NodeName类型（包括 hostname,lan-ip两种形式，默认为lan-ip。如果开启了hostname模式，创建节点时需要设置HostName参数，并且InstanceName需要和HostName一致）
	NodeNameType *string `json:"NodeNameType,omitempty" name:"NodeNameType"`

	// 集群自定义参数
	ExtraArgs *ClusterExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`

	// 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR。
	NetworkType *string `json:"NetworkType,omitempty" name:"NetworkType"`

	// 集群VPC-CNI模式是否为非固定IP，默认: FALSE 固定IP。
	IsNonStaticIpMode *bool `json:"IsNonStaticIpMode,omitempty" name:"IsNonStaticIpMode"`

	// 是否启用集群删除保护
	DeletionProtection *bool `json:"DeletionProtection,omitempty" name:"DeletionProtection"`

	// 集群的网络代理模型，目前tke集群支持的网络代理模式有三种：iptables,ipvs,ipvs-bpf，此参数仅在使用ipvs-bpf模式时使用，三种网络模式的参数设置关系如下：
	// iptables模式：IPVS和KubeProxyMode都不设置
	// ipvs模式: 设置IPVS为true, KubeProxyMode不设置
	// ipvs-bpf模式: 设置KubeProxyMode为kube-proxy-bpf
	// 使用ipvs-bpf的网络模式需要满足以下条件：
	// 1. 集群版本必须为1.14及以上；
	// 2. 系统镜像必须是: Tencent Linux 2.4；
	KubeProxyMode *string `json:"KubeProxyMode,omitempty" name:"KubeProxyMode"`

	// 是否开启审计开关
	AuditEnabled *bool `json:"AuditEnabled,omitempty" name:"AuditEnabled"`

	// 审计日志上传到的logset日志集
	AuditLogsetId *string `json:"AuditLogsetId,omitempty" name:"AuditLogsetId"`

	// 审计日志上传到的topic
	AuditLogTopicId *string `json:"AuditLogTopicId,omitempty" name:"AuditLogTopicId"`

	// 区分共享网卡多IP模式和独立网卡模式，共享网卡多 IP 模式填写"tke-route-eni"，独立网卡模式填写"tke-direct-eni"，默认为共享网卡模式
	VpcCniType *string `json:"VpcCniType,omitempty" name:"VpcCniType"`

	// 运行时版本
	RuntimeVersion *string `json:"RuntimeVersion,omitempty" name:"RuntimeVersion"`

	// 是否开节点podCIDR大小的自定义模式
	EnableCustomizedPodCIDR *bool `json:"EnableCustomizedPodCIDR,omitempty" name:"EnableCustomizedPodCIDR"`

	// 自定义模式下的基础pod数量
	BasePodNumber *int64 `json:"BasePodNumber,omitempty" name:"BasePodNumber"`

	// 启用 CiliumMode 的模式，空值表示不启用，“clusterIP” 表示启用 Cilium 支持 ClusterIP
	CiliumMode *string `json:"CiliumMode,omitempty" name:"CiliumMode"`
}

type ClusterAsGroup struct {

	// 伸缩组ID
	AutoScalingGroupId *string `json:"AutoScalingGroupId,omitempty" name:"AutoScalingGroupId"`

	// 伸缩组状态(开启 enabled 开启中 enabling 关闭 disabled 关闭中 disabling 更新中 updating 删除中 deleting 开启缩容中 scaleDownEnabling 关闭缩容中 scaleDownDisabling)
	Status *string `json:"Status,omitempty" name:"Status"`

	// 节点是否设置成不可调度
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsUnschedulable *bool `json:"IsUnschedulable,omitempty" name:"IsUnschedulable"`

	// 伸缩组的label列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 创建时间
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`
}

type ClusterAsGroupAttribute struct {

	// 伸缩组ID
	AutoScalingGroupId *string `json:"AutoScalingGroupId,omitempty" name:"AutoScalingGroupId"`

	// 是否开启
	AutoScalingGroupEnabled *bool `json:"AutoScalingGroupEnabled,omitempty" name:"AutoScalingGroupEnabled"`

	// 伸缩组最大最小实例数
	AutoScalingGroupRange *AutoScalingGroupRange `json:"AutoScalingGroupRange,omitempty" name:"AutoScalingGroupRange"`
}

type ClusterAsGroupOption struct {

	// 是否开启缩容
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsScaleDownEnabled *bool `json:"IsScaleDownEnabled,omitempty" name:"IsScaleDownEnabled"`

	// 多伸缩组情况下扩容选择算法(random 随机选择，most-pods 最多类型的Pod least-waste 最少的资源浪费，默认为random)
	// 注意：此字段可能返回 null，表示取不到有效值。
	Expander *string `json:"Expander,omitempty" name:"Expander"`

	// 最大并发缩容数
	// 注意：此字段可能返回 null，表示取不到有效值。
	MaxEmptyBulkDelete *int64 `json:"MaxEmptyBulkDelete,omitempty" name:"MaxEmptyBulkDelete"`

	// 集群扩容后多少分钟开始判断缩容（默认为10分钟）
	// 注意：此字段可能返回 null，表示取不到有效值。
	ScaleDownDelay *int64 `json:"ScaleDownDelay,omitempty" name:"ScaleDownDelay"`

	// 节点连续空闲多少分钟后被缩容（默认为 10分钟）
	// 注意：此字段可能返回 null，表示取不到有效值。
	ScaleDownUnneededTime *int64 `json:"ScaleDownUnneededTime,omitempty" name:"ScaleDownUnneededTime"`

	// 节点资源使用量低于多少(百分比)时认为空闲(默认: 50(百分比))
	// 注意：此字段可能返回 null，表示取不到有效值。
	ScaleDownUtilizationThreshold *int64 `json:"ScaleDownUtilizationThreshold,omitempty" name:"ScaleDownUtilizationThreshold"`

	// 含有本地存储Pod的节点是否不缩容(默认： FALSE)
	// 注意：此字段可能返回 null，表示取不到有效值。
	SkipNodesWithLocalStorage *bool `json:"SkipNodesWithLocalStorage,omitempty" name:"SkipNodesWithLocalStorage"`

	// 含有kube-system namespace下非DaemonSet管理的Pod的节点是否不缩容 (默认： FALSE)
	// 注意：此字段可能返回 null，表示取不到有效值。
	SkipNodesWithSystemPods *bool `json:"SkipNodesWithSystemPods,omitempty" name:"SkipNodesWithSystemPods"`

	// 计算资源使用量时是否默认忽略DaemonSet的实例(默认值: False，不忽略)
	// 注意：此字段可能返回 null，表示取不到有效值。
	IgnoreDaemonSetsUtilization *bool `json:"IgnoreDaemonSetsUtilization,omitempty" name:"IgnoreDaemonSetsUtilization"`

	// CA做健康性判断的个数，默认3，即超过OkTotalUnreadyCount个数后，CA会进行健康性判断。
	// 注意：此字段可能返回 null，表示取不到有效值。
	OkTotalUnreadyCount *int64 `json:"OkTotalUnreadyCount,omitempty" name:"OkTotalUnreadyCount"`

	// 未就绪节点的最大百分比，此后CA会停止操作
	// 注意：此字段可能返回 null，表示取不到有效值。
	MaxTotalUnreadyPercentage *int64 `json:"MaxTotalUnreadyPercentage,omitempty" name:"MaxTotalUnreadyPercentage"`

	// 表示未准备就绪的节点在有资格进行缩减之前应该停留多长时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	ScaleDownUnreadyTime *int64 `json:"ScaleDownUnreadyTime,omitempty" name:"ScaleDownUnreadyTime"`

	// CA删除未在Kubernetes中注册的节点之前等待的时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	UnregisteredNodeRemovalTime *int64 `json:"UnregisteredNodeRemovalTime,omitempty" name:"UnregisteredNodeRemovalTime"`
}

type ClusterBasicSettings struct {

	// 集群系统。centos7.2x86_64 或者 ubuntu16.04.1 LTSx86_64，默认取值为ubuntu16.04.1 LTSx86_64
	ClusterOs *string `json:"ClusterOs,omitempty" name:"ClusterOs"`

	// 集群版本,默认值为1.10.5
	ClusterVersion *string `json:"ClusterVersion,omitempty" name:"ClusterVersion"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 集群描述
	ClusterDescription *string `json:"ClusterDescription,omitempty" name:"ClusterDescription"`

	// 私有网络ID，形如vpc-xxx。创建托管空集群时必传。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 集群内新增资源所属项目ID。
	ProjectId *int64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 标签描述列表。通过指定该参数可以同时绑定标签到相应的资源实例，当前仅支持绑定标签到集群实例。
	TagSpecification []*TagSpecification `json:"TagSpecification,omitempty" name:"TagSpecification"`

	// 容器的镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 是否开启节点的默认安全组(默认: 否，Aphla特性)
	NeedWorkSecurityGroup *bool `json:"NeedWorkSecurityGroup,omitempty" name:"NeedWorkSecurityGroup"`

	// 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`
}

type ClusterCIDRSettings struct {

	// 用于分配集群容器和服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突。且网段范围必须在内网网段内，例如:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16。
	ClusterCIDR *string `json:"ClusterCIDR,omitempty" name:"ClusterCIDR"`

	// 是否忽略 ClusterCIDR 冲突错误, 默认不忽略
	IgnoreClusterCIDRConflict *bool `json:"IgnoreClusterCIDRConflict,omitempty" name:"IgnoreClusterCIDRConflict"`

	// 集群中每个Node上最大的Pod数量。取值范围4～256。不为2的幂值时会向上取最接近的2的幂值。
	MaxNodePodNum *uint64 `json:"MaxNodePodNum,omitempty" name:"MaxNodePodNum"`

	// 集群最大的service数量。取值范围32～32768，不为2的幂值时会向上取最接近的2的幂值。默认值256
	MaxClusterServiceNum *uint64 `json:"MaxClusterServiceNum,omitempty" name:"MaxClusterServiceNum"`

	// 用于分配集群服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突。且网段范围必须在内网网段内，例如:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16。
	ServiceCIDR *string `json:"ServiceCIDR,omitempty" name:"ServiceCIDR"`

	// VPC-CNI网络模式下，弹性网卡的子网Id。
	EniSubnetIds []*string `json:"EniSubnetIds,omitempty" name:"EniSubnetIds"`

	// VPC-CNI网络模式下，弹性网卡IP的回收时间，取值范围[300,15768000)
	ClaimExpiredSeconds *int64 `json:"ClaimExpiredSeconds,omitempty" name:"ClaimExpiredSeconds"`
}

type ClusterCredential struct {

	// CA 根证书
	CACert *string `json:"CACert,omitempty" name:"CACert"`

	// 认证用的Token
	Token *string `json:"Token,omitempty" name:"Token"`
}

type ClusterExtraArgs struct {

	// kube-apiserver自定义参数，参数格式为["k1=v1", "k1=v2"]， 例如["max-requests-inflight=500","feature-gates=PodShareProcessNamespace=true,DynamicKubeletConfig=true"]
	// 注意：此字段可能返回 null，表示取不到有效值。
	KubeAPIServer []*string `json:"KubeAPIServer,omitempty" name:"KubeAPIServer"`

	// kube-controller-manager自定义参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	KubeControllerManager []*string `json:"KubeControllerManager,omitempty" name:"KubeControllerManager"`

	// kube-scheduler自定义参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	KubeScheduler []*string `json:"KubeScheduler,omitempty" name:"KubeScheduler"`

	// etcd自定义参数，只支持独立集群
	// 注意：此字段可能返回 null，表示取不到有效值。
	Etcd []*string `json:"Etcd,omitempty" name:"Etcd"`
}

type ClusterInternalLB struct {

	// 是否开启内网访问LB
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`

	// 内网访问LB关联的子网Id
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`
}

type ClusterNetworkSettings struct {

	// 用于分配集群容器和服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突
	ClusterCIDR *string `json:"ClusterCIDR,omitempty" name:"ClusterCIDR"`

	// 是否忽略 ClusterCIDR 冲突错误, 默认不忽略
	IgnoreClusterCIDRConflict *bool `json:"IgnoreClusterCIDRConflict,omitempty" name:"IgnoreClusterCIDRConflict"`

	// 集群中每个Node上最大的Pod数量(默认为256)
	MaxNodePodNum *uint64 `json:"MaxNodePodNum,omitempty" name:"MaxNodePodNum"`

	// 集群最大的service数量(默认为256)
	MaxClusterServiceNum *uint64 `json:"MaxClusterServiceNum,omitempty" name:"MaxClusterServiceNum"`

	// 是否启用IPVS(默认不开启)
	Ipvs *bool `json:"Ipvs,omitempty" name:"Ipvs"`

	// 集群的VPCID（如果创建空集群，为必传值，否则自动设置为和集群的节点保持一致）
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 网络插件是否启用CNI(默认开启)
	Cni *bool `json:"Cni,omitempty" name:"Cni"`

	// service的网络模式，当前参数只适用于ipvs+bpf模式
	// 注意：此字段可能返回 null，表示取不到有效值。
	KubeProxyMode *string `json:"KubeProxyMode,omitempty" name:"KubeProxyMode"`

	// 用于分配service的IP range，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突
	// 注意：此字段可能返回 null，表示取不到有效值。
	ServiceCIDR *string `json:"ServiceCIDR,omitempty" name:"ServiceCIDR"`

	// 集群关联的容器子网
	// 注意：此字段可能返回 null，表示取不到有效值。
	Subnets []*string `json:"Subnets,omitempty" name:"Subnets"`
}

type ClusterPublicLB struct {

	// 是否开启公网访问LB
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`

	// 允许访问的来源CIDR列表
	AllowFromCidrs []*string `json:"AllowFromCidrs,omitempty" name:"AllowFromCidrs"`

	// 安全策略放通单个IP或CIDR(例如: "192.168.1.0/24",默认为拒绝所有)
	SecurityPolicies []*string `json:"SecurityPolicies,omitempty" name:"SecurityPolicies"`

	// 外网访问相关的扩展参数，格式为json
	ExtraParam *string `json:"ExtraParam,omitempty" name:"ExtraParam"`

	// 新内外网功能，需要传递安全组
	SecurityGroup *string `json:"SecurityGroup,omitempty" name:"SecurityGroup"`
}

type ClusterVersion struct {

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群主版本号列表，例如1.18.4
	Versions []*string `json:"Versions,omitempty" name:"Versions"`
}

type CommonName struct {

	// 子账户UIN
	SubaccountUin *string `json:"SubaccountUin,omitempty" name:"SubaccountUin"`

	// 子账户客户端证书中的CommonName字段
	CN *string `json:"CN,omitempty" name:"CN"`
}

type Container struct {

	// 镜像
	Image *string `json:"Image,omitempty" name:"Image"`

	// 容器名
	Name *string `json:"Name,omitempty" name:"Name"`

	// 容器启动命令
	Commands []*string `json:"Commands,omitempty" name:"Commands"`

	// 容器启动参数
	Args []*string `json:"Args,omitempty" name:"Args"`

	// 容器内操作系统的环境变量
	EnvironmentVars []*EnvironmentVariable `json:"EnvironmentVars,omitempty" name:"EnvironmentVars"`

	// CPU，制改容器最多可使用的核数，该值不可超过容器实例的总核数。单位：核。
	Cpu *float64 `json:"Cpu,omitempty" name:"Cpu"`

	// 内存，限制该容器最多可使用的内存值，该值不可超过容器实例的总内存值。单位：GiB
	Memory *float64 `json:"Memory,omitempty" name:"Memory"`

	// 数据卷挂载信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	VolumeMounts []*VolumeMount `json:"VolumeMounts,omitempty" name:"VolumeMounts"`

	// 当前状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	CurrentState *ContainerState `json:"CurrentState,omitempty" name:"CurrentState"`

	// 重启次数
	// 注意：此字段可能返回 null，表示取不到有效值。
	RestartCount *uint64 `json:"RestartCount,omitempty" name:"RestartCount"`

	// 容器工作目录
	// 注意：此字段可能返回 null，表示取不到有效值。
	WorkingDir *string `json:"WorkingDir,omitempty" name:"WorkingDir"`

	// 存活探针
	// 注意：此字段可能返回 null，表示取不到有效值。
	LivenessProbe *LivenessOrReadinessProbe `json:"LivenessProbe,omitempty" name:"LivenessProbe"`

	// 就绪探针
	// 注意：此字段可能返回 null，表示取不到有效值。
	ReadinessProbe *LivenessOrReadinessProbe `json:"ReadinessProbe,omitempty" name:"ReadinessProbe"`

	// Gpu限制
	// 注意：此字段可能返回 null，表示取不到有效值。
	GpuLimit *uint64 `json:"GpuLimit,omitempty" name:"GpuLimit"`
}

type ContainerState struct {

	// 容器运行开始时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

	// 容器状态：created, running, exited, unknown
	State *string `json:"State,omitempty" name:"State"`

	// 容器运行结束时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	FinishTime *string `json:"FinishTime,omitempty" name:"FinishTime"`

	// 容器运行退出码
	// 注意：此字段可能返回 null，表示取不到有效值。
	ExitCode *int64 `json:"ExitCode,omitempty" name:"ExitCode"`

	// 容器状态 Reason
	// 注意：此字段可能返回 null，表示取不到有效值。
	Reason *string `json:"Reason,omitempty" name:"Reason"`

	// 容器状态信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Message *string `json:"Message,omitempty" name:"Message"`

	// 容器重启次数
	// 注意：此字段可能返回 null，表示取不到有效值。
	RestartCount *int64 `json:"RestartCount,omitempty" name:"RestartCount"`
}

type ControllerStatus struct {

	// 控制器的名字
	Name *string `json:"Name,omitempty" name:"Name"`

	// 控制器是否开启
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

type CreateClusterAsGroupRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 伸缩组创建透传参数，json化字符串格式，详见[伸缩组创建实例](https://cloud.tencent.com/document/api/377/20440)接口。LaunchConfigurationId由LaunchConfigurePara参数创建，不支持填写
	AutoScalingGroupPara *string `json:"AutoScalingGroupPara,omitempty" name:"AutoScalingGroupPara"`

	// 启动配置创建透传参数，json化字符串格式，详见[创建启动配置](https://cloud.tencent.com/document/api/377/20447)接口。另外ImageId参数由于集群维度已经有的ImageId信息，这个字段不需要填写。UserData字段设置通过UserScript设置，这个字段不需要填写。
	LaunchConfigurePara *string `json:"LaunchConfigurePara,omitempty" name:"LaunchConfigurePara"`

	// 节点高级配置信息
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 节点Label数组
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`
}

func (r *CreateClusterAsGroupRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterAsGroupRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "AutoScalingGroupPara")
	delete(f, "LaunchConfigurePara")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "Labels")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterAsGroupRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterAsGroupResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 启动配置ID
		LaunchConfigurationId *string `json:"LaunchConfigurationId,omitempty" name:"LaunchConfigurationId"`

		// 伸缩组ID
		AutoScalingGroupId *string `json:"AutoScalingGroupId,omitempty" name:"AutoScalingGroupId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterAsGroupResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterAsGroupResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterEndpointRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群端口所在的子网ID  (仅在开启非外网访问时需要填，必须为集群所在VPC内的子网)
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 是否为外网访问（TRUE 外网访问 FALSE 内网访问，默认值： FALSE）
	IsExtranet *bool `json:"IsExtranet,omitempty" name:"IsExtranet"`
}

func (r *CreateClusterEndpointRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterEndpointRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "SubnetId")
	delete(f, "IsExtranet")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterEndpointRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterEndpointResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterEndpointResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterEndpointResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterEndpointVipRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 安全策略放通单个IP或CIDR(例如: "192.168.1.0/24",默认为拒绝所有)
	SecurityPolicies []*string `json:"SecurityPolicies,omitempty" name:"SecurityPolicies"`
}

func (r *CreateClusterEndpointVipRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterEndpointVipRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "SecurityPolicies")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterEndpointVipRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterEndpointVipResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 请求任务的FlowId
		RequestFlowId *int64 `json:"RequestFlowId,omitempty" name:"RequestFlowId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterEndpointVipResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterEndpointVipResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群 ID，请填写 查询集群列表 接口中返回的 clusterId 字段
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// CVM创建透传参数，json化字符串格式，如需要保证扩展集群节点请求幂等性需要在此参数添加ClientToken字段，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口。
	RunInstancePara *string `json:"RunInstancePara,omitempty" name:"RunInstancePara"`

	// 实例额外需要设置参数信息
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 校验规则相关选项，可配置跳过某些校验规则。目前支持GlobalRouteCIDRCheck（跳过GlobalRouter的相关校验），VpcCniCIDRCheck（跳过VpcCni相关校验）
	SkipValidateOptions []*string `json:"SkipValidateOptions,omitempty" name:"SkipValidateOptions"`
}

func (r *CreateClusterInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "RunInstancePara")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "SkipValidateOptions")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 节点实例ID
		InstanceIdSet []*string `json:"InstanceIdSet,omitempty" name:"InstanceIdSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterNodePoolFromExistingAsgRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 伸缩组ID
	AutoscalingGroupId *string `json:"AutoscalingGroupId,omitempty" name:"AutoscalingGroupId"`
}

func (r *CreateClusterNodePoolFromExistingAsgRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterNodePoolFromExistingAsgRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "AutoscalingGroupId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterNodePoolFromExistingAsgRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterNodePoolFromExistingAsgResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 节点池ID
		NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterNodePoolFromExistingAsgResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterNodePoolFromExistingAsgResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterNodePoolRequest struct {
	*tchttp.BaseRequest

	// cluster id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// AutoScalingGroupPara AS组参数
	AutoScalingGroupPara *string `json:"AutoScalingGroupPara,omitempty" name:"AutoScalingGroupPara"`

	// LaunchConfigurePara 运行参数
	LaunchConfigurePara *string `json:"LaunchConfigurePara,omitempty" name:"LaunchConfigurePara"`

	// InstanceAdvancedSettings 示例参数
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 是否启用自动伸缩
	EnableAutoscale *bool `json:"EnableAutoscale,omitempty" name:"EnableAutoscale"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// Labels标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// Taints互斥
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 节点池os
	NodePoolOs *string `json:"NodePoolOs,omitempty" name:"NodePoolOs"`

	// 容器的镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 资源标签
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`
}

func (r *CreateClusterNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "AutoScalingGroupPara")
	delete(f, "LaunchConfigurePara")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "EnableAutoscale")
	delete(f, "Name")
	delete(f, "Labels")
	delete(f, "Taints")
	delete(f, "NodePoolOs")
	delete(f, "OsCustomizeType")
	delete(f, "Tags")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 节点池id
		NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterRequest struct {
	*tchttp.BaseRequest

	// 集群容器网络配置信息
	ClusterCIDRSettings *ClusterCIDRSettings `json:"ClusterCIDRSettings,omitempty" name:"ClusterCIDRSettings"`

	// 集群类型，托管集群：MANAGED_CLUSTER，独立集群：INDEPENDENT_CLUSTER。
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`

	// CVM创建透传参数，json化字符串格式，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口。总机型(包括地域)数量不超过10个，相同机型(地域)购买多台机器可以通过设置参数中RunInstances中InstanceCount来实现。
	RunInstancesForNode []*RunInstancesForNode `json:"RunInstancesForNode,omitempty" name:"RunInstancesForNode"`

	// 集群的基本配置信息
	ClusterBasicSettings *ClusterBasicSettings `json:"ClusterBasicSettings,omitempty" name:"ClusterBasicSettings"`

	// 集群高级配置信息
	ClusterAdvancedSettings *ClusterAdvancedSettings `json:"ClusterAdvancedSettings,omitempty" name:"ClusterAdvancedSettings"`

	// 节点高级配置信息
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 已存在实例的配置信息。所有实例必须在同一个VPC中，最大数量不超过100，不支持添加竞价实例。
	ExistedInstancesForNode []*ExistedInstancesForNode `json:"ExistedInstancesForNode,omitempty" name:"ExistedInstancesForNode"`

	// CVM类型和其对应的数据盘挂载配置信息
	InstanceDataDiskMountSettings []*InstanceDataDiskMountSetting `json:"InstanceDataDiskMountSettings,omitempty" name:"InstanceDataDiskMountSettings"`

	// 需要安装的扩展组件信息
	ExtensionAddons []*ExtensionAddon `json:"ExtensionAddons,omitempty" name:"ExtensionAddons"`
}

func (r *CreateClusterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterCIDRSettings")
	delete(f, "ClusterType")
	delete(f, "RunInstancesForNode")
	delete(f, "ClusterBasicSettings")
	delete(f, "ClusterAdvancedSettings")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "ExistedInstancesForNode")
	delete(f, "InstanceDataDiskMountSettings")
	delete(f, "ExtensionAddons")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群ID
		ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterRouteRequest struct {
	*tchttp.BaseRequest

	// 路由表名称。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 目的端CIDR。
	DestinationCidrBlock *string `json:"DestinationCidrBlock,omitempty" name:"DestinationCidrBlock"`

	// 下一跳地址。
	GatewayIp *string `json:"GatewayIp,omitempty" name:"GatewayIp"`
}

func (r *CreateClusterRouteRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterRouteRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableName")
	delete(f, "DestinationCidrBlock")
	delete(f, "GatewayIp")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterRouteRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterRouteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterRouteResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterRouteResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterRouteTableRequest struct {
	*tchttp.BaseRequest

	// 路由表名称
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 路由表CIDR
	RouteTableCidrBlock *string `json:"RouteTableCidrBlock,omitempty" name:"RouteTableCidrBlock"`

	// 路由表绑定的VPC
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 是否忽略CIDR冲突
	IgnoreClusterCidrConflict *int64 `json:"IgnoreClusterCidrConflict,omitempty" name:"IgnoreClusterCidrConflict"`
}

func (r *CreateClusterRouteTableRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterRouteTableRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableName")
	delete(f, "RouteTableCidrBlock")
	delete(f, "VpcId")
	delete(f, "IgnoreClusterCidrConflict")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateClusterRouteTableRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateClusterRouteTableResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateClusterRouteTableResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateClusterRouteTableResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateEKSClusterRequest struct {
	*tchttp.BaseRequest

	// k8s版本号。可为1.14.4, 1.12.8。
	K8SVersion *string `json:"K8SVersion,omitempty" name:"K8SVersion"`

	// vpc 的Id
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 子网Id 列表
	SubnetIds []*string `json:"SubnetIds,omitempty" name:"SubnetIds"`

	// 集群描述信息
	ClusterDesc *string `json:"ClusterDesc,omitempty" name:"ClusterDesc"`

	// Serivce 所在子网Id
	ServiceSubnetId *string `json:"ServiceSubnetId,omitempty" name:"ServiceSubnetId"`

	// 集群自定义的Dns服务器信息
	DnsServers []*DnsServerConf `json:"DnsServers,omitempty" name:"DnsServers"`

	// 扩展参数。须是map[string]string 的json 格式。
	ExtraParam *string `json:"ExtraParam,omitempty" name:"ExtraParam"`

	// 是否在用户集群内开启Dns。默认为true
	EnableVpcCoreDNS *bool `json:"EnableVpcCoreDNS,omitempty" name:"EnableVpcCoreDNS"`

	// 标签描述列表。通过指定该参数可以同时绑定标签到相应的资源实例，当前仅支持绑定标签到集群实例。
	TagSpecification []*TagSpecification `json:"TagSpecification,omitempty" name:"TagSpecification"`
}

func (r *CreateEKSClusterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateEKSClusterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "K8SVersion")
	delete(f, "VpcId")
	delete(f, "ClusterName")
	delete(f, "SubnetIds")
	delete(f, "ClusterDesc")
	delete(f, "ServiceSubnetId")
	delete(f, "DnsServers")
	delete(f, "ExtraParam")
	delete(f, "EnableVpcCoreDNS")
	delete(f, "TagSpecification")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateEKSClusterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateEKSClusterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 弹性集群Id
		ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateEKSClusterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateEKSClusterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateEKSContainerInstancesRequest struct {
	*tchttp.BaseRequest

	// 容器组
	Containers []*Container `json:"Containers,omitempty" name:"Containers"`

	// EKS Container Instance容器实例名称
	EksCiName *string `json:"EksCiName,omitempty" name:"EksCiName"`

	// 指定新创建实例所属于的安全组Id
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`

	// 实例所属子网Id
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 实例所属VPC的Id
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 内存，单位：GiB。可参考[资源规格](https://cloud.tencent.com/document/product/457/39808)文档
	Memory *float64 `json:"Memory,omitempty" name:"Memory"`

	// CPU，单位：核。可参考[资源规格](https://cloud.tencent.com/document/product/457/39808)文档
	Cpu *float64 `json:"Cpu,omitempty" name:"Cpu"`

	// 实例重启策略： Always(总是重启)、Never(从不重启)、OnFailure(失败时重启)，默认：Always。
	RestartPolicy *string `json:"RestartPolicy,omitempty" name:"RestartPolicy"`

	// 镜像仓库凭证数组
	ImageRegistryCredentials []*ImageRegistryCredential `json:"ImageRegistryCredentials,omitempty" name:"ImageRegistryCredentials"`

	// 数据卷，包含NfsVolume数组和CbsVolume数组
	EksCiVolume *EksCiVolume `json:"EksCiVolume,omitempty" name:"EksCiVolume"`

	// 实例副本数，默认为1
	Replicas *int64 `json:"Replicas,omitempty" name:"Replicas"`

	// Init 容器
	InitContainers []*Container `json:"InitContainers,omitempty" name:"InitContainers"`

	// 自定义DNS配置
	DnsConfig *DNSConfig `json:"DnsConfig,omitempty" name:"DnsConfig"`

	// 用来绑定容器实例的已有EIP的列表。如传值，需要保证数值和Replicas相等。
	// 另外此参数和AutoCreateEipAttribute互斥。
	ExistedEipIds []*string `json:"ExistedEipIds,omitempty" name:"ExistedEipIds"`

	// 自动创建EIP的可选参数。若传此参数，则会自动创建EIP。
	// 另外此参数和ExistedEipIds互斥
	AutoCreateEipAttribute *EipAttribute `json:"AutoCreateEipAttribute,omitempty" name:"AutoCreateEipAttribute"`

	// 是否为容器实例自动创建EIP，默认为false。若传true，则此参数和ExistedEipIds互斥
	AutoCreateEip *bool `json:"AutoCreateEip,omitempty" name:"AutoCreateEip"`

	// Pod 所需的 CPU 资源型号，如果不填写则默认不强制指定 CPU 类型。目前支持型号如下：
	// intel
	// amd
	// - 支持优先级顺序写法，如 “amd,intel” 表示优先创建 amd 资源 Pod，如果所选地域可用区 amd 资源不足，则会创建 intel 资源 Pod。
	CpuType *string `json:"CpuType,omitempty" name:"CpuType"`

	// 容器实例所需的 GPU 资源型号，目前支持型号如下：
	// 1/4\*V100
	// 1/2\*V100
	// V100
	// 1/4\*T4
	// 1/2\*T4
	// T4
	GpuType *string `json:"GpuType,omitempty" name:"GpuType"`

	// Pod 所需的 GPU 数量，如填写，请确保为支持的规格。默认单位为卡，无需再次注明。
	GpuCount *uint64 `json:"GpuCount,omitempty" name:"GpuCount"`

	// 为容器实例关联 CAM 角色，value 填写 CAM 角色名称，容器实例可获取该 CAM 角色包含的权限策略，方便 容器实例 内的程序进行如购买资源、读写存储等云资源操作。
	CamRoleName *string `json:"CamRoleName,omitempty" name:"CamRoleName"`
}

func (r *CreateEKSContainerInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateEKSContainerInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Containers")
	delete(f, "EksCiName")
	delete(f, "SecurityGroupIds")
	delete(f, "SubnetId")
	delete(f, "VpcId")
	delete(f, "Memory")
	delete(f, "Cpu")
	delete(f, "RestartPolicy")
	delete(f, "ImageRegistryCredentials")
	delete(f, "EksCiVolume")
	delete(f, "Replicas")
	delete(f, "InitContainers")
	delete(f, "DnsConfig")
	delete(f, "ExistedEipIds")
	delete(f, "AutoCreateEipAttribute")
	delete(f, "AutoCreateEip")
	delete(f, "CpuType")
	delete(f, "GpuType")
	delete(f, "GpuCount")
	delete(f, "CamRoleName")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateEKSContainerInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateEKSContainerInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// EKS Container Instance Id集合，格式为eksci-xxx，是容器实例的唯一标识。
		EksCiIds []*string `json:"EksCiIds,omitempty" name:"EksCiIds"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateEKSContainerInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateEKSContainerInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusAlertRuleRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 告警配置
	AlertRule *PrometheusAlertRuleDetail `json:"AlertRule,omitempty" name:"AlertRule"`
}

func (r *CreatePrometheusAlertRuleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusAlertRuleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "AlertRule")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreatePrometheusAlertRuleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusAlertRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 告警id
		Id *string `json:"Id,omitempty" name:"Id"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreatePrometheusAlertRuleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusAlertRuleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusDashboardRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 面板组名称
	DashboardName *string `json:"DashboardName,omitempty" name:"DashboardName"`

	// 面板列表
	// 每一项是一个grafana dashboard的json定义
	Contents []*string `json:"Contents,omitempty" name:"Contents"`
}

func (r *CreatePrometheusDashboardRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusDashboardRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "DashboardName")
	delete(f, "Contents")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreatePrometheusDashboardRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusDashboardResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreatePrometheusDashboardResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusDashboardResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusTemplateRequest struct {
	*tchttp.BaseRequest

	// 模板设置
	Template *PrometheusTemplate `json:"Template,omitempty" name:"Template"`
}

func (r *CreatePrometheusTemplateRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusTemplateRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Template")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreatePrometheusTemplateRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreatePrometheusTemplateResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 模板Id
		TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreatePrometheusTemplateResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreatePrometheusTemplateResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DNSConfig struct {

	// DNS 服务器IP地址列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Nameservers []*string `json:"Nameservers,omitempty" name:"Nameservers"`

	// DNS搜索域列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Searches []*string `json:"Searches,omitempty" name:"Searches"`

	// 对象选项列表，每个对象由name和value（可选）构成
	// 注意：此字段可能返回 null，表示取不到有效值。
	Options []*DNSConfigOption `json:"Options,omitempty" name:"Options"`
}

type DNSConfigOption struct {

	// 配置项名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	Name *string `json:"Name,omitempty" name:"Name"`

	// 项值
	// 注意：此字段可能返回 null，表示取不到有效值。
	Value *string `json:"Value,omitempty" name:"Value"`
}

type DataDisk struct {

	// 云盘类型
	// 注意：此字段可能返回 null，表示取不到有效值。
	DiskType *string `json:"DiskType,omitempty" name:"DiskType"`

	// 文件系统(ext3/ext4/xfs)
	// 注意：此字段可能返回 null，表示取不到有效值。
	FileSystem *string `json:"FileSystem,omitempty" name:"FileSystem"`

	// 云盘大小(G）
	// 注意：此字段可能返回 null，表示取不到有效值。
	DiskSize *int64 `json:"DiskSize,omitempty" name:"DiskSize"`

	// 是否自动化格式盘并挂载
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoFormatAndMount *bool `json:"AutoFormatAndMount,omitempty" name:"AutoFormatAndMount"`

	// 挂载目录
	// 注意：此字段可能返回 null，表示取不到有效值。
	MountTarget *string `json:"MountTarget,omitempty" name:"MountTarget"`

	// 挂载设备名或分区名
	// 注意：此字段可能返回 null，表示取不到有效值。
	DiskPartition *string `json:"DiskPartition,omitempty" name:"DiskPartition"`
}

type DeleteClusterAsGroupsRequest struct {
	*tchttp.BaseRequest

	// 集群ID，通过[DescribeClusters](https://cloud.tencent.com/document/api/457/31862)接口获取。
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群伸缩组ID的列表
	AutoScalingGroupIds []*string `json:"AutoScalingGroupIds,omitempty" name:"AutoScalingGroupIds"`

	// 是否保留伸缩组中的节点(默认值： false(不保留))
	KeepInstance *bool `json:"KeepInstance,omitempty" name:"KeepInstance"`
}

func (r *DeleteClusterAsGroupsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterAsGroupsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "AutoScalingGroupIds")
	delete(f, "KeepInstance")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterAsGroupsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterAsGroupsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterAsGroupsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterAsGroupsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterEndpointRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 是否为外网访问（TRUE 外网访问 FALSE 内网访问，默认值： FALSE）
	IsExtranet *bool `json:"IsExtranet,omitempty" name:"IsExtranet"`
}

func (r *DeleteClusterEndpointRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterEndpointRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "IsExtranet")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterEndpointRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterEndpointResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterEndpointResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterEndpointResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterEndpointVipRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DeleteClusterEndpointVipRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterEndpointVipRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterEndpointVipRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterEndpointVipResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterEndpointVipResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterEndpointVipResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 主机InstanceId列表
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 集群实例删除时的策略：terminate（销毁实例，仅支持按量计费云主机实例） retain （仅移除，保留实例）
	InstanceDeleteMode *string `json:"InstanceDeleteMode,omitempty" name:"InstanceDeleteMode"`

	// 是否强制删除(当节点在初始化时，可以指定参数为TRUE)
	ForceDelete *bool `json:"ForceDelete,omitempty" name:"ForceDelete"`
}

func (r *DeleteClusterInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "InstanceIds")
	delete(f, "InstanceDeleteMode")
	delete(f, "ForceDelete")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 删除成功的实例ID列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		SuccInstanceIds []*string `json:"SuccInstanceIds,omitempty" name:"SuccInstanceIds"`

		// 删除失败的实例ID列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		FailedInstanceIds []*string `json:"FailedInstanceIds,omitempty" name:"FailedInstanceIds"`

		// 未匹配到的实例ID列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		NotFoundInstanceIds []*string `json:"NotFoundInstanceIds,omitempty" name:"NotFoundInstanceIds"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterNodePoolRequest struct {
	*tchttp.BaseRequest

	// 节点池对应的 ClusterId
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 需要删除的节点池 Id 列表
	NodePoolIds []*string `json:"NodePoolIds,omitempty" name:"NodePoolIds"`

	// 删除节点池时是否保留节点池内节点(节点仍然会被移出集群，但对应的实例不会被销毁)
	KeepInstance *bool `json:"KeepInstance,omitempty" name:"KeepInstance"`
}

func (r *DeleteClusterNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolIds")
	delete(f, "KeepInstance")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群实例删除时的策略：terminate（销毁实例，仅支持按量计费云主机实例） retain （仅移除，保留实例）
	InstanceDeleteMode *string `json:"InstanceDeleteMode,omitempty" name:"InstanceDeleteMode"`

	// 集群删除时资源的删除策略，目前支持CBS（默认保留CBS）
	ResourceDeleteOptions []*ResourceDeleteOption `json:"ResourceDeleteOptions,omitempty" name:"ResourceDeleteOptions"`
}

func (r *DeleteClusterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "InstanceDeleteMode")
	delete(f, "ResourceDeleteOptions")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterRouteRequest struct {
	*tchttp.BaseRequest

	// 路由表名称。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 下一跳地址。
	GatewayIp *string `json:"GatewayIp,omitempty" name:"GatewayIp"`

	// 目的端CIDR。
	DestinationCidrBlock *string `json:"DestinationCidrBlock,omitempty" name:"DestinationCidrBlock"`
}

func (r *DeleteClusterRouteRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterRouteRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableName")
	delete(f, "GatewayIp")
	delete(f, "DestinationCidrBlock")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterRouteRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterRouteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterRouteResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterRouteResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterRouteTableRequest struct {
	*tchttp.BaseRequest

	// 路由表名称
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`
}

func (r *DeleteClusterRouteTableRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterRouteTableRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableName")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteClusterRouteTableRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteClusterRouteTableResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteClusterRouteTableResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteClusterRouteTableResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteEKSClusterRequest struct {
	*tchttp.BaseRequest

	// 弹性集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DeleteEKSClusterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteEKSClusterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteEKSClusterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteEKSClusterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteEKSClusterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteEKSClusterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteEKSContainerInstancesRequest struct {
	*tchttp.BaseRequest

	// 需要删除的EksCi的Id。 最大数量不超过20
	EksCiIds []*string `json:"EksCiIds,omitempty" name:"EksCiIds"`

	// 是否释放为EksCi自动创建的Eip
	ReleaseAutoCreatedEip *bool `json:"ReleaseAutoCreatedEip,omitempty" name:"ReleaseAutoCreatedEip"`
}

func (r *DeleteEKSContainerInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteEKSContainerInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "EksCiIds")
	delete(f, "ReleaseAutoCreatedEip")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteEKSContainerInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteEKSContainerInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteEKSContainerInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteEKSContainerInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusAlertRuleRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 告警规则id列表
	AlertIds []*string `json:"AlertIds,omitempty" name:"AlertIds"`
}

func (r *DeletePrometheusAlertRuleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusAlertRuleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "AlertIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeletePrometheusAlertRuleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusAlertRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeletePrometheusAlertRuleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusAlertRuleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusTemplateRequest struct {
	*tchttp.BaseRequest

	// 模板id
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`
}

func (r *DeletePrometheusTemplateRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusTemplateRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "TemplateId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeletePrometheusTemplateRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusTemplateResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeletePrometheusTemplateResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusTemplateResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusTemplateSyncRequest struct {
	*tchttp.BaseRequest

	// 模板id
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

	// 取消同步的对象列表
	Targets []*PrometheusTemplateSyncTarget `json:"Targets,omitempty" name:"Targets"`
}

func (r *DeletePrometheusTemplateSyncRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusTemplateSyncRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "TemplateId")
	delete(f, "Targets")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeletePrometheusTemplateSyncRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeletePrometheusTemplateSyncResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeletePrometheusTemplateSyncResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeletePrometheusTemplateSyncResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeAvailableClusterVersionRequest struct {
	*tchttp.BaseRequest

	// 集群 Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群 Id 列表
	ClusterIds []*string `json:"ClusterIds,omitempty" name:"ClusterIds"`
}

func (r *DescribeAvailableClusterVersionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeAvailableClusterVersionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeAvailableClusterVersionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeAvailableClusterVersionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 可升级的集群版本号
		// 注意：此字段可能返回 null，表示取不到有效值。
		Versions []*string `json:"Versions,omitempty" name:"Versions"`

		// 集群信息
		// 注意：此字段可能返回 null，表示取不到有效值。
		Clusters []*ClusterVersion `json:"Clusters,omitempty" name:"Clusters"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeAvailableClusterVersionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeAvailableClusterVersionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAsGroupOptionRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterAsGroupOptionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAsGroupOptionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterAsGroupOptionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAsGroupOptionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群弹性伸缩属性
		// 注意：此字段可能返回 null，表示取不到有效值。
		ClusterAsGroupOption *ClusterAsGroupOption `json:"ClusterAsGroupOption,omitempty" name:"ClusterAsGroupOption"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterAsGroupOptionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAsGroupOptionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAsGroupsRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 伸缩组ID列表，如果为空，表示拉取集群关联的所有伸缩组。
	AutoScalingGroupIds []*string `json:"AutoScalingGroupIds,omitempty" name:"AutoScalingGroupIds"`

	// 偏移量，默认为0。关于Offset的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`

	// 返回数量，默认为20，最大值为100。关于Limit的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`
}

func (r *DescribeClusterAsGroupsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAsGroupsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "AutoScalingGroupIds")
	delete(f, "Offset")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterAsGroupsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAsGroupsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群关联的伸缩组总数
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群关联的伸缩组列表
		ClusterAsGroupSet []*ClusterAsGroup `json:"ClusterAsGroupSet,omitempty" name:"ClusterAsGroupSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterAsGroupsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAsGroupsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAuthenticationOptionsRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterAuthenticationOptionsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAuthenticationOptionsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterAuthenticationOptionsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterAuthenticationOptionsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// ServiceAccount认证配置
		// 注意：此字段可能返回 null，表示取不到有效值。
		ServiceAccounts *ServiceAccountAuthenticationOptions `json:"ServiceAccounts,omitempty" name:"ServiceAccounts"`

		// 最近一次修改操作结果，返回值可能为：Updating，Success，Failed，TimeOut
		// 注意：此字段可能返回 null，表示取不到有效值。
		LatestOperationState *string `json:"LatestOperationState,omitempty" name:"LatestOperationState"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterAuthenticationOptionsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterAuthenticationOptionsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterCommonNamesRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 子账户列表，不可超出最大值50
	SubaccountUins []*string `json:"SubaccountUins,omitempty" name:"SubaccountUins"`

	// 角色ID列表，不可超出最大值50
	RoleIds []*string `json:"RoleIds,omitempty" name:"RoleIds"`
}

func (r *DescribeClusterCommonNamesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterCommonNamesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "SubaccountUins")
	delete(f, "RoleIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterCommonNamesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterCommonNamesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 子账户Uin与其客户端证书的CN字段映射
		CommonNames []*CommonName `json:"CommonNames,omitempty" name:"CommonNames"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterCommonNamesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterCommonNamesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterControllersRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterControllersRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterControllersRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterControllersRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterControllersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 描述集群中各个控制器的状态
		ControllerStatusSet []*ControllerStatus `json:"ControllerStatusSet,omitempty" name:"ControllerStatusSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterControllersResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterControllersResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterEndpointStatusRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 是否为外网访问（TRUE 外网访问 FALSE 内网访问，默认值： FALSE）
	IsExtranet *bool `json:"IsExtranet,omitempty" name:"IsExtranet"`
}

func (r *DescribeClusterEndpointStatusRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterEndpointStatusRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "IsExtranet")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterEndpointStatusRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterEndpointStatusResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 查询集群访问端口状态（Created 开启成功，Creating 开启中，NotFound 未开启）
		// 注意：此字段可能返回 null，表示取不到有效值。
		Status *string `json:"Status,omitempty" name:"Status"`

		// 开启访问入口失败信息
		// 注意：此字段可能返回 null，表示取不到有效值。
		ErrorMsg *string `json:"ErrorMsg,omitempty" name:"ErrorMsg"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterEndpointStatusResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterEndpointStatusResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterEndpointVipStatusRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterEndpointVipStatusRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterEndpointVipStatusRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterEndpointVipStatusRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterEndpointVipStatusResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 端口操作状态 (Creating 创建中  CreateFailed 创建失败 Created 创建完成 Deleting 删除中 DeletedFailed 删除失败 Deleted 已删除 NotFound 未发现操作 )
		Status *string `json:"Status,omitempty" name:"Status"`

		// 操作失败的原因
		// 注意：此字段可能返回 null，表示取不到有效值。
		ErrorMsg *string `json:"ErrorMsg,omitempty" name:"ErrorMsg"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterEndpointVipStatusResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterEndpointVipStatusResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 偏移量，默认为0。关于Offset的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`

	// 返回数量，默认为20，最大值为100。关于Limit的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`

	// 需要获取的节点实例Id列表。如果为空，表示拉取集群下所有节点实例。
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 节点角色, MASTER, WORKER, ETCD, MASTER_ETCD,ALL, 默认为WORKER。默认为WORKER类型。
	InstanceRole *string `json:"InstanceRole,omitempty" name:"InstanceRole"`

	// 过滤条件列表；Name的可选值为nodepool-id、nodepool-instance-type；Name为nodepool-id表示根据节点池id过滤机器，Value的值为具体的节点池id，Name为nodepool-instance-type表示节点加入节点池的方式，Value的值为MANUALLY_ADDED（手动加入节点池）、AUTOSCALING_ADDED（伸缩组扩容方式加入节点池）、ALL（手动加入节点池 和 伸缩组扩容方式加入节点池）
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribeClusterInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "InstanceIds")
	delete(f, "InstanceRole")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群中实例总数
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群中实例列表
		InstanceSet []*Instance `json:"InstanceSet,omitempty" name:"InstanceSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterKubeconfigRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 默认false 获取内网，是否获取外网访问的kubeconfig
	IsExtranet *bool `json:"IsExtranet,omitempty" name:"IsExtranet"`
}

func (r *DescribeClusterKubeconfigRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterKubeconfigRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "IsExtranet")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterKubeconfigRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterKubeconfigResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 子账户kubeconfig文件，可用于直接访问集群kube-apiserver
		Kubeconfig *string `json:"Kubeconfig,omitempty" name:"Kubeconfig"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterKubeconfigResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterKubeconfigResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterNodePoolDetailRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`
}

func (r *DescribeClusterNodePoolDetailRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterNodePoolDetailRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterNodePoolDetailRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterNodePoolDetailResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 节点池详情
		NodePool *NodePool `json:"NodePool,omitempty" name:"NodePool"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterNodePoolDetailResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterNodePoolDetailResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterNodePoolsRequest struct {
	*tchttp.BaseRequest

	// ClusterId（集群id）
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterNodePoolsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterNodePoolsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterNodePoolsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterNodePoolsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// NodePools（节点池列表）
		// 注意：此字段可能返回 null，表示取不到有效值。
		NodePoolSet []*NodePool `json:"NodePoolSet,omitempty" name:"NodePoolSet"`

		// 资源总数
		TotalCount *int64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterNodePoolsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterNodePoolsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterRouteTablesRequest struct {
	*tchttp.BaseRequest
}

func (r *DescribeClusterRouteTablesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterRouteTablesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterRouteTablesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterRouteTablesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 符合条件的实例数量。
		TotalCount *int64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群路由表对象。
		RouteTableSet []*RouteTableInfo `json:"RouteTableSet,omitempty" name:"RouteTableSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterRouteTablesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterRouteTablesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterRoutesRequest struct {
	*tchttp.BaseRequest

	// 路由表名称。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 过滤条件,当前只支持按照单个条件GatewayIP进行过滤（可选）
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribeClusterRoutesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterRoutesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableName")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterRoutesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterRoutesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 符合条件的实例数量。
		TotalCount *int64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群路由对象。
		RouteSet []*RouteInfo `json:"RouteSet,omitempty" name:"RouteSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterRoutesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterRoutesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterSecurityRequest struct {
	*tchttp.BaseRequest

	// 集群 ID，请填写 查询集群列表 接口中返回的 clusterId 字段
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeClusterSecurityRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterSecurityRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClusterSecurityRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClusterSecurityResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群的账号名称
		UserName *string `json:"UserName,omitempty" name:"UserName"`

		// 集群的访问密码
		Password *string `json:"Password,omitempty" name:"Password"`

		// 集群访问CA证书
		CertificationAuthority *string `json:"CertificationAuthority,omitempty" name:"CertificationAuthority"`

		// 集群访问的地址
		ClusterExternalEndpoint *string `json:"ClusterExternalEndpoint,omitempty" name:"ClusterExternalEndpoint"`

		// 集群访问的域名
		Domain *string `json:"Domain,omitempty" name:"Domain"`

		// 集群Endpoint地址
		PgwEndpoint *string `json:"PgwEndpoint,omitempty" name:"PgwEndpoint"`

		// 集群访问策略组
		// 注意：此字段可能返回 null，表示取不到有效值。
		SecurityPolicy []*string `json:"SecurityPolicy,omitempty" name:"SecurityPolicy"`

		// 集群Kubeconfig文件
		// 注意：此字段可能返回 null，表示取不到有效值。
		Kubeconfig *string `json:"Kubeconfig,omitempty" name:"Kubeconfig"`

		// 集群JnsGw的访问地址
		// 注意：此字段可能返回 null，表示取不到有效值。
		JnsGwEndpoint *string `json:"JnsGwEndpoint,omitempty" name:"JnsGwEndpoint"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClusterSecurityResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClusterSecurityResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClustersRequest struct {
	*tchttp.BaseRequest

	// 集群ID列表(为空时，
	// 表示获取账号下所有集群)
	ClusterIds []*string `json:"ClusterIds,omitempty" name:"ClusterIds"`

	// 偏移量,默认0
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`

	// 最大输出条数，默认20，最大为100
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`

	// ·  ClusterName
	//     按照【集群名】进行过滤。
	//     类型：String
	//     必选：否
	//
	// ·  Tags
	//     按照【标签键值对】进行过滤。
	//     类型：String
	//     必选：否
	//
	// ·  vpc-id
	//     按照【VPC】进行过滤。
	//     类型：String
	//     必选：否
	//
	// ·  tag-key
	//     按照【标签键】进行过滤。
	//     类型：String
	//     必选：否
	//
	// ·  tag-value
	//     按照【标签值】进行过滤。
	//     类型：String
	//     必选：否
	//
	// ·  tag:tag-key
	//     按照【标签键值对】进行过滤。
	//     类型：String
	//     必选：否
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`

	// 集群类型，例如：MANAGED_CLUSTER
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`
}

func (r *DescribeClustersRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClustersRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterIds")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "Filters")
	delete(f, "ClusterType")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeClustersRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeClustersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群总个数
		TotalCount *int64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群信息列表
		Clusters []*Cluster `json:"Clusters,omitempty" name:"Clusters"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClustersResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeClustersResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSClusterCredentialRequest struct {
	*tchttp.BaseRequest

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeEKSClusterCredentialRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSClusterCredentialRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEKSClusterCredentialRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSClusterCredentialResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群的接入地址信息
		Addresses []*IPAddress `json:"Addresses,omitempty" name:"Addresses"`

		// 集群的认证信息
		Credential *ClusterCredential `json:"Credential,omitempty" name:"Credential"`

		// 集群的公网访问信息
		PublicLB *ClusterPublicLB `json:"PublicLB,omitempty" name:"PublicLB"`

		// 集群的内网访问信息
		InternalLB *ClusterInternalLB `json:"InternalLB,omitempty" name:"InternalLB"`

		// 标记是否新的内外网功能
		ProxyLB *bool `json:"ProxyLB,omitempty" name:"ProxyLB"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEKSClusterCredentialResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSClusterCredentialResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSClustersRequest struct {
	*tchttp.BaseRequest

	// 集群ID列表(为空时，
	// 表示获取账号下所有集群)
	ClusterIds []*string `json:"ClusterIds,omitempty" name:"ClusterIds"`

	// 偏移量,默认0
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 最大输出条数，默认20
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`

	// 过滤条件,当前只支持按照单个条件ClusterName进行过滤
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribeEKSClustersRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSClustersRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterIds")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEKSClustersRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSClustersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群总个数
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 集群信息列表
		Clusters []*EksCluster `json:"Clusters,omitempty" name:"Clusters"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEKSClustersResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSClustersResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstanceEventRequest struct {
	*tchttp.BaseRequest

	// 容器实例id
	EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

	// 最大事件数量。默认为50，最大取值100。
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`
}

func (r *DescribeEKSContainerInstanceEventRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstanceEventRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "EksCiId")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEKSContainerInstanceEventRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstanceEventResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 事件集合
		Events []*Event `json:"Events,omitempty" name:"Events"`

		// 容器实例id
		EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEKSContainerInstanceEventResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstanceEventResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstanceRegionsRequest struct {
	*tchttp.BaseRequest
}

func (r *DescribeEKSContainerInstanceRegionsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstanceRegionsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEKSContainerInstanceRegionsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstanceRegionsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// EKS Container Instance支持的地域信息
		// 注意：此字段可能返回 null，表示取不到有效值。
		Regions []*EksCiRegionInfo `json:"Regions,omitempty" name:"Regions"`

		// 总数
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEKSContainerInstanceRegionsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstanceRegionsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstancesRequest struct {
	*tchttp.BaseRequest

	// 限定此次返回资源的数量。如果不设定，默认返回20，最大不能超过100
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`

	// 偏移量,默认0
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 过滤条件，可条件：
	// (1)实例名称
	// KeyName: eks-ci-name
	// 类型：String
	//
	// (2)实例状态
	// KeyName: status
	// 类型：String
	// 可选值："Pending", "Running", "Succeeded", "Failed"
	//
	// (3)内网ip
	// KeyName: private-ip
	// 类型：String
	//
	// (4)EIP地址
	// KeyName: eip-address
	// 类型：String
	//
	// (5)VpcId
	// KeyName: vpc-id
	// 类型：String
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`

	// 容器实例 ID 数组
	EksCiIds []*string `json:"EksCiIds,omitempty" name:"EksCiIds"`
}

func (r *DescribeEKSContainerInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Limit")
	delete(f, "Offset")
	delete(f, "Filters")
	delete(f, "EksCiIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEKSContainerInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEKSContainerInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 容器组总数
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 容器组列表
		EksCis []*EksCi `json:"EksCis,omitempty" name:"EksCis"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEKSContainerInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEKSContainerInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEksContainerInstanceLogRequest struct {
	*tchttp.BaseRequest

	// Eks Container Instance Id，即容器实例Id
	EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

	// 容器名称，单容器的实例可选填。如果为多容器实例，请指定容器名称。
	ContainerName *string `json:"ContainerName,omitempty" name:"ContainerName"`

	// 返回最新日志行数，默认500，最大2000。日志内容最大返回 1M 数据。
	Tail *uint64 `json:"Tail,omitempty" name:"Tail"`

	// UTC时间，RFC3339标准
	StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

	// 是否是查上一个容器（如果容器退出重启了）
	Previous *bool `json:"Previous,omitempty" name:"Previous"`

	// 查询最近多少秒内的日志
	SinceSeconds *uint64 `json:"SinceSeconds,omitempty" name:"SinceSeconds"`

	// 日志总大小限制
	LimitBytes *uint64 `json:"LimitBytes,omitempty" name:"LimitBytes"`
}

func (r *DescribeEksContainerInstanceLogRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEksContainerInstanceLogRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "EksCiId")
	delete(f, "ContainerName")
	delete(f, "Tail")
	delete(f, "StartTime")
	delete(f, "Previous")
	delete(f, "SinceSeconds")
	delete(f, "LimitBytes")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEksContainerInstanceLogRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEksContainerInstanceLogResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 容器名称
		ContainerName *string `json:"ContainerName,omitempty" name:"ContainerName"`

		// 日志内容
		LogContent *string `json:"LogContent,omitempty" name:"LogContent"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEksContainerInstanceLogResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEksContainerInstanceLogResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEnableVpcCniProgressRequest struct {
	*tchttp.BaseRequest

	// 开启vpc-cni的集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribeEnableVpcCniProgressRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEnableVpcCniProgressRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeEnableVpcCniProgressRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeEnableVpcCniProgressResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 任务进度的描述：Running/Succeed/Failed
		Status *string `json:"Status,omitempty" name:"Status"`

		// 当任务进度为Failed时，对任务状态的进一步描述，例如IPAMD组件安装失败
		// 注意：此字段可能返回 null，表示取不到有效值。
		ErrorMessage *string `json:"ErrorMessage,omitempty" name:"ErrorMessage"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeEnableVpcCniProgressResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeEnableVpcCniProgressResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeExistedInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群 ID，请填写查询集群列表 接口中返回的 ClusterId 字段（仅通过ClusterId获取需要过滤条件中的VPCID。节点状态比较时会使用该地域下所有集群中的节点进行比较。参数不支持同时指定InstanceIds和ClusterId。
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 按照一个或者多个实例ID查询。实例ID形如：ins-xxxxxxxx。（此参数的具体格式可参考API简介的id.N一节）。每次请求的实例的上限为100。参数不支持同时指定InstanceIds和Filters。
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 过滤条件,字段和详见[CVM查询实例](https://cloud.tencent.com/document/api/213/15728)如果设置了ClusterId，会附加集群的VPCID作为查询字段，在此情况下如果在Filter中指定了"vpc-id"作为过滤字段，指定的VPCID必须与集群的VPCID相同。
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`

	// 实例IP进行过滤(同时支持内网IP和外网IP)
	VagueIpAddress *string `json:"VagueIpAddress,omitempty" name:"VagueIpAddress"`

	// 实例名称进行过滤
	VagueInstanceName *string `json:"VagueInstanceName,omitempty" name:"VagueInstanceName"`

	// 偏移量，默认为0。关于Offset的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 返回数量，默认为20，最大值为100。关于Limit的更进一步介绍请参考 API [简介](https://cloud.tencent.com/document/api/213/15688)中的相关小节。
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`

	// 根据多个实例IP进行过滤
	IpAddresses []*string `json:"IpAddresses,omitempty" name:"IpAddresses"`
}

func (r *DescribeExistedInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExistedInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "InstanceIds")
	delete(f, "Filters")
	delete(f, "VagueIpAddress")
	delete(f, "VagueInstanceName")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "IpAddresses")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeExistedInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeExistedInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 已经存在的实例信息数组。
		// 注意：此字段可能返回 null，表示取不到有效值。
		ExistedInstanceSet []*ExistedInstance `json:"ExistedInstanceSet,omitempty" name:"ExistedInstanceSet"`

		// 符合条件的实例数量。
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeExistedInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExistedInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeExternalClusterSpecRequest struct {
	*tchttp.BaseRequest

	// 注册集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 默认false 获取内网，是否获取外网版注册命令
	IsExtranet *bool `json:"IsExtranet,omitempty" name:"IsExtranet"`

	// 默认false 不刷新有效时间 ，true刷新有效时间
	IsRefreshExpirationTime *bool `json:"IsRefreshExpirationTime,omitempty" name:"IsRefreshExpirationTime"`
}

func (r *DescribeExternalClusterSpecRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalClusterSpecRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "IsExtranet")
	delete(f, "IsRefreshExpirationTime")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeExternalClusterSpecRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeExternalClusterSpecResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 导入第三方集群YAML定义
		Spec *string `json:"Spec,omitempty" name:"Spec"`

		// agent.yaml文件过期时间字符串，时区UTC
		Expiration *string `json:"Expiration,omitempty" name:"Expiration"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeExternalClusterSpecResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalClusterSpecResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeImagesRequest struct {
	*tchttp.BaseRequest
}

func (r *DescribeImagesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeImagesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeImagesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeImagesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 镜像数量
		// 注意：此字段可能返回 null，表示取不到有效值。
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 镜像信息列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		ImageInstanceSet []*ImageInstance `json:"ImageInstanceSet,omitempty" name:"ImageInstanceSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeImagesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeImagesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAgentInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群id
	// 可以是tke, eks, edge的集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DescribePrometheusAgentInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAgentInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusAgentInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAgentInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 关联该集群的实例列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		Instances []*string `json:"Instances,omitempty" name:"Instances"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusAgentInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAgentInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAgentsRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 用于分页
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 用于分页
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`
}

func (r *DescribePrometheusAgentsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAgentsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "Offset")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusAgentsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAgentsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 被关联集群信息
		Agents []*PrometheusAgentOverview `json:"Agents,omitempty" name:"Agents"`

		// 被关联集群总量
		Total *uint64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusAgentsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAgentsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAlertHistoryRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 告警名称
	RuleName *string `json:"RuleName,omitempty" name:"RuleName"`

	// 开始时间
	StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

	// 结束时间
	EndTime *string `json:"EndTime,omitempty" name:"EndTime"`

	// label集合
	Labels *string `json:"Labels,omitempty" name:"Labels"`

	// 分片
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 分片
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`
}

func (r *DescribePrometheusAlertHistoryRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAlertHistoryRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "RuleName")
	delete(f, "StartTime")
	delete(f, "EndTime")
	delete(f, "Labels")
	delete(f, "Offset")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusAlertHistoryRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAlertHistoryResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 告警历史
		Items []*PrometheusAlertHistoryItem `json:"Items,omitempty" name:"Items"`

		// 总数
		Total *uint64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusAlertHistoryResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAlertHistoryResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAlertRuleRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 分页
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 分页
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`

	// 过滤
	// 支持ID，Name
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribePrometheusAlertRuleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAlertRuleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusAlertRuleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusAlertRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 告警详情
		AlertRules []*PrometheusAlertRuleDetail `json:"AlertRules,omitempty" name:"AlertRules"`

		// 总数
		Total *uint64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusAlertRuleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusAlertRuleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusInstanceRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`
}

func (r *DescribePrometheusInstanceRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusInstanceRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusInstanceRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusInstanceResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 实例id
		InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

		// 实例名称
		Name *string `json:"Name,omitempty" name:"Name"`

		// 私有网络id
		VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

		// 子网id
		SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

		// cos桶名称
		COSBucket *string `json:"COSBucket,omitempty" name:"COSBucket"`

		// 数据查询地址
		QueryAddress *string `json:"QueryAddress,omitempty" name:"QueryAddress"`

		// 实例中grafana相关的信息
		// 注意：此字段可能返回 null，表示取不到有效值。
		Grafana *PrometheusGrafanaInfo `json:"Grafana,omitempty" name:"Grafana"`

		// 用户自定义alertmanager
		// 注意：此字段可能返回 null，表示取不到有效值。
		AlertManagerUrl *string `json:"AlertManagerUrl,omitempty" name:"AlertManagerUrl"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusInstanceResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusInstanceResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusOverviewsRequest struct {
	*tchttp.BaseRequest

	// 用于分页
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 用于分页
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`

	// 过滤实例，目前支持：
	// ID: 通过实例ID来过滤
	// Name: 通过实例名称来过滤
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribePrometheusOverviewsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusOverviewsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Offset")
	delete(f, "Limit")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusOverviewsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusOverviewsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 实例列表
		Instances []*PrometheusInstanceOverview `json:"Instances,omitempty" name:"Instances"`

		// 实例总数
		// 注意：此字段可能返回 null，表示取不到有效值。
		Total *uint64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusOverviewsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusOverviewsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTargetsRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 集群类型
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 过滤条件，当前支持
	// Name=state
	// Value=up, down, unknown
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`
}

func (r *DescribePrometheusTargetsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTargetsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "ClusterType")
	delete(f, "ClusterId")
	delete(f, "Filters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusTargetsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 所有Job的targets信息
		Jobs []*PrometheusJobTargets `json:"Jobs,omitempty" name:"Jobs"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusTargetsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTargetsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTemplateSyncRequest struct {
	*tchttp.BaseRequest

	// 模板ID
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`
}

func (r *DescribePrometheusTemplateSyncRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTemplateSyncRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "TemplateId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusTemplateSyncRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTemplateSyncResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 同步目标详情
		Targets []*PrometheusTemplateSyncTarget `json:"Targets,omitempty" name:"Targets"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusTemplateSyncResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTemplateSyncResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTemplatesRequest struct {
	*tchttp.BaseRequest

	// 模糊过滤条件，支持
	// Level 按模板级别过滤
	// Name 按名称过滤
	// Describe 按描述过滤
	// ID 按templateId过滤
	Filters []*Filter `json:"Filters,omitempty" name:"Filters"`

	// 分页偏移
	Offset *uint64 `json:"Offset,omitempty" name:"Offset"`

	// 总数限制
	Limit *uint64 `json:"Limit,omitempty" name:"Limit"`
}

func (r *DescribePrometheusTemplatesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTemplatesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Filters")
	delete(f, "Offset")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribePrometheusTemplatesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribePrometheusTemplatesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 模板列表
		Templates []*PrometheusTemplate `json:"Templates,omitempty" name:"Templates"`

		// 总数
		Total *uint64 `json:"Total,omitempty" name:"Total"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribePrometheusTemplatesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribePrometheusTemplatesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRegionsRequest struct {
	*tchttp.BaseRequest
}

func (r *DescribeRegionsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRegionsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeRegionsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRegionsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 地域的数量
		// 注意：此字段可能返回 null，表示取不到有效值。
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 地域列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		RegionInstanceSet []*RegionInstance `json:"RegionInstanceSet,omitempty" name:"RegionInstanceSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeRegionsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRegionsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRouteTableConflictsRequest struct {
	*tchttp.BaseRequest

	// 路由表CIDR
	RouteTableCidrBlock *string `json:"RouteTableCidrBlock,omitempty" name:"RouteTableCidrBlock"`

	// 路由表绑定的VPC
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`
}

func (r *DescribeRouteTableConflictsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRouteTableConflictsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "RouteTableCidrBlock")
	delete(f, "VpcId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeRouteTableConflictsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRouteTableConflictsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 路由表是否冲突。
		HasConflict *bool `json:"HasConflict,omitempty" name:"HasConflict"`

		// 路由表冲突列表。
		// 注意：此字段可能返回 null，表示取不到有效值。
		RouteTableConflictSet []*RouteTableConflict `json:"RouteTableConflictSet,omitempty" name:"RouteTableConflictSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeRouteTableConflictsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRouteTableConflictsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeVersionsRequest struct {
	*tchttp.BaseRequest
}

func (r *DescribeVersionsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeVersionsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeVersionsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeVersionsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 版本数量
		// 注意：此字段可能返回 null，表示取不到有效值。
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 版本列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		VersionInstanceSet []*VersionInstance `json:"VersionInstanceSet,omitempty" name:"VersionInstanceSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeVersionsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeVersionsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeVpcCniPodLimitsRequest struct {
	*tchttp.BaseRequest

	// 查询的机型所在可用区，如：ap-guangzhou-3，默认为空，即不按可用区过滤信息
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 查询的实例机型系列信息，如：S5，默认为空，即不按机型系列过滤信息
	InstanceFamily *string `json:"InstanceFamily,omitempty" name:"InstanceFamily"`

	// 查询的实例机型信息，如：S5.LARGE8，默认为空，即不按机型过滤信息
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`
}

func (r *DescribeVpcCniPodLimitsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeVpcCniPodLimitsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Zone")
	delete(f, "InstanceFamily")
	delete(f, "InstanceType")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeVpcCniPodLimitsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeVpcCniPodLimitsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 机型数据数量
		// 注意：此字段可能返回 null，表示取不到有效值。
		TotalCount *int64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 机型信息及其可支持的最大VPC-CNI模式Pod数量信息
		// 注意：此字段可能返回 null，表示取不到有效值。
		PodLimitsInstanceSet []*PodLimitsInstance `json:"PodLimitsInstanceSet,omitempty" name:"PodLimitsInstanceSet"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeVpcCniPodLimitsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeVpcCniPodLimitsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DisableClusterDeletionProtectionRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DisableClusterDeletionProtectionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DisableClusterDeletionProtectionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DisableClusterDeletionProtectionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DisableClusterDeletionProtectionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DisableClusterDeletionProtectionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DisableClusterDeletionProtectionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DisableVpcCniNetworkTypeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *DisableVpcCniNetworkTypeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DisableVpcCniNetworkTypeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DisableVpcCniNetworkTypeRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DisableVpcCniNetworkTypeResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DisableVpcCniNetworkTypeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DisableVpcCniNetworkTypeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DnsServerConf struct {

	// 域名。空字符串表示所有域名。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// dns 服务器地址列表。地址格式 ip:port
	DnsServers []*string `json:"DnsServers,omitempty" name:"DnsServers"`
}

type EipAttribute struct {

	// 容器实例删除后，EIP是否释放。
	// Never表示不释放，其他任意值（包括空字符串）表示释放。
	DeletePolicy *string `json:"DeletePolicy,omitempty" name:"DeletePolicy"`

	// EIP线路类型。默认值：BGP。
	// 已开通静态单线IP白名单的用户，可选值：
	// CMCC：中国移动
	// CTCC：中国电信
	// CUCC：中国联通
	// 注意：仅部分地域支持静态单线IP。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InternetServiceProvider *string `json:"InternetServiceProvider,omitempty" name:"InternetServiceProvider"`

	// EIP出带宽上限，单位：Mbps。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InternetMaxBandwidthOut *uint64 `json:"InternetMaxBandwidthOut,omitempty" name:"InternetMaxBandwidthOut"`
}

type EksCi struct {

	// EKS Cotainer Instance Id
	EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

	// EKS Cotainer Instance Name
	EksCiName *string `json:"EksCiName,omitempty" name:"EksCiName"`

	// 内存大小
	Memory *float64 `json:"Memory,omitempty" name:"Memory"`

	// CPU大小
	Cpu *float64 `json:"Cpu,omitempty" name:"Cpu"`

	// 安全组ID
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`

	// 容器组的重启策略
	// 注意：此字段可能返回 null，表示取不到有效值。
	RestartPolicy *string `json:"RestartPolicy,omitempty" name:"RestartPolicy"`

	// 返回容器组创建状态：Pending，Running，Succeeded，Failed。其中：
	// Failed （运行失败）指的容器组退出，RestartPolilcy为Never， 有容器exitCode非0；
	// Succeeded（运行成功）指的是容器组退出了，RestartPolicy为Never或onFailure，所有容器exitCode都为0；
	// Failed和Succeeded这两种状态都会停止运行，停止计费。
	// Pending是创建中，Running是 运行中。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Status *string `json:"Status,omitempty" name:"Status"`

	// 接到请求后的系统创建时间。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreationTime *string `json:"CreationTime,omitempty" name:"CreationTime"`

	// 容器全部成功退出后的时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	SucceededTime *string `json:"SucceededTime,omitempty" name:"SucceededTime"`

	// 容器列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Containers []*Container `json:"Containers,omitempty" name:"Containers"`

	// 数据卷信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	EksCiVolume *EksCiVolume `json:"EksCiVolume,omitempty" name:"EksCiVolume"`

	// 容器组运行的安全上下文
	// 注意：此字段可能返回 null，表示取不到有效值。
	SecurityContext *SecurityContext `json:"SecurityContext,omitempty" name:"SecurityContext"`

	// 内网ip地址
	// 注意：此字段可能返回 null，表示取不到有效值。
	PrivateIp *string `json:"PrivateIp,omitempty" name:"PrivateIp"`

	// 容器实例绑定的Eip地址，注意可能为空
	// 注意：此字段可能返回 null，表示取不到有效值。
	EipAddress *string `json:"EipAddress,omitempty" name:"EipAddress"`

	// GPU类型。如无使用GPU则不返回
	// 注意：此字段可能返回 null，表示取不到有效值。
	GpuType *string `json:"GpuType,omitempty" name:"GpuType"`

	// CPU类型
	// 注意：此字段可能返回 null，表示取不到有效值。
	CpuType *string `json:"CpuType,omitempty" name:"CpuType"`

	// GPU卡数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	GpuCount *uint64 `json:"GpuCount,omitempty" name:"GpuCount"`

	// 实例所属VPC的Id
	// 注意：此字段可能返回 null，表示取不到有效值。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 实例所属子网Id
	// 注意：此字段可能返回 null，表示取不到有效值。
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 初始化容器列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	InitContainers []*Container `json:"InitContainers,omitempty" name:"InitContainers"`

	// 为容器实例关联 CAM 角色，value 填写 CAM 角色名称，容器实例可获取该 CAM 角色包含的权限策略，方便 容器实例 内的程序进行如购买资源、读写存储等云资源操作。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CamRoleName *string `json:"CamRoleName,omitempty" name:"CamRoleName"`

	// 自动为用户创建的EipId
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoCreatedEipId *string `json:"AutoCreatedEipId,omitempty" name:"AutoCreatedEipId"`

	// 容器状态是否持久化
	// 注意：此字段可能返回 null，表示取不到有效值。
	PersistStatus *bool `json:"PersistStatus,omitempty" name:"PersistStatus"`
}

type EksCiRegionInfo struct {

	// 地域别名，形如gz
	Alias *string `json:"Alias,omitempty" name:"Alias"`

	// 地域名，形如ap-guangzhou
	RegionName *string `json:"RegionName,omitempty" name:"RegionName"`

	// 地域ID
	RegionId *uint64 `json:"RegionId,omitempty" name:"RegionId"`
}

type EksCiVolume struct {

	// Cbs Volume
	// 注意：此字段可能返回 null，表示取不到有效值。
	CbsVolumes []*CbsVolume `json:"CbsVolumes,omitempty" name:"CbsVolumes"`

	// Nfs Volume
	// 注意：此字段可能返回 null，表示取不到有效值。
	NfsVolumes []*NfsVolume `json:"NfsVolumes,omitempty" name:"NfsVolumes"`
}

type EksCluster struct {

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// Vpc Id
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 子网列表
	SubnetIds []*string `json:"SubnetIds,omitempty" name:"SubnetIds"`

	// k8s 版本号
	K8SVersion *string `json:"K8SVersion,omitempty" name:"K8SVersion"`

	// 集群状态(running运行中，initializing 初始化中，failed异常)
	Status *string `json:"Status,omitempty" name:"Status"`

	// 集群描述信息
	ClusterDesc *string `json:"ClusterDesc,omitempty" name:"ClusterDesc"`

	// 集群创建时间
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// Service 子网Id
	ServiceSubnetId *string `json:"ServiceSubnetId,omitempty" name:"ServiceSubnetId"`

	// 集群的自定义dns 服务器信息
	DnsServers []*DnsServerConf `json:"DnsServers,omitempty" name:"DnsServers"`

	// 将来删除集群时是否要删除cbs。默认为 FALSE
	NeedDeleteCbs *bool `json:"NeedDeleteCbs,omitempty" name:"NeedDeleteCbs"`

	// 是否在用户集群内开启Dns。默认为TRUE
	EnableVpcCoreDNS *bool `json:"EnableVpcCoreDNS,omitempty" name:"EnableVpcCoreDNS"`

	// 标签描述列表。
	// 注意：此字段可能返回 null，表示取不到有效值。
	TagSpecification []*TagSpecification `json:"TagSpecification,omitempty" name:"TagSpecification"`
}

type EnableClusterDeletionProtectionRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

func (r *EnableClusterDeletionProtectionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *EnableClusterDeletionProtectionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "EnableClusterDeletionProtectionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type EnableClusterDeletionProtectionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *EnableClusterDeletionProtectionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *EnableClusterDeletionProtectionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type EnableVpcCniNetworkTypeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 开启vpc-cni的模式，tke-route-eni开启的是策略路由模式，tke-direct-eni开启的是独立网卡模式
	VpcCniType *string `json:"VpcCniType,omitempty" name:"VpcCniType"`

	// 是否开启固定IP模式
	EnableStaticIp *bool `json:"EnableStaticIp,omitempty" name:"EnableStaticIp"`

	// 使用的容器子网
	Subnets []*string `json:"Subnets,omitempty" name:"Subnets"`

	// 在固定IP模式下，Pod销毁后退还IP的时间，传参必须大于300；不传默认IP永不销毁。
	ExpiredSeconds *uint64 `json:"ExpiredSeconds,omitempty" name:"ExpiredSeconds"`
}

func (r *EnableVpcCniNetworkTypeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *EnableVpcCniNetworkTypeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "VpcCniType")
	delete(f, "EnableStaticIp")
	delete(f, "Subnets")
	delete(f, "ExpiredSeconds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "EnableVpcCniNetworkTypeRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type EnableVpcCniNetworkTypeResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *EnableVpcCniNetworkTypeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *EnableVpcCniNetworkTypeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type EnhancedService struct {

	// 开启云安全服务。若不指定该参数，则默认开启云安全服务。
	SecurityService *RunSecurityServiceEnabled `json:"SecurityService,omitempty" name:"SecurityService"`

	// 开启云监控服务。若不指定该参数，则默认开启云监控服务。
	MonitorService *RunMonitorServiceEnabled `json:"MonitorService,omitempty" name:"MonitorService"`

	// 开启云自动化助手服务。若不指定该参数，则默认不开启云自动化助手服务。
	AutomationService *RunAutomationServiceEnabled `json:"AutomationService,omitempty" name:"AutomationService"`
}

type EnvironmentVariable struct {

	// key
	Name *string `json:"Name,omitempty" name:"Name"`

	// val
	Value *string `json:"Value,omitempty" name:"Value"`
}

type Event struct {

	// pod名称
	PodName *string `json:"PodName,omitempty" name:"PodName"`

	// 事件原因内容
	Reason *string `json:"Reason,omitempty" name:"Reason"`

	// 事件类型
	Type *string `json:"Type,omitempty" name:"Type"`

	// 事件出现次数
	Count *int64 `json:"Count,omitempty" name:"Count"`

	// 事件第一次出现时间
	FirstTimestamp *string `json:"FirstTimestamp,omitempty" name:"FirstTimestamp"`

	// 事件最后一次出现时间
	LastTimestamp *string `json:"LastTimestamp,omitempty" name:"LastTimestamp"`

	// 事件内容
	Message *string `json:"Message,omitempty" name:"Message"`
}

type Exec struct {

	// 容器内检测的命令
	// 注意：此字段可能返回 null，表示取不到有效值。
	Commands []*string `json:"Commands,omitempty" name:"Commands"`
}

type ExistedInstance struct {

	// 实例是否支持加入集群(TRUE 可以加入 FALSE 不能加入)。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Usable *bool `json:"Usable,omitempty" name:"Usable"`

	// 实例不支持加入的原因。
	// 注意：此字段可能返回 null，表示取不到有效值。
	UnusableReason *string `json:"UnusableReason,omitempty" name:"UnusableReason"`

	// 实例已经所在的集群ID。
	// 注意：此字段可能返回 null，表示取不到有效值。
	AlreadyInCluster *string `json:"AlreadyInCluster,omitempty" name:"AlreadyInCluster"`

	// 实例ID形如：ins-xxxxxxxx。
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 实例名称。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceName *string `json:"InstanceName,omitempty" name:"InstanceName"`

	// 实例主网卡的内网IP列表。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PrivateIpAddresses []*string `json:"PrivateIpAddresses,omitempty" name:"PrivateIpAddresses"`

	// 实例主网卡的公网IP列表。
	// 注意：此字段可能返回 null，表示取不到有效值。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PublicIpAddresses []*string `json:"PublicIpAddresses,omitempty" name:"PublicIpAddresses"`

	// 创建时间。按照ISO8601标准表示，并且使用UTC时间。格式为：YYYY-MM-DDThh:mm:ssZ。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// 实例的CPU核数，单位：核。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CPU *uint64 `json:"CPU,omitempty" name:"CPU"`

	// 实例内存容量，单位：GB。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Memory *uint64 `json:"Memory,omitempty" name:"Memory"`

	// 操作系统名称。
	// 注意：此字段可能返回 null，表示取不到有效值。
	OsName *string `json:"OsName,omitempty" name:"OsName"`

	// 实例机型。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 伸缩组ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoscalingGroupId *string `json:"AutoscalingGroupId,omitempty" name:"AutoscalingGroupId"`

	// 实例计费模式。取值范围： PREPAID：表示预付费，即包年包月 POSTPAID_BY_HOUR：表示后付费，即按量计费 CDHPAID：CDH付费，即只对CDH计费，不对CDH上的实例计费。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceChargeType *string `json:"InstanceChargeType,omitempty" name:"InstanceChargeType"`
}

type ExistedInstancesForNode struct {

	// 节点角色，取值:MASTER_ETCD, WORKER。MASTER_ETCD只有在创建 INDEPENDENT_CLUSTER 独立集群时需要指定。MASTER_ETCD节点数量为3～7，建议为奇数。MASTER_ETCD最小配置为4C8G。
	NodeRole *string `json:"NodeRole,omitempty" name:"NodeRole"`

	// 已存在实例的重装参数
	ExistedInstancesPara *ExistedInstancesPara `json:"ExistedInstancesPara,omitempty" name:"ExistedInstancesPara"`

	// 节点高级设置，会覆盖集群级别设置的InstanceAdvancedSettings（当前只对节点自定义参数ExtraArgs生效）
	InstanceAdvancedSettingsOverride *InstanceAdvancedSettings `json:"InstanceAdvancedSettingsOverride,omitempty" name:"InstanceAdvancedSettingsOverride"`

	// 自定义模式集群，可指定每个节点的pod数量
	DesiredPodNumbers []*int64 `json:"DesiredPodNumbers,omitempty" name:"DesiredPodNumbers"`
}

type ExistedInstancesPara struct {

	// 集群ID
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 实例额外需要设置参数信息
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 增强服务。通过该参数可以指定是否开启云安全、云监控等服务。若不指定该参数，则默认开启云监控、云安全服务。
	EnhancedService *EnhancedService `json:"EnhancedService,omitempty" name:"EnhancedService"`

	// 节点登录信息（目前仅支持使用Password或者单个KeyIds）
	LoginSettings *LoginSettings `json:"LoginSettings,omitempty" name:"LoginSettings"`

	// 实例所属安全组。该参数可以通过调用 DescribeSecurityGroups 的返回值中的sgId字段来获取。若不指定该参数，则绑定默认安全组。
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`

	// 重装系统时，可以指定修改实例的HostName(集群为HostName模式时，此参数必传，规则名称除不支持大写字符外与[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口HostName一致)
	HostName *string `json:"HostName,omitempty" name:"HostName"`
}

type ExtensionAddon struct {

	// 扩展组件名称
	AddonName *string `json:"AddonName,omitempty" name:"AddonName"`

	// 扩展组件信息(扩展组件资源对象的json字符串描述)
	AddonParam *string `json:"AddonParam,omitempty" name:"AddonParam"`
}

type Filter struct {

	// 需要过滤的字段。
	Name *string `json:"Name,omitempty" name:"Name"`

	// 字段的过滤值。
	Values []*string `json:"Values,omitempty" name:"Values"`
}

type ForwardApplicationRequestV3Request struct {
	*tchttp.BaseRequest

	// 请求集群addon的访问
	Method *string `json:"Method,omitempty" name:"Method"`

	// 请求集群addon的路径
	Path *string `json:"Path,omitempty" name:"Path"`

	// 请求集群addon后允许接收的数据格式
	Accept *string `json:"Accept,omitempty" name:"Accept"`

	// 请求集群addon的数据格式
	ContentType *string `json:"ContentType,omitempty" name:"ContentType"`

	// 请求集群addon的数据
	RequestBody *string `json:"RequestBody,omitempty" name:"RequestBody"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 是否编码请求内容
	EncodedBody *string `json:"EncodedBody,omitempty" name:"EncodedBody"`
}

func (r *ForwardApplicationRequestV3Request) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ForwardApplicationRequestV3Request) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Method")
	delete(f, "Path")
	delete(f, "Accept")
	delete(f, "ContentType")
	delete(f, "RequestBody")
	delete(f, "ClusterName")
	delete(f, "EncodedBody")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ForwardApplicationRequestV3Request has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ForwardApplicationRequestV3Response struct {
	*tchttp.BaseResponse
	Response *struct {

		// 请求集群addon后返回的数据
		ResponseBody *string `json:"ResponseBody,omitempty" name:"ResponseBody"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ForwardApplicationRequestV3Response) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ForwardApplicationRequestV3Response) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type GetTkeAppChartListRequest struct {
	*tchttp.BaseRequest

	// app类型，取值log,scheduler,network,storage,monitor,dns,image,other,invisible
	Kind *string `json:"Kind,omitempty" name:"Kind"`

	// app支持的操作系统，取值arm32、arm64、amd64
	Arch *string `json:"Arch,omitempty" name:"Arch"`

	// 集群类型，取值tke、eks
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`
}

func (r *GetTkeAppChartListRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetTkeAppChartListRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Kind")
	delete(f, "Arch")
	delete(f, "ClusterType")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "GetTkeAppChartListRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type GetTkeAppChartListResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 所支持的chart列表
		// 注意：此字段可能返回 null，表示取不到有效值。
		AppCharts []*AppChart `json:"AppCharts,omitempty" name:"AppCharts"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *GetTkeAppChartListResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetTkeAppChartListResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type GetUpgradeInstanceProgressRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 最多获取多少个节点的进度
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`

	// 从第几个节点开始获取进度
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`
}

func (r *GetUpgradeInstanceProgressRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetUpgradeInstanceProgressRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Limit")
	delete(f, "Offset")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "GetUpgradeInstanceProgressRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type GetUpgradeInstanceProgressResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 升级节点总数
		Total *int64 `json:"Total,omitempty" name:"Total"`

		// 已升级节点总数
		Done *int64 `json:"Done,omitempty" name:"Done"`

		// 升级任务生命周期
		// process 运行中
		// paused 已停止
		// pauing 正在停止
		// done  已完成
		// timeout 已超时
		// aborted 已取消
		LifeState *string `json:"LifeState,omitempty" name:"LifeState"`

		// 各节点升级进度详情
		Instances []*InstanceUpgradeProgressItem `json:"Instances,omitempty" name:"Instances"`

		// 集群当前状态
		ClusterStatus *InstanceUpgradeClusterStatus `json:"ClusterStatus,omitempty" name:"ClusterStatus"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *GetUpgradeInstanceProgressResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetUpgradeInstanceProgressResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type HttpGet struct {

	// HttpGet检测的路径
	// 注意：此字段可能返回 null，表示取不到有效值。
	Path *string `json:"Path,omitempty" name:"Path"`

	// HttpGet检测的端口号
	// 注意：此字段可能返回 null，表示取不到有效值。
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// HTTP or HTTPS
	// 注意：此字段可能返回 null，表示取不到有效值。
	Scheme *string `json:"Scheme,omitempty" name:"Scheme"`
}

type IPAddress struct {

	// Ip 地址的类型。可为 advertise, public 等
	Type *string `json:"Type,omitempty" name:"Type"`

	// Ip 地址
	Ip *string `json:"Ip,omitempty" name:"Ip"`

	// 网络端口
	Port *uint64 `json:"Port,omitempty" name:"Port"`
}

type ImageInstance struct {

	// 镜像别名
	// 注意：此字段可能返回 null，表示取不到有效值。
	Alias *string `json:"Alias,omitempty" name:"Alias"`

	// 操作系统名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	OsName *string `json:"OsName,omitempty" name:"OsName"`

	// 镜像ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	ImageId *string `json:"ImageId,omitempty" name:"ImageId"`

	// 容器的镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	// 注意：此字段可能返回 null，表示取不到有效值。
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`
}

type ImageRegistryCredential struct {

	// 镜像仓库地址
	Server *string `json:"Server,omitempty" name:"Server"`

	// 用户名
	Username *string `json:"Username,omitempty" name:"Username"`

	// 密码
	Password *string `json:"Password,omitempty" name:"Password"`

	// ImageRegistryCredential的名字
	Name *string `json:"Name,omitempty" name:"Name"`
}

type Instance struct {

	// 实例ID
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 节点角色, MASTER, WORKER, ETCD, MASTER_ETCD,ALL, 默认为WORKER
	InstanceRole *string `json:"InstanceRole,omitempty" name:"InstanceRole"`

	// 实例异常(或者处于初始化中)的原因
	FailedReason *string `json:"FailedReason,omitempty" name:"FailedReason"`

	// 实例的状态（running 运行中，initializing 初始化中，failed 异常）
	InstanceState *string `json:"InstanceState,omitempty" name:"InstanceState"`

	// 实例是否封锁状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	DrainStatus *string `json:"DrainStatus,omitempty" name:"DrainStatus"`

	// 节点配置
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 添加时间
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// 节点内网IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	LanIP *string `json:"LanIP,omitempty" name:"LanIP"`

	// 资源池ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 自动伸缩组ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoscalingGroupId *string `json:"AutoscalingGroupId,omitempty" name:"AutoscalingGroupId"`
}

type InstanceAdvancedSettings struct {

	// 数据盘挂载点, 默认不挂载数据盘. 已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载，其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载，请注意备份数据! 无数据盘或有多块数据盘的云主机此设置不生效。
	// 注意，注意，多盘场景请使用下方的DataDisks数据结构，设置对应的云盘类型、云盘大小、挂载路径、是否格式化等信息。
	// 注意：此字段可能返回 null，表示取不到有效值。
	MountTarget *string `json:"MountTarget,omitempty" name:"MountTarget"`

	// dockerd --graph 指定值, 默认为 /var/lib/docker
	// 注意：此字段可能返回 null，表示取不到有效值。
	DockerGraphPath *string `json:"DockerGraphPath,omitempty" name:"DockerGraphPath"`

	// base64 编码的用户脚本, 此脚本会在 k8s 组件运行后执行, 需要用户保证脚本的可重入及重试逻辑, 脚本及其生成的日志文件可在节点的 /data/ccs_userscript/ 路径查看, 如果要求节点需要在进行初始化完成后才可加入调度, 可配合 unschedulable 参数使用, 在 userScript 最后初始化完成后, 添加 kubectl uncordon nodename --kubeconfig=/root/.kube/config 命令使节点加入调度
	// 注意：此字段可能返回 null，表示取不到有效值。
	UserScript *string `json:"UserScript,omitempty" name:"UserScript"`

	// 设置加入的节点是否参与调度，默认值为0，表示参与调度；非0表示不参与调度, 待节点初始化完成之后, 可执行kubectl uncordon nodename使node加入调度.
	Unschedulable *int64 `json:"Unschedulable,omitempty" name:"Unschedulable"`

	// 节点Label数组
	// 注意：此字段可能返回 null，表示取不到有效值。
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 多盘数据盘挂载信息：新建节点时请确保购买CVM的参数传递了购买多个数据盘的信息，如CreateClusterInstances API的RunInstancesPara下的DataDisks也需要设置购买多个数据盘, 具体可以参考CreateClusterInstances接口的添加集群节点(多块数据盘)样例；添加已有节点时，请确保填写的分区信息在节点上真实存在
	// 注意：此字段可能返回 null，表示取不到有效值。
	DataDisks []*DataDisk `json:"DataDisks,omitempty" name:"DataDisks"`

	// 节点相关的自定义参数信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	ExtraArgs *InstanceExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`

	// 该节点属于podCIDR大小自定义模式时，可指定节点上运行的pod数量上限
	// 注意：此字段可能返回 null，表示取不到有效值。
	DesiredPodNumber *int64 `json:"DesiredPodNumber,omitempty" name:"DesiredPodNumber"`

	// base64 编码的用户脚本，在初始化节点之前执行，目前只对添加已有节点生效
	// 注意：此字段可能返回 null，表示取不到有效值。
	PreStartUserScript *string `json:"PreStartUserScript,omitempty" name:"PreStartUserScript"`
}

type InstanceDataDiskMountSetting struct {

	// CVM实例类型
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 数据盘挂载信息
	DataDisks []*DataDisk `json:"DataDisks,omitempty" name:"DataDisks"`

	// CVM实例所属可用区
	Zone *string `json:"Zone,omitempty" name:"Zone"`
}

type InstanceExtraArgs struct {

	// kubelet自定义参数，参数格式为["k1=v1", "k1=v2"]， 例如["root-dir=/var/lib/kubelet","feature-gates=PodShareProcessNamespace=true,DynamicKubeletConfig=true"]
	// 注意：此字段可能返回 null，表示取不到有效值。
	Kubelet []*string `json:"Kubelet,omitempty" name:"Kubelet"`
}

type InstanceUpgradeClusterStatus struct {

	// pod总数
	PodTotal *int64 `json:"PodTotal,omitempty" name:"PodTotal"`

	// NotReady pod总数
	NotReadyPod *int64 `json:"NotReadyPod,omitempty" name:"NotReadyPod"`
}

type InstanceUpgradePreCheckResult struct {

	// 检查是否通过
	CheckPass *bool `json:"CheckPass,omitempty" name:"CheckPass"`

	// 检查项数组
	Items []*InstanceUpgradePreCheckResultItem `json:"Items,omitempty" name:"Items"`

	// 本节点独立pod列表
	SinglePods []*string `json:"SinglePods,omitempty" name:"SinglePods"`
}

type InstanceUpgradePreCheckResultItem struct {

	// 工作负载的命名空间
	Namespace *string `json:"Namespace,omitempty" name:"Namespace"`

	// 工作负载类型
	WorkLoadKind *string `json:"WorkLoadKind,omitempty" name:"WorkLoadKind"`

	// 工作负载名称
	WorkLoadName *string `json:"WorkLoadName,omitempty" name:"WorkLoadName"`

	// 驱逐节点前工作负载running的pod数目
	Before *uint64 `json:"Before,omitempty" name:"Before"`

	// 驱逐节点后工作负载running的pod数目
	After *uint64 `json:"After,omitempty" name:"After"`

	// 工作负载在本节点上的pod列表
	Pods []*string `json:"Pods,omitempty" name:"Pods"`
}

type InstanceUpgradeProgressItem struct {

	// 节点instanceID
	InstanceID *string `json:"InstanceID,omitempty" name:"InstanceID"`

	// 任务生命周期
	// process 运行中
	// paused 已停止
	// pauing 正在停止
	// done  已完成
	// timeout 已超时
	// aborted 已取消
	// pending 还未开始
	LifeState *string `json:"LifeState,omitempty" name:"LifeState"`

	// 升级开始时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	StartAt *string `json:"StartAt,omitempty" name:"StartAt"`

	// 升级结束时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	EndAt *string `json:"EndAt,omitempty" name:"EndAt"`

	// 升级前检查结果
	CheckResult *InstanceUpgradePreCheckResult `json:"CheckResult,omitempty" name:"CheckResult"`

	// 升级步骤详情
	Detail []*TaskStepInfo `json:"Detail,omitempty" name:"Detail"`
}

type Label struct {

	// map表中的Name
	Name *string `json:"Name,omitempty" name:"Name"`

	// map表中的Value
	Value *string `json:"Value,omitempty" name:"Value"`
}

type LivenessOrReadinessProbe struct {

	// 探针参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	Probe *Probe `json:"Probe,omitempty" name:"Probe"`

	// HttpGet检测参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpGet *HttpGet `json:"HttpGet,omitempty" name:"HttpGet"`

	// 容器内检测命令参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	Exec *Exec `json:"Exec,omitempty" name:"Exec"`

	// TcpSocket检测的端口参数
	// 注意：此字段可能返回 null，表示取不到有效值。
	TcpSocket *TcpSocket `json:"TcpSocket,omitempty" name:"TcpSocket"`
}

type LoginSettings struct {

	// 实例登录密码。不同操作系统类型密码复杂度限制不一样，具体如下：<br><li>Linux实例密码必须8到30位，至少包括两项[a-z]，[A-Z]、[0-9] 和 [( ) \` ~ ! @ # $ % ^ & *  - + = | { } [ ] : ; ' , . ? / ]中的特殊符号。<br><li>Windows实例密码必须12到30位，至少包括三项[a-z]，[A-Z]，[0-9] 和 [( ) \` ~ ! @ # $ % ^ & * - + = | { } [ ] : ; ' , . ? /]中的特殊符号。<br><br>若不指定该参数，则由系统随机生成密码，并通过站内信方式通知到用户。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Password *string `json:"Password,omitempty" name:"Password"`

	// 密钥ID列表。关联密钥后，就可以通过对应的私钥来访问实例；KeyId可通过接口[DescribeKeyPairs](https://cloud.tencent.com/document/api/213/15699)获取，密钥与密码不能同时指定，同时Windows操作系统不支持指定密钥。当前仅支持购买的时候指定一个密钥。
	// 注意：此字段可能返回 null，表示取不到有效值。
	KeyIds []*string `json:"KeyIds,omitempty" name:"KeyIds"`

	// 保持镜像的原始设置。该参数与Password或KeyIds.N不能同时指定。只有使用自定义镜像、共享镜像或外部导入镜像创建实例时才能指定该参数为TRUE。取值范围：<br><li>TRUE：表示保持镜像的登录设置<br><li>FALSE：表示不保持镜像的登录设置<br><br>默认取值：FALSE。
	// 注意：此字段可能返回 null，表示取不到有效值。
	KeepImageLogin *string `json:"KeepImageLogin,omitempty" name:"KeepImageLogin"`
}

type ManuallyAdded struct {

	// 加入中的节点数量
	Joining *int64 `json:"Joining,omitempty" name:"Joining"`

	// 初始化中的节点数量
	Initializing *int64 `json:"Initializing,omitempty" name:"Initializing"`

	// 正常的节点数量
	Normal *int64 `json:"Normal,omitempty" name:"Normal"`

	// 节点总数
	Total *int64 `json:"Total,omitempty" name:"Total"`
}

type ModifyClusterAsGroupAttributeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群关联的伸缩组属性
	ClusterAsGroupAttribute *ClusterAsGroupAttribute `json:"ClusterAsGroupAttribute,omitempty" name:"ClusterAsGroupAttribute"`
}

func (r *ModifyClusterAsGroupAttributeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAsGroupAttributeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterAsGroupAttribute")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterAsGroupAttributeRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAsGroupAttributeResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterAsGroupAttributeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAsGroupAttributeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAsGroupOptionAttributeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群弹性伸缩属性
	ClusterAsGroupOption *ClusterAsGroupOption `json:"ClusterAsGroupOption,omitempty" name:"ClusterAsGroupOption"`
}

func (r *ModifyClusterAsGroupOptionAttributeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAsGroupOptionAttributeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterAsGroupOption")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterAsGroupOptionAttributeRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAsGroupOptionAttributeResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterAsGroupOptionAttributeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAsGroupOptionAttributeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAttributeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 集群所属项目
	ProjectId *int64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 集群描述
	ClusterDesc *string `json:"ClusterDesc,omitempty" name:"ClusterDesc"`
}

func (r *ModifyClusterAttributeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAttributeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ProjectId")
	delete(f, "ClusterName")
	delete(f, "ClusterDesc")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterAttributeRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAttributeResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 集群所属项目
		// 注意：此字段可能返回 null，表示取不到有效值。
		ProjectId *int64 `json:"ProjectId,omitempty" name:"ProjectId"`

		// 集群名称
		// 注意：此字段可能返回 null，表示取不到有效值。
		ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

		// 集群描述
		// 注意：此字段可能返回 null，表示取不到有效值。
		ClusterDesc *string `json:"ClusterDesc,omitempty" name:"ClusterDesc"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterAttributeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAttributeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAuthenticationOptionsRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// ServiceAccount认证配置
	ServiceAccounts *ServiceAccountAuthenticationOptions `json:"ServiceAccounts,omitempty" name:"ServiceAccounts"`
}

func (r *ModifyClusterAuthenticationOptionsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAuthenticationOptionsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ServiceAccounts")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterAuthenticationOptionsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterAuthenticationOptionsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterAuthenticationOptionsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterAuthenticationOptionsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterEndpointSPRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 安全策略放通单个IP或CIDR(例如: "192.168.1.0/24",默认为拒绝所有)
	SecurityPolicies []*string `json:"SecurityPolicies,omitempty" name:"SecurityPolicies"`
}

func (r *ModifyClusterEndpointSPRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterEndpointSPRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "SecurityPolicies")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterEndpointSPRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterEndpointSPResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterEndpointSPResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterEndpointSPResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 最大节点数
	MaxNodesNum *int64 `json:"MaxNodesNum,omitempty" name:"MaxNodesNum"`

	// 最小节点数
	MinNodesNum *int64 `json:"MinNodesNum,omitempty" name:"MinNodesNum"`

	// 标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 污点
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 是否开启伸缩
	EnableAutoscale *bool `json:"EnableAutoscale,omitempty" name:"EnableAutoscale"`

	// 操作系统名称
	OsName *string `json:"OsName,omitempty" name:"OsName"`

	// 镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 节点自定义参数
	ExtraArgs *InstanceExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`

	// 资源标签
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`

	// 设置加入的节点是否参与调度，默认值为0，表示参与调度；非0表示不参与调度, 待节点初始化完成之后, 可执行kubectl uncordon nodename使node加入调度.
	Unschedulable *int64 `json:"Unschedulable,omitempty" name:"Unschedulable"`
}

func (r *ModifyClusterNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "Name")
	delete(f, "MaxNodesNum")
	delete(f, "MinNodesNum")
	delete(f, "Labels")
	delete(f, "Taints")
	delete(f, "EnableAutoscale")
	delete(f, "OsName")
	delete(f, "OsCustomizeType")
	delete(f, "ExtraArgs")
	delete(f, "Tags")
	delete(f, "Unschedulable")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyClusterNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyClusterNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyClusterNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyClusterNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyNodePoolDesiredCapacityAboutAsgRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点池所关联的伸缩组的期望实例数
	DesiredCapacity *int64 `json:"DesiredCapacity,omitempty" name:"DesiredCapacity"`
}

func (r *ModifyNodePoolDesiredCapacityAboutAsgRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyNodePoolDesiredCapacityAboutAsgRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "DesiredCapacity")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyNodePoolDesiredCapacityAboutAsgRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyNodePoolDesiredCapacityAboutAsgResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyNodePoolDesiredCapacityAboutAsgResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyNodePoolDesiredCapacityAboutAsgResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyNodePoolInstanceTypesRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 机型列表
	InstanceTypes []*string `json:"InstanceTypes,omitempty" name:"InstanceTypes"`
}

func (r *ModifyNodePoolInstanceTypesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyNodePoolInstanceTypesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "InstanceTypes")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyNodePoolInstanceTypesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyNodePoolInstanceTypesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyNodePoolInstanceTypesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyNodePoolInstanceTypesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyPrometheusAlertRuleRequest struct {
	*tchttp.BaseRequest

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 告警配置
	AlertRule *PrometheusAlertRuleDetail `json:"AlertRule,omitempty" name:"AlertRule"`
}

func (r *ModifyPrometheusAlertRuleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyPrometheusAlertRuleRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "InstanceId")
	delete(f, "AlertRule")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyPrometheusAlertRuleRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyPrometheusAlertRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyPrometheusAlertRuleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyPrometheusAlertRuleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyPrometheusTemplateRequest struct {
	*tchttp.BaseRequest

	// 模板ID
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

	// 修改内容
	Template *PrometheusTemplateModify `json:"Template,omitempty" name:"Template"`
}

func (r *ModifyPrometheusTemplateRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyPrometheusTemplateRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "TemplateId")
	delete(f, "Template")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyPrometheusTemplateRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyPrometheusTemplateResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyPrometheusTemplateResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyPrometheusTemplateResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type NfsVolume struct {

	// nfs volume 数据卷名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// NFS 服务器地址
	Server *string `json:"Server,omitempty" name:"Server"`

	// NFS 数据卷路径
	Path *string `json:"Path,omitempty" name:"Path"`

	// 默认为 False
	ReadOnly *bool `json:"ReadOnly,omitempty" name:"ReadOnly"`
}

type NodeCountSummary struct {

	// 手动管理的节点
	// 注意：此字段可能返回 null，表示取不到有效值。
	ManuallyAdded *ManuallyAdded `json:"ManuallyAdded,omitempty" name:"ManuallyAdded"`

	// 自动管理的节点
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoscalingAdded *AutoscalingAdded `json:"AutoscalingAdded,omitempty" name:"AutoscalingAdded"`
}

type NodePool struct {

	// NodePoolId 资源池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// Name 资源池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// ClusterInstanceId 集群实例id
	ClusterInstanceId *string `json:"ClusterInstanceId,omitempty" name:"ClusterInstanceId"`

	// LifeState 状态，当前节点池生命周期状态包括：creating，normal，updating，deleting，deleted
	LifeState *string `json:"LifeState,omitempty" name:"LifeState"`

	// LaunchConfigurationId 配置
	LaunchConfigurationId *string `json:"LaunchConfigurationId,omitempty" name:"LaunchConfigurationId"`

	// AutoscalingGroupId 分组id
	AutoscalingGroupId *string `json:"AutoscalingGroupId,omitempty" name:"AutoscalingGroupId"`

	// Labels 标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// Taints 污点标记
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// NodeCountSummary 节点列表
	NodeCountSummary *NodeCountSummary `json:"NodeCountSummary,omitempty" name:"NodeCountSummary"`

	// 状态信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoscalingGroupStatus *string `json:"AutoscalingGroupStatus,omitempty" name:"AutoscalingGroupStatus"`

	// 最大节点数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	MaxNodesNum *int64 `json:"MaxNodesNum,omitempty" name:"MaxNodesNum"`

	// 最小节点数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	MinNodesNum *int64 `json:"MinNodesNum,omitempty" name:"MinNodesNum"`

	// 期望的节点数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	DesiredNodesNum *int64 `json:"DesiredNodesNum,omitempty" name:"DesiredNodesNum"`

	// 节点池osName
	// 注意：此字段可能返回 null，表示取不到有效值。
	NodePoolOs *string `json:"NodePoolOs,omitempty" name:"NodePoolOs"`

	// 容器的镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	// 注意：此字段可能返回 null，表示取不到有效值。
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 镜像id
	// 注意：此字段可能返回 null，表示取不到有效值。
	ImageId *string `json:"ImageId,omitempty" name:"ImageId"`

	// 集群属于节点podCIDR大小自定义模式时，节点池需要带上pod数量属性
	// 注意：此字段可能返回 null，表示取不到有效值。
	DesiredPodNum *int64 `json:"DesiredPodNum,omitempty" name:"DesiredPodNum"`

	// 用户自定义脚本
	// 注意：此字段可能返回 null，表示取不到有效值。
	UserScript *string `json:"UserScript,omitempty" name:"UserScript"`

	// 资源标签
	// 注意：此字段可能返回 null，表示取不到有效值。
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`
}

type NodePoolOption struct {

	// 是否加入节点池
	AddToNodePool *bool `json:"AddToNodePool,omitempty" name:"AddToNodePool"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 是否继承节点池相关配置
	InheritConfigurationFromNodePool *bool `json:"InheritConfigurationFromNodePool,omitempty" name:"InheritConfigurationFromNodePool"`
}

type PodLimitsByType struct {

	// TKE共享网卡非固定IP模式可支持的Pod数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	TKERouteENINonStaticIP *int64 `json:"TKERouteENINonStaticIP,omitempty" name:"TKERouteENINonStaticIP"`

	// TKE共享网卡固定IP模式可支持的Pod数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	TKERouteENIStaticIP *int64 `json:"TKERouteENIStaticIP,omitempty" name:"TKERouteENIStaticIP"`

	// TKE独立网卡模式可支持的Pod数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	TKEDirectENI *int64 `json:"TKEDirectENI,omitempty" name:"TKEDirectENI"`
}

type PodLimitsInstance struct {

	// 机型所在可用区
	// 注意：此字段可能返回 null，表示取不到有效值。
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 机型所属机型族
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceFamily *string `json:"InstanceFamily,omitempty" name:"InstanceFamily"`

	// 实例机型名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 机型可支持的最大VPC-CNI模式Pod数量信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	PodLimits *PodLimitsByType `json:"PodLimits,omitempty" name:"PodLimits"`
}

type Probe struct {

	// Number of seconds after the container has started before liveness probes are initiated.
	// 注意：此字段可能返回 null，表示取不到有效值。
	InitialDelaySeconds *int64 `json:"InitialDelaySeconds,omitempty" name:"InitialDelaySeconds"`

	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	// 注意：此字段可能返回 null，表示取不到有效值。
	TimeoutSeconds *int64 `json:"TimeoutSeconds,omitempty" name:"TimeoutSeconds"`

	// How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
	// 注意：此字段可能返回 null，表示取不到有效值。
	PeriodSeconds *int64 `json:"PeriodSeconds,omitempty" name:"PeriodSeconds"`

	// Minimum consecutive successes for the probe to be considered successful after having failed.Defaults to 1. Must be 1 for liveness. Minimum value is 1.
	// 注意：此字段可能返回 null，表示取不到有效值。
	SuccessThreshold *int64 `json:"SuccessThreshold,omitempty" name:"SuccessThreshold"`

	// Minimum consecutive failures for the probe to be considered failed after having succeeded.Defaults to 3. Minimum value is 1.
	// 注意：此字段可能返回 null，表示取不到有效值。
	FailureThreshold *int64 `json:"FailureThreshold,omitempty" name:"FailureThreshold"`
}

type PrometheusAgentOverview struct {

	// 集群类型
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// agent状态
	// normal = 正常
	// abnormal = 异常
	Status *string `json:"Status,omitempty" name:"Status"`

	// 集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 额外labels
	// 本集群的所有指标都会带上这几个label
	// 注意：此字段可能返回 null，表示取不到有效值。
	ExternalLabels []*Label `json:"ExternalLabels,omitempty" name:"ExternalLabels"`
}

type PrometheusAlertHistoryItem struct {

	// 告警名称
	RuleName *string `json:"RuleName,omitempty" name:"RuleName"`

	// 告警开始时间
	StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

	// 告警内容
	Content *string `json:"Content,omitempty" name:"Content"`

	// 告警状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	State *string `json:"State,omitempty" name:"State"`

	// 触发的规则名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	RuleItem *string `json:"RuleItem,omitempty" name:"RuleItem"`

	// 告警渠道的id
	// 注意：此字段可能返回 null，表示取不到有效值。
	TopicId *string `json:"TopicId,omitempty" name:"TopicId"`

	// 告警渠道的名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	TopicName *string `json:"TopicName,omitempty" name:"TopicName"`
}

type PrometheusAlertRule struct {

	// 规则名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// prometheus语句
	Rule *string `json:"Rule,omitempty" name:"Rule"`

	// 额外标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 告警发送模板
	Template *string `json:"Template,omitempty" name:"Template"`

	// 持续时间
	For *string `json:"For,omitempty" name:"For"`

	// 该条规则的描述信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Describe *string `json:"Describe,omitempty" name:"Describe"`

	// 参考prometheus rule中的annotations
	// 注意：此字段可能返回 null，表示取不到有效值。
	Annotations []*Label `json:"Annotations,omitempty" name:"Annotations"`

	// 告警规则状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	RuleState *int64 `json:"RuleState,omitempty" name:"RuleState"`
}

type PrometheusAlertRuleDetail struct {

	// 规则名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 规则列表
	Rules []*PrometheusAlertRule `json:"Rules,omitempty" name:"Rules"`

	// 最后修改时间
	UpdatedAt *string `json:"UpdatedAt,omitempty" name:"UpdatedAt"`

	// 告警渠道
	Notification *PrometheusNotification `json:"Notification,omitempty" name:"Notification"`

	// 告警 id
	Id *string `json:"Id,omitempty" name:"Id"`

	// 如果该告警来至模板下发，则TemplateId为模板id
	// 注意：此字段可能返回 null，表示取不到有效值。
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

	// 计算周期
	// 注意：此字段可能返回 null，表示取不到有效值。
	Interval *string `json:"Interval,omitempty" name:"Interval"`
}

type PrometheusConfigItem struct {

	// 名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 配置内容
	Config *string `json:"Config,omitempty" name:"Config"`

	// 用于出参，如果该配置来至模板，则为模板id
	// 注意：此字段可能返回 null，表示取不到有效值。
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`
}

type PrometheusGrafanaInfo struct {

	// 是否启用
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`

	// 域名，只有开启外网访问才有效果
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 内网地址，或者外网地址
	Address *string `json:"Address,omitempty" name:"Address"`

	// 是否开启了外网访问
	// close = 未开启外网访问
	// opening = 正在开启外网访问
	// open  = 已开启外网访问
	Internet *string `json:"Internet,omitempty" name:"Internet"`

	// grafana管理员用户名
	AdminUser *string `json:"AdminUser,omitempty" name:"AdminUser"`
}

type PrometheusInstanceOverview struct {

	// 实例id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 实例名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 实例vpcId
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 实例子网Id
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 实例当前的状态
	// prepare_env = 初始化环境
	// install_suit = 安装组件
	// running = 运行中
	Status *string `json:"Status,omitempty" name:"Status"`

	// COS桶存储
	COSBucket *string `json:"COSBucket,omitempty" name:"COSBucket"`

	// grafana默认地址，如果开启外网访问得为域名，否则为内网地址
	// 注意：此字段可能返回 null，表示取不到有效值。
	GrafanaURL *string `json:"GrafanaURL,omitempty" name:"GrafanaURL"`

	// 关联集群总数
	// 注意：此字段可能返回 null，表示取不到有效值。
	BoundTotal *uint64 `json:"BoundTotal,omitempty" name:"BoundTotal"`

	// 运行正常的集群数
	// 注意：此字段可能返回 null，表示取不到有效值。
	BoundNormal *uint64 `json:"BoundNormal,omitempty" name:"BoundNormal"`
}

type PrometheusJobTargets struct {

	// 该Job的targets列表
	Targets []*PrometheusTarget `json:"Targets,omitempty" name:"Targets"`

	// job的名称
	JobName *string `json:"JobName,omitempty" name:"JobName"`

	// targets总数
	Total *uint64 `json:"Total,omitempty" name:"Total"`

	// 健康的target总数
	Up *uint64 `json:"Up,omitempty" name:"Up"`
}

type PrometheusNotification struct {

	// 是否启用
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`

	// 收敛时间
	RepeatInterval *string `json:"RepeatInterval,omitempty" name:"RepeatInterval"`

	// 生效起始时间
	TimeRangeStart *string `json:"TimeRangeStart,omitempty" name:"TimeRangeStart"`

	// 生效结束时间
	TimeRangeEnd *string `json:"TimeRangeEnd,omitempty" name:"TimeRangeEnd"`

	// 告警通知方式。目前有SMS、EMAIL、CALL、WECHAT方式。
	// 分别代表：短信、邮件、电话、微信
	// 注意：此字段可能返回 null，表示取不到有效值。
	NotifyWay []*string `json:"NotifyWay,omitempty" name:"NotifyWay"`

	// 告警接收组（用户组）
	// 注意：此字段可能返回 null，表示取不到有效值。
	ReceiverGroups []*uint64 `json:"ReceiverGroups,omitempty" name:"ReceiverGroups"`

	// 电话告警顺序。
	// 注：NotifyWay选择CALL，采用该参数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PhoneNotifyOrder []*uint64 `json:"PhoneNotifyOrder,omitempty" name:"PhoneNotifyOrder"`

	// 电话告警次数。
	// 注：NotifyWay选择CALL，采用该参数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PhoneCircleTimes *int64 `json:"PhoneCircleTimes,omitempty" name:"PhoneCircleTimes"`

	// 电话告警轮内间隔。单位：秒
	// 注：NotifyWay选择CALL，采用该参数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PhoneInnerInterval *int64 `json:"PhoneInnerInterval,omitempty" name:"PhoneInnerInterval"`

	// 电话告警轮外间隔。单位：秒
	// 注：NotifyWay选择CALL，采用该参数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PhoneCircleInterval *int64 `json:"PhoneCircleInterval,omitempty" name:"PhoneCircleInterval"`

	// 电话告警触达通知
	// 注：NotifyWay选择CALL，采用该参数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PhoneArriveNotice *bool `json:"PhoneArriveNotice,omitempty" name:"PhoneArriveNotice"`

	// 通道类型，默认为amp，支持以下
	// amp
	// webhook
	// 注意：此字段可能返回 null，表示取不到有效值。
	Type *string `json:"Type,omitempty" name:"Type"`

	// 如果Type为webhook, 则该字段为必填项
	// 注意：此字段可能返回 null，表示取不到有效值。
	WebHook *string `json:"WebHook,omitempty" name:"WebHook"`
}

type PrometheusTarget struct {

	// 抓取目标的URL
	Url *string `json:"Url,omitempty" name:"Url"`

	// target当前状态,当前支持
	// up = 健康
	// down = 不健康
	// unknown = 未知
	State *string `json:"State,omitempty" name:"State"`

	// target的元label
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 上一次抓取的时间
	LastScrape *string `json:"LastScrape,omitempty" name:"LastScrape"`

	// 上一次抓取的耗时，单位是s
	ScrapeDuration *float64 `json:"ScrapeDuration,omitempty" name:"ScrapeDuration"`

	// 上一次抓取如果错误，该字段存储错误信息
	Error *string `json:"Error,omitempty" name:"Error"`
}

type PrometheusTemplate struct {

	// 模板名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 模板维度，支持以下类型
	// instance 实例级别
	// cluster 集群级别
	Level *string `json:"Level,omitempty" name:"Level"`

	// 模板描述
	// 注意：此字段可能返回 null，表示取不到有效值。
	Describe *string `json:"Describe,omitempty" name:"Describe"`

	// 当Level为instance时有效，
	// 模板中的告警配置列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	AlertRules []*PrometheusAlertRule `json:"AlertRules,omitempty" name:"AlertRules"`

	// 当Level为instance时有效，
	// 模板中的聚合规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	RecordRules []*PrometheusConfigItem `json:"RecordRules,omitempty" name:"RecordRules"`

	// 当Level为cluster时有效，
	// 模板中的ServiceMonitor规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	ServiceMonitors []*PrometheusConfigItem `json:"ServiceMonitors,omitempty" name:"ServiceMonitors"`

	// 当Level为cluster时有效，
	// 模板中的PodMonitors规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	PodMonitors []*PrometheusConfigItem `json:"PodMonitors,omitempty" name:"PodMonitors"`

	// 当Level为cluster时有效，
	// 模板中的RawJobs规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	RawJobs []*PrometheusConfigItem `json:"RawJobs,omitempty" name:"RawJobs"`

	// 模板的ID, 用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

	// 最近更新时间，用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	UpdateTime *string `json:"UpdateTime,omitempty" name:"UpdateTime"`

	// 当前版本，用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	Version *string `json:"Version,omitempty" name:"Version"`

	// 是否系统提供的默认模板，用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsDefault *bool `json:"IsDefault,omitempty" name:"IsDefault"`

	// 当Level为instance时有效，
	// 模板中的告警配置列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	AlertDetailRules []*PrometheusAlertRuleDetail `json:"AlertDetailRules,omitempty" name:"AlertDetailRules"`
}

type PrometheusTemplateModify struct {

	// 修改名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 修改描述
	// 注意：此字段可能返回 null，表示取不到有效值。
	Describe *string `json:"Describe,omitempty" name:"Describe"`

	// 修改内容，只有当模板类型是Alert时生效
	// 注意：此字段可能返回 null，表示取不到有效值。
	AlertRules []*PrometheusAlertRule `json:"AlertRules,omitempty" name:"AlertRules"`

	// 当Level为instance时有效，
	// 模板中的聚合规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	RecordRules []*PrometheusConfigItem `json:"RecordRules,omitempty" name:"RecordRules"`

	// 当Level为cluster时有效，
	// 模板中的ServiceMonitor规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	ServiceMonitors []*PrometheusConfigItem `json:"ServiceMonitors,omitempty" name:"ServiceMonitors"`

	// 当Level为cluster时有效，
	// 模板中的PodMonitors规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	PodMonitors []*PrometheusConfigItem `json:"PodMonitors,omitempty" name:"PodMonitors"`

	// 当Level为cluster时有效，
	// 模板中的RawJobs规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	RawJobs []*PrometheusConfigItem `json:"RawJobs,omitempty" name:"RawJobs"`

	// 修改内容，只有当模板类型是Alert时生效
	// 注意：此字段可能返回 null，表示取不到有效值。
	AlertDetailRules []*PrometheusAlertRuleDetail `json:"AlertDetailRules,omitempty" name:"AlertDetailRules"`
}

type PrometheusTemplateSyncTarget struct {

	// 目标所在地域
	Region *string `json:"Region,omitempty" name:"Region"`

	// 目标实例
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 集群id，只有当采集模板的Level为cluster的时候需要
	// 注意：此字段可能返回 null，表示取不到有效值。
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 最后一次同步时间， 用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	SyncTime *string `json:"SyncTime,omitempty" name:"SyncTime"`

	// 当前使用的模板版本，用于出参
	// 注意：此字段可能返回 null，表示取不到有效值。
	Version *string `json:"Version,omitempty" name:"Version"`

	// 集群类型，只有当采集模板的Level为cluster的时候需要
	// 注意：此字段可能返回 null，表示取不到有效值。
	ClusterType *string `json:"ClusterType,omitempty" name:"ClusterType"`

	// 用于出参，实例名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceName *string `json:"InstanceName,omitempty" name:"InstanceName"`

	// 用于出参，集群名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`
}

type RegionInstance struct {

	// 地域名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	RegionName *string `json:"RegionName,omitempty" name:"RegionName"`

	// 地域ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	RegionId *int64 `json:"RegionId,omitempty" name:"RegionId"`

	// 地域状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	Status *string `json:"Status,omitempty" name:"Status"`

	// 地域特性开关(按照JSON的形式返回所有属性)
	// 注意：此字段可能返回 null，表示取不到有效值。
	FeatureGates *string `json:"FeatureGates,omitempty" name:"FeatureGates"`

	// 地域简称
	// 注意：此字段可能返回 null，表示取不到有效值。
	Alias *string `json:"Alias,omitempty" name:"Alias"`

	// 地域白名单
	// 注意：此字段可能返回 null，表示取不到有效值。
	Remark *string `json:"Remark,omitempty" name:"Remark"`
}

type RemoveNodeFromNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点id列表，一次最多支持100台
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`
}

func (r *RemoveNodeFromNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *RemoveNodeFromNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "InstanceIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "RemoveNodeFromNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type RemoveNodeFromNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *RemoveNodeFromNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *RemoveNodeFromNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ResourceDeleteOption struct {

	// 资源类型，例如CBS
	ResourceType *string `json:"ResourceType,omitempty" name:"ResourceType"`

	// 集群删除时资源的删除模式：terminate（销毁），retain （保留）
	DeleteMode *string `json:"DeleteMode,omitempty" name:"DeleteMode"`
}

type RestartEKSContainerInstancesRequest struct {
	*tchttp.BaseRequest

	// EKS instance ids
	EksCiIds []*string `json:"EksCiIds,omitempty" name:"EksCiIds"`
}

func (r *RestartEKSContainerInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *RestartEKSContainerInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "EksCiIds")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "RestartEKSContainerInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type RestartEKSContainerInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *RestartEKSContainerInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *RestartEKSContainerInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type RouteInfo struct {

	// 路由表名称。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 目的端CIDR。
	DestinationCidrBlock *string `json:"DestinationCidrBlock,omitempty" name:"DestinationCidrBlock"`

	// 下一跳地址。
	GatewayIp *string `json:"GatewayIp,omitempty" name:"GatewayIp"`
}

type RouteTableConflict struct {

	// 路由表类型。
	RouteTableType *string `json:"RouteTableType,omitempty" name:"RouteTableType"`

	// 路由表CIDR。
	// 注意：此字段可能返回 null，表示取不到有效值。
	RouteTableCidrBlock *string `json:"RouteTableCidrBlock,omitempty" name:"RouteTableCidrBlock"`

	// 路由表名称。
	// 注意：此字段可能返回 null，表示取不到有效值。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 路由表ID。
	// 注意：此字段可能返回 null，表示取不到有效值。
	RouteTableId *string `json:"RouteTableId,omitempty" name:"RouteTableId"`
}

type RouteTableInfo struct {

	// 路由表名称。
	RouteTableName *string `json:"RouteTableName,omitempty" name:"RouteTableName"`

	// 路由表CIDR。
	RouteTableCidrBlock *string `json:"RouteTableCidrBlock,omitempty" name:"RouteTableCidrBlock"`

	// VPC实例ID。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`
}

type RunAutomationServiceEnabled struct {

	// 是否开启云自动化助手。取值范围：<br><li>TRUE：表示开启云自动化助手服务<br><li>FALSE：表示不开启云自动化助手服务<br><br>默认取值：FALSE。
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

type RunInstancesForNode struct {

	// 节点角色，取值:MASTER_ETCD, WORKER。MASTER_ETCD只有在创建 INDEPENDENT_CLUSTER 独立集群时需要指定。MASTER_ETCD节点数量为3～7，建议为奇数。MASTER_ETCD节点最小配置为4C8G。
	NodeRole *string `json:"NodeRole,omitempty" name:"NodeRole"`

	// CVM创建透传参数，json化字符串格式，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口，传入公共参数外的其他参数即可，其中ImageId会替换为TKE集群OS对应的镜像。
	RunInstancesPara []*string `json:"RunInstancesPara,omitempty" name:"RunInstancesPara"`

	// 节点高级设置，该参数会覆盖集群级别设置的InstanceAdvancedSettings，和上边的RunInstancesPara按照顺序一一对应（当前只对节点自定义参数ExtraArgs生效）。
	InstanceAdvancedSettingsOverrides []*InstanceAdvancedSettings `json:"InstanceAdvancedSettingsOverrides,omitempty" name:"InstanceAdvancedSettingsOverrides"`
}

type RunMonitorServiceEnabled struct {

	// 是否开启[云监控](/document/product/248)服务。取值范围：<br><li>TRUE：表示开启云监控服务<br><li>FALSE：表示不开启云监控服务<br><br>默认取值：TRUE。
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

type RunSecurityServiceEnabled struct {

	// 是否开启[云安全](/document/product/296)服务。取值范围：<br><li>TRUE：表示开启云安全服务<br><li>FALSE：表示不开启云安全服务<br><br>默认取值：TRUE。
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

type ScaleInClusterMasterRequest struct {
	*tchttp.BaseRequest

	// 集群实例ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// master缩容选项
	ScaleInMasters []*ScaleInMaster `json:"ScaleInMasters,omitempty" name:"ScaleInMasters"`
}

func (r *ScaleInClusterMasterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ScaleInClusterMasterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ScaleInMasters")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ScaleInClusterMasterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ScaleInClusterMasterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ScaleInClusterMasterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ScaleInClusterMasterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ScaleInMaster struct {

	// 实例ID
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 缩容的实例角色：MASTER,ETCD,MASTER_ETCD
	NodeRole *string `json:"NodeRole,omitempty" name:"NodeRole"`

	// 实例的保留模式
	InstanceDeleteMode *string `json:"InstanceDeleteMode,omitempty" name:"InstanceDeleteMode"`
}

type ScaleOutClusterMasterRequest struct {
	*tchttp.BaseRequest

	// 集群实例ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 新建节点参数
	RunInstancesForNode []*RunInstancesForNode `json:"RunInstancesForNode,omitempty" name:"RunInstancesForNode"`

	// 添加已有节点相关参数
	ExistedInstancesForNode []*ExistedInstancesForNode `json:"ExistedInstancesForNode,omitempty" name:"ExistedInstancesForNode"`

	// 实例高级设置
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 集群master组件自定义参数
	ExtraArgs *ClusterExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`
}

func (r *ScaleOutClusterMasterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ScaleOutClusterMasterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "RunInstancesForNode")
	delete(f, "ExistedInstancesForNode")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "ExtraArgs")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ScaleOutClusterMasterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ScaleOutClusterMasterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ScaleOutClusterMasterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ScaleOutClusterMasterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type SecurityContext struct {

	// 安全能力清单
	// 注意：此字段可能返回 null，表示取不到有效值。
	Capabilities *Capabilities `json:"Capabilities,omitempty" name:"Capabilities"`
}

type ServiceAccountAuthenticationOptions struct {

	// service-account-issuer
	// 注意：此字段可能返回 null，表示取不到有效值。
	Issuer *string `json:"Issuer,omitempty" name:"Issuer"`

	// service-account-jwks-uri
	// 注意：此字段可能返回 null，表示取不到有效值。
	JWKSURI *string `json:"JWKSURI,omitempty" name:"JWKSURI"`

	// 如果为true，则会自动创建允许匿名用户访问'/.well-known/openid-configuration'和/openid/v1/jwks的rbac规则
	// 注意：此字段可能返回 null，表示取不到有效值。
	AutoCreateDiscoveryAnonymousAuth *bool `json:"AutoCreateDiscoveryAnonymousAuth,omitempty" name:"AutoCreateDiscoveryAnonymousAuth"`
}

type SetNodePoolNodeProtectionRequest struct {
	*tchttp.BaseRequest

	// 集群id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池id
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点id
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 节点是否需要移出保护
	ProtectedFromScaleIn *bool `json:"ProtectedFromScaleIn,omitempty" name:"ProtectedFromScaleIn"`
}

func (r *SetNodePoolNodeProtectionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *SetNodePoolNodeProtectionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "InstanceIds")
	delete(f, "ProtectedFromScaleIn")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "SetNodePoolNodeProtectionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type SetNodePoolNodeProtectionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 成功设置的节点id
		// 注意：此字段可能返回 null，表示取不到有效值。
		SucceedInstanceIds []*string `json:"SucceedInstanceIds,omitempty" name:"SucceedInstanceIds"`

		// 没有成功设置的节点id
		// 注意：此字段可能返回 null，表示取不到有效值。
		FailedInstanceIds []*string `json:"FailedInstanceIds,omitempty" name:"FailedInstanceIds"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *SetNodePoolNodeProtectionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *SetNodePoolNodeProtectionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type SyncPrometheusTemplateRequest struct {
	*tchttp.BaseRequest

	// 实例id
	TemplateId *string `json:"TemplateId,omitempty" name:"TemplateId"`

	// 同步目标
	Targets []*PrometheusTemplateSyncTarget `json:"Targets,omitempty" name:"Targets"`
}

func (r *SyncPrometheusTemplateRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *SyncPrometheusTemplateRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "TemplateId")
	delete(f, "Targets")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "SyncPrometheusTemplateRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type SyncPrometheusTemplateResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *SyncPrometheusTemplateResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *SyncPrometheusTemplateResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type Tag struct {

	// 标签键
	Key *string `json:"Key,omitempty" name:"Key"`

	// 标签值
	Value *string `json:"Value,omitempty" name:"Value"`
}

type TagSpecification struct {

	// 标签绑定的资源类型，当前支持类型："cluster"
	// 注意：此字段可能返回 null，表示取不到有效值。
	ResourceType *string `json:"ResourceType,omitempty" name:"ResourceType"`

	// 标签对列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`
}

type Taint struct {

	// Key
	Key *string `json:"Key,omitempty" name:"Key"`

	// Value
	Value *string `json:"Value,omitempty" name:"Value"`

	// Effect
	Effect *string `json:"Effect,omitempty" name:"Effect"`
}

type TaskStepInfo struct {

	// 步骤名称
	Step *string `json:"Step,omitempty" name:"Step"`

	// 生命周期
	// pending : 步骤未开始
	// running: 步骤执行中
	// success: 步骤成功完成
	// failed: 步骤失败
	LifeState *string `json:"LifeState,omitempty" name:"LifeState"`

	// 步骤开始时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	StartAt *string `json:"StartAt,omitempty" name:"StartAt"`

	// 步骤结束时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	EndAt *string `json:"EndAt,omitempty" name:"EndAt"`

	// 若步骤生命周期为failed,则此字段显示错误信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	FailedMsg *string `json:"FailedMsg,omitempty" name:"FailedMsg"`
}

type TcpSocket struct {

	// TcpSocket检测的端口
	// 注意：此字段可能返回 null，表示取不到有效值。
	Port *uint64 `json:"Port,omitempty" name:"Port"`
}

type UpdateClusterVersionRequest struct {
	*tchttp.BaseRequest

	// 集群 Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 需要升级到的版本
	DstVersion *string `json:"DstVersion,omitempty" name:"DstVersion"`

	// 集群自定义参数
	ExtraArgs *ClusterExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`

	// 可容忍的最大不可用pod数目
	MaxNotReadyPercent *float64 `json:"MaxNotReadyPercent,omitempty" name:"MaxNotReadyPercent"`

	// 是否跳过预检查阶段
	SkipPreCheck *bool `json:"SkipPreCheck,omitempty" name:"SkipPreCheck"`
}

func (r *UpdateClusterVersionRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateClusterVersionRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "DstVersion")
	delete(f, "ExtraArgs")
	delete(f, "MaxNotReadyPercent")
	delete(f, "SkipPreCheck")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "UpdateClusterVersionRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type UpdateClusterVersionResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *UpdateClusterVersionResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateClusterVersionResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type UpdateEKSClusterRequest struct {
	*tchttp.BaseRequest

	// 弹性集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 弹性集群名称
	ClusterName *string `json:"ClusterName,omitempty" name:"ClusterName"`

	// 弹性集群描述信息
	ClusterDesc *string `json:"ClusterDesc,omitempty" name:"ClusterDesc"`

	// 子网Id 列表
	SubnetIds []*string `json:"SubnetIds,omitempty" name:"SubnetIds"`

	// 弹性容器集群公网访问LB信息
	PublicLB *ClusterPublicLB `json:"PublicLB,omitempty" name:"PublicLB"`

	// 弹性容器集群内网访问LB信息
	InternalLB *ClusterInternalLB `json:"InternalLB,omitempty" name:"InternalLB"`

	// Service 子网Id
	ServiceSubnetId *string `json:"ServiceSubnetId,omitempty" name:"ServiceSubnetId"`

	// 集群自定义的dns 服务器信息
	DnsServers []*DnsServerConf `json:"DnsServers,omitempty" name:"DnsServers"`

	// 是否清空自定义dns 服务器设置。为1 表示 是。其他表示 否。
	ClearDnsServer *string `json:"ClearDnsServer,omitempty" name:"ClearDnsServer"`

	// 将来删除集群时是否要删除cbs。默认为 FALSE
	NeedDeleteCbs *bool `json:"NeedDeleteCbs,omitempty" name:"NeedDeleteCbs"`

	// 标记是否是新的内外网。默认为false
	ProxyLB *bool `json:"ProxyLB,omitempty" name:"ProxyLB"`

	// 扩展参数。须是map[string]string 的json 格式。
	ExtraParam *string `json:"ExtraParam,omitempty" name:"ExtraParam"`
}

func (r *UpdateEKSClusterRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateEKSClusterRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterName")
	delete(f, "ClusterDesc")
	delete(f, "SubnetIds")
	delete(f, "PublicLB")
	delete(f, "InternalLB")
	delete(f, "ServiceSubnetId")
	delete(f, "DnsServers")
	delete(f, "ClearDnsServer")
	delete(f, "NeedDeleteCbs")
	delete(f, "ProxyLB")
	delete(f, "ExtraParam")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "UpdateEKSClusterRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type UpdateEKSClusterResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *UpdateEKSClusterResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateEKSClusterResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type UpdateEKSContainerInstanceRequest struct {
	*tchttp.BaseRequest

	// 容器实例 ID
	EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

	// 实例重启策略： Always(总是重启)、Never(从不重启)、OnFailure(失败时重启)
	RestartPolicy *string `json:"RestartPolicy,omitempty" name:"RestartPolicy"`

	// 数据卷，包含NfsVolume数组和CbsVolume数组
	EksCiVolume *EksCiVolume `json:"EksCiVolume,omitempty" name:"EksCiVolume"`

	// 容器组
	Containers []*Container `json:"Containers,omitempty" name:"Containers"`

	// Init 容器组
	InitContainers []*Container `json:"InitContainers,omitempty" name:"InitContainers"`

	// 容器实例名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 镜像仓库凭证数组
	ImageRegistryCredentials []*ImageRegistryCredential `json:"ImageRegistryCredentials,omitempty" name:"ImageRegistryCredentials"`
}

func (r *UpdateEKSContainerInstanceRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateEKSContainerInstanceRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "EksCiId")
	delete(f, "RestartPolicy")
	delete(f, "EksCiVolume")
	delete(f, "Containers")
	delete(f, "InitContainers")
	delete(f, "Name")
	delete(f, "ImageRegistryCredentials")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "UpdateEKSContainerInstanceRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type UpdateEKSContainerInstanceResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 容器实例 ID
		// 注意：此字段可能返回 null，表示取不到有效值。
		EksCiId *string `json:"EksCiId,omitempty" name:"EksCiId"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *UpdateEKSContainerInstanceResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpdateEKSContainerInstanceResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type UpgradeAbleInstancesItem struct {

	// 节点Id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 节点的当前版本
	Version *string `json:"Version,omitempty" name:"Version"`

	// 当前版本的最新小版本
	// 注意：此字段可能返回 null，表示取不到有效值。
	LatestVersion *string `json:"LatestVersion,omitempty" name:"LatestVersion"`
}

type UpgradeClusterInstancesRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// create 表示开始一次升级任务
	// pause 表示停止任务
	// resume表示继续任务
	// abort表示终止任务
	Operation *string `json:"Operation,omitempty" name:"Operation"`

	// 升级类型，只有Operation是create需要设置
	// reset 大版本重装升级
	// hot 小版本热升级
	// major 大版本原地升级
	UpgradeType *string `json:"UpgradeType,omitempty" name:"UpgradeType"`

	// 需要升级的节点列表
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds"`

	// 当节点重新加入集群时候所使用的参数，参考添加已有节点接口
	ResetParam *UpgradeNodeResetParam `json:"ResetParam,omitempty" name:"ResetParam"`

	// 是否忽略节点升级前检查
	SkipPreCheck *bool `json:"SkipPreCheck,omitempty" name:"SkipPreCheck"`

	// 最大可容忍的不可用Pod比例
	MaxNotReadyPercent *float64 `json:"MaxNotReadyPercent,omitempty" name:"MaxNotReadyPercent"`
}

func (r *UpgradeClusterInstancesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpgradeClusterInstancesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Operation")
	delete(f, "UpgradeType")
	delete(f, "InstanceIds")
	delete(f, "ResetParam")
	delete(f, "SkipPreCheck")
	delete(f, "MaxNotReadyPercent")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "UpgradeClusterInstancesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type UpgradeClusterInstancesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *UpgradeClusterInstancesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *UpgradeClusterInstancesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type UpgradeNodeResetParam struct {

	// 实例额外需要设置参数信息
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"`

	// 增强服务。通过该参数可以指定是否开启云安全、云监控等服务。若不指定该参数，则默认开启云监控、云安全服务。
	EnhancedService *EnhancedService `json:"EnhancedService,omitempty" name:"EnhancedService"`

	// 节点登录信息（目前仅支持使用Password或者单个KeyIds）
	LoginSettings *LoginSettings `json:"LoginSettings,omitempty" name:"LoginSettings"`

	// 实例所属安全组。该参数可以通过调用 DescribeSecurityGroups 的返回值中的sgId字段来获取。若不指定该参数，则绑定默认安全组。（目前仅支持设置单个sgId）
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`
}

type VersionInstance struct {

	// 版本名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	Name *string `json:"Name,omitempty" name:"Name"`

	// 版本信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Version *string `json:"Version,omitempty" name:"Version"`

	// Remark
	// 注意：此字段可能返回 null，表示取不到有效值。
	Remark *string `json:"Remark,omitempty" name:"Remark"`
}

type VolumeMount struct {

	// volume名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	Name *string `json:"Name,omitempty" name:"Name"`

	// 挂载路径
	// 注意：此字段可能返回 null，表示取不到有效值。
	MountPath *string `json:"MountPath,omitempty" name:"MountPath"`

	// 是否只读
	// 注意：此字段可能返回 null，表示取不到有效值。
	ReadOnly *bool `json:"ReadOnly,omitempty" name:"ReadOnly"`

	// 子路径
	// 注意：此字段可能返回 null，表示取不到有效值。
	SubPath *string `json:"SubPath,omitempty" name:"SubPath"`

	// 传播挂载方式
	// 注意：此字段可能返回 null，表示取不到有效值。
	MountPropagation *string `json:"MountPropagation,omitempty" name:"MountPropagation"`

	// 子路径表达式
	// 注意：此字段可能返回 null，表示取不到有效值。
	SubPathExpr *string `json:"SubPathExpr,omitempty" name:"SubPathExpr"`
}
