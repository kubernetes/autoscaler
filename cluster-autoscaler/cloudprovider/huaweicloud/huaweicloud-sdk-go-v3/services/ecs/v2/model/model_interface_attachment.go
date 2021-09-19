/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type InterfaceAttachment struct {
	// 网卡私网IP信息列表。
	FixedIps *[]ServerInterfaceFixedIp `json:"fixed_ips,omitempty"`
	// 网卡Mac地址信息。
	MacAddr *string `json:"mac_addr,omitempty"`
	// 网卡端口所属网络ID。
	NetId *string `json:"net_id,omitempty"`
	// 网卡端口ID。
	PortId *string `json:"port_id,omitempty"`
	// 网卡端口状态。
	PortState *string `json:"port_state,omitempty"`
}

func (o InterfaceAttachment) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"InterfaceAttachment", string(data)}, " ")
}
