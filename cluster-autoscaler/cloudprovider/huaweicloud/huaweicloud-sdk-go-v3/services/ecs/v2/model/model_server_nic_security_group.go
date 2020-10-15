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
type ServerNicSecurityGroup struct {
	// 安全组ID。
	Id string `json:"id"`
}

func (o ServerNicSecurityGroup) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerNicSecurityGroup", string(data)}, " ")
}
