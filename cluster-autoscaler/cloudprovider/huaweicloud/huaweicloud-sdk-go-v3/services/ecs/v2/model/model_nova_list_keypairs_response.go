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
type NovaListKeypairsResponse struct {
	// 密钥信息列表。
	Keypairs *[]NovaListKeypairsResult `json:"keypairs,omitempty"`
}

func (o NovaListKeypairsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListKeypairsResponse", string(data)}, " ")
}
