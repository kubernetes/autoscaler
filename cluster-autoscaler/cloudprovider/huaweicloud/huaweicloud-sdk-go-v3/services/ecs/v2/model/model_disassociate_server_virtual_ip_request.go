package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DisassociateServerVirtualIpRequest struct {
	// 云服务器网卡ID。

	NicId string `json:"nic_id"`

	Body *DisassociateServerVirtualIpRequestBody `json:"body,omitempty"`
}

func (o DisassociateServerVirtualIpRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DisassociateServerVirtualIpRequest struct{}"
	}

	return strings.Join([]string{"DisassociateServerVirtualIpRequest", string(data)}, " ")
}
