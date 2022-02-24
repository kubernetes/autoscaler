package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ResumeScalingPolicyRequest struct {
	// 伸缩策略ID。

	ScalingPolicyId string `json:"scaling_policy_id"`

	Body *ResumeScalingPolicyOption `json:"body,omitempty"`
}

func (o ResumeScalingPolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingPolicyRequest struct{}"
	}

	return strings.Join([]string{"ResumeScalingPolicyRequest", string(data)}, " ")
}
