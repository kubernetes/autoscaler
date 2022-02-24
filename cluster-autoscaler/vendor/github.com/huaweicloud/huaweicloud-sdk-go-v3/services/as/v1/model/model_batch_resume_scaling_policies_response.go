package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchResumeScalingPoliciesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchResumeScalingPoliciesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchResumeScalingPoliciesResponse struct{}"
	}

	return strings.Join([]string{"BatchResumeScalingPoliciesResponse", string(data)}, " ")
}
