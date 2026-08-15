package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchResumeScalingPoliciesRequest struct {
	Body *BatchResumeScalingPoliciesOption `json:"body,omitempty"`
}

func (o BatchResumeScalingPoliciesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchResumeScalingPoliciesRequest struct{}"
	}

	return strings.Join([]string{"BatchResumeScalingPoliciesRequest", string(data)}, " ")
}
