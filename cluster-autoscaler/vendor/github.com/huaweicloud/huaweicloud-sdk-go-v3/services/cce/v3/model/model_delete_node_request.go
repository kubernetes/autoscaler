package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type DeleteNodeRequest struct {
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`
	// 节点ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	NodeId string `json:"node_id"`
	// 集群状态兼容Error参数，用于API平滑切换。 兼容场景下，errorStatus为空则屏蔽Error状态为Deleting状态。

	ErrorStatus *string `json:"errorStatus,omitempty"`
	// 标明是否为nodepool下发的请求。若不为“NoScaleDown”将自动更新对应节点池的实例数

	NodepoolScaleDown *DeleteNodeRequestNodepoolScaleDown `json:"nodepoolScaleDown,omitempty"`
}

func (o DeleteNodeRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteNodeRequest struct{}"
	}

	return strings.Join([]string{"DeleteNodeRequest", string(data)}, " ")
}

type DeleteNodeRequestNodepoolScaleDown struct {
	value string
}

type DeleteNodeRequestNodepoolScaleDownEnum struct {
	NO_SCALE_DOWN DeleteNodeRequestNodepoolScaleDown
}

func GetDeleteNodeRequestNodepoolScaleDownEnum() DeleteNodeRequestNodepoolScaleDownEnum {
	return DeleteNodeRequestNodepoolScaleDownEnum{
		NO_SCALE_DOWN: DeleteNodeRequestNodepoolScaleDown{
			value: "NoScaleDown",
		},
	}
}

func (c DeleteNodeRequestNodepoolScaleDown) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteNodeRequestNodepoolScaleDown) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
