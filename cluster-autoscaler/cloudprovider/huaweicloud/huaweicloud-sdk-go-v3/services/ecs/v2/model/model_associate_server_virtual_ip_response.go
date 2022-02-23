package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type AssociateServerVirtualIpResponse struct {
	// 云服务器网卡ID。

	PortId         *string `json:"port_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AssociateServerVirtualIpResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AssociateServerVirtualIpResponse struct{}"
	}

	return strings.Join([]string{"AssociateServerVirtualIpResponse", string(data)}, " ")
}
