package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 插件状态信息
type AddonInstanceStatus struct {
	// 插件实例状态

	Status AddonInstanceStatusStatus `json:"status"`
	// 插件安装失败原因

	Reason string `json:"Reason"`
	// 安装错误详情

	Message string `json:"message"`
	// 此插件版本，支持升级的集群版本

	TargetVersions *[]string `json:"targetVersions,omitempty"`

	CurrentVersion *Versions `json:"currentVersion"`
}

func (o AddonInstanceStatus) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AddonInstanceStatus struct{}"
	}

	return strings.Join([]string{"AddonInstanceStatus", string(data)}, " ")
}

type AddonInstanceStatusStatus struct {
	value string
}

type AddonInstanceStatusStatusEnum struct {
	RUNNING        AddonInstanceStatusStatus
	ABNORMAL       AddonInstanceStatusStatus
	INSTALLING     AddonInstanceStatusStatus
	INSTALL_FAILED AddonInstanceStatusStatus
	UPGRADING      AddonInstanceStatusStatus
	UPGRADE_FAILED AddonInstanceStatusStatus
	DELETING       AddonInstanceStatusStatus
	DELETE_SUCCESS AddonInstanceStatusStatus
	DELETE_FAILED  AddonInstanceStatusStatus
	AVAILABLE      AddonInstanceStatusStatus
	ROLLBACKING    AddonInstanceStatusStatus
}

func GetAddonInstanceStatusStatusEnum() AddonInstanceStatusStatusEnum {
	return AddonInstanceStatusStatusEnum{
		RUNNING: AddonInstanceStatusStatus{
			value: "running",
		},
		ABNORMAL: AddonInstanceStatusStatus{
			value: "abnormal",
		},
		INSTALLING: AddonInstanceStatusStatus{
			value: "installing",
		},
		INSTALL_FAILED: AddonInstanceStatusStatus{
			value: "installFailed",
		},
		UPGRADING: AddonInstanceStatusStatus{
			value: "upgrading",
		},
		UPGRADE_FAILED: AddonInstanceStatusStatus{
			value: "upgradeFailed",
		},
		DELETING: AddonInstanceStatusStatus{
			value: "deleting",
		},
		DELETE_SUCCESS: AddonInstanceStatusStatus{
			value: "deleteSuccess",
		},
		DELETE_FAILED: AddonInstanceStatusStatus{
			value: "deleteFailed",
		},
		AVAILABLE: AddonInstanceStatusStatus{
			value: "available",
		},
		ROLLBACKING: AddonInstanceStatusStatus{
			value: "rollbacking",
		},
	}
}

func (c AddonInstanceStatusStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *AddonInstanceStatusStatus) UnmarshalJSON(b []byte) error {
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
