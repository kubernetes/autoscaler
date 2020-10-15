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
type Ipv6Bandwidth struct {
	// IPv6带宽的ID。
	Id *string `json:"id,omitempty"`
}

func (o Ipv6Bandwidth) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Ipv6Bandwidth", string(data)}, " ")
}
