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

// 磁盘组信息，系统盘必选，数据盘可选。
type Disk struct {
	// 磁盘大小，容量单位为GB。系统盘输入大小范围为40~32768，且不小于镜像中系统盘的最小(min_disk属性)值。数据盘输入大小范围为10~32768。
	Size int32 `json:"size"`
	// 云服务器数据盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。磁盘类型枚举值：SATA：普通IO磁盘类型。SAS：高IO磁盘类型。SSD：超高IO磁盘类型。co-pl：高IO (性能优化Ⅰ型)磁盘类型。uh-l1：超高 IO (时延优化)磁盘类型。说明：对于HANA云服务器和HL1型云服务器，需使用co-p1和uh-l1两种磁盘类型。对于其他类型的云服务器，不能使用co-p1和uh-l1两种磁盘类型。
	VolumeType DiskVolumeType `json:"volume_type"`
	// 系统盘还是数据盘，DATA表示为数据盘，SYS表示为系统盘。
	DiskType DiskDiskType `json:"disk_type"`
	// 云服务器的磁盘可指定创建在用户的专属存储中，需要指定专属存储ID。说明：同一个伸缩配置中的磁盘需统一指定或统一不指定专属存储，不支持混用；当指定专属存储时，所有专属存储需要属于同一个可用分区，且每个磁盘选择的专属存储支持的磁盘类型都需要和参数volume_type保持一致。
	DedicateStorageId *string `json:"dedicate_storage_id,omitempty"`
	// 云服务器的数据盘可指定从数据盘镜像导出，需要指定数据盘镜像ID。
	DataDiskImageId *string `json:"data_disk_image_id,omitempty"`
	// 当选择使用整机镜像时，云服务器的系统盘及数据盘将通过整机备份恢复，需要指定磁盘备份的快照ID。说明：磁盘备份的快照ID可通过镜像的整机备份ID在CSBS查询备份详情获得；一个伸缩配置中的每一个disk需要通过snapshot_id和整机备份中的磁盘备份一一对应。
	SnapshotId *string `json:"snapshot_id,omitempty"`
}

func (o Disk) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Disk", string(data)}, " ")
}

type DiskVolumeType struct {
	value string
}

type DiskVolumeTypeEnum struct {
	SATA  DiskVolumeType
	SAS   DiskVolumeType
	SSD   DiskVolumeType
	CO_PL DiskVolumeType
	UH_11 DiskVolumeType
}

func GetDiskVolumeTypeEnum() DiskVolumeTypeEnum {
	return DiskVolumeTypeEnum{
		SATA: DiskVolumeType{
			value: "SATA",
		},
		SAS: DiskVolumeType{
			value: "SAS",
		},
		SSD: DiskVolumeType{
			value: "SSD",
		},
		CO_PL: DiskVolumeType{
			value: "co-pl",
		},
		UH_11: DiskVolumeType{
			value: "uh-11",
		},
	}
}

func (c DiskVolumeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DiskVolumeType) UnmarshalJSON(b []byte) error {
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

type DiskDiskType struct {
	value string
}

type DiskDiskTypeEnum struct {
	SYS  DiskDiskType
	DATA DiskDiskType
}

func GetDiskDiskTypeEnum() DiskDiskTypeEnum {
	return DiskDiskTypeEnum{
		SYS: DiskDiskType{
			value: "SYS",
		},
		DATA: DiskDiskType{
			value: "DATA",
		},
	}
}

func (c DiskDiskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DiskDiskType) UnmarshalJSON(b []byte) error {
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
