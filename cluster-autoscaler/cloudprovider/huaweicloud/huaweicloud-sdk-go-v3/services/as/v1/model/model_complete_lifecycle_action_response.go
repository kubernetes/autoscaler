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
type CompleteLifecycleActionResponse struct {
}

func (o CompleteLifecycleActionResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CompleteLifecycleActionResponse", string(data)}, " ")
}
