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
type CreatePostPaidServersRequestBody struct {
	Server *PostPaidServer `json:"server"`
}

func (o CreatePostPaidServersRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreatePostPaidServersRequestBody", string(data)}, " ")
}
