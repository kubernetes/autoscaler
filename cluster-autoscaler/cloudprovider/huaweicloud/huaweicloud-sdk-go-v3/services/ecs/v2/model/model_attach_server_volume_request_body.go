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
type AttachServerVolumeRequestBody struct {
	VolumeAttachment *AttachServerVolumeOption `json:"volumeAttachment"`
}

func (o AttachServerVolumeRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AttachServerVolumeRequestBody", string(data)}, " ")
}
