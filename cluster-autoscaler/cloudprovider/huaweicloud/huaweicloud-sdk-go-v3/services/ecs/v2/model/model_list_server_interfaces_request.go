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
type ListServerInterfacesRequest struct {
	ServerId string `json:"server_id"`
}

func (o ListServerInterfacesRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListServerInterfacesRequest", string(data)}, " ")
}
