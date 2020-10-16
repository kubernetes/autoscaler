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
type NovaCreateServersRequest struct {
	OpenStackAPIVersion *string                       `json:"OpenStack-API-Version,omitempty"`
	Body                *NovaCreateServersRequestBody `json:"body,omitempty"`
}

func (o NovaCreateServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaCreateServersRequest", string(data)}, " ")
}
