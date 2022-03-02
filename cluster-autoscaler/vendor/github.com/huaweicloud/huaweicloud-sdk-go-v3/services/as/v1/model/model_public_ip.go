package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 配置云服务器的弹性IP信息，弹性IP有两种配置方式。详情请参考表 public_ip字段数据结构说明。  不使用（无该字段） 自动分配，需要指定新创建弹性IP的信息 说明： 当用户开通了细粒度策略，并且要将配置了弹性IP的伸缩配置关联到某个伸缩组时，这个用户被授予的细粒度策略中必须包含允许“vpc:publicIps:create”的授权项。
type PublicIp struct {
	Eip *EipInfo `json:"eip"`
}

func (o PublicIp) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PublicIp struct{}"
	}

	return strings.Join([]string{"PublicIp", string(data)}, " ")
}
