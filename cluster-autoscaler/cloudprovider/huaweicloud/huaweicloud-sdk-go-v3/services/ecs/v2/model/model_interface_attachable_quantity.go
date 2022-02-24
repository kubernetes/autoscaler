package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type InterfaceAttachableQuantity struct {
	// 云服务器剩余可挂载网卡数量

	FreeNic *int32 `json:"free_nic,omitempty"`
}

func (o InterfaceAttachableQuantity) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "InterfaceAttachableQuantity struct{}"
	}

	return strings.Join([]string{"InterfaceAttachableQuantity", string(data)}, " ")
}
