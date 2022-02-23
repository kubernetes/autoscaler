package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListScalingV2PoliciesRequest struct {
	// 伸缩组ID。

	ScalingResourceId string `json:"scaling_resource_id"`
	// 伸缩策略名称。

	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 策略类型。

	ScalingPolicyType *string `json:"scaling_policy_type,omitempty"`
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数，默认20，最大100。

	Limit *int32 `json:"limit,omitempty"`
}

func (o ListScalingV2PoliciesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingV2PoliciesRequest struct{}"
	}

	return strings.Join([]string{"ListScalingV2PoliciesRequest", string(data)}, " ")
}
