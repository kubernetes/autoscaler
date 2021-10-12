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
type ShowServerTagsResponse struct {
	// 标签列表
	Tags *[]ServerTag `json:"tags,omitempty"`
}

func (o ShowServerTagsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerTagsResponse", string(data)}, " ")
}
