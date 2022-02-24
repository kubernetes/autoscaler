package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateAddonInstanceRequest struct {
	Body *InstanceRequest `json:"body,omitempty"`
}

func (o CreateAddonInstanceRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateAddonInstanceRequest struct{}"
	}

	return strings.Join([]string{"CreateAddonInstanceRequest", string(data)}, " ")
}
