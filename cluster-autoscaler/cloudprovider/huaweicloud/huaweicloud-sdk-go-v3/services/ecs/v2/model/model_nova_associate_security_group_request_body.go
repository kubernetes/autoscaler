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

// This is a auto create Body Object
type NovaAssociateSecurityGroupRequestBody struct {
	AddSecurityGroup *NovaAddSecurityGroupOption `json:"addSecurityGroup"`
}

func (o NovaAssociateSecurityGroupRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaAssociateSecurityGroupRequestBody", string(data)}, " ")
}
