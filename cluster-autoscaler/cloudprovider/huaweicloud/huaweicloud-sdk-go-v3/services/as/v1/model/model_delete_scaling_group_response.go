package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteScalingGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingGroupResponse", string(data)}, " ")
}
