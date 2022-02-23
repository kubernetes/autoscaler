package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchDeleteScalingPoliciesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteScalingPoliciesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingPoliciesResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingPoliciesResponse", string(data)}, " ")
}
