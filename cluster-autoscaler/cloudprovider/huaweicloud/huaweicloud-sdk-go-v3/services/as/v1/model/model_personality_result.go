package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 个人信息
type PersonalityResult struct {
	// 注入文件路径信息。

	Path *string `json:"path,omitempty"`
	// 注入文件内容，base64格式编码。

	Content *string `json:"content,omitempty"`
}

func (o PersonalityResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PersonalityResult struct{}"
	}

	return strings.Join([]string{"PersonalityResult", string(data)}, " ")
}
