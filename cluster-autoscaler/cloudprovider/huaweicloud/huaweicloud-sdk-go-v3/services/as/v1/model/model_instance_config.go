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

// 实例配置信息
type InstanceConfig struct {
	// 云服务器ID，当使用已存在的云服务器的规格为模板创建弹性伸缩配置时传入该字段，此时flavorRef、imageRef、disk、security_groups、tenancy和dedicated_host_id字段不生效。当不传入instance_id字段时flavorRef、imageRef、disk字段为必选。
	InstanceId *string `json:"instance_id,omitempty"`
	// 云服务器的规格ID。最多支持选择10个规格，多个规格ID以逗号分隔。云服务器的ID通过查询弹性云服务器规格详情和扩展信息列表接口获取，详情请参考查询云服务器规格详情和扩展信息列表。
	FlavorRef *string `json:"flavorRef,omitempty"`
	// 镜像ID，同image_id，通过查询镜像服务镜像列表接口获取，详见《镜像服务API参考》的“查询镜像列表”。
	ImageRef *string `json:"imageRef,omitempty"`
	// 磁盘组信息，系统盘必选，数据盘可选。
	Disk *[]Disk `json:"disk,omitempty"`
	// 登录云服务器的SSH密钥名称，与adminPass互斥，且必选一个。Windoes弹性云服务器不支持使用密钥登陆方式。
	KeyName *string `json:"key_name,omitempty"`
	// 注入文件信息。仅支持注入文本文件，最大支持注入5个文件，每个文件最大1KB。
	Personality *[]Personality `json:"personality,omitempty"`
	PublicIp    *PublicIp      `json:"public_ip,omitempty"`
	// cloud-init用户数据。支持注入文本、文本文件或gzip文件。文件内容需要进行base64格式编码，注入内容（编码之前的内容）最大为32KB。说明：当key_name没有指定时，user_data注入的数据默认为云服务器root账号的登录密码。创建密码方式鉴权的Linux弹性云服务器时为必填项，为root用户注入自定义初始化密码。
	UserData *string   `json:"user_data,omitempty"`
	Metadata *MetaData `json:"metadata,omitempty"`
	// 安全组信息。使用vpc_id通过查询VPC服务安全组列表接口获取，详见《虚拟私有云API参考》的“查询安全组列表”。当伸缩配置和伸缩组同时指定安全组时，将以伸缩配置中的安全组为准；当伸缩配置和伸缩组都没有指定安全组时，将使用默认安全组。为了使用灵活性更高，推荐在伸缩配置中指定安全组。
	SecurityGroups *[]SecurityGroups `json:"security_groups,omitempty"`
	// 云服务器组ID。
	ServerGroupId *string `json:"server_group_id,omitempty"`
	// 在专属主机上创建弹性云服务器。参数取值为dedicated。
	Tenancy *InstanceConfigTenancy `json:"tenancy,omitempty"`
	// 专属主机的ID。 说明：该字段仅在tenancy为dedicated时生效；如果指定该字段，云服务器将被创建到指定的专属主机上；如果不指定该字段，此时系统会将云服务器创建在符合规格的专属主机中剩余内存最大的那一台上，以使各专属主机尽量均衡负载。
	DedicatedHostId *string `json:"dedicated_host_id,omitempty"`
	// 使用伸缩配置创建云主机的时候，多规格使用的优先级策略。PICK_FIRST（默认）：选择优先，虚拟机扩容时规格的选择按照flavorRef列表的顺序进行优先级排序。COST_FIRST：成本优化，虚拟机扩容时规格的选择按照价格最优原则进行优先级排序。
	MultiFlavorPriorityPolicy *InstanceConfigMultiFlavorPriorityPolicy `json:"multi_flavor_priority_policy,omitempty"`
	// 云服务器的计费模式，可以选择竞价计费或按需计费，取值如下：按需计费：不指定该字段。竞价计费：spot
	MarketType *InstanceConfigMarketType `json:"market_type,omitempty"`
}

func (o InstanceConfig) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"InstanceConfig", string(data)}, " ")
}

type InstanceConfigTenancy struct {
	value string
}

type InstanceConfigTenancyEnum struct {
	DEDICATED InstanceConfigTenancy
}

func GetInstanceConfigTenancyEnum() InstanceConfigTenancyEnum {
	return InstanceConfigTenancyEnum{
		DEDICATED: InstanceConfigTenancy{
			value: "dedicated",
		},
	}
}

func (c InstanceConfigTenancy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceConfigTenancy) UnmarshalJSON(b []byte) error {
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

type InstanceConfigMultiFlavorPriorityPolicy struct {
	value string
}

type InstanceConfigMultiFlavorPriorityPolicyEnum struct {
	PICK_FIRST InstanceConfigMultiFlavorPriorityPolicy
	COST_FIRST InstanceConfigMultiFlavorPriorityPolicy
}

func GetInstanceConfigMultiFlavorPriorityPolicyEnum() InstanceConfigMultiFlavorPriorityPolicyEnum {
	return InstanceConfigMultiFlavorPriorityPolicyEnum{
		PICK_FIRST: InstanceConfigMultiFlavorPriorityPolicy{
			value: "PICK_FIRST",
		},
		COST_FIRST: InstanceConfigMultiFlavorPriorityPolicy{
			value: "COST_FIRST",
		},
	}
}

func (c InstanceConfigMultiFlavorPriorityPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceConfigMultiFlavorPriorityPolicy) UnmarshalJSON(b []byte) error {
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

type InstanceConfigMarketType struct {
	value string
}

type InstanceConfigMarketTypeEnum struct {
	SPOT InstanceConfigMarketType
}

func GetInstanceConfigMarketTypeEnum() InstanceConfigMarketTypeEnum {
	return InstanceConfigMarketTypeEnum{
		SPOT: InstanceConfigMarketType{
			value: "spot",
		},
	}
}

func (c InstanceConfigMarketType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceConfigMarketType) UnmarshalJSON(b []byte) error {
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
