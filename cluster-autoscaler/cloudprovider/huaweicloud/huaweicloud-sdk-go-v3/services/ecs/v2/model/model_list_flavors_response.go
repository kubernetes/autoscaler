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
type ListFlavorsResponse struct {
	// 云服务器规格列表。
	Flavors *[]Flavor `json:"flavors,omitempty"`
}

func (o ListFlavorsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListFlavorsResponse", string(data)}, " ")
}
