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

// 云服务器规格。
type SimpleFlavor struct {
	// 云服务器规格的ID。
	Id string `json:"id"`
	// 规格相关快捷链接地址。
	Links []Link `json:"links"`
}

func (o SimpleFlavor) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"SimpleFlavor", string(data)}, " ")
}
