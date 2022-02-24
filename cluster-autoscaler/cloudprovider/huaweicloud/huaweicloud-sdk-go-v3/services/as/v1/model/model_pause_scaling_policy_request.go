package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type PauseScalingPolicyRequest struct {
	// 伸缩策略ID。

	ScalingPolicyId string `json:"scaling_policy_id"`

	Body *PauseScalingPolicyOption `json:"body,omitempty"`
}

func (o PauseScalingPolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PauseScalingPolicyRequest struct{}"
	}

	return strings.Join([]string{"PauseScalingPolicyRequest", string(data)}, " ")
}
