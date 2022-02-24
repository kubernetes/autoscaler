package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type Login struct {
	// 选择密钥对方式登录时的密钥对名称。密钥对和密码登录方式二者必选其一。

	SshKey *string `json:"sshKey,omitempty"`

	UserPassword *UserPassword `json:"userPassword,omitempty"`
}

func (o Login) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Login struct{}"
	}

	return strings.Join([]string{"Login", string(data)}, " ")
}
