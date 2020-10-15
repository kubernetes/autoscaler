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
type ShowResetPasswordFlagRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowResetPasswordFlagRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowResetPasswordFlagRequest", string(data)}, " ")
}
