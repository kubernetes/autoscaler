package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type UpdateScalingV2PolicyRequest struct {
	// 伸缩策略ID。

	ScalingPolicyId string `json:"scaling_policy_id"`

	Body *UpdateScalingV2PolicyOption `json:"body,omitempty"`
}

func (o UpdateScalingV2PolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateScalingV2PolicyRequest struct{}"
	}

	return strings.Join([]string{"UpdateScalingV2PolicyRequest", string(data)}, " ")
}
