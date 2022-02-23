package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteLifecycleHookRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`
	// 生命周期挂钩标识。

	LifecycleHookName string `json:"lifecycle_hook_name"`
}

func (o DeleteLifecycleHookRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteLifecycleHookRequest struct{}"
	}

	return strings.Join([]string{"DeleteLifecycleHookRequest", string(data)}, " ")
}
