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
type DeleteScalingPolicyResponse struct {
}

func (o DeleteScalingPolicyResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingPolicyResponse", string(data)}, " ")
}
