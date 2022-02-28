package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type UserPassword struct {
	// 登录帐号，默认为“root”

	Username *string `json:"username,omitempty"`
	// 登录密码，取值请参见[[创建云服务器](https://support.huaweicloud.com/api-ecs/zh-cn_topic_0020212668.html)](tag:hws)[[创建云服务器](https://support.huaweicloud.com/intl/zh-cn/api-ecs/zh-cn_topic_0020212668.html)](tag:hws_hk)中**adminPass**参数的描述。若创建节点通过用户名密码方式，即使用该字段，则响应体中该字段作屏蔽展示。创建节点时password字段需要加盐加密，具体方法请参见[[创建节点时password字段加盐加密](https://support.huaweicloud.com/bestpractice-cce/cce_bestpractice_0058.html)](tag:hws)[[创建节点时password字段加盐加密](https://support.huaweicloud.com/intl/zh-cn/bestpractice-cce/cce_bestpractice_0058.html)](tag:hws_hk)。

	Password string `json:"password"`
}

func (o UserPassword) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UserPassword struct{}"
	}

	return strings.Join([]string{"UserPassword", string(data)}, " ")
}
