package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点池自动伸缩相关配置
type NodePoolNodeAutoscaling struct {
	// 是否开启自动扩缩容

	Enable *bool `json:"enable,omitempty"`
	// 若开启自动扩缩容，最小能缩容的节点个数。不可大于集群规格所允许的节点上限

	MinNodeCount *int32 `json:"minNodeCount,omitempty"`
	// 若开启自动扩缩容，最大能扩容的节点个数，应大于等于 minNodeCount，且不超过集群规格对应的节点数量上限。

	MaxNodeCount *int32 `json:"maxNodeCount,omitempty"`
	// 节点保留时间，单位为分钟，扩容出来的节点在这个时间内不会被缩掉

	ScaleDownCooldownTime *int32 `json:"scaleDownCooldownTime,omitempty"`
	// 节点池权重，更高的权重在扩容时拥有更高的优先级

	Priority *int32 `json:"priority,omitempty"`
}

func (o NodePoolNodeAutoscaling) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodePoolNodeAutoscaling struct{}"
	}

	return strings.Join([]string{"NodePoolNodeAutoscaling", string(data)}, " ")
}
