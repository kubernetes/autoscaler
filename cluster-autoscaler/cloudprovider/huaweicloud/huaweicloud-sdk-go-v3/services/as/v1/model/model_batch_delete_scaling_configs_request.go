package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchDeleteScalingConfigsRequest struct {
	Body *BatchDeleteScalingConfigOption `json:"body,omitempty"`
}

func (o BatchDeleteScalingConfigsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingConfigsRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingConfigsRequest", string(data)}, " ")
}
