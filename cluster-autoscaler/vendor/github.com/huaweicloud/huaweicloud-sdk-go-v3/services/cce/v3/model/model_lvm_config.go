package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type LvmConfig struct {
	// LVM写入模式：linear、striped。linear：线性模式；striped：条带模式，使用多块磁盘组成条带模式，能够提升磁盘性能。

	LvType string `json:"lvType"`
	// 磁盘挂载路径。仅在用户配置中生效。支持包含：数字、大小写字母、点、中划线、下划线的绝对路径。

	Path *string `json:"path,omitempty"`
}

func (o LvmConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "LvmConfig struct{}"
	}

	return strings.Join([]string{"LvmConfig", string(data)}, " ")
}
