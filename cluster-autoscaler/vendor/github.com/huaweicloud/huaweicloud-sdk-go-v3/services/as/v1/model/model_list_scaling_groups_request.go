package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingGroupsRequest struct {
	// 伸缩组名称

	ScalingGroupName *string `json:"scaling_group_name,omitempty"`
	// 伸缩配置ID，通过查询弹性伸缩配置列表接口获取，详见[查询弹性伸缩配置列表](https://support.huaweicloud.com/api-as/as_06_0202.html)。

	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 伸缩组状态，取值如下：  - INSERVICE：正常状态 - PAUSED：停用状态 - ERROR：异常状态 - DELETING：删除中 - FREEZED：已冻结

	ScalingGroupStatus *ListScalingGroupsRequestScalingGroupStatus `json:"scaling_group_status,omitempty"`
	// 查询的起始行号，默认为0。最小值为0，最大值没有限制。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询的记录条数，默认为20。取值范围为：0~100。

	Limit *int32 `json:"limit,omitempty"`
	// 企业项目ID，当传入all_granted_eps时表示查询该用户所有授权的企业项目下的伸缩组列表，如何获取企业项目ID，请参考[查询企业项目列表](https://support.huaweicloud.com/api-em/zh-cn_topic_0121230880.html)。  说明： 华为云帐号和拥有全局权限的IAM用户可以查询该用户所有伸缩组列表。  授予部分企业项目的IAM用户，如果拥有超过100个企业项目，则只能返回有权限的前100个企业项目对应的伸缩组列表。

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
