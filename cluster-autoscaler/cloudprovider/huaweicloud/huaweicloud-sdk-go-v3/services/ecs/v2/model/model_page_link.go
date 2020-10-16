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
type PageLink struct {
	// 相应资源的链接。
	Href string `json:"href"`
	// 对应快捷链接。
	Rel string `json:"rel"`
}

func (o PageLink) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PageLink", string(data)}, " ")
}
