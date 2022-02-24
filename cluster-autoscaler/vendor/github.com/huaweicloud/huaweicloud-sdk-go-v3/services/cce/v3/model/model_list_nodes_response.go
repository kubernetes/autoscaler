package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListNodesResponse struct {
	// API类型，固定值“List”

	Kind *string `json:"kind,omitempty"`
	// API版本，固定值“v3”

	ApiVersion *string `json:"apiVersion,omitempty"`
	// 节点对象列表，包含了当前集群下所有节点的详细信息。可通过items.metadata.name下的值来找到对应的节点。

	Items          *[]Node `json:"items,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListNodesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListNodesResponse struct{}"
	}

	return strings.Join([]string{"ListNodesResponse", string(data)}, " ")
}
