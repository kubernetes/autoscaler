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
type NovaRemoveSecurityGroupOption struct {
	// 弹性云服务器移除的安全组名称，会对云服务器中配置的网卡生效。
	Name string `json:"name"`
}

func (o NovaRemoveSecurityGroupOption) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaRemoveSecurityGroupOption", string(data)}, " ")
}
