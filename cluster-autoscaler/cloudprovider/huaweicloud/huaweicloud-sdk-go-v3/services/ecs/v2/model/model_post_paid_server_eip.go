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
type PostPaidServerEip struct {
	// 弹性IP地址类型。
	Iptype      string                        `json:"iptype"`
	Bandwidth   *PostPaidServerEipBandwidth   `json:"bandwidth"`
	Extendparam *PostPaidServerEipExtendParam `json:"extendparam,omitempty"`
}

func (o PostPaidServerEip) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PostPaidServerEip", string(data)}, " ")
}
