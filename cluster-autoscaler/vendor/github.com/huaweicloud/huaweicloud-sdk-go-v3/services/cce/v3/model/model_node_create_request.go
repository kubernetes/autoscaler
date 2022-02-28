package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type NodeCreateRequest struct {
	// API类型，固定值“Node”，该值不可修改。

	Kind string `json:"kind"`
	// API版本，固定值“v3”，该值不可修改。

	ApiVersion string `json:"apiVersion"`

	Metadata *NodeMetadata `json:"metadata,omitempty"`

	Spec *NodeSpec `json:"spec"`
}

func (o NodeCreateRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeCreateRequest struct{}"
	}

	return strings.Join([]string{"NodeCreateRequest", string(data)}, " ")
}
