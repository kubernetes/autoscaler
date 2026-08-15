package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListScalingConfigsRequest struct {
	// 伸缩配置名称。

	ScalingConfigurationName *string `json:"scaling_configuration_name,omitempty"`
	// 镜像ID，同imageRef。

	ImageId *string `json:"image_id,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询的记录条数，默认为20。

	Limit *int32 `json:"limit,omitempty"`
}

func (o ListScalingConfigsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingConfigsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingConfigsRequest", string(data)}, " ")
}
