package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type MigrateServerRequest struct {
	// 云服务器ID。

	ServerId string `json:"server_id"`

	Body *MigrateServerRequestBody `json:"body,omitempty"`
}

func (o MigrateServerRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "MigrateServerRequest struct{}"
	}

	return strings.Join([]string{"MigrateServerRequest", string(data)}, " ")
}
