package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 纳管节点参数。满足条件的已有服务器，支持通过纳管节点方式安装并接入集群，重置过程将清理节点上系统盘、数据盘数据，并作为新节点接入Kuberntes集群，请提前备份迁移关键数据。其中节点池内节点重置时不支持外部指定配置，将以节点池配置进行校验并重装，以保证同节点池节点一致性。
type ResetNodeList struct {
	// API版本，固定值“v3”。

	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“List”。

	Kind string `json:"kind"`
	// 重置节点列表

	NodeList []ResetNode `json:"nodeList"`
}

func (o ResetNodeList) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResetNodeList struct{}"
	}

	return strings.Join([]string{"ResetNodeList", string(data)}, " ")
}
