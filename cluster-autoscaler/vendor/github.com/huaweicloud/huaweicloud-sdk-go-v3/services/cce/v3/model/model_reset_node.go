package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 重置节点参数。集群内已有节点通过重置进行重新安装并接入集群。
type ResetNode struct {
	// 节点ID，获取方式请参见[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)

	NodeID string `json:"nodeID"`

	Spec *ReinstallNodeSpec `json:"spec"`
}

func (o ResetNode) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResetNode struct{}"
	}

	return strings.Join([]string{"ResetNode", string(data)}, " ")
}
