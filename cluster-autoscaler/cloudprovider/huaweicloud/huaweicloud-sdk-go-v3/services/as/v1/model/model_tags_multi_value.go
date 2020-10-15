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

type TagsMultiValue struct {
	// 资源标签键。最大长度127个unicode字符。key不能为空。（搜索时不对此参数做校验）。最多为10个，不能为空或者空字符串。且不能重复。
	Key *string `json:"key,omitempty"`
	// 资源标签值列表。每个值最大长度255个unicode字符，每个key下最多为10个，同一个key中values不能重复。
	Values *[]string `json:"values,omitempty"`
}

func (o TagsMultiValue) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"TagsMultiValue", string(data)}, " ")
}
