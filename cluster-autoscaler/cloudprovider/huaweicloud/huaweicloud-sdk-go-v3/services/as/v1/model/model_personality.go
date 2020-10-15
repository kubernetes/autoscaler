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

// 注入文件信息。仅支持注入文本文件，最大支持注入5个文件，每个文件最大1KB。
type Personality struct {
	// 注入文件路径信息。Linux系统请输入注入文件保存路径，例如 “/etc/foo.txt”。Windows系统注入文件自动保存在C盘根目录，只需要输入保存文件名，例如 “foo”，文件名只能包含字母（a~zA~Z）和数字（0~9）。
	Path string `json:"path"`
	// 注入文件内容。该值应指定为注入文件的内容进行base64格式编码后的信息。
	Content string `json:"content"`
}

func (o Personality) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Personality", string(data)}, " ")
}
