package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type VirtualSpace struct {
	// virtualSpace的名称，当前仅支持三种类型：kubernetes、runtime、user。kubernetes：k8s空间配置，需配置lvmConfig；runtime：运行时空间配置，需配置runtimeConfig；user：用户空间配置，需配置lvmConfig

	Name string `json:"name"`
	// virtualSpace的大小，仅支持整数百分比。例如：90%。 需要注意：一个group中所有virtualSpace的百分比之和不得超过100%

	Size string `json:"size"`

	LvmConfig *LvmConfig `json:"lvmConfig,omitempty"`

	RuntimeConfig *RuntimeConfig `json:"runtimeConfig,omitempty"`
}

func (o VirtualSpace) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "VirtualSpace struct{}"
	}

	return strings.Join([]string{"VirtualSpace", string(data)}, " ")
}
