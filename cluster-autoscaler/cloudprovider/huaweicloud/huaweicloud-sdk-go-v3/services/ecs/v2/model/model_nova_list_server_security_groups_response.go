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

// Response Object
type NovaListServerSecurityGroupsResponse struct {
	// security_group列表
	SecurityGroups *[]NovaSecurityGroup `json:"security_groups,omitempty"`
}

func (o NovaListServerSecurityGroupsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListServerSecurityGroupsResponse", string(data)}, " ")
}
