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
type SubJobEntities struct {
	// 云服务器相关操作显示server_id。
	ServerId *string `json:"server_id,omitempty"`
	// 网卡相关操作显示nic_id。
	NicId *string `json:"nic_id,omitempty"`
	// 子任务执行失败的具体原因。
	ErrorcodeMessage *string `json:"errorcode_message,omitempty"`
}

func (o SubJobEntities) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"SubJobEntities", string(data)}, " ")
}
