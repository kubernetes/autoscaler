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

// This is a auto create Body Object
type CreateServersRequestBody struct {
	Server *PrePaidServer `json:"server"`
}

func (o CreateServersRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateServersRequestBody", string(data)}, " ")
}
