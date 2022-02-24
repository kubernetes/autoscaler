package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListClustersRequest struct {
	// 集群状态兼容Error参数，用于API平滑切换。 兼容场景下，errorStatus为空则屏蔽Error状态为Deleting状态。

	ErrorStatus *string `json:"errorStatus,omitempty"`
	// 查询集群详细信息。若设置为true，获取集群下节点总数(totalNodesNumber)、正常节点数(activeNodesNumber)、CPU总量(totalNodesCPU)、内存总量(totalNodesMemory)、已安装插件列表(installedAddonInstances)，已安装插件列表中包含名称(addonTemplateName)、版本号(version)、插件的状态信息(status)，放入到annotation中。

	Detail *string `json:"detail,omitempty"`
	// 集群状态，取值如下 - Available：可用，表示集群处于正常状态。 - Unavailable：不可用，表示集群异常，需手动删除或联系管理员删除。 - ScalingUp：扩容中，表示集群正处于扩容过程中。 - ScalingDown：缩容中，表示集群正处于缩容过程中。 - Creating：创建中，表示集群正处于创建过程中。 - Deleting：删除中，表示集群正处于删除过程中。 - Upgrading：升级中，表示集群正处于升级过程中。 - Resizing：规格变更中，表示集群正处于变更规格中。 - RollingBack：回滚中，表示集群正处于回滚过程中。 - RollbackFailed：回滚异常，表示集群回滚异常，需联系管理员进行回滚重试。 - Empty：集群无任何资源

	Status *ListClustersRequestStatus `json:"status,omitempty"`
	// 集群类型： - VirtualMachine：CCE集群 - ARM64：鲲鹏集群

	Type *ListClustersRequestType `json:"type,omitempty"`
	// 集群版本过滤

	Version *string `json:"version,omitempty"`
}

func (o ListClustersRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListClustersRequest struct{}"
	}

	return strings.Join([]string{"ListClustersRequest", string(data)}, " ")
}

type ListClustersRequestStatus struct {
	value string
}

type ListClustersRequestStatusEnum struct {
	AVAILABLE       ListClustersRequestStatus
	UNAVAILABLE     ListClustersRequestStatus
	SCALING_UP      ListClustersRequestStatus
	SCALING_DOWN    ListClustersRequestStatus
	CREATING        ListClustersRequestStatus
	DELETING        ListClustersRequestStatus
	UPGRADING       ListClustersRequestStatus
	RESIZING        ListClustersRequestStatus
	ROLLING_BACK    ListClustersRequestStatus
	ROLLBACK_FAILED ListClustersRequestStatus
	EMPTY           ListClustersRequestStatus
}

func GetListClustersRequestStatusEnum() ListClustersRequestStatusEnum {
	return ListClustersRequestStatusEnum{
		AVAILABLE: ListClustersRequestStatus{
			value: "Available",
		},
		UNAVAILABLE: ListClustersRequestStatus{
			value: "Unavailable",
		},
		SCALING_UP: ListClustersRequestStatus{
			value: "ScalingUp",
		},
		SCALING_DOWN: ListClustersRequestStatus{
			value: "ScalingDown",
		},
		CREATING: ListClustersRequestStatus{
			value: "Creating",
		},
		DELETING: ListClustersRequestStatus{
			value: "Deleting",
		},
		UPGRADING: ListClustersRequestStatus{
			value: "Upgrading",
		},
		RESIZING: ListClustersRequestStatus{
			value: "Resizing",
		},
		ROLLING_BACK: ListClustersRequestStatus{
			value: "RollingBack",
		},
		ROLLBACK_FAILED: ListClustersRequestStatus{
			value: "RollbackFailed",
		},
		EMPTY: ListClustersRequestStatus{
			value: "Empty",
		},
	}
}

func (c ListClustersRequestStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListClustersRequestStatus) UnmarshalJSON(b []byte) error {
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

type ListClustersRequestType struct {
	value string
}

type ListClustersRequestTypeEnum struct {
	VIRTUAL_MACHINE ListClustersRequestType
	ARM64           ListClustersRequestType
}

func GetListClustersRequestTypeEnum() ListClustersRequestTypeEnum {
	return ListClustersRequestTypeEnum{
		VIRTUAL_MACHINE: ListClustersRequestType{
			value: "VirtualMachine",
		},
		ARM64: ListClustersRequestType{
			value: "ARM64",
		},
	}
}

func (c ListClustersRequestType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListClustersRequestType) UnmarshalJSON(b []byte) error {
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
