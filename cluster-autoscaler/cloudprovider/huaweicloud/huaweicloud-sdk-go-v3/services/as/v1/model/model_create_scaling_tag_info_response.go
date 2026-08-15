package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateScalingTagInfoResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateScalingTagInfoResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingTagInfoResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingTagInfoResponse", string(data)}, " ")
}
