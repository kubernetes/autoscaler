package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowScalingGroupResponse struct {
	ScalingGroup   *ScalingGroups `json:"scaling_group,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"ShowScalingGroupResponse", string(data)}, " ")
}
