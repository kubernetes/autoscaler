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

// > 说明： >  > 如果是从镜像创建云硬盘，则不支持传入“__system__encrypted”和“__system__cmkid”字段，创建出来的云硬盘与镜像的加密属性一致。
type PostPaidServerDataVolumeMetadata struct {
	// metadata中的表示加密功能的字段，0代表不加密，1代表加密。  该字段不存在时，云硬盘默认为不加密。
	SystemEncrypted *string `json:"__system__encrypted,omitempty"`
	// metadata中的加密cmkid字段，与__system__encrypted配合表示需要加密，cmkid长度固定为36个字节。
	SystemCmkid *string `json:"__system__cmkid,omitempty"`
}

func (o PostPaidServerDataVolumeMetadata) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PostPaidServerDataVolumeMetadata", string(data)}, " ")
}
