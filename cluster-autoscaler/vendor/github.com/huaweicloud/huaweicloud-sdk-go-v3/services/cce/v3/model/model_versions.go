package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 具体插件版本信息
type Versions struct {
	// 插件版本号

	Version string `json:"version"`
	// 插件安装参数

	Input *interface{} `json:"input"`
	// 是否为稳定版本

	Stable bool `json:"stable"`
	// 供界面使用的翻译信息

	Translate *interface{} `json:"translate"`
	// 支持集群版本号

	SupportVersions []SupportVersions `json:"supportVersions"`
	// 创建时间

	CreationTimestamp *string `json:"creationTimestamp,omitempty"`
	// 更新时间

	UpdateTimestamp string `json:"updateTimestamp"`
}

func (o Versions) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Versions struct{}"
	}

	return strings.Join([]string{"Versions", string(data)}, " ")
}
