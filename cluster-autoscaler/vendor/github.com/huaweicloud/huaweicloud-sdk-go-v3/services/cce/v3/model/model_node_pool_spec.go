package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

//
type NodePoolSpec struct {
	// 节点池类型。不填写时默认为vm。  - vm：弹性云服务器 - ElasticBMS：C6型弹性裸金属通用计算增强型云服务器，规格示例：c6.22xlarge.2.physical

	Type *NodePoolSpecType `json:"type,omitempty"`

	NodeTemplate *NodeSpec `json:"nodeTemplate"`
	// 节点池初始化节点个数。查询时为节点池目标节点数量。

	InitialNodeCount *int32 `json:"initialNodeCount,omitempty"`

	Autoscaling *NodePoolNodeAutoscaling `json:"autoscaling,omitempty"`

	NodeManagement *NodeManagement `json:"nodeManagement,omitempty"`
	// 1.21版本集群节点池支持绑定安全组，最多五个。

	PodSecurityGroups *[]SecurityId `json:"podSecurityGroups,omitempty"`
}

func (o NodePoolSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodePoolSpec struct{}"
	}

	return strings.Join([]string{"NodePoolSpec", string(data)}, " ")
}

type NodePoolSpecType struct {
	value string
}

type NodePoolSpecTypeEnum struct {
	VM          NodePoolSpecType
	ELASTIC_BMS NodePoolSpecType
}

func GetNodePoolSpecTypeEnum() NodePoolSpecTypeEnum {
	return NodePoolSpecTypeEnum{
		VM: NodePoolSpecType{
			value: "vm",
		},
		ELASTIC_BMS: NodePoolSpecType{
			value: "ElasticBMS",
		},
	}
}

func (c NodePoolSpecType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *NodePoolSpecType) UnmarshalJSON(b []byte) error {
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
