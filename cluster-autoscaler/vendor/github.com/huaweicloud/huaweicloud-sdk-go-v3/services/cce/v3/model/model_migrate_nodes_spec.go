package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type MigrateNodesSpec struct {
	// 操作系统类型，须精确到版本号。 当指定“alpha.cce/NodeImageID”参数时，“os”参数必须和用户自定义镜像的操作系统一致。

	Os string `json:"os"`

	ExtendParam *MigrateNodeExtendParam `json:"extendParam,omitempty"`

	Login *Login `json:"login"`
	// 待操作节点列表

	Nodes []NodeItem `json:"nodes"`
}

func (o MigrateNodesSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "MigrateNodesSpec struct{}"
	}

	return strings.Join([]string{"MigrateNodesSpec", string(data)}, " ")
}
