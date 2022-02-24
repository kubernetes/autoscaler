package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ExecuteScalingPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ExecuteScalingPolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ExecuteScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"ExecuteScalingPolicyResponse", string(data)}, " ")
}
