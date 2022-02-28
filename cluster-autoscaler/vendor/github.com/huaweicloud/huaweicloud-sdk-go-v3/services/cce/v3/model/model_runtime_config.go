package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type RuntimeConfig struct {
	// LVM写入模式：linear、striped。linear：线性模式；striped：条带模式，使用多块磁盘组成条带模式，能够提升磁盘性能。

	LvType string `json:"lvType"`
}

func (o RuntimeConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RuntimeConfig struct{}"
	}

	return strings.Join([]string{"RuntimeConfig", string(data)}, " ")
}
