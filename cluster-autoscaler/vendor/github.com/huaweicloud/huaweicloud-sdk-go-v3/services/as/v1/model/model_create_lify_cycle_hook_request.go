package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateLifyCycleHookRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`

	Body *CreateLifeCycleHookOption `json:"body,omitempty"`
}

func (o CreateLifyCycleHookRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateLifyCycleHookRequest struct{}"
	}

	return strings.Join([]string{"CreateLifyCycleHookRequest", string(data)}, " ")
}
