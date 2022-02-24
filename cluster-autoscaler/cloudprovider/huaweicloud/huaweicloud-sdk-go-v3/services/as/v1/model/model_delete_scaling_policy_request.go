package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteScalingPolicyRequest struct {
	// 伸缩策略ID。

	ScalingPolicyId string `json:"scaling_policy_id"`
}

func (o DeleteScalingPolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingPolicyRequest struct{}"
	}

	return strings.Join([]string{"DeleteScalingPolicyRequest", string(data)}, " ")
}
