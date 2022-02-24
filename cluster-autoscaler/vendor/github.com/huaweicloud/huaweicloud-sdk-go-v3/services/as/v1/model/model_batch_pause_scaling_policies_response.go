package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchPauseScalingPoliciesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchPauseScalingPoliciesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchPauseScalingPoliciesResponse struct{}"
	}

	return strings.Join([]string{"BatchPauseScalingPoliciesResponse", string(data)}, " ")
}
