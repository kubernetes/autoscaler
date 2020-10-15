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

// 用户自定义键值对
type MetaData struct {
	// 用户自定义数据总长度不大于512字节。用户自定义键不能是空串，不能包含符号.，以及不能以符号$开头。说明：Windows弹性云服务器Administrator用户的密码。示例：键（固定）：admin_pass值：cloud.1234创建密码方式鉴权的Windows弹性云服务器时为必选字段。
	CustomizeKey *string `json:"customize_key,omitempty"`
}

func (o MetaData) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"MetaData", string(data)}, " ")
}
