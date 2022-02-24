package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// metadata是集群对象的元数据定义，是集合类的元素类型，包含一组由不同名称定义的属性。
type PersistentVolumeClaimMetadata struct {
	// PersistentVolumeClaim名称，可以包含小写字母、数字、连字符和点，开头和结尾必须是字母或数字，最长253个字符，同一namespace下name不能重复。

	Name string `json:"name"`
	// PersistentVolumeClaim标签，key/value对格式。  - Key：必须以字母或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符；另外可以使用DNS子域作为前缀，例如example.com/my-key， DNS子域最长253个字符。  - Value：可以为空或者非空字符串，非空字符串必须以字符或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符。

	Labels *string `json:"labels,omitempty"`
}

func (o PersistentVolumeClaimMetadata) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PersistentVolumeClaimMetadata struct{}"
	}

	return strings.Join([]string{"PersistentVolumeClaimMetadata", string(data)}, " ")
}
