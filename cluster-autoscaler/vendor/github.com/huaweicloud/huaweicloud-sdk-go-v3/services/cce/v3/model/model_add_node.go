package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 纳管节点参数。集群内已有节点支持通过重置进行重新安装并接入集群。
type AddNode struct {
	// 服务器ID，获取方式请参见ECS/BMS相关资料。

	ServerID string `json:"serverID"`

	Spec *ReinstallNodeSpec `json:"spec"`
}

func (o AddNode) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AddNode struct{}"
	}

	return strings.Join([]string{"AddNode", string(data)}, " ")
}
