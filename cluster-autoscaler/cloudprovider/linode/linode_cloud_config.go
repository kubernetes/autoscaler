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

package linode

import (
	"fmt"
	"io"
	"strconv"

	"gopkg.in/gcfg.v1"
)

const (
	// defaultMinSizePerLinodeType is the min size of the node groups
	// if no other value is defined for a specific node group.
	defaultMinSizePerLinodeType int = 1
	// defaultMaxSizePerLinodeType is the max size of the node groups
	// if no other value is defined for a specific node group.
	defaultMaxSizePerLinodeType int = 254
)

// nodeGroupConfig is the configuration for a specific node group.
type nodeGroupConfig struct {
	minSize int
	maxSize int
}

// linodeConfig holds the configuration for the linode provider.
type linodeConfig struct {
	clusterID       int
	token           string
	defaultMinSize  int
	defaultMaxSize  int
	excludedPoolIDs map[int]bool
	nodeGroupCfg    map[string]*nodeGroupConfig
}

// GcfgGlobalConfig is the gcfg representation of the global section in the cloud config file for linode.
type GcfgGlobalConfig struct {
	ClusterID       string   `gcfg:"lke-cluster-id"`
	Token           string   `gcfg:"linode-token"`
	DefaultMinSize  string   `gcfg:"defaut-min-size-per-linode-type"`
	DefaultMaxSize  string   `gcfg:"defaut-max-size-per-linode-type"`
	ExcludedPoolIDs []string `gcfg:"do-not-import-pool-id"`
}

// GcfgNodeGroupConfig is the gcfg representation of the section in the cloud config file to change defaults for a node group.
type GcfgNodeGroupConfig struct {
	MinSize string `gcfg:"min-size"`
	MaxSize string `gcfg:"max-size"`
}

// gcfgCloudConfig is the gcfg representation of the cloud config file for linode.
type gcfgCloudConfig struct {
	Global     GcfgGlobalConfig                `gcfg:"global"`
	NodeGroups map[string]*GcfgNodeGroupConfig `gcfg:"nodegroup"`
}

// buildCloudConfig creates the configuration struct for the provider.
func buildCloudConfig(config io.Reader) (*linodeConfig, error) {

	// read the config and get the gcfg struct
	var gcfgCloudConfig gcfgCloudConfig
	if err := gcfg.ReadInto(&gcfgCloudConfig, config); err != nil {
		return nil, err
	}

	// get the clusterID
	clusterID, err := strconv.Atoi(gcfgCloudConfig.Global.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("LKE Cluster ID %q is not a number: %v", gcfgCloudConfig.Global.ClusterID, err)
	}

	// get the linode token
	token := gcfgCloudConfig.Global.Token
	if len(gcfgCloudConfig.Global.Token) == 0 {
		return nil, fmt.Errorf("linode token not present in global section")
	}

	// get the list of LKE pools that must not be imported
	excludedPoolIDs := make(map[int]bool)
	for _, excludedPoolIDStr := range gcfgCloudConfig.Global.ExcludedPoolIDs {
		excludedPoolID, err := strconv.Atoi(excludedPoolIDStr)
		if err != nil {
			return nil, fmt.Errorf("excluded pool id %q is not a number: %v", excludedPoolIDStr, err)
		}
		excludedPoolIDs[excludedPoolID] = true
	}

	// get the default min and max size as defined in the global section of the config file
	defaultMinSize, defaultMaxSize, err := getSizeLimits(
		gcfgCloudConfig.Global.DefaultMinSize,
		gcfgCloudConfig.Global.DefaultMaxSize,
		defaultMinSizePerLinodeType,
		defaultMaxSizePerLinodeType)
	if err != nil {
		return nil, fmt.Errorf("cannot get default size values in global section: %v", err)
	}

	// get the specific configuration of a node group
	nodeGroupCfg := make(map[string]*nodeGroupConfig)
	for nodeType, gcfgNodeGroup := range gcfgCloudConfig.NodeGroups {
		minSize, maxSize, err := getSizeLimits(gcfgNodeGroup.MinSize, gcfgNodeGroup.MaxSize, defaultMinSize, defaultMaxSize)
		if err != nil {
			return nil, fmt.Errorf("cannot get size values for node group %q: %v", nodeType, err)
		}
		ngc := &nodeGroupConfig{
			maxSize: maxSize,
			minSize: minSize,
		}
		nodeGroupCfg[nodeType] = ngc
	}

	return &linodeConfig{
		clusterID:       clusterID,
		token:           token,
		defaultMinSize:  defaultMinSize,
		defaultMaxSize:  defaultMaxSize,
		excludedPoolIDs: excludedPoolIDs,
		nodeGroupCfg:    nodeGroupCfg,
	}, nil
}

// getSizeLimits takes the max, min size of a node group as strings (empty if no values are provided)
// and default sizes, validates them an returns them as integer, or an error if such occurred
func getSizeLimits(minStr string, maxStr string, defaultMin int, defaultMax int) (int, int, error) {
	var err error
	min := defaultMin
	if len(minStr) != 0 {
		min, err = strconv.Atoi(minStr)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse min size for node group: %v", err)
		}
	}
	if min < 1 {
		return 0, 0, fmt.Errorf("min size for node group cannot be < 1")
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
