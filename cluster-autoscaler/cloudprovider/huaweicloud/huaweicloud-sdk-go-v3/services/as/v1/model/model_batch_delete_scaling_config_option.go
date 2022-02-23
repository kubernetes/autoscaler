package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 批量删除伸缩配置请求
type BatchDeleteScalingConfigOption struct {
	// 伸缩配置ID。

	ScalingConfigurationId []string `json:"scaling_configuration_id"`
}

func (o BatchDeleteScalingConfigOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingConfigOption struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingConfigOption", string(data)}, " ")
}
