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
type NovaDisassociateSecurityGroupResponse struct {
}

func (o NovaDisassociateSecurityGroupResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaDisassociateSecurityGroupResponse", string(data)}, " ")
}
