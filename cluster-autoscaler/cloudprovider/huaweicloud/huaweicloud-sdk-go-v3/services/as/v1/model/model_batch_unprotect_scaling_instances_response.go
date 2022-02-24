package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

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
