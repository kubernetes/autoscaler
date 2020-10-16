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
type PrePaidServerIpv6Bandwidth struct {
	// 绑定的共享带宽ID。
	Id *string `json:"id,omitempty"`
}

func (o PrePaidServerIpv6Bandwidth) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PrePaidServerIpv6Bandwidth", string(data)}, " ")
}
