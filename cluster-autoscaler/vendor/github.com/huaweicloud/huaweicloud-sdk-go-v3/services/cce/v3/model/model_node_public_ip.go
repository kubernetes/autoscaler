package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type NodePublicIp struct {
	// 已有的弹性IP的ID列表。数量不得大于待创建节点数 > 若已配置ids参数，则无需配置count和eip参数

	Ids *[]string `json:"ids,omitempty"`
	// 要动态创建的弹性IP个数。 > count参数与eip参数必须同时配置。

	Count *int32 `json:"count,omitempty"`

	Eip *NodeEipSpec `json:"eip,omitempty"`
}

func (o NodePublicIp) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodePublicIp struct{}"
	}

	return strings.Join([]string{"NodePublicIp", string(data)}, " ")
}
