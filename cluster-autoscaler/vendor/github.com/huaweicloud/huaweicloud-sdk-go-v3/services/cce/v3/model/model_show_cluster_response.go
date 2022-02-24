package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowClusterResponse struct {
	// API类型，固定值“Cluster”或“cluster”，该值不可修改。

	Kind *string `json:"kind,omitempty"`
	// API版本，固定值“v3”，该值不可修改。

	ApiVersion *string `json:"apiVersion,omitempty"`

	Metadata *ClusterMetadata `json:"metadata,omitempty"`

	Spec *ClusterSpec `json:"spec,omitempty"`

	Status         *ClusterStatus `json:"status,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowClusterResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowClusterResponse struct{}"
	}

	return strings.Join([]string{"ShowClusterResponse", string(data)}, " ")
}
