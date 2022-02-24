package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 插件实例详细信息-response结构体
type AddonInstance struct {
	// API类型，固定值“Addon”，该值不可修改。

	Kind string `json:"kind"`
	// API版本，固定值“v3”，该值不可修改。

	ApiVersion string `json:"apiVersion"`

	Metadata *Metadata `json:"metadata,omitempty"`

	Spec *InstanceSpec `json:"spec"`

	Status *AddonInstanceStatus `json:"status"`
}

func (o AddonInstance) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AddonInstance struct{}"
	}

	return strings.Join([]string{"AddonInstance", string(data)}, " ")
}
