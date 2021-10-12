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

// 重装操作系统body体。
type ChangeServerOsWithCloudInitOption struct {
	// 云服务器管理员帐户的初始登录密码。  其中，Windows管理员帐户的用户名为Administrator。  建议密码复杂度如下：  - 长度为8-26位。 - 密码至少必须包含大写字母、小写字母、数字和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。  > 说明： >  - Windows云服务器的密码，不能包含用户名或用户名的逆序，不能包含用户名中超过两个连续字符的部分。 - 对于Linux弹性云服务器也可使用user_data字段实现密码注入，此时adminpass字段无效。 - adminpass和keyname不能同时有值。 - adminpass和keyname如果同时为空，此时，metadata中的user_data属性必须有值。 - 对于已安装Cloud-init的云服务器，使用adminpass字段切换操作系统时，系统如果提示您使用keypair方式切换操作系统，表示当前区域暂不支持使用密码方式。
	Adminpass *string `json:"adminpass,omitempty"`
	// 密钥名称。
	Keyname *string `json:"keyname,omitempty"`
	// 用户ID。 说明 如果使用秘钥方式切换操作系统，则该字段为必选字段。
	Userid *string `json:"userid,omitempty"`
	// 切换系统所使用的新镜像的ID，格式为UUID。
	Imageid  string                  `json:"imageid"`
	Metadata *ChangeSeversOsMetadata `json:"metadata,omitempty"`
	// 取值为withStopServer ，支持开机状态下切换弹性云服务器操作系统。 mode取值为withStopServer时，对开机状态的弹性云服务器执行切换操作系统操作，系统自动对云服务器先执行关机，再切换操作系统。
	Mode *string `json:"mode,omitempty"`
}

func (o ChangeServerOsWithCloudInitOption) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ChangeServerOsWithCloudInitOption", string(data)}, " ")
}
