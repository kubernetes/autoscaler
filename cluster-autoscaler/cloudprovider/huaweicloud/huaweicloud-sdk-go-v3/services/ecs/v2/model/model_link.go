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

// 相关快捷链接地址。
type Link struct {
	// 对应快捷链接。
	Href string `json:"href"`
	// 快捷链接标记名称。
	Rel string `json:"rel"`
}

func (o Link) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Link", string(data)}, " ")
}
