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
type ListScalingTagInfosByResourceIdResponse struct {
	// 资源标签列表。
	Tags *[]TagsSingleValue `json:"tags,omitempty"`
	// 系统资源标签列表。
	SysTags *[]TagsSingleValue `json:"sys_tags,omitempty"`
}

func (o ListScalingTagInfosByResourceIdResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingTagInfosByResourceIdResponse", string(data)}, " ")
}
