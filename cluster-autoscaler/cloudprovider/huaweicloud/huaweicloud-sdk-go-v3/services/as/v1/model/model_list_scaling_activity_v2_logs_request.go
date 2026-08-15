package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingActivityV2LogsRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
	// 伸缩活动日志ID

	LogId *string `json:"log_id,omitempty"`
	// 查询的起始时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	StartTime *string `json:"start_time,omitempty"`
	// 查询的截止时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	EndTime *string `json:"end_time,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数，默认20，最大100。

	Limit *int32 `json:"limit,omitempty"`
	// 查询的伸缩活动类型（查询多类型使用逗号分隔）：NORMAL：普通伸缩活动；MANNUAL_REMOVE：从伸缩组手动移除实例；MANNUAL_DELETE：从伸缩组手动移除实例并删除实例；ELB_CHECK_DELETE：ELB检查移除并删除实例；DIFF：期望实例数与实际实例 不一致；MODIFY_ELB：LB迁移

	Type *ListScalingActivityV2LogsRequestType `json:"type,omitempty"`
	// 查询的伸缩活动状态：SUCCESS：成功；FAIL：失败；DOING：伸缩中

	Status *ListScalingActivityV2LogsRequestStatus `json:"status,omitempty"`
}

func (o ListScalingActivityV2LogsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingActivityV2LogsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingActivityV2LogsRequest", string(data)}, " ")
}

type ListScalingActivityV2LogsRequestType struct {
	value string
}

type ListScalingActivityV2LogsRequestTypeEnum struct {
	NORMAL             ListScalingActivityV2LogsRequestType
	MANNUAL_REMOVE     ListScalingActivityV2LogsRequestType
	MANNUAL_DELETE     ListScalingActivityV2LogsRequestType
	MANNUAL_ADD        ListScalingActivityV2LogsRequestType
	ELB_CHECK_DELETE   ListScalingActivityV2LogsRequestType
	AUDIT_CHECK_DELETE ListScalingActivityV2LogsRequestType
	MODIFY_ELB         ListScalingActivityV2LogsRequestType
}

func GetListScalingActivityV2LogsRequestTypeEnum() ListScalingActivityV2LogsRequestTypeEnum {
	return ListScalingActivityV2LogsRequestTypeEnum{
		NORMAL: ListScalingActivityV2LogsRequestType{
			value: "NORMAL",
		},
		MANNUAL_REMOVE: ListScalingActivityV2LogsRequestType{
			value: "MANNUAL_REMOVE",
		},
		MANNUAL_DELETE: ListScalingActivityV2LogsRequestType{
			value: "MANNUAL_DELETE",
		},
		MANNUAL_ADD: ListScalingActivityV2LogsRequestType{
			value: "MANNUAL_ADD",
		},
		ELB_CHECK_DELETE: ListScalingActivityV2LogsRequestType{
			value: "ELB_CHECK_DELETE",
		},
		AUDIT_CHECK_DELETE: ListScalingActivityV2LogsRequestType{
			value: "AUDIT_CHECK_DELETE",
		},
		MODIFY_ELB: ListScalingActivityV2LogsRequestType{
			value: "MODIFY_ELB",
		},
	}
}

func (c ListScalingActivityV2LogsRequestType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingActivityV2LogsRequestType) UnmarshalJSON(b []byte) error {
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

type ListScalingActivityV2LogsRequestStatus struct {
	value string
}

type ListScalingActivityV2LogsRequestStatusEnum struct {
	SUCCESS ListScalingActivityV2LogsRequestStatus
	FAIL    ListScalingActivityV2LogsRequestStatus
	DOING   ListScalingActivityV2LogsRequestStatus
}

func GetListScalingActivityV2LogsRequestStatusEnum() ListScalingActivityV2LogsRequestStatusEnum {
	return ListScalingActivityV2LogsRequestStatusEnum{
		SUCCESS: ListScalingActivityV2LogsRequestStatus{
			value: "SUCCESS",
		},
		FAIL: ListScalingActivityV2LogsRequestStatus{
			value: "FAIL",
		},
		DOING: ListScalingActivityV2LogsRequestStatus{
			value: "DOING",
		},
	}
}

func (c ListScalingActivityV2LogsRequestStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingActivityV2LogsRequestStatus) UnmarshalJSON(b []byte) error {
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
