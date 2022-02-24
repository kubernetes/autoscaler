package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type NodeMetadata struct {
	// 节点名称 > 命名规则：以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-56位，且不能以中划线(-)结尾。

	Name *string `json:"name,omitempty"`
	// 节点ID，资源唯一标识，创建成功后自动生成，填写无效

	Uid *string `json:"uid,omitempty"`
	// CCE自有节点标签，非Kubernetes原生labels。  标签可用于选择对象并查找满足某些条件的对象集合，格式为key/value键值对。  示例：  ``` \"labels\": {   \"key\" : \"value\" } ```

	Labels map[string]string `json:"labels,omitempty"`
	// CCE自有节点注解，非Kubernetes原生annotations，格式为key/value键值对。   示例：  ```  \"annotations\": {   \"key1\" : \"value1\",   \"key2\" : \"value2\" }  ```   > Annotations不用于标识和选择对象。Annotations中的元数据可以是small 或large，structured 或unstructured，并且可以包括标签不允许使用的字符。

	Annotations map[string]string `json:"annotations,omitempty"`
	// 创建时间，创建成功后自动生成，填写无效

	CreationTimestamp *string `json:"creationTimestamp,omitempty"`
	// 更新时间，创建成功后自动生成，填写无效

	UpdateTimestamp *string `json:"updateTimestamp,omitempty"`
}

func (o NodeMetadata) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeMetadata struct{}"
	}

	return strings.Join([]string{"NodeMetadata", string(data)}, " ")
}
