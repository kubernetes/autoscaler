package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 磁盘信息
type DiskResult struct {
	// 磁盘大小，容量单位为GB。

	Size *int32 `json:"size,omitempty"`
	// 磁盘类型。

	VolumeType *DiskResultVolumeType `json:"volume_type,omitempty"`
	// 系统盘还是数据盘，DATA表示为数据盘，SYS表示为系统盘。

	DiskType *DiskResultDiskType `json:"disk_type,omitempty"`
	// 磁盘所属的专属存储ID。

	DedicatedStorageId *string `json:"dedicated_storage_id,omitempty"`
	// 导入数据盘的数据盘镜像ID。

	DataDiskImageId *string `json:"data_disk_image_id,omitempty"`
	// 磁盘备份的快照ID。

	SnapshotId *string `json:"snapshot_id,omitempty"`

	Metadata *MetaData `json:"metadata,omitempty"`
}

func (o DiskResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DiskResult struct{}"
	}

	return strings.Join([]string{"DiskResult", string(data)}, " ")
}

type DiskResultVolumeType struct {
	value string
}

type DiskResultVolumeTypeEnum struct {
	SATA  DiskResultVolumeType
	SAS   DiskResultVolumeType
	SSD   DiskResultVolumeType
	CO_PL DiskResultVolumeType
	UH_11 DiskResultVolumeType
}

func GetDiskResultVolumeTypeEnum() DiskResultVolumeTypeEnum {
	return DiskResultVolumeTypeEnum{
		SATA: DiskResultVolumeType{
			value: "SATA",
		},
		SAS: DiskResultVolumeType{
			value: "SAS",
		},
		SSD: DiskResultVolumeType{
			value: "SSD",
		},
		CO_PL: DiskResultVolumeType{
			value: "co-pl",
		},
		UH_11: DiskResultVolumeType{
			value: "uh-11",
		},
	}
}

func (c DiskResultVolumeType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DiskResultVolumeType) UnmarshalJSON(b []byte) error {
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

type DiskResultDiskType struct {
	value string
}

type DiskResultDiskTypeEnum struct {
	SYS  DiskResultDiskType
	DATA DiskResultDiskType
}

func GetDiskResultDiskTypeEnum() DiskResultDiskTypeEnum {
	return DiskResultDiskTypeEnum{
		SYS: DiskResultDiskType{
			value: "SYS",
		},
		DATA: DiskResultDiskType{
			value: "DATA",
		},
	}
}

func (c DiskResultDiskType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DiskResultDiskType) UnmarshalJSON(b []byte) error {
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
