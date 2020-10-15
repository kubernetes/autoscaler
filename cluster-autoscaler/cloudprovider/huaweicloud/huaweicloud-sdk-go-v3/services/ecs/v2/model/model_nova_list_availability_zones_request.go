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
type NovaListAvailabilityZonesRequest struct {
}

func (o NovaListAvailabilityZonesRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListAvailabilityZonesRequest", string(data)}, " ")
}
