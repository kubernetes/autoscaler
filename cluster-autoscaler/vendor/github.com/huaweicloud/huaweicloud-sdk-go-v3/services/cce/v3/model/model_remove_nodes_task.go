package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type RemoveNodesTask struct {
	// API版本，固定值“v3”。

	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“RemoveNodesTask”。

	Kind *string `json:"kind,omitempty"`

	Spec *RemoveNodesSpec `json:"spec"`

	Status *TaskStatus `json:"status,omitempty"`
}

func (o RemoveNodesTask) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RemoveNodesTask struct{}"
	}

	return strings.Join([]string{"RemoveNodesTask", string(data)}, " ")
}
