package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type AwakeClusterResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AwakeClusterResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AwakeClusterResponse struct{}"
	}

	return strings.Join([]string{"AwakeClusterResponse", string(data)}, " ")
}
