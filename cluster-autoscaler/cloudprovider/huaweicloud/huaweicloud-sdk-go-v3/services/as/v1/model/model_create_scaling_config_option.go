package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 创建伸缩配置请求
type CreateScalingConfigOption struct {
	// 伸缩配置名称(1-64个字符)，只能包含中文、字母、数字、下划线或中划线。

	ScalingConfigurationName string `json:"scaling_configuration_name"`

	InstanceConfig *InstanceConfig `json:"instance_config"`
}

func (o CreateScalingConfigOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingConfigOption struct{}"
	}

	return strings.Join([]string{"CreateScalingConfigOption", string(data)}, " ")
}
