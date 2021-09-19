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

// Response Object
type UpdateServerMetadataResponse struct {
	// 用户自定义metadata键值对。
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (o UpdateServerMetadataResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateServerMetadataResponse", string(data)}, " ")
}
