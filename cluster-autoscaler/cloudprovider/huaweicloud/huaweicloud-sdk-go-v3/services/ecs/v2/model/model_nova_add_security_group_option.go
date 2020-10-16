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
type NovaAddSecurityGroupOption struct {
	// 弹性云服务器添加的安全组名称，会对云服务器中配置的网卡生效。
	Name string `json:"name"`
}

func (o NovaAddSecurityGroupOption) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaAddSecurityGroupOption", string(data)}, " ")
}
