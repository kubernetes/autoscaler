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
type DeleteServersRequest struct {
	Body *DeleteServersRequestBody `json:"body,omitempty"`
}

func (o DeleteServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteServersRequest", string(data)}, " ")
}
