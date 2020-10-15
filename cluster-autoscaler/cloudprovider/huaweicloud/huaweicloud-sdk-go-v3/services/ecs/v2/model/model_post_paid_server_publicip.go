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
type PostPaidServerPublicip struct {
	// 为待创建云服务器分配已有弹性IP时，分配的弹性IP的ID，UUID格式。  约束：只能分配状态（status）为DOWN的弹性IP。
	Id  *string            `json:"id,omitempty"`
	Eip *PostPaidServerEip `json:"eip,omitempty"`
}

func (o PostPaidServerPublicip) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PostPaidServerPublicip", string(data)}, " ")
}
