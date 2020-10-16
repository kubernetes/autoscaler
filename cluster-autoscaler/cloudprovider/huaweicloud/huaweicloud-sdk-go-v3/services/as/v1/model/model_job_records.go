/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 策略执行动作包含的具体任务
type JobRecords struct {
	// 任务名称
	JobName *string `json:"job_name,omitempty"`
	// 记录类型。API：接口调用类型。MEG：消息类型。
	RecordType *JobRecordsRecordType `json:"record_type,omitempty"`
	// 记录时间。
	RecordTime *string `json:"record_time,omitempty"`
	// 请求体，仅当record_type为API时有效
	Request *string `json:"request,omitempty"`
	// 返回体，仅当record_type为API时有效
	Response *string `json:"response,omitempty"`
	// 返回码，仅当record_type为API时有效
	Code *int32 `json:"code,omitempty"`
	// 消息，仅当record_type为MEG时有效
	Message *string `json:"message,omitempty"`
	// job执行状态：SUCCESS：成功。FAIL：失败。
	JobStatus *JobRecordsJobStatus `json:"job_status,omitempty"`
}

func (o JobRecords) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"JobRecords", string(data)}, " ")
}

type JobRecordsRecordType struct {
	value string
}

type JobRecordsRecordTypeEnum struct {
	API JobRecordsRecordType
	MEG JobRecordsRecordType
}

func GetJobRecordsRecordTypeEnum() JobRecordsRecordTypeEnum {
	return JobRecordsRecordTypeEnum{
		API: JobRecordsRecordType{
			value: "API",
		},
		MEG: JobRecordsRecordType{
			value: "MEG",
		},
	}
}

func (c JobRecordsRecordType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *JobRecordsRecordType) UnmarshalJSON(b []byte) error {
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

type JobRecordsJobStatus struct {
	value string
}

type JobRecordsJobStatusEnum struct {
	SUCCESS JobRecordsJobStatus
	FAIL    JobRecordsJobStatus
}

func GetJobRecordsJobStatusEnum() JobRecordsJobStatusEnum {
	return JobRecordsJobStatusEnum{
		SUCCESS: JobRecordsJobStatus{
			value: "SUCCESS",
		},
		FAIL: JobRecordsJobStatus{
			value: "FAIL",
		},
	}
}

func (c JobRecordsJobStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *JobRecordsJobStatus) UnmarshalJSON(b []byte) error {
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
