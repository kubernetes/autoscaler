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

// Request Object
type NovaDeleteKeypairRequest struct {
	KeypairName string `json:"keypair_name"`
}

func (o NovaDeleteKeypairRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaDeleteKeypairRequest", string(data)}, " ")
}
