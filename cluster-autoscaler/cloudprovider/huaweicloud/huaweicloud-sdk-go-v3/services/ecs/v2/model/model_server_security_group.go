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

// 弹性云服务器所属安全组列表。
type ServerSecurityGroup struct {
	// 安全组名称或者UUID。
	Name string `json:"name"`
	// 安全组ID。
	Id string `json:"id"`
}

func (o ServerSecurityGroup) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerSecurityGroup", string(data)}, " ")
}
