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
type DeleteScalingTagsResponse struct {
}

func (o DeleteScalingTagsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingTagsResponse", string(data)}, " ")
}
