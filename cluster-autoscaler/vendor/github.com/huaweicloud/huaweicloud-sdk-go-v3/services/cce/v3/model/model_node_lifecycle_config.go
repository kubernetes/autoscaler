package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点自定义生命周期配置
type NodeLifecycleConfig struct {
	// 安装前执行脚本 > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。

	PreInstall *string `json:"preInstall,omitempty"`
	// 安装后执行脚本 > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。

	PostInstall *string `json:"postInstall,omitempty"`
}

func (o NodeLifecycleConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeLifecycleConfig struct{}"
	}

	return strings.Join([]string{"NodeLifecycleConfig", string(data)}, " ")
}
