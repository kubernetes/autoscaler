package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowNodePoolResponse struct {
	// API类型，固定值“NodePool”。

	Kind *string `json:"kind,omitempty"`
	// API版本，固定值“v3”。

	ApiVersion *string `json:"apiVersion,omitempty"`

	Metadata *NodePoolMetadata `json:"metadata,omitempty"`

	Spec *NodePoolSpec `json:"spec,omitempty"`

	Status         *NodePoolStatus `json:"status,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ShowNodePoolResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowNodePoolResponse struct{}"
	}

	return strings.Join([]string{"ShowNodePoolResponse", string(data)}, " ")
}
