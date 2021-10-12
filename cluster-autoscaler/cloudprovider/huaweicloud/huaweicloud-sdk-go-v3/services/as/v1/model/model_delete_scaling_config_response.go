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
type DeleteScalingConfigResponse struct {
}

func (o DeleteScalingConfigResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingConfigResponse", string(data)}, " ")
}
