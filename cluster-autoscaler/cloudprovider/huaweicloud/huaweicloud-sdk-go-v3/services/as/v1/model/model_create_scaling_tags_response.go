/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateScalingTagsResponse struct {
}

func (o CreateScalingTagsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingTagsResponse", string(data)}, " ")
}
