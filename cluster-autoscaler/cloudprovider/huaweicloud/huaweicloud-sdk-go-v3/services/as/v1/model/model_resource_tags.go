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

type ResourceTags struct {
	// 资源标签值。最大长度36个unicode字符。
	Key *string `json:"key,omitempty"`
	// 资源标签值。
	Value *string `json:"value,omitempty"`
}

func (o ResourceTags) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ResourceTags", string(data)}, " ")
}
