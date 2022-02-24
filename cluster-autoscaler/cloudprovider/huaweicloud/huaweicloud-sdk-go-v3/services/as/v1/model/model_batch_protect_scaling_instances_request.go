package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchProtectScalingInstancesRequest struct {
	// 实例ID。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *BatchProtectInstancesOption `json:"body,omitempty"`
}

func (o BatchProtectScalingInstancesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchProtectScalingInstancesRequest struct{}"
	}

	return strings.Join([]string{"BatchProtectScalingInstancesRequest", string(data)}, " ")
}
