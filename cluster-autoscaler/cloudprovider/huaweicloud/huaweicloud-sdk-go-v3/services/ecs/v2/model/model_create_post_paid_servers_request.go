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
type CreatePostPaidServersRequest struct {
	Body *CreatePostPaidServersRequestBody `json:"body,omitempty"`
}

func (o CreatePostPaidServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreatePostPaidServersRequest", string(data)}, " ")
}
