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

// 弹性云服务器规格信息。
type ServerFlavor struct {
	// 弹性云服务器规格ID。
	Id string `json:"id"`
	// 弹性云服务器规格名称。
	Name string `json:"name"`
	// 该云服务器规格对应要求系统盘大小，0为不限制。此字段在本系统中无效。
	Disk string `json:"disk"`
	// 该云服务器规格对应的CPU核数。
	Vcpus string `json:"vcpus"`
	// 该云服务器规格对应的内存大小，单位为MB。
	Ram string `json:"ram"`
}

func (o ServerFlavor) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerFlavor", string(data)}, " ")
}
