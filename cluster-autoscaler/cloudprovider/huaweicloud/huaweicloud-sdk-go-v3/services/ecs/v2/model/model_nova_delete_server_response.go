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
type NovaDeleteServerResponse struct {
}

func (o NovaDeleteServerResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaDeleteServerResponse", string(data)}, " ")
}
