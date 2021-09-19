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
type ListServersDetailsRequest struct {
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	Flavor              *string `json:"flavor,omitempty"`
	Ip                  *string `json:"ip,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
	Name                *string `json:"name,omitempty"`
	NotTags             *string `json:"not-tags,omitempty"`
	Offset              *int32  `json:"offset,omitempty"`
	ReservationId       *string `json:"reservation_id,omitempty"`
	Status              *string `json:"status,omitempty"`
	Tags                *string `json:"tags,omitempty"`
}

func (o ListServersDetailsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListServersDetailsRequest", string(data)}, " ")
}
