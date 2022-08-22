/*
Copyright 2016 The Kubernetes Authors.

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

package kamatera

import (
	"fmt"
	"io"
	"strconv"

	"gopkg.in/gcfg.v1"
)

const (
	defaultMinSize int    = 1
	defaultMaxSize int    = 254
	defaultApiUrl  string = "https://cloudcli.cloudwm.com"
)

// nodeGroupConfig is the configuration for a specific node group.
type nodeGroupConfig struct {
	minSize        int
	maxSize        int
	NamePrefix     string
	Password       string
	SshKey         string
	Datacenter     string
	Image          string
	Cpu            string
	Ram            string
	Disks          []string
	Dailybackup    bool
	Managed        bool
	Networks       []string
	BillingCycle   string
	MonthlyPackage string
	ScriptBase64   string
}

// kamateraConfig holds the configuration for the Kamatera provider.
type kamateraConfig struct {
	apiClientId    string
	apiSecret      string
	apiUrl         string
	clusterName    string
	defaultMinSize int
	defaultMaxSize int
	nodeGroupCfg   map[string]*nodeGroupConfig // key is the node group name
}

// GcfgGlobalConfig is the gcfg representation of the global section in the cloud config file for Kamatera.
type GcfgGlobalConfig struct {
	KamateraApiClientId   string   `gcfg:"kamatera-api-client-id"`
	KamateraApiSecret     string   `gcfg:"kamatera-api-secret"`
	KamateraApiUrl        string   `gcfg:"kamatera-api-url"`
	ClusterName           string   `gcfg:"cluster-name"`
	DefaultMinSize        string   `gcfg:"default-min-size"`
	DefaultMaxSize        string   `gcfg:"default-max-size"`
	DefaultNamePrefix     string   `gcfg:"default-name-prefix"`
	DefaultPassword       string   `gcfg:"default-password"`
	DefaultSshKey         string   `gcfg:"default-ssh-key"`
	DefaultDatacenter     string   `gcfg:"default-datacenter"`
	DefaultImage          string   `gcfg:"default-image"`
	DefaultCpu            string   `gcfg:"default-cpu"`
	DefaultRam            string   `gcfg:"default-ram"`
	DefaultDisks          []string `gcfg:"default-disk"`
	DefaultDailybackup    bool     `gcfg:"default-dailybackup"`
	DefaultManaged        bool     `gcfg:"default-managed"`
	DefaultNetworks       []string `gcfg:"default-network"`
	DefaultBillingCycle   string   `gcfg:"default-billingcycle"`
	DefaultMonthlyPackage string   `gcfg:"default-monthlypackage"`
	DefaultScriptBase64   string   `gcfg:"default-script-base64"`
}

// GcfgNodeGroupConfig is the gcfg representation of the section in the cloud config file to change defaults for a node group.
type GcfgNodeGroupConfig struct {
	MinSize        string   `gcfg:"min-size"`
	MaxSize        string   `gcfg:"max-size"`
	NamePrefix     string   `gcfg:"name-prefix"`
	Password       string   `gcfg:"password"`
	SshKey         string   `gcfg:"ssh-key"`
	Datacenter     string   `gcfg:"datacenter"`
	Image          string   `gcfg:"image"`
	Cpu            string   `gcfg:"cpu"`
	Ram            string   `gcfg:"ram"`
	Disks          []string `gcfg:"disk"`
	Dailybackup    bool     `gcfg:"dailybackup"`
	Managed        bool     `gcfg:"managed"`
	Networks       []string `gcfg:"network"`
	BillingCycle   string   `gcfg:"billingcycle"`
	MonthlyPackage string   `gcfg:"monthlypackage"`
	ScriptBase64   string   `gcfg:"script-base64"`
}

// gcfgCloudConfig is the gcfg representation of the cloud config file for Kamatera.
type gcfgCloudConfig struct {
	Global     GcfgGlobalConfig                `gcfg:"global"`
	NodeGroups map[string]*GcfgNodeGroupConfig `gcfg:"nodegroup"` // key is the node group name
}

// buildCloudConfig creates the configuration struct for the provider.
func buildCloudConfig(config io.Reader) (*kamateraConfig, error) {

	// read the config and get the gcfg struct
	var gcfgCloudConfig gcfgCloudConfig
	if err := gcfg.ReadInto(&gcfgCloudConfig, config); err != nil {
		return nil, err
	}

	// get the clusterName and Kamatera tokens
	clusterName := gcfgCloudConfig.Global.ClusterName
	if len(clusterName) == 0 {
		return nil, fmt.Errorf("cluster name is not set")
	}
	apiClientId := gcfgCloudConfig.Global.KamateraApiClientId
	if len(apiClientId) == 0 {
		return nil, fmt.Errorf("kamatera api client id is not set")
	}
	apiSecret := gcfgCloudConfig.Global.KamateraApiSecret
	if len(apiSecret) == 0 {
		return nil, fmt.Errorf("kamatera api secret is not set")
	}
	apiUrl := gcfgCloudConfig.Global.KamateraApiUrl
	if len(apiUrl) == 0 {
		apiUrl = defaultApiUrl
	}

	// Cluster name must be max 15 characters due to limitation of Kamatera server tags
	if len(clusterName) > 15 {
		return nil, fmt.Errorf("cluster name must be at most 15 characters long")
	}

	// get the default min and max size as defined in the global section of the config file
	defaultMinSize, defaultMaxSize, err := getSizeLimits(
		gcfgCloudConfig.Global.DefaultMinSize,
		gcfgCloudConfig.Global.DefaultMaxSize,
		defaultMinSize,
		defaultMaxSize)
	if err != nil {
		return nil, fmt.Errorf("cannot get default size values in global section: %v", err)
	}

	// get the specific configuration of a node group
	nodeGroupCfg := make(map[string]*nodeGroupConfig)
	for nodeGroupName, gcfgNodeGroup := range gcfgCloudConfig.NodeGroups {
		// node group name must be max 15 characters due to limitation of Kamatera server tags
		if len(nodeGroupName) > 15 {
			return nil, fmt.Errorf("node group name must be at most 15 characters long")
		}
		minSize, maxSize, err := getSizeLimits(gcfgNodeGroup.MinSize, gcfgNodeGroup.MaxSize, defaultMinSize, defaultMaxSize)
		if err != nil {
			return nil, fmt.Errorf("cannot get size values for node group %s: %v", nodeGroupName, err)
		}
		namePrefix := gcfgCloudConfig.Global.DefaultNamePrefix
		if len(gcfgNodeGroup.NamePrefix) > 0 {
			namePrefix = gcfgNodeGroup.NamePrefix
		}
		password := gcfgCloudConfig.Global.DefaultPassword
		if len(gcfgNodeGroup.Password) > 0 {
			password = gcfgNodeGroup.Password
		}
		sshKey := gcfgCloudConfig.Global.DefaultSshKey
		if len(gcfgNodeGroup.SshKey) > 0 {
			sshKey = gcfgNodeGroup.SshKey
		}
		datacenter := gcfgCloudConfig.Global.DefaultDatacenter
		if len(gcfgNodeGroup.Datacenter) > 0 {
			datacenter = gcfgNodeGroup.Datacenter
		}
		image := gcfgCloudConfig.Global.DefaultImage
		if len(gcfgNodeGroup.Image) > 0 {
			image = gcfgNodeGroup.Image
		}
		cpu := gcfgCloudConfig.Global.DefaultCpu
		if len(gcfgNodeGroup.Cpu) > 0 {
			cpu = gcfgNodeGroup.Cpu
		}
		ram := gcfgCloudConfig.Global.DefaultRam
		if len(gcfgNodeGroup.Ram) > 0 {
			ram = gcfgNodeGroup.Ram
		}
		disks := gcfgCloudConfig.Global.DefaultDisks
		if gcfgNodeGroup.Disks != nil {
			disks = gcfgNodeGroup.Disks
		}
		dailybackup := gcfgCloudConfig.Global.DefaultDailybackup
		if gcfgNodeGroup.Dailybackup {
			dailybackup = gcfgNodeGroup.Dailybackup
		}
		managed := gcfgCloudConfig.Global.DefaultManaged
		if gcfgNodeGroup.Managed {
			managed = gcfgNodeGroup.Managed
		}
		networks := gcfgCloudConfig.Global.DefaultNetworks
		if gcfgNodeGroup.Networks != nil {
			networks = gcfgNodeGroup.Networks
		}
		billingCycle := gcfgCloudConfig.Global.DefaultBillingCycle
		if len(gcfgNodeGroup.BillingCycle) > 0 {
			billingCycle = gcfgNodeGroup.BillingCycle
		}
		monthlyPackage := gcfgCloudConfig.Global.DefaultMonthlyPackage
		if len(gcfgNodeGroup.MonthlyPackage) > 0 {
			monthlyPackage = gcfgNodeGroup.MonthlyPackage
		}
		scriptBase64 := gcfgCloudConfig.Global.DefaultScriptBase64
		if len(gcfgNodeGroup.ScriptBase64) > 0 {
			scriptBase64 = gcfgNodeGroup.ScriptBase64
		}
		ngc := &nodeGroupConfig{
			maxSize:        maxSize,
			minSize:        minSize,
			NamePrefix:     namePrefix,
			Password:       password,
			SshKey:         sshKey,
			Datacenter:     datacenter,
			Image:          image,
			Cpu:            cpu,
			Ram:            ram,
			Disks:          disks,
			Dailybackup:    dailybackup,
			Managed:        managed,
			Networks:       networks,
			BillingCycle:   billingCycle,
			MonthlyPackage: monthlyPackage,
			ScriptBase64:   scriptBase64,
		}
		nodeGroupCfg[nodeGroupName] = ngc
	}

	return &kamateraConfig{
		clusterName:    clusterName,
		apiClientId:    apiClientId,
		apiSecret:      apiSecret,
		apiUrl:         apiUrl,
		defaultMinSize: defaultMinSize,
		defaultMaxSize: defaultMaxSize,
		nodeGroupCfg:   nodeGroupCfg,
	}, nil
}

// getSizeLimits takes the max, min size of a node group as strings (empty if no values are provided)
// and default sizes, validates them and returns them as integer, or an error if such occurred
func getSizeLimits(minStr string, maxStr string, defaultMin int, defaultMax int) (int, int, error) {
	var err error
	min := defaultMin
	if len(minStr) != 0 {
		min, err = strconv.Atoi(minStr)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse min size for node group: %v", err)
		}
	}
	if min < 0 {
		return 0, 0, fmt.Errorf("min size for node group cannot be < 0")
	}
	max := defaultMax
	if len(maxStr) != 0 {
		max, err = strconv.Atoi(maxStr)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse max size for node group: %v", err)
		}
	}
	if min > max {
		return 0, 0, fmt.Errorf("min size for a node group must be less than its max size (got min: %d, max: %d)",
			min, max)
	}
	return min, max, nil
}
