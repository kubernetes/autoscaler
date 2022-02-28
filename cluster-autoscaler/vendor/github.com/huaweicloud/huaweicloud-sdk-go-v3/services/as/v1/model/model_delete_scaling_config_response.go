package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteScalingConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingConfigResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingConfigResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingConfigResponse", string(data)}, " ")
}
