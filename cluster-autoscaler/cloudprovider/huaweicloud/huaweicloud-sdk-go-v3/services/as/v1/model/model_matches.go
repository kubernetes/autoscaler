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

type Matches struct {
	// 键。暂限定为resource_name。
	Key *string `json:"key,omitempty"`
	// 值。为固定字典值。每个值最大长度255个unicode字符。若为空字符串、resource_id时为精确匹配。
	Value *string `json:"value,omitempty"`
}

func (o Matches) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Matches", string(data)}, " ")
}
