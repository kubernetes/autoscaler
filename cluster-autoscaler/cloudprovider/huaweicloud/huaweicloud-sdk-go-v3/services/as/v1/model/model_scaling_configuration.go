package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 伸缩配置详情
type ScalingConfiguration struct {
	// 伸缩配置ID，全局唯一。

	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 租户ID。

	Tenant *string `json:"tenant,omitempty"`
	// 伸缩配置名称。

	ScalingConfigurationName *string `json:"scaling_configuration_name,omitempty"`

	InstanceConfig *InstanceConfigResult `json:"instance_config,omitempty"`
	// 创建伸缩配置的时间，遵循UTC时间。

	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 绑定该伸缩配置的伸缩组ID

	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
}

func (o ScalingConfiguration) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingConfiguration struct{}"
	}

	return strings.Join([]string{"ScalingConfiguration", string(data)}, " ")
}
