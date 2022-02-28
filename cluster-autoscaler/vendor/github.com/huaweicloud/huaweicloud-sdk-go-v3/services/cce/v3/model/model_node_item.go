package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type NodeItem struct {
	// 节点ID

	Uid string `json:"uid"`
}

func (o NodeItem) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeItem struct{}"
	}

	return strings.Join([]string{"NodeItem", string(data)}, " ")
}
