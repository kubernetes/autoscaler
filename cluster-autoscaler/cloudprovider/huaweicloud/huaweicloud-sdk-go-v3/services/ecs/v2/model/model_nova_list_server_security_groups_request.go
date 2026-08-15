package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaListServerSecurityGroupsRequest struct {
	// 云服务器ID。

	ServerId string `json:"server_id"`
}

func (o NovaListServerSecurityGroupsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaListServerSecurityGroupsRequest struct{}"
	}

	return strings.Join([]string{"NovaListServerSecurityGroupsRequest", string(data)}, " ")
}
