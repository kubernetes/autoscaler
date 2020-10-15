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

type Resources struct {
	// 资源详情ID。
	ResourceId *string `json:"resource_id,omitempty"`
	// 资源详情。
	ResourceDetail *string `json:"resource_detail,omitempty"`
	// 标签列表，没有标签默认为空数组。
	Tags *[]ResourceTags `json:"tags,omitempty"`
	// 资源名称，没有默认为空字符串。
	ResourceName *string `json:"resource_name,omitempty"`
}

func (o Resources) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Resources", string(data)}, " ")
}
