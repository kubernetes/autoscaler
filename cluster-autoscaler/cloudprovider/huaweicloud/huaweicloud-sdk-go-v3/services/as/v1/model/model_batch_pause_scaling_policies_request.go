package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchPauseScalingPoliciesRequest struct {
	Body *BatchPauseScalingPoliciesOption `json:"body,omitempty"`
}

func (o BatchPauseScalingPoliciesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchPauseScalingPoliciesRequest struct{}"
	}

	return strings.Join([]string{"BatchPauseScalingPoliciesRequest", string(data)}, " ")
}
