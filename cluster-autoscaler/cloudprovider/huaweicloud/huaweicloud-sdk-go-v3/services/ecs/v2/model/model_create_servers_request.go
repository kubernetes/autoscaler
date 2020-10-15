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
type CreateServersRequest struct {
	Body *CreateServersRequestBody `json:"body,omitempty"`
}

func (o CreateServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateServersRequest", string(data)}, " ")
}
