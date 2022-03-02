package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 磁盘组信息，系统盘必选，数据盘可选。
type DiskInfo struct {
	// 磁盘大小，容量单位为GB。系统盘输入大小范围为1~1024，且不小于镜像中系统盘的最小(min_disk属性)值。数据盘输入大小范围为10~32768。

	Size int32 `json:"size"`
	// 云服务器系统盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。  SATA：普通IO磁盘类型。 SAS：高IO磁盘类型。 SSD：超高IO磁盘类型。 GPSSD：通用型SSD磁盘类型。 co-p1：高IO (性能优化Ⅰ型) uh-l1：超高IO (时延优化) 当指定的云硬盘类型在avaliability_zone内不存在时，则创建云硬盘失败。  说明： 对于HANA云服务器、HL1型云服务器、HL2型云服务器，需使用co-p1和uh-l1两种磁盘类型。对于其他类型的云服务器，不能使用co-p1和uh-l1两种磁盘类型。  了解不同磁盘类型的详细信息，请参见[磁盘类型及性能介绍](https://support.huaweicloud.com/productdesc-evs/zh-cn_topic_0044524691.html)。

	VolumeType DiskInfoVolumeType `json:"volume_type"`
	// 系统盘还是数据盘，DATA表示为数据盘，SYS表示为系统盘。 说明： 系统盘不支持加密。

	DiskType DiskInfoDiskType `json:"disk_type"`
	// 云服务器的磁盘可指定创建在用户的专属存储中，需要指定专属存储ID。说明：同一个伸缩配置中的磁盘需统一指定或统一不指定专属存储，不支持混用；当指定专属存储时，所有专属存储需要属于同一个可用分区，且每个磁盘选择的专属存储支持的磁盘类型都需要和参数volume_type保持一致。

	DedicatedStorageId *string `json:"dedicated_storage_id,omitempty"`
	// 云服务器的数据盘可指定从数据盘镜像导出，需要指定数据盘镜像ID。

	DataDiskImageId *string `json:"data_disk_image_id,omitempty"`
	// 当选择使用整机镜像时，云服务器的系统盘及数据盘将通过整机备份恢复，需要指定磁盘备份的快照ID。说明：磁盘备份的快照ID可通过镜像的整机备份ID在CSBS查询备份详情获得；一个伸缩配置中的每一个disk需要通过snapshot_id和整机备份中的磁盘备份一一对应。

	SnapshotId *string `json:"snapshot_id,omitempty"`

	Metadata *MetaData `json:"metadata,omitempty"`
}

func (o DiskInfo) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DiskInfo struct{}"
	}

	return strings.Join([]string{"DiskInfo", string(data)}, " ")
}

type DiskInfoVolumeType struct {
	value string
}

type DiskInfoVolumeTypeEnum struct {
	SATA  DiskInfoVolumeType
	SAS   DiskInfoVolumeType
	SSD   DiskInfoVolumeType
	CO_PL DiskInfoVolumeType
	UH_11 DiskInfoVolumeType
	GPSSD DiskInfoVolumeType
}

func GetDiskInfoVolumeTypeEnum() DiskInfoVolumeTypeEnum {
	return DiskInfoVolumeTypeEnum{
		SATA: DiskInfoVolumeType{
			value: "SATA",
		},
		SAS: DiskInfoVolumeType{
			value: "SAS",
		},
		SSD: DiskInfoVolumeType{
			value: "SSD",
		},
		CO_PL: DiskInfoVolumeType{
			value: "co-pl",
		},
		UH_11: DiskInfoVolumeType{
			value: "uh-11",
		},
		GPSSD: DiskInfoVolumeType{
			value: "GPSSD",
		},
	}
}

func (c DiskInfoVolumeType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DiskInfoVolumeType) UnmarshalJSON(b []byte) error {
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

type DiskInfoDiskType struct {
	value string
}

type DiskInfoDiskTypeEnum struct {
	SYS  DiskInfoDiskType
	DATA DiskInfoDiskType
}

func GetDiskInfoDiskTypeEnum() DiskInfoDiskTypeEnum {
	return DiskInfoDiskTypeEnum{
		SYS: DiskInfoDiskType{
			value: "SYS",
		},
		DATA: DiskInfoDiskType{
			value: "DATA",
		},
	}
}

func (c DiskInfoDiskType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DiskInfoDiskType) UnmarshalJSON(b []byte) error {
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
