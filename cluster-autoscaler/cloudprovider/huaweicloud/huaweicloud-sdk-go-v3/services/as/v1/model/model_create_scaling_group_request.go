package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateScalingGroupRequest struct {
	Body *CreateScalingGroupOption `json:"body,omitempty"`
}

func (o CreateScalingGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateScalingGroupRequest", string(data)}, " ")
}
