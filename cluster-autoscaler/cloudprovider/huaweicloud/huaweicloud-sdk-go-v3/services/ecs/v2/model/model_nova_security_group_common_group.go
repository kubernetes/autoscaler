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
type NovaSecurityGroupCommonGroup struct {
	// 对端安全组的名称
	Name string `json:"name"`
	// 对端安全组所属租户的租户ID
	TenantId string `json:"tenant_id"`
}

func (o NovaSecurityGroupCommonGroup) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaSecurityGroupCommonGroup", string(data)}, " ")
}
