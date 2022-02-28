package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListAddonTemplatesResponse struct {
	// API类型，固定值“Addon”，该值不可修改。

	Kind *string `json:"kind,omitempty"`
	// API版本，固定值“v3”，该值不可修改。

	ApiVersion *string `json:"apiVersion,omitempty"`
	// 插件模板列表

	Items          *[]AddonTemplate `json:"items,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListAddonTemplatesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListAddonTemplatesResponse struct{}"
	}

	return strings.Join([]string{"ListAddonTemplatesResponse", string(data)}, " ")
}
