package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type QuotaResource struct {
	// 资源类型

	QuotaKey *string `json:"quotaKey,omitempty"`
	// 配额值

	QuotaLimit *int32 `json:"quotaLimit,omitempty"`
	// 已创建的资源个数

	Used *int32 `json:"used,omitempty"`
	// 单位

	Unit *string `json:"unit,omitempty"`
	// 局点ID。若资源不涉及此参数，则不返回该参数。

	RegionId *string `json:"regionId,omitempty"`
	// 可用区ID。若资源不涉及此参数，则不返回该参数。

	AvailabilityZoneId *string `json:"availabilityZoneId,omitempty"`
}

func (o QuotaResource) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "QuotaResource struct{}"
	}

	return strings.Join([]string{"QuotaResource", string(data)}, " ")
}
