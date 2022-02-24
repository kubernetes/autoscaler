package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 云硬盘加密信息，仅在创建节点系统盘或数据盘需加密时须填写。
type VolumeMetadata struct {
	// 表示云硬盘加密功能的字段，'0'代表不加密，'1'代表加密。  该字段不存在时，云硬盘默认为不加密。

	SystemEncrypted *string `json:"__system__encrypted,omitempty"`
	// 用户主密钥ID，是metadata中的表示加密功能的字段，与__system__encrypted配合使用。

	SystemCmkid *string `json:"__system__cmkid,omitempty"`
}

func (o VolumeMetadata) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "VolumeMetadata struct{}"
	}

	return strings.Join([]string{"VolumeMetadata", string(data)}, " ")
}
