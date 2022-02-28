package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowScalingGroupRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
}

func (o ShowScalingGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"ShowScalingGroupRequest", string(data)}, " ")
}
