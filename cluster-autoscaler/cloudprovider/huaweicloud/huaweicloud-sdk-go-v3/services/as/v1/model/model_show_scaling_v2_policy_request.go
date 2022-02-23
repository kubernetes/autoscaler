package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowScalingV2PolicyRequest struct {
	// 伸缩组ID。

	ScalingPolicyId string `json:"scaling_policy_id"`
}

func (o ShowScalingV2PolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingV2PolicyRequest struct{}"
	}

	return strings.Join([]string{"ShowScalingV2PolicyRequest", string(data)}, " ")
}
