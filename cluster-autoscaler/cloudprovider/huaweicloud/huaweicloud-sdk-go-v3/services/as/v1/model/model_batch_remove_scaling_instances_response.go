package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchRemoveScalingInstancesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchRemoveScalingInstancesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchRemoveScalingInstancesResponse struct{}"
	}

	return strings.Join([]string{"BatchRemoveScalingInstancesResponse", string(data)}, " ")
}
