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

// This is a auto create Body Object
type BatchDeleteServerNicsRequestBody struct {
	// 需要删除的网卡列表信息。  说明： 主网卡是弹性云服务器上配置了路由规则的网卡，不可删除。
	Nics []BatchDeleteServerNicOption `json:"nics"`
}

func (o BatchDeleteServerNicsRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchDeleteServerNicsRequestBody", string(data)}, " ")
}
