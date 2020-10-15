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

//
type NovaServerNetwork struct {
	// 网络port uuid。  没有指定网络uuid时必须指定。
	Port *string `json:"port,omitempty"`
	// 网络uuid。  没有指定网络port时必须指定。
	Uuid *string `json:"uuid,omitempty"`
	// 指定的IP地址。网络的三个参数（port、uuid和fixed_ip）中，port优先级最高；指定fixed_ip时必须指明uuid。
	FixedIp *string `json:"fixed_ip,omitempty"`
}

func (o NovaServerNetwork) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaServerNetwork", string(data)}, " ")
}
