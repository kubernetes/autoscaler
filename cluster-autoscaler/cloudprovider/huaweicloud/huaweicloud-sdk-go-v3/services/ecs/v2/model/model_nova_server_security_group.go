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
type NovaServerSecurityGroup struct {
	// 安全组名称或者uuid。
	Name *string `json:"name,omitempty"`
}

func (o NovaServerSecurityGroup) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaServerSecurityGroup", string(data)}, " ")
}
