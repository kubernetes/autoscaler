/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ess

// ScalingGroup is a nested struct in ess response
type ScalingGroup struct {
	VpcId                               string `json:"VpcId" xml:"VpcId"`
	CreationTime                        string `json:"CreationTime" xml:"CreationTime"`
	TotalInstanceCount                  int    `json:"TotalInstanceCount" xml:"TotalInstanceCount"`
	ScalingGroupName                    string `json:"ScalingGroupName" xml:"ScalingGroupName"`
	Weighted                            bool   `json:"Weighted" xml:"Weighted"`
	SpotInstancePools                   int    `json:"SpotInstancePools" xml:"SpotInstancePools"`
	StoppedCapacity                     int    `json:"StoppedCapacity" xml:"StoppedCapacity"`
	OnDemandPercentageAboveBaseCapacity int    `json:"OnDemandPercentageAboveBaseCapacity" xml:"OnDemandPercentageAboveBaseCapacity"`
	ModificationTime                    string `json:"ModificationTime" xml:"ModificationTime"`
	MinSize                             int    `json:"MinSize" xml:"MinSize"`
	ScalingGroupId                      string `json:"ScalingGroupId" xml:"ScalingGroupId"`
	CompensateWithOnDemand              bool   `json:"CompensateWithOnDemand" xml:"CompensateWithOnDemand"`
	ScalingPolicy                       string `json:"ScalingPolicy" xml:"ScalingPolicy"`
	RemovingWaitCapacity                int    `json:"RemovingWaitCapacity" xml:"RemovingWaitCapacity"`
	ActiveCapacity                      int    `json:"ActiveCapacity" xml:"ActiveCapacity"`
	OnDemandBaseCapacity                int    `json:"OnDemandBaseCapacity" xml:"OnDemandBaseCapacity"`
	ProtectedCapacity                   int    `json:"ProtectedCapacity" xml:"ProtectedCapacity"`
	HealthCheckType                     string `json:"HealthCheckType" xml:"HealthCheckType"`
	LifecycleState                      string `json:"LifecycleState" xml:"LifecycleState"`
	GroupDeletionProtection             bool   `json:"GroupDeletionProtection" xml:"GroupDeletionProtection"`
	ActiveScalingConfigurationId        string `json:"ActiveScalingConfigurationId" xml:"ActiveScalingConfigurationId"`
	GroupType                           string `json:"GroupType" xml:"GroupType"`
	MultiAZPolicy                       string `json:"MultiAZPolicy" xml:"MultiAZPolicy"`
	RemovingCapacity                    int    `json:"RemovingCapacity" xml:"RemovingCapacity"`
	PendingWaitCapacity                 int    `json:"PendingWaitCapacity" xml:"PendingWaitCapacity"`
	StandbyCapacity                     int    `json:"StandbyCapacity" xml:"StandbyCapacity"`
	CurrentHostName                     string `json:"CurrentHostName" xml:"CurrentHostName"`
	PendingCapacity                     int    `json:"PendingCapacity" xml:"PendingCapacity"`
	LaunchTemplateId                    string `json:"LaunchTemplateId" xml:"LaunchTemplateId"`
	TotalCapacity                       int    `json:"TotalCapacity" xml:"TotalCapacity"`
	DesiredCapacity                     int    `json:"DesiredCapacity" xml:"DesiredCapacity"`
	SpotInstanceRemedy                  bool   `json:"SpotInstanceRemedy" xml:"SpotInstanceRemedy"`
	LaunchTemplateVersion               string `json:"LaunchTemplateVersion" xml:"LaunchTemplateVersion"`
	RegionId                            string `json:"RegionId" xml:"RegionId"`
	VSwitchId                           string `json:"VSwitchId" xml:"VSwitchId"`
	MaxSize                             int    `json:"MaxSize" xml:"MaxSize"`
	ScaleOutAmountCheck                 bool   `json:"ScaleOutAmountCheck" xml:"ScaleOutAmountCheck"`
	DefaultCooldown                     int    `json:"DefaultCooldown" xml:"DefaultCooldown"`
	SystemSuspended                     bool   `json:"SystemSuspended" xml:"SystemSuspended"`
	IsElasticStrengthInAlarm            bool   `json:"IsElasticStrengthInAlarm" xml:"IsElasticStrengthInAlarm"`
	MonitorGroupId                      string `json:"MonitorGroupId" xml:"MonitorGroupId"`
	AzBalance                           bool   `json:"AzBalance" xml:"AzBalance"`
	AllocationStrategy                  string `json:"AllocationStrategy" xml:"AllocationStrategy"`
	SpotAllocationStrategy              string `json:"SpotAllocationStrategy" xml:"SpotAllocationStrategy"`
	MaxInstanceLifetime                 int    `json:"MaxInstanceLifetime" xml:"MaxInstanceLifetime"`
	CustomPolicyARN                     string `json:"CustomPolicyARN" xml:"CustomPolicyARN"`
	InitCapacity                        int    `json:"InitCapacity" xml:"InitCapacity"`
	ResourceGroupId                     string `json:"ResourceGroupId" xml:"ResourceGroupId"`
	EnableDesiredCapacity               bool   `json:"EnableDesiredCapacity" xml:"EnableDesiredCapacity"`
	//RemovalPolicies                     RemovalPolicies         `json:"RemovalPolicies" xml:"RemovalPolicies"`
	//DBInstanceIds                       DBInstanceIds           `json:"DBInstanceIds" xml:"DBInstanceIds"`
	//LoadBalancerIds                     LoadBalancerIds         `json:"LoadBalancerIds" xml:"LoadBalancerIds"`
	//VSwitchIds                          VSwitchIds              `json:"VSwitchIds" xml:"VSwitchIds"`
	//SuspendedProcesses                  SuspendedProcesses      `json:"SuspendedProcesses" xml:"SuspendedProcesses"`
	//VServerGroups                       VServerGroups           `json:"VServerGroups" xml:"VServerGroups"`
	//LaunchTemplateOverrides             LaunchTemplateOverrides `json:"LaunchTemplateOverrides" xml:"LaunchTemplateOverrides"`
	//AlbServerGroups                     AlbServerGroups         `json:"AlbServerGroups" xml:"AlbServerGroups"`
	//ServerGroups                        ServerGroups            `json:"ServerGroups" xml:"ServerGroups"`
	//LoadBalancerConfigs                 LoadBalancerConfigs     `json:"LoadBalancerConfigs" xml:"LoadBalancerConfigs"`
}
