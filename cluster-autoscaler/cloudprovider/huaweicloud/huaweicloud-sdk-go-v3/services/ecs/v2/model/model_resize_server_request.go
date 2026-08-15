package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ResizeServerRequest struct {
	// 云服务器ID。

	ServerId string `json:"server_id"`

	Body *ResizeServerRequestBody `json:"body,omitempty"`
}

func (o ResizeServerRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResizeServerRequest struct{}"
	}

	return strings.Join([]string{"ResizeServerRequest", string(data)}, " ")
}
