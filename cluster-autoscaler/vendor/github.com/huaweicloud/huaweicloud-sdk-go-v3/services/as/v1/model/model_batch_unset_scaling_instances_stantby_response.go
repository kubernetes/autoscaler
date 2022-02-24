package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchUnsetScalingInstancesStantbyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchUnsetScalingInstancesStantbyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchUnsetScalingInstancesStantbyResponse struct{}"
	}

	return strings.Join([]string{"BatchUnsetScalingInstancesStantbyResponse", string(data)}, " ")
}
