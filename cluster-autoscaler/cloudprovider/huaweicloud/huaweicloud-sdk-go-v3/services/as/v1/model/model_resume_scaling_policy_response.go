package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ResumeScalingPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResumeScalingPolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"ResumeScalingPolicyResponse", string(data)}, " ")
}
