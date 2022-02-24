package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ResumeScalingGroupRequest struct {
	// 伸缩组ID

	ScalingGroupId string `json:"scaling_group_id"`

	Body *ResumeScalingGroupOption `json:"body,omitempty"`
}

func (o ResumeScalingGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"ResumeScalingGroupRequest", string(data)}, " ")
}
