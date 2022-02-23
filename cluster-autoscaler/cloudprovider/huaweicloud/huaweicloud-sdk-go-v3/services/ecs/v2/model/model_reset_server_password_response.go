package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ResetServerPasswordResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetServerPasswordResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResetServerPasswordResponse struct{}"
	}

	return strings.Join([]string{"ResetServerPasswordResponse", string(data)}, " ")
}
