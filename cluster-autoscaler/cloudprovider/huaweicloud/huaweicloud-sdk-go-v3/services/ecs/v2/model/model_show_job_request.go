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
type ShowJobRequest struct {
	JobId string `json:"job_id"`
}

func (o ShowJobRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowJobRequest", string(data)}, " ")
}
