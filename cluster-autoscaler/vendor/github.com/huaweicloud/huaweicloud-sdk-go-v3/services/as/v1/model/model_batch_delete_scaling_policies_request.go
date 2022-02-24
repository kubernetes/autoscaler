package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchDeleteScalingPoliciesRequest struct {
	Body *BatchDeleteScalingPoliciesOption `json:"body,omitempty"`
}

func (o BatchDeleteScalingPoliciesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingPoliciesRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingPoliciesRequest", string(data)}, " ")
}
