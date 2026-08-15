package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type BatchDeleteServerTagsRequest struct {
	// 云服务器ID。

	ServerId string `json:"server_id"`

	Body *BatchDeleteServerTagsRequestBody `json:"body,omitempty"`
}

func (o BatchDeleteServerTagsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeleteServerTagsRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteServerTagsRequest", string(data)}, " ")
}
