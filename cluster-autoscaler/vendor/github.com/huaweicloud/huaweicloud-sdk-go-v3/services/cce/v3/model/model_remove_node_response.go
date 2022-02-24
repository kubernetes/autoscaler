package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type RemoveNodeResponse struct {
	// API版本，固定值“v3”。

	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“RemoveNodesTask”。

	Kind *string `json:"kind,omitempty"`

	Spec *RemoveNodesSpec `json:"spec,omitempty"`

	Status         *TaskStatus `json:"status,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o RemoveNodeResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RemoveNodeResponse struct{}"
	}

	return strings.Join([]string{"RemoveNodeResponse", string(data)}, " ")
}
