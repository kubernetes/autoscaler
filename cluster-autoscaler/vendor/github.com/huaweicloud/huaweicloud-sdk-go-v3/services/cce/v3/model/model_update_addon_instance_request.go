package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type UpdateAddonInstanceRequest struct {
	// 插件实例id

	Id string `json:"id"`

	Body *InstanceRequest `json:"body,omitempty"`
}

func (o UpdateAddonInstanceRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateAddonInstanceRequest struct{}"
	}

	return strings.Join([]string{"UpdateAddonInstanceRequest", string(data)}, " ")
}
