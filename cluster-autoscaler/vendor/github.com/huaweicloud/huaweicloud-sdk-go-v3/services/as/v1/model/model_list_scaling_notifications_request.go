package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListScalingNotificationsRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`
}

func (o ListScalingNotificationsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingNotificationsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingNotificationsRequest", string(data)}, " ")
}
