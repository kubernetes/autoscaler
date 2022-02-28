package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 重置节点参数。集群内已有节点，支持通过重置节点方式进行重新安装并接入集群，纳管过程将清理节点上系统盘、数据盘数据，并作为新节点接入Kuberntes集群，请提前备份迁移关键数据。
type AddNodeList struct {
	// API版本，固定值“v3”。

	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“List”。

	Kind string `json:"kind"`
	// 纳管节点列表

	NodeList []AddNode `json:"nodeList"`
}

func (o AddNodeList) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AddNodeList struct{}"
	}

	return strings.Join([]string{"AddNodeList", string(data)}, " ")
}
