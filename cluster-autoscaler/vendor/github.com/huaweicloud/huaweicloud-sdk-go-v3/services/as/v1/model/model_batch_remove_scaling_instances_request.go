package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchRemoveScalingInstancesRequest struct {
	// 实例ID。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *BatchRemoveInstancesOption `json:"body,omitempty"`
}

func (o BatchRemoveScalingInstancesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchRemoveScalingInstancesRequest struct{}"
	}

	return strings.Join([]string{"BatchRemoveScalingInstancesRequest", string(data)}, " ")
}
