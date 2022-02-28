package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type DeleteClusterRequest struct {
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`
	// 集群状态兼容Error参数，用于API平滑切换。 兼容场景下，errorStatus为空则屏蔽Error状态为Deleting状态。

	ErrorStatus *string `json:"errorStatus,omitempty"`
	// 是否删除SFS Turbo（极速文件存储卷）， 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程，默认选项)

	DeleteEfs *DeleteClusterRequestDeleteEfs `json:"delete_efs,omitempty"`
	// 是否删除eni ports（原生弹性网卡）， 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程，默认选项) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程)

	DeleteEni *DeleteClusterRequestDeleteEni `json:"delete_eni,omitempty"`
	// 是否删除evs（云硬盘）， 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程，默认选项)

	DeleteEvs *DeleteClusterRequestDeleteEvs `json:"delete_evs,omitempty"`
	// 是否删除elb（弹性负载均衡）等集群Service/Ingress相关资源。 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程，默认选项) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程)

	DeleteNet *DeleteClusterRequestDeleteNet `json:"delete_net,omitempty"`
	// 是否删除obs（对象存储卷）， 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程，默认选项)

	DeleteObs *DeleteClusterRequestDeleteObs `json:"delete_obs,omitempty"`
	// 是否删除sfs（文件存储卷）， 枚举取值： - true或block (执行删除流程，失败则阻塞后续流程) - try (执行删除流程，失败则忽略，并继续执行后续流程) - false或skip (跳过删除流程，默认选项)

	DeleteSfs *DeleteClusterRequestDeleteSfs `json:"delete_sfs,omitempty"`
	// 是否使用包周期集群删除参数预置模式（仅对包周期集群生效）。 需要和其他删除选项参数一起使用，未指定的参数，则使用默认值。 使用该参数，集群不执行真正的删除，仅将本次请求的全部query参数都预置到集群数据库中，用于包周期集群退订时识别用户要删除的资源。 允许重复执行，覆盖预置的删除参数。 枚举取值： - true  (预置模式，仅预置query参数，不执行删除)

	Tobedeleted *DeleteClusterRequestTobedeleted `json:"tobedeleted,omitempty"`
}

func (o DeleteClusterRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteClusterRequest struct{}"
	}

	return strings.Join([]string{"DeleteClusterRequest", string(data)}, " ")
}

type DeleteClusterRequestDeleteEfs struct {
	value string
}

type DeleteClusterRequestDeleteEfsEnum struct {
	TRUE  DeleteClusterRequestDeleteEfs
	BLOCK DeleteClusterRequestDeleteEfs
	TRY   DeleteClusterRequestDeleteEfs
	FALSE DeleteClusterRequestDeleteEfs
	SKIP  DeleteClusterRequestDeleteEfs
}

func GetDeleteClusterRequestDeleteEfsEnum() DeleteClusterRequestDeleteEfsEnum {
	return DeleteClusterRequestDeleteEfsEnum{
		TRUE: DeleteClusterRequestDeleteEfs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEfs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEfs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEfs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEfs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEfs) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEfs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteEni struct {
	value string
}

type DeleteClusterRequestDeleteEniEnum struct {
	TRUE  DeleteClusterRequestDeleteEni
	BLOCK DeleteClusterRequestDeleteEni
	TRY   DeleteClusterRequestDeleteEni
	FALSE DeleteClusterRequestDeleteEni
	SKIP  DeleteClusterRequestDeleteEni
}

func GetDeleteClusterRequestDeleteEniEnum() DeleteClusterRequestDeleteEniEnum {
	return DeleteClusterRequestDeleteEniEnum{
		TRUE: DeleteClusterRequestDeleteEni{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEni{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEni{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEni{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEni{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEni) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEni) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteEvs struct {
	value string
}

type DeleteClusterRequestDeleteEvsEnum struct {
	TRUE  DeleteClusterRequestDeleteEvs
	BLOCK DeleteClusterRequestDeleteEvs
	TRY   DeleteClusterRequestDeleteEvs
	FALSE DeleteClusterRequestDeleteEvs
	SKIP  DeleteClusterRequestDeleteEvs
}

func GetDeleteClusterRequestDeleteEvsEnum() DeleteClusterRequestDeleteEvsEnum {
	return DeleteClusterRequestDeleteEvsEnum{
		TRUE: DeleteClusterRequestDeleteEvs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteEvs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteEvs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteEvs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteEvs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteEvs) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteEvs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteNet struct {
	value string
}

type DeleteClusterRequestDeleteNetEnum struct {
	TRUE  DeleteClusterRequestDeleteNet
	BLOCK DeleteClusterRequestDeleteNet
	TRY   DeleteClusterRequestDeleteNet
	FALSE DeleteClusterRequestDeleteNet
	SKIP  DeleteClusterRequestDeleteNet
}

func GetDeleteClusterRequestDeleteNetEnum() DeleteClusterRequestDeleteNetEnum {
	return DeleteClusterRequestDeleteNetEnum{
		TRUE: DeleteClusterRequestDeleteNet{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteNet{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteNet{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteNet{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteNet{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteNet) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteNet) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteObs struct {
	value string
}

type DeleteClusterRequestDeleteObsEnum struct {
	TRUE  DeleteClusterRequestDeleteObs
	BLOCK DeleteClusterRequestDeleteObs
	TRY   DeleteClusterRequestDeleteObs
	FALSE DeleteClusterRequestDeleteObs
	SKIP  DeleteClusterRequestDeleteObs
}

func GetDeleteClusterRequestDeleteObsEnum() DeleteClusterRequestDeleteObsEnum {
	return DeleteClusterRequestDeleteObsEnum{
		TRUE: DeleteClusterRequestDeleteObs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteObs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteObs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteObs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteObs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteObs) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteObs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestDeleteSfs struct {
	value string
}

type DeleteClusterRequestDeleteSfsEnum struct {
	TRUE  DeleteClusterRequestDeleteSfs
	BLOCK DeleteClusterRequestDeleteSfs
	TRY   DeleteClusterRequestDeleteSfs
	FALSE DeleteClusterRequestDeleteSfs
	SKIP  DeleteClusterRequestDeleteSfs
}

func GetDeleteClusterRequestDeleteSfsEnum() DeleteClusterRequestDeleteSfsEnum {
	return DeleteClusterRequestDeleteSfsEnum{
		TRUE: DeleteClusterRequestDeleteSfs{
			value: "true",
		},
		BLOCK: DeleteClusterRequestDeleteSfs{
			value: "block",
		},
		TRY: DeleteClusterRequestDeleteSfs{
			value: "try",
		},
		FALSE: DeleteClusterRequestDeleteSfs{
			value: "false",
		},
		SKIP: DeleteClusterRequestDeleteSfs{
			value: "skip",
		},
	}
}

func (c DeleteClusterRequestDeleteSfs) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestDeleteSfs) UnmarshalJSON(b []byte) error {
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

type DeleteClusterRequestTobedeleted struct {
	value string
}

type DeleteClusterRequestTobedeletedEnum struct {
	TRUE DeleteClusterRequestTobedeleted
}

func GetDeleteClusterRequestTobedeletedEnum() DeleteClusterRequestTobedeletedEnum {
	return DeleteClusterRequestTobedeletedEnum{
		TRUE: DeleteClusterRequestTobedeleted{
			value: "true",
		},
	}
}

func (c DeleteClusterRequestTobedeleted) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteClusterRequestTobedeleted) UnmarshalJSON(b []byte) error {
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
