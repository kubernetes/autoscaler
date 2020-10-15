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

// 规格相关快捷链接地址。
type FlavorLink struct {
	// 对应快捷链接。
	Href string `json:"href"`
	// 快捷链接标记名称。
	Rel string `json:"rel"`
	// 快捷链接类型，当前接口未使用，缺省值为null。
	Type string `json:"type"`
}

func (o FlavorLink) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"FlavorLink", string(data)}, " ")
}
