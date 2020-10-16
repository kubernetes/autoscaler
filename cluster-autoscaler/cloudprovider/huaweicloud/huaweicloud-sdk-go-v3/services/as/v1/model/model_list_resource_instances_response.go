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

// Response Object
type ListResourceInstancesResponse struct {
	// 标签资源实例。
	Resources *[]Resources `json:"resources,omitempty"`
	// 总记录数。
	TotalCount *int32 `json:"total_count,omitempty"`
	// 分页位置标识。
	Marker *int32 `json:"marker,omitempty"`
}

func (o ListResourceInstancesResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListResourceInstancesResponse", string(data)}, " ")
}
