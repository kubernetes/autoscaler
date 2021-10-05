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

// 资源标签键
type TagsSingleValue struct {
	// 资源标签键。最大长度36个Unicode字符，不能为空，不能包含非打印字符ASCII(0-31)，“=”,“*”,“<”,“>”,“\\”,“,”,“|”,“/”。同一资源的key值不能重复。action为delete时，不校验标签字符集，最大长度127个Unicode字符。
	Key *string `json:"key,omitempty"`
	// 资源标签值。每个值最大长度43个Unicode字符，可以为空字符串，不能包含非打印字符ASCII(0-31), “=”,“*”,“<”,“>”,“\\”,“,”,“|”,“/”。action为delete时，不校验标签字符集，每个值最大长度255个Unicode字符。如果value有值按照key/value删除，如果value没值，则按照key删除。
	Value *string `json:"value,omitempty"`
}

func (o TagsSingleValue) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"TagsSingleValue", string(data)}, " ")
}
