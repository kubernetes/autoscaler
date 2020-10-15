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

// 重装操作系统body。
type ReinstallServerWithCloudInitOption struct {
	// 云服务器管理员帐户的初始登录密码。 其中，Windows管理员帐户的用户名为Administrator。 建议密码复杂度如下：  - 长度为8-26位。 - 密码至少必须包含大写字母、小写字母、数字和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。   > 说明：  - 对于Windows弹性云服务器，密码不能包含用户名或用户名的逆序，不能包含用户名中超过两个连续字符的部分。 - 对于Linux弹性云服务器也可使用user_data字段实现密码注入，此时adminpass字段无效。 - adminpass和keyname不能同时有值。 - adminpass和keyname如果同时为空，此时，metadata中的user_data属性必须有值。
	Adminpass *string `json:"adminpass,omitempty"`
	// 密钥名称。
	Keyname *string `json:"keyname,omitempty"`
	// 用户ID。当传入keyname参数时，此参数为必选。
	Userid   *string                 `json:"userid,omitempty"`
	Metadata *ReinstallSeverMetadata `json:"metadata,omitempty"`
	// 取值为withStopServer ，支持开机状态下重装弹性云服务器。 mode取值为withStopServer时，对开机状态的弹性云服务器执行重装操作，系统自动对云服务器先执行关机，再重装操作系统。
	Mode *string `json:"mode,omitempty"`
}

func (o ReinstallServerWithCloudInitOption) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ReinstallServerWithCloudInitOption", string(data)}, " ")
}
