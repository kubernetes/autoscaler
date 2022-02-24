package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type StorageSelectors struct {
	// selector的名字，作为storageGroup中selectorNames的索引，因此各个selector间的名字不能重复。

	Name string `json:"name"`
	// 存储类型，当前仅支持evs（云硬盘）或local（本地盘）；local存储类型不支持磁盘选择，所有本地盘将被组成一个VG，因此也仅允许只有一个local类型的storageSelector。

	StorageType string `json:"storageType"`

	MatchLabels *StorageSelectorsMatchLabels `json:"matchLabels,omitempty"`
}

func (o StorageSelectors) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "StorageSelectors struct{}"
	}

	return strings.Join([]string{"StorageSelectors", string(data)}, " ")
}
