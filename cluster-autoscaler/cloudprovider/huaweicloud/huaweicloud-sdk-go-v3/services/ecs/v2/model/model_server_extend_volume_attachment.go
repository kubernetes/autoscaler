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

// 挂载到弹性云服务器上的磁盘。
type ServerExtendVolumeAttachment struct {
	// 磁盘ID，格式为UUID。
	Id string `json:"id"`
	// 删弹性云服务器时是否一并删除该磁盘。  - true：是 - false：否
	DeleteOnTermination string `json:"delete_on_termination"`
	// 启动标识，“0”代表启动盘，“-1“代表非启动盘。
	BootIndex *string `json:"bootIndex,omitempty"`
	// 云硬盘挂载盘符，即磁盘挂载点。
	Device string `json:"device"`
}

func (o ServerExtendVolumeAttachment) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ServerExtendVolumeAttachment", string(data)}, " ")
}
