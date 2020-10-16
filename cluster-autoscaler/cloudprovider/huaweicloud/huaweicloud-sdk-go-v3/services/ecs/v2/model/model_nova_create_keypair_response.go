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
type NovaCreateKeypairResponse struct {
	Keypair *NovaKeypair `json:"keypair,omitempty"`
}

func (o NovaCreateKeypairResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaCreateKeypairResponse", string(data)}, " ")
}
