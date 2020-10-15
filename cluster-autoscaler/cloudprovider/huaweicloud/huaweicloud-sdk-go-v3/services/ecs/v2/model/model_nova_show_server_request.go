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
type NovaShowServerRequest struct {
	ServerId            string  `json:"server_id"`
	OpenStackAPIVersion *string `json:"OpenStack-API-Version,omitempty"`
}

func (o NovaShowServerRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaShowServerRequest", string(data)}, " ")
}
