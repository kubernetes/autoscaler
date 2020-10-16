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
type ListResizeFlavorsResponse struct {
	// 云服务器规格列表。
	Flavors *[]ListResizeFlavorsResult `json:"flavors,omitempty"`
}

func (o ListResizeFlavorsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListResizeFlavorsResponse", string(data)}, " ")
}
