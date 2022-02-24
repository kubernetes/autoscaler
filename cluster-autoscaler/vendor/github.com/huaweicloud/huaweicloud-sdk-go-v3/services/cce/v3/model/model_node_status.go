package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

//
type NodeStatus struct {
	// 节点状态。

	Phase *NodeStatusPhase `json:"phase,omitempty"`
	// 创建或删除时的任务ID。

	JobID *string `json:"jobID,omitempty"`
	// 底层云服务器或裸金属节点ID。

	ServerId *string `json:"serverId,omitempty"`
	// 节点主网卡私有网段IP地址。

	PrivateIP *string `json:"privateIP,omitempty"`
	// 节点弹性公网IP地址。如果ECS的数据没有实时同步，可在界面上通过“同步节点信息”手动进行更新。

	PublicIP *string `json:"publicIP,omitempty"`

	DeleteStatus *DeleteStatus `json:"deleteStatus,omitempty"`
}

func (o NodeStatus) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeStatus struct{}"
	}

	return strings.Join([]string{"NodeStatus", string(data)}, " ")
}

type NodeStatusPhase struct {
	value string
}

type NodeStatusPhaseEnum struct {
	BUILD      NodeStatusPhase
	INSTALLING NodeStatusPhase
	INSTALLED  NodeStatusPhase
	SHUT_DOWN  NodeStatusPhase
	UPGRADING  NodeStatusPhase
	ACTIVE     NodeStatusPhase
	ABNORMAL   NodeStatusPhase
	DELETING   NodeStatusPhase
	ERROR      NodeStatusPhase
}

func GetNodeStatusPhaseEnum() NodeStatusPhaseEnum {
	return NodeStatusPhaseEnum{
		BUILD: NodeStatusPhase{
			value: "Build",
		},
		INSTALLING: NodeStatusPhase{
			value: "Installing",
		},
		INSTALLED: NodeStatusPhase{
			value: "Installed",
		},
		SHUT_DOWN: NodeStatusPhase{
			value: "ShutDown",
		},
		UPGRADING: NodeStatusPhase{
			value: "Upgrading",
		},
		ACTIVE: NodeStatusPhase{
			value: "Active",
		},
		ABNORMAL: NodeStatusPhase{
			value: "Abnormal",
		},
		DELETING: NodeStatusPhase{
			value: "Deleting",
		},
		ERROR: NodeStatusPhase{
			value: "Error",
		},
	}
}

func (c NodeStatusPhase) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *NodeStatusPhase) UnmarshalJSON(b []byte) error {
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
