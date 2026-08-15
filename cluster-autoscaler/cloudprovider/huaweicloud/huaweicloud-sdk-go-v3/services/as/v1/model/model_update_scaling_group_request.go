package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type UpdateScalingGroupRequest struct {
	// 伸缩组ID

	ScalingGroupId string `json:"scaling_group_id"`

	Body *UpdateScalingGroupOption `json:"body,omitempty"`
}

func (o UpdateScalingGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"UpdateScalingGroupRequest", string(data)}, " ")
}
