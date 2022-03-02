package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 云服务器元数据
type VmMetaData struct {
	// 如果需要使用密码方式登录云服务器，可使用adminPass字段指定云服务器管理员帐户初始登录密码。其中，Linux管理员帐户为root，Windows管理员帐户为Administrator。  密码复杂度要求： - 长度为8-26位。 - 密码至少必须包含大写字母、小写字母、数字和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。 - 密码不能包含用户名或用户名的逆序。 - Windows系统密码不能包含用户名或用户名的逆序，不能包含用户名中超过两个连续字符的部分。

	AdminPass *string `json:"admin_pass,omitempty"`
}

func (o VmMetaData) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "VmMetaData struct{}"
	}

	return strings.Join([]string{"VmMetaData", string(data)}, " ")
}
