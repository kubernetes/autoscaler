package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 实例配置信息
type InstanceConfigResult struct {
	// 云服务器的规格ID。

	FlavorRef *string `json:"flavorRef,omitempty"`
	// 镜像ID，同image_id。

	ImageRef *string `json:"imageRef,omitempty"`
	// 磁盘组信息。

	Disk *[]DiskResult `json:"disk,omitempty"`
	// 登录云服务器的SSH密钥名称。

	KeyName *string `json:"key_name,omitempty"`
	// 登录云服务器的SSH密钥指纹。

	KeyFingerprint *string `json:"key_fingerprint,omitempty"`
	// 该参数为预留字段。

	InstanceName *string `json:"instance_name,omitempty"`
	// 该参数为预留字段。

	InstanceId *string `json:"instance_id,omitempty"`
	// 登录云服务器的密码，非明文回显。

	AdminPass *string `json:"adminPass,omitempty"`
	// 个人信息

	Personality *[]PersonalityResult `json:"personality,omitempty"`

	PublicIp *PublicipResult `json:"public_ip,omitempty"`
	// cloud-init用户数据，base64格式编码。

	UserData *string `json:"user_data,omitempty"`

	Metadata *VmMetaData `json:"metadata,omitempty"`
	// 安全组信息。

	SecurityGroups *[]SecurityGroups `json:"security_groups,omitempty"`
	// 云服务器组ID。

	ServerGroupId *string `json:"server_group_id,omitempty"`
	// 在专属主机上创建弹性云服务器。

	Tenancy *string `json:"tenancy,omitempty"`
	// 专属主机的ID。

	DedicatedHostId *string `json:"dedicated_host_id,omitempty"`
	// 云服务器的计费模式，可以选择竞价计费或按需计费。

	MarketType *string `json:"market_type,omitempty"`
	// 使用伸缩配置创建云主机的时候，多规格使用的优先级策略。

	MultiFlavorPriorityPolicy *string `json:"multi_flavor_priority_policy,omitempty"`
}

func (o InstanceConfigResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "InstanceConfigResult struct{}"
	}

	return strings.Join([]string{"InstanceConfigResult", string(data)}, " ")
}
