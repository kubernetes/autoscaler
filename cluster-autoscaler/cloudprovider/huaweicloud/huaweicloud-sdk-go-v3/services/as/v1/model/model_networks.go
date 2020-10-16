/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 网络信息
type Networks struct {
	// 网络ID。
	Id string `json:"id"`
	// 是否启用IPv6。取值为true时，标识此网卡已启用IPv6。
	Ipv6Enable    *bool          `json:"ipv6_enable,omitempty"`
	Ipv6Bandwidth *Ipv6Bandwidth `json:"ipv6_bandwidth,omitempty"`
}

func (o Networks) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Networks", string(data)}, " ")
}
