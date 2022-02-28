package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// master的配置，支持指定可用区、规格和故障域。若指定故障域，则必须所有master节点都需要指定故障字段。
type MasterSpec struct {
	// 可用区

	AvailabilityZone *string `json:"availabilityZone,omitempty"`
	// 规格

	Flavor *string `json:"flavor,omitempty"`
	// 故障域。 1. 指定该字段需要当前系统已开启故障域特性，否则校验失败。 2. 仅单az场景支持且必须显式指定az。

	FaultDomain *string `json:"faultDomain,omitempty"`
}

func (o MasterSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "MasterSpec struct{}"
	}

	return strings.Join([]string{"MasterSpec", string(data)}, " ")
}
