package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点重装场景容器运行时配置
type ReinstallRuntimeConfig struct {
	// Device mapper模式下，节点上Docker单容器的可用磁盘空间大小，OverlayFS模式(CCE Turbo集群中CentOS 7.6和Ubuntu 18.04节点，以及混合集群中Ubuntu 18.04节点)下不支持此字段。Device mapper模式下建议dockerBaseSize配置不超过80G，设置过大时可能会导致docker初始化时间过长而启动失败，若对容器磁盘大小有特殊要求，可考虑使用挂载外部或本地存储方式代替。

	DockerBaseSize *int32 `json:"dockerBaseSize,omitempty"`

	Runtime *Runtime `json:"runtime,omitempty"`
}

func (o ReinstallRuntimeConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ReinstallRuntimeConfig struct{}"
	}

	return strings.Join([]string{"ReinstallRuntimeConfig", string(data)}, " ")
}
