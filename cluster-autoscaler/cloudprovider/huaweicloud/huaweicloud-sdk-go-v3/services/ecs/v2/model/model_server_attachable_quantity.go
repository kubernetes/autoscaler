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

// 云服务器可挂载网卡和卷数。
type ServerAttachableQuantity struct {
	// 可挂载scsi卷数。
	FreeScsi int32 `json:"free_scsi"`
	// 可挂载vbd卷数。
	FreeBlk int32 `json:"free_blk"`
	// 可挂载卷数。
	FreeDisk int32 `json:"free_disk"`
	// 可挂载网卡数。
	FreeNic int32 `json:"free_nic"`
}

func (o ServerAttachableQuantity) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerAttachableQuantity", string(data)}, " ")
}
