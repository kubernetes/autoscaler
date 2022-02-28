package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListNodePoolsResponse struct {
	// API type. The value is fixed to List.

	Kind *string `json:"kind,omitempty"`
	// API version. The value is fixed to v3.

	ApiVersion *string `json:"apiVersion,omitempty"`
	// /

	Items          *[]NodePool `json:"items,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ListNodePoolsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListNodePoolsResponse struct{}"
	}

	return strings.Join([]string{"ListNodePoolsResponse", string(data)}, " ")
}
