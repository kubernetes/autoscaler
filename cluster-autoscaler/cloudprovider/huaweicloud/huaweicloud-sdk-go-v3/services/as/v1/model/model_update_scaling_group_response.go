package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type UpdateScalingGroupResponse struct {
	// 伸缩组ID

	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"UpdateScalingGroupResponse", string(data)}, " ")
}
