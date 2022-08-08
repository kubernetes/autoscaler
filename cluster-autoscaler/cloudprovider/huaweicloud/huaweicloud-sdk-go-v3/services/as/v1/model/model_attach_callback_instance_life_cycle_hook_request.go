package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type AttachCallbackInstanceLifeCycleHookRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *CallbackLifeCycleHookOption `json:"body,omitempty"`
}

func (o AttachCallbackInstanceLifeCycleHookRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AttachCallbackInstanceLifeCycleHookRequest struct{}"
	}

	return strings.Join([]string{"AttachCallbackInstanceLifeCycleHookRequest", string(data)}, " ")
}
