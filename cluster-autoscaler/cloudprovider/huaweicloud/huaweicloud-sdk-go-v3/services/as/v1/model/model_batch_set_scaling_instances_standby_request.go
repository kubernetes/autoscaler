package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchSetScalingInstancesStandbyRequest struct {
	// 实例ID。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *BatchEnterStandbyInstancesOption `json:"body,omitempty"`
}

func (o BatchSetScalingInstancesStandbyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchSetScalingInstancesStandbyRequest struct{}"
	}

	return strings.Join([]string{"BatchSetScalingInstancesStandbyRequest", string(data)}, " ")
}
