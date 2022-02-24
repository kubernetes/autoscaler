package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListAddonInstancesResponse struct {
	// API类型，固定值“Addon”，该值不可修改。

	Kind *string `json:"kind,omitempty"`
	// API版本，固定值“v3”，该值不可修改。

	ApiVersion *string `json:"apiVersion,omitempty"`
	// 插件实例列表

	Items          *[]AddonInstance `json:"items,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListAddonInstancesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListAddonInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListAddonInstancesResponse", string(data)}, " ")
}
