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
type NovaDisassociateSecurityGroupRequest struct {
	ServerId string                                    `json:"server_id"`
	Body     *NovaDisassociateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o NovaDisassociateSecurityGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaDisassociateSecurityGroupRequest", string(data)}, " ")
}
