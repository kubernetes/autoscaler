package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteAddonInstanceResponse struct {
	Body           *string `json:"body,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteAddonInstanceResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteAddonInstanceResponse struct{}"
	}

	return strings.Join([]string{"DeleteAddonInstanceResponse", string(data)}, " ")
}
