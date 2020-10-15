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
type NovaDeleteKeypairResponse struct {
}

func (o NovaDeleteKeypairResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaDeleteKeypairResponse", string(data)}, " ")
}
