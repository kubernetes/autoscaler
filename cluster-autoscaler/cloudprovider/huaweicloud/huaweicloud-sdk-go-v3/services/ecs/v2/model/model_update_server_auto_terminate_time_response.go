package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type UpdateServerAutoTerminateTimeResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateServerAutoTerminateTimeResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateServerAutoTerminateTimeResponse struct{}"
	}

	return strings.Join([]string{"UpdateServerAutoTerminateTimeResponse", string(data)}, " ")
}
