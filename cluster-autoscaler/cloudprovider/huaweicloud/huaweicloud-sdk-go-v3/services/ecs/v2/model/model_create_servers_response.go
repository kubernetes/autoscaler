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

// Response Object
type CreateServersResponse struct {
	// 提交任务成功后返回的任务ID，用户可以使用该ID对任务执行情况进行查询。
	JobId *string `json:"job_id,omitempty"`
	// 订单号，创建包年包月的弹性云服务器时返回该参数。
	OrderId *string `json:"order_id,omitempty"`
	// 云服务器ID列表。  通过云服务器ID查询云服务器详情 ，若返回404 可能云服务器还在创建或者已经创建失败。
	ServerIds *[]string `json:"serverIds,omitempty"`
}

func (o CreateServersResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateServersResponse", string(data)}, " ")
}
