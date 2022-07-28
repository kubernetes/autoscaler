package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateScalingNotificationRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *CreateNotificationOption `json:"body,omitempty"`
}

func (o CreateScalingNotificationRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingNotificationRequest struct{}"
	}

	return strings.Join([]string{"CreateScalingNotificationRequest", string(data)}, " ")
}
