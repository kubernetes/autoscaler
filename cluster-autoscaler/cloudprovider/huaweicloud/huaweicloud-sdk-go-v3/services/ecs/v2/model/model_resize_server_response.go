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
type ResizeServerResponse struct {
	// 订单号，创建包年包月的弹性云服务器时返回该参数。
	OrderId *string `json:"order_id,omitempty"`
	// 任务ID，变更按需的弹性云服务器规格时返回该参数。
	JobId *string `json:"job_id,omitempty"`
}

func (o ResizeServerResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ResizeServerResponse", string(data)}, " ")
}
