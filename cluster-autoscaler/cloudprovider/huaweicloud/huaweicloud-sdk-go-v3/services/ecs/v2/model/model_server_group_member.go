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

// 云服务器组添加、删除成员列表
type ServerGroupMember struct {
	// 云服务器UUID。
	InstanceUuid string `json:"instance_uuid"`
}

func (o ServerGroupMember) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerGroupMember", string(data)}, " ")
}
