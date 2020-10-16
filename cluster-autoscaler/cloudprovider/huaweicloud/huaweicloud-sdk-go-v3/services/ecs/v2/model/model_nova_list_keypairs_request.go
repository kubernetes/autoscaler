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
type NovaListKeypairsRequest struct {
	Limit               *int32  `json:"limit,omitempty"`
	Marker              *string `json:"marker,omitempty"`
	OpenStackAPIVersion *string `json:"OpenStack-API-Version,omitempty"`
}

func (o NovaListKeypairsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListKeypairsRequest", string(data)}, " ")
}
