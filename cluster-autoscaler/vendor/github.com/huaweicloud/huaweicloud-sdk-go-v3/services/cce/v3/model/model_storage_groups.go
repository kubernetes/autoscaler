package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type StorageGroups struct {
	// storageGroups的名字，作为虚拟存储组的名字，因此各个group个名字不能重复。

	Name string `json:"name"`
	// k8s及runtime所属存储空间。有且仅有一个group被设置为true，不填默认false。

	CceManaged *bool `json:"cceManaged,omitempty"`
	// 对应storageSelectors中的name，一个group可选择多个selector；但一个selector只能被一个group选择。

	SelectorNames []string `json:"selectorNames"`
	// group中空间配置的详细管理。

	VirtualSpaces []VirtualSpace `json:"virtualSpaces"`
}

func (o StorageGroups) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "StorageGroups struct{}"
	}

	return strings.Join([]string{"StorageGroups", string(data)}, " ")
}
