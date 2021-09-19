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
type NovaAssociateSecurityGroupRequest struct {
	ServerId string                                 `json:"server_id"`
	Body     *NovaAssociateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o NovaAssociateSecurityGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaAssociateSecurityGroupRequest", string(data)}, " ")
}
