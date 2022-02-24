package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 云服务器元数据
type VmMetaData struct {
	// Windows弹性云服务器Administrator用户的密码。

	AdminPass *string `json:"admin_pass,omitempty"`
}

func (o VmMetaData) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "VmMetaData struct{}"
	}

	return strings.Join([]string{"VmMetaData", string(data)}, " ")
}
