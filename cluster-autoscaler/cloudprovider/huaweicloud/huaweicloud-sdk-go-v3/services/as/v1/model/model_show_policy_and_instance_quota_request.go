package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowPolicyAndInstanceQuotaRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
}

func (o ShowPolicyAndInstanceQuotaRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowPolicyAndInstanceQuotaRequest struct{}"
	}

	return strings.Join([]string{"ShowPolicyAndInstanceQuotaRequest", string(data)}, " ")
}
