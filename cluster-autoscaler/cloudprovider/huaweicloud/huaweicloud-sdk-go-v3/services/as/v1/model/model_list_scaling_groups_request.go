package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingGroupsRequest struct {
	// 伸缩组名称

	ScalingGroupName *string `json:"scaling_group_name,omitempty"`
	// 伸缩配置ID，通过查询弹性伸缩配置列表接口获取，详见查询弹性伸缩配置列表。

	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 伸缩组状态，包括INSERVICE，PAUSED，ERROR，DELETING。

	ScalingGroupStatus *ListScalingGroupsRequestScalingGroupStatus `json:"scaling_group_status,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询的记录条数，默认为20。

	Limit *int32 `json:"limit,omitempty"`
	// 企业项目ID，当传入all_granted_eps时表示查询该用户所有授权的企业项目下的伸缩组列表

	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o ListScalingGroupsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingGroupsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingGroupsRequest", string(data)}, " ")
}

type ListScalingGroupsRequestScalingGroupStatus struct {
	value string
}

type ListScalingGroupsRequestScalingGroupStatusEnum struct {
	INSERVICE ListScalingGroupsRequestScalingGroupStatus
	PAUSED    ListScalingGroupsRequestScalingGroupStatus
	ERROR     ListScalingGroupsRequestScalingGroupStatus
	DELETING  ListScalingGroupsRequestScalingGroupStatus
}

func GetListScalingGroupsRequestScalingGroupStatusEnum() ListScalingGroupsRequestScalingGroupStatusEnum {
	return ListScalingGroupsRequestScalingGroupStatusEnum{
		INSERVICE: ListScalingGroupsRequestScalingGroupStatus{
			value: "INSERVICE",
		},
		PAUSED: ListScalingGroupsRequestScalingGroupStatus{
			value: "PAUSED",
		},
		ERROR: ListScalingGroupsRequestScalingGroupStatus{
			value: "ERROR",
		},
		DELETING: ListScalingGroupsRequestScalingGroupStatus{
			value: "DELETING",
		},
	}
}

func (c ListScalingGroupsRequestScalingGroupStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingGroupsRequestScalingGroupStatus) UnmarshalJSON(b []byte) error {
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
