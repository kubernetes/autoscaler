package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteScalingPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingPolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingPolicyResponse", string(data)}, " ")
}
