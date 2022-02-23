package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

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
