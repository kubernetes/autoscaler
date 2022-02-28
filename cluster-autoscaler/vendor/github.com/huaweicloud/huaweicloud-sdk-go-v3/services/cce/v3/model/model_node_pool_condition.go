package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点池详细状态。
type NodePoolCondition struct {
	// Condition类型，当前支持类型如下 - \"Scalable\"：节点池实际的可扩容状态，如果状态为\"False\"时则不会再次触发节点池扩容行为。 - \"QuotaInsufficient\"：节点池扩容依赖的配额不足，影响节点池可扩容状态。 - \"ResourceInsufficient\"：节点池扩容依赖的资源不足，影响节点池可扩容状态。 - \"UnexpectedError\"：节点池非预期扩容失败，影响节点池可扩容状态。 - \"LockedByOrder\"：包周期节点池被订单锁定，此时Reason为待支付订单ID。 - \"Error\"：节点池错误，通常由于删除失败触发。

	Type *string `json:"type,omitempty"`
	// Condition当前状态，取值如下 - \"True\" - \"False\"

	Status *string `json:"status,omitempty"`
	// 上次状态检查时间。

	LastProbeTime *string `json:"lastProbeTime,omitempty"`
	// 上次状态变更时间。

	LastTransitTime *string `json:"lastTransitTime,omitempty"`
	// 上次状态变更原因。

	Reason *string `json:"reason,omitempty"`
	// Condition详细描述。

	Message *string `json:"message,omitempty"`
}

func (o NodePoolCondition) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodePoolCondition struct{}"
	}

	return strings.Join([]string{"NodePoolCondition", string(data)}, " ")
}
