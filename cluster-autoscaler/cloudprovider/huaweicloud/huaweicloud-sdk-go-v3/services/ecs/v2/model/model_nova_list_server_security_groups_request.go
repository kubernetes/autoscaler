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
type NovaListServerSecurityGroupsRequest struct {
	ServerId string `json:"server_id"`
}

func (o NovaListServerSecurityGroupsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListServerSecurityGroupsRequest", string(data)}, " ")
}
