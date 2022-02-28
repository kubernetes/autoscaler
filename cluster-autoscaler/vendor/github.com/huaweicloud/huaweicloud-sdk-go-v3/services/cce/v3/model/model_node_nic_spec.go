package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点网卡的描述信息。
type NodeNicSpec struct {
	PrimaryNic *NicSpec `json:"primaryNic,omitempty"`
	// 扩展网卡

	ExtNics *[]NicSpec `json:"extNics,omitempty"`
}

func (o NodeNicSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeNicSpec struct{}"
	}

	return strings.Join([]string{"NodeNicSpec", string(data)}, " ")
}
