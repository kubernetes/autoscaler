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

// 弹性云服务器的网络属性。
type UpdateServerAddress struct {
	// IP地址版本。  - 4：代表IPv4。 - 6：代表IPv6。
	Version int32 `json:"version"`
	// IP地址。
	Addr string `json:"addr"`
}

func (o UpdateServerAddress) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateServerAddress", string(data)}, " ")
}
