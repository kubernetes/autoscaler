package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type Contexts struct {
	// 上下文的名称。 - 若不存在publicIp（虚拟机弹性IP），则集群列表的集群数量为1，该字段值为“internal”。 - 若存在publicIp，则集群列表的集群数量大于1，所有扩展的context的name的值为“external”。

	Name *string `json:"name,omitempty"`

	Context *Context `json:"context,omitempty"`
}

func (o Contexts) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Contexts struct{}"
	}

	return strings.Join([]string{"Contexts", string(data)}, " ")
}
