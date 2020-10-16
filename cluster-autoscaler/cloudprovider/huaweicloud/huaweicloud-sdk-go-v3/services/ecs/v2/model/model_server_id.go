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
type ServerId struct {
	// 云服务器ID。
	Id string `json:"id"`
}

func (o ServerId) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerId", string(data)}, " ")
}
