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

// Response Object
type BatchCreateServerTagsResponse struct {
}

func (o BatchCreateServerTagsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchCreateServerTagsResponse", string(data)}, " ")
}
