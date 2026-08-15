package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateScalingV2PolicyResponse struct {
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	HttpStatusCode  int     `json:"-"`
}

func (o CreateScalingV2PolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingV2PolicyResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingV2PolicyResponse", string(data)}, " ")
}
