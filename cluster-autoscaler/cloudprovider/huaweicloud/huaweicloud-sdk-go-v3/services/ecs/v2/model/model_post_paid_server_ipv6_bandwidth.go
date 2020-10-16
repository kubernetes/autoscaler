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

// IPV6共享带宽。
type PostPaidServerIpv6Bandwidth struct {
	// 绑定的共享带宽ID。
	Id *string `json:"id,omitempty"`
}

func (o PostPaidServerIpv6Bandwidth) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PostPaidServerIpv6Bandwidth", string(data)}, " ")
}
