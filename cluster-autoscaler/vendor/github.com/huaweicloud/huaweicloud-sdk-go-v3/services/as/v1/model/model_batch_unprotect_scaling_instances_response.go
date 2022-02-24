package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchUnprotectScalingInstancesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchUnprotectScalingInstancesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchUnprotectScalingInstancesResponse struct{}"
	}

	return strings.Join([]string{"BatchUnprotectScalingInstancesResponse", string(data)}, " ")
}
