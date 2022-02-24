package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type User struct {
	// 客户端证书。

	ClientCertificateData *string `json:"client-certificate-data,omitempty"`
	// 包含来自TLS客户端密钥文件的PEM编码数据。

	ClientKeyData *string `json:"client-key-data,omitempty"`
}

func (o User) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "User struct{}"
	}

	return strings.Join([]string{"User", string(data)}, " ")
}
