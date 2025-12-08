/*
Copyright 2025 The Kubernetes Authors.

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

package slicer

import (
	"fmt"
	"io"
	"strconv"

	"gopkg.in/gcfg.v1"
)

const (
	// defaultMinSize is the min size of the node groups
	// if no other value is defined for a specific node group.
	defaultMinSize int = 0
	// defaultMaxSize is the max size of the node groups
	// if no other value is defined for a specific node group.
	defaultMaxSize int = 8
)

// nodeGroupConfig is the configuration for a specific node group.
type nodeGroupConfig struct {
	slicerUrl   string
	slicerToken string
	minSize     int
	maxSize     int
	arch        string
}

// MinSize returns the MinSize value
func (ng *nodeGroupConfig) MinSize() int {
	return ng.minSize
}

// MaxSize returns the MaxSize value
func (ng *nodeGroupConfig) MaxSize() int {
	return ng.maxSize
}

// slicerConfig holds the configuration for the slicer provider.
type slicerConfig struct {
	k3sURL         string
	k3sToken       string
	caBundle       string
	defaultMinSize int
	defaultMaxSize int
	nodeGroupCfg   map[string]*nodeGroupConfig
}

// DefaultMinSize returns the DefaultMinSize value
func (sc *slicerConfig) DefaultMinSize() int {
	return sc.defaultMinSize
}

// DefaultMaxSize returns the DefaultMaxSize value
func (sc *slicerConfig) DefaultMaxSize() int {
	return sc.defaultMaxSize
}

// CABundle returns the CA bundle path
func (sc *slicerConfig) CABundle() string {
	return sc.caBundle
}

// GcfgGlobalConfig is the gcfg representation of the global section in the cloud config file for slicer.
type GcfgGlobalConfig struct {
	K3SURL         string `gcfg:"k3s-url"`
	K3SToken       string `gcfg:"k3s-token"`
	CABundle       string `gcfg:"ca-bundle"`
	DefaultMinSize string `gcfg:"default-min-size"`
	DefaultMaxSize string `gcfg:"default-max-size"`
}

// GcfgNodeGroupConfig is the gcfg representation of the section in the cloud config file to change defaults for a node group.
type GcfgNodeGroupConfig struct {
	SlicerUrl   string `gcfg:"slicer-url"`
	SlicerToken string `gcfg:"slicer-token"`
	MinSize     string `gcfg:"min-size"`
	MaxSize     string `gcfg:"max-size"`
	Arch        string `gcfg:"arch"`
}

// gcfgCloudConfig is the gcfg representation of the cloud config file for slicer.
type gcfgCloudConfig struct {
	Global     GcfgGlobalConfig                `gcfg:"global"`
	NodeGroups map[string]*GcfgNodeGroupConfig `gcfg:"nodegroup"`
}

// buildConfig creates the configuration struct for the provider.
func buildConfig(config io.Reader) (*slicerConfig, error) {

	// read the config and get the gcfg struct
	var gcfgCloudConfig gcfgCloudConfig
	if err := gcfg.ReadInto(&gcfgCloudConfig, config); err != nil {
		return nil, err
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
		minSize, maxSize, err := getSizeLimits(gcfgNodeGroup.MinSize, gcfgNodeGroup.MaxSize, defaultMinSize, defaultMaxSize)
		if err != nil {
			return nil, fmt.Errorf("cannot get size values for node group %q: %v", nodeGroupName, err)
		}

		ngc := &nodeGroupConfig{
			slicerUrl:   gcfgNodeGroup.SlicerUrl,
			slicerToken: gcfgNodeGroup.SlicerToken,
			minSize:     minSize,
			maxSize:     maxSize,
		}
		nodeGroupCfg[nodeGroupName] = ngc
	}

	// Validate required K3S configuration
	if gcfgCloudConfig.Global.K3SURL == "" {
		return nil, fmt.Errorf("K3S URL not configured, nodes may not be able to join the cluster")
	}

	if gcfgCloudConfig.Global.K3SToken == "" {
		return nil, fmt.Errorf("K3S token not configured, nodes may not be able to join the cluster")
	}

	return &slicerConfig{
		k3sURL:         gcfgCloudConfig.Global.K3SURL,
		k3sToken:       gcfgCloudConfig.Global.K3SToken,
		caBundle:       gcfgCloudConfig.Global.CABundle,
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
