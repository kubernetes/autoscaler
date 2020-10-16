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

//
type ReinstallSeverMetadata struct {
	// 重装云服务器过程中注入用户数据。  支持注入文本、文本文件或gzip文件。注入内容最大长度32KB。注入内容，需要进行base64格式编码。
	UserData *string `json:"user_data,omitempty"`
}

func (o ReinstallSeverMetadata) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ReinstallSeverMetadata", string(data)}, " ")
}
