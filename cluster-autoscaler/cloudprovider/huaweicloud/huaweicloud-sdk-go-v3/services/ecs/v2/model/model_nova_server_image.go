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
type NovaServerImage struct {
	// 镜像ID。
	Id string `json:"id"`
	// 云服务器类型相关标记快捷链接信息。
	Links []NovaLink `json:"links"`
}

func (o NovaServerImage) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaServerImage", string(data)}, " ")
}
