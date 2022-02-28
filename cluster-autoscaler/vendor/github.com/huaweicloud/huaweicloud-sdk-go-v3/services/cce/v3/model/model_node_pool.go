package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type NodePool struct {
	// API类型，固定值“NodePool”。

	Kind string `json:"kind"`
	// API版本，固定值“v3”。

	ApiVersion string `json:"apiVersion"`

	Metadata *NodePoolMetadata `json:"metadata"`

	Spec *NodePoolSpec `json:"spec"`

	Status *NodePoolStatus `json:"status,omitempty"`
}

func (o NodePool) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodePool struct{}"
	}

	return strings.Join([]string{"NodePool", string(data)}, " ")
}
