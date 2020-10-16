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
type ListServersDetailsResponse struct {
	// 弹性云服务器的列表总数。
	Count *int32 `json:"count,omitempty"`
	// 弹性云服务器详情列表，具体参照-查询云服务器详情接口。查询级别不同，返回的详情不同。
	Servers *[]ServerDetail `json:"servers,omitempty"`
}

func (o ListServersDetailsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListServersDetailsResponse", string(data)}, " ")
}
