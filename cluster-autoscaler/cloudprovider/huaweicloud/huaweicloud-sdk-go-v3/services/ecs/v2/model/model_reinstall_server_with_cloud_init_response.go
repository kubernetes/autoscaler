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
type ReinstallServerWithCloudInitResponse struct {
	// 提交任务成功后返回的任务ID，用户可以使用该ID对任务执行情况进行查询。
	JobId *string `json:"job_id,omitempty"`
}

func (o ReinstallServerWithCloudInitResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ReinstallServerWithCloudInitResponse", string(data)}, " ")
}
