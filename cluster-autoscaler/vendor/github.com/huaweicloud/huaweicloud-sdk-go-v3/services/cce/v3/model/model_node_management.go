package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type NodeManagement struct {
	// 云服务器组ID，若指定，节点池中所有节点将创建在该云服务器组下，节点池的云服务器组只能在创建时指定，无法修改。指定云服务器组时节点池中的节点数量不允许超出云服务器组的配额限制。

	ServerGroupReference *string `json:"serverGroupReference,omitempty"`
}

func (o NodeManagement) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeManagement struct{}"
	}

	return strings.Join([]string{"NodeManagement", string(data)}, " ")
}
