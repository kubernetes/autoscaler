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
type NovaSecurityGroupCommonIpRange struct {
	// 对端IP网段，cidr格式。
	Cidr string `json:"cidr"`
}

func (o NovaSecurityGroupCommonIpRange) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaSecurityGroupCommonIpRange", string(data)}, " ")
}
