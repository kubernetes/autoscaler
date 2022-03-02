package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

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
