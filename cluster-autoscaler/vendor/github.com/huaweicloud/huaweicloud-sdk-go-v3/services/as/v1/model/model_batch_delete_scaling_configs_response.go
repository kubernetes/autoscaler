package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchDeleteScalingConfigsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteScalingConfigsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingConfigsResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingConfigsResponse", string(data)}, " ")
}
