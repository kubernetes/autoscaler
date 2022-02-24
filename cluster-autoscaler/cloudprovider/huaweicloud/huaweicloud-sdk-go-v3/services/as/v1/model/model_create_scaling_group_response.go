package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateScalingGroupResponse struct {
	// 伸缩组ID

	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingGroupResponse", string(data)}, " ")
}
