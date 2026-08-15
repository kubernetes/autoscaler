package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchUpdateServersNameRequest struct {
	Body *BatchUpdateServersNameRequestBody `json:"body,omitempty"`
}

func (o BatchUpdateServersNameRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchUpdateServersNameRequest struct{}"
	}

	return strings.Join([]string{"BatchUpdateServersNameRequest", string(data)}, " ")
}
