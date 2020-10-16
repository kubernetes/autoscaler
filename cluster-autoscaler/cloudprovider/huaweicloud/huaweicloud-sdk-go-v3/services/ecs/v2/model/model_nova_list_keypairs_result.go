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
type NovaListKeypairsResult struct {
	Keypair *NovaSimpleKeypair `json:"keypair"`
}

func (o NovaListKeypairsResult) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListKeypairsResult", string(data)}, " ")
}
