package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type ClusterNodeInformationMetadata struct {
	// 节点名称  > 修改节点名称后，弹性云服务器名称（虚拟机名称）会同步修改。 > > 命名规则：以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-56位，且不能以中划线(-)结尾。

	Name string `json:"name"`
}

func (o ClusterNodeInformationMetadata) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ClusterNodeInformationMetadata struct{}"
	}

	return strings.Join([]string{"ClusterNodeInformationMetadata", string(data)}, " ")
}
