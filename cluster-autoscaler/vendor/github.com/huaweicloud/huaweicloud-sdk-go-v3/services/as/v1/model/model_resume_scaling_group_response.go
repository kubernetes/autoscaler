package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ResumeScalingGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResumeScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"ResumeScalingGroupResponse", string(data)}, " ")
}
