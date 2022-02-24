package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 重装拓展参数，已废弃。
type ReinstallExtendParam struct {
	// 指定待切换目标操作系统所使用的用户镜像ID，已废弃。 指定此参数等价于指定ReinstallVolumeSpec中imageID，原取值将被覆盖。

	AlphaCceNodeImageID *string `json:"alpha.cce/NodeImageID,omitempty"`
}

func (o ReinstallExtendParam) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ReinstallExtendParam struct{}"
	}

	return strings.Join([]string{"ReinstallExtendParam", string(data)}, " ")
}
