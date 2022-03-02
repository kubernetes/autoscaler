package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 创建磁盘的元数据
type MetaData struct {
	// metadata中的表示加密功能的字段，0代表不加密，1代表加密。  该字段不存在时，云硬盘默认为不加密。 说明： 系统盘不支持加密。

	SystemEncrypted *string `json:"__system__encrypted,omitempty"`
	// 用户主密钥ID，是metadata中的表示加密功能的字段，与__system__encrypted配合使用。 说明： - 系统盘不支持加密。 - 请参考[查询密钥列表](https://apiexplorer.developer.huaweicloud.com/apiexplorer/doc?product=KMS&api=ListKeys&version=v2)，通过HTTPS请求获取密钥ID。

	SystemCmkid *string `json:"__system__cmkid,omitempty"`
}

func (o MetaData) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "MetaData struct{}"
	}

	return strings.Join([]string{"MetaData", string(data)}, " ")
}
