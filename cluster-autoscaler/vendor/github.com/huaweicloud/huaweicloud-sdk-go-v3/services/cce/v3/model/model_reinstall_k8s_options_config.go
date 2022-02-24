package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点重装场景Kubernetes节点配置
type ReinstallK8sOptionsConfig struct {
	// 格式为key/value键值对。键值对个数不超过20条。  - Key：必须以字母或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符；另外可以使用DNS子域作为前缀，例如example.com/my-key， DNS子域最长253个字符。 - Value：可以为空或者非空字符串，非空字符串必须以字符或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符。  示例：  ``` \"k8sTags\": {  \"key\": \"value\" } ```

	Labels map[string]string `json:"labels,omitempty"`
	// 支持给创建出来的节点加Taints来设置反亲和性，taints配置不超过20条。每条Taints包含以下3个参数：  - Key：必须以字母或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符；另外可以使用DNS子域作为前缀。 - Value：必须以字符或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符。 - Effect：只可选NoSchedule，PreferNoSchedule或NoExecute。  示例：  ``` \"taints\": [{  \"key\": \"status\",  \"value\": \"unavailable\",  \"effect\": \"NoSchedule\" }, {  \"key\": \"looks\",  \"value\": \"bad\",  \"effect\": \"NoSchedule\" }] ```

	Taints *[]Taint `json:"taints,omitempty"`
	// 节点最大允许创建的实例数(Pod)，该数量包含系统默认实例，取值范围为16~256。 该设置的目的为防止节点因管理过多实例而负载过重，请根据您的业务需要进行设置。

	MaxPods *int32 `json:"maxPods,omitempty"`
	// - 弹性网卡队列数配置，默认配置示例如下： ``` \"[{\\\"queue\\\":4}]\" ``` 包含如下字段：   - queue: 弹性网卡队列数。 - 仅在turbo集群的BMS节点时，该字段才可配置。 - 当前支持可配置队列数以及弹性网卡数：{\"1\":128, \"2\":92, \"4\":92, \"8\":32, \"16\":16, \"28\":9}, 既1弹性网卡队列可绑定128张弹性网卡，2队列弹性网卡可绑定92张，以此类推。 - 弹性网卡队列数越多，性能越强，但可绑定弹性网卡数越少，请根据您的需求进行配置（创建后不可修改）。

	NicMultiqueue *string `json:"nicMultiqueue,omitempty"`
	// - 弹性网卡预绑定比例配置，默认配置示例如下： ``` \"0.3:0.6\" ```   - 第一位小数：预绑定低水位，弹性网卡预绑定的最低比例（最小预绑定弹性网卡数 = ⌊节点的总弹性网卡数 * 预绑定低水位⌋）   - 第二位小数：预绑定高水位，弹性网卡预绑定的最高比例（最大预绑定弹性网卡数 = ⌊节点的总弹性网卡数 * 预绑定高水位⌋）   - BMS节点上绑定的弹性网卡数：Pod正在使用的弹性网卡数 + 最小预绑定弹性网卡数 < BMS节点上绑定的弹性网卡数 < Pod正在使用的弹性网卡数 + 最大预绑定弹性网卡数   - BMS节点上当预绑定弹性网卡数 < 最小预绑定弹性网卡数时：会绑定弹性网卡，使得预绑定弹性网卡数 = 最小预绑定弹性网卡数   - BMS节点上当预绑定弹性网卡数 > 最大预绑定弹性网卡数时：会定时解绑弹性网卡（约2分钟一次），直到预绑定弹性网卡数 = 最大预绑定弹性网卡数   - 取值范围：[0.0, 1.0]; 一位小数; 低水位 <= 高水位 - 仅在turbo集群的BMS节点时，该字段才可配置。 - 弹性网卡预绑定能加快工作负载的创建，但会占用IP，请根据您的需求进行配置。

	NicThreshold *string `json:"nicThreshold,omitempty"`
}

func (o ReinstallK8sOptionsConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ReinstallK8sOptionsConfig struct{}"
	}

	return strings.Join([]string{"ReinstallK8sOptionsConfig", string(data)}, " ")
}
