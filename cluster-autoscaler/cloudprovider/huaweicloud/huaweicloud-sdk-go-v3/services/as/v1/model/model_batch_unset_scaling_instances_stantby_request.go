package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchUnsetScalingInstancesStantbyRequest struct {
	// 实例ID。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *BatchExitStandByInstancesOption `json:"body,omitempty"`
}

func (o BatchUnsetScalingInstancesStantbyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchUnsetScalingInstancesStantbyRequest struct{}"
	}

	return strings.Join([]string{"BatchUnsetScalingInstancesStantbyRequest", string(data)}, " ")
}
