package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 伸缩实例。
type ScalingInstance struct {
	// 云服务器名称。

	InstanceName *string `json:"instance_name,omitempty"`
	// 云服务器id。

	InstanceId *string `json:"instance_id,omitempty"`
	// 实例伸缩失败原因。

	FailedReason *string `json:"failed_reason,omitempty"`
	// 实例伸缩失败详情。

	FailedDetails *string `json:"failed_details,omitempty"`
	// 实例配置信息。

	InstanceConfig *string `json:"instance_config,omitempty"`
}

func (o ScalingInstance) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingInstance struct{}"
	}

	return strings.Join([]string{"ScalingInstance", string(data)}, " ")
}
