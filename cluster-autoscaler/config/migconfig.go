/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"fmt"
	"strconv"
	"strings"

	gceurl "k8s.io/contrib/cluster-autoscaler/utils/gce_url"
)

// InstanceConfig contains instance configuration details.
type InstanceConfig struct {
	Project string
	Zone    string
	Name    string
}

// InstanceConfigFromProviderId creates InstanceConfig object
// from provider id which must be in format:
// gce://<project-id>/<zone>/<name>
// TODO(piosz): add better check whether the id is correct
func InstanceConfigFromProviderId(id string) (*InstanceConfig, error) {
	splitted := strings.Split(id[6:], "/")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("Wrong id: expected format gce://<project-id>/<zone>/<name>, got %v", id)
	}
	return &InstanceConfig{
		Project: splitted[0],
		Zone:    splitted[1],
		Name:    splitted[2],
	}, nil
}

// MigConfig contains managed instance group configuration details.
type MigConfig struct {
	MinSize int
	MaxSize int
	Project string
	Zone    string
	Name    string
}

// MigConfigFlag is an array of MIG configuration details. Working as a multi-value flag.
type MigConfigFlag []MigConfig

func (migconfigflag *MigConfigFlag) String() string {
	configs := make([]string, len(*migconfigflag))
	for _, migconfig := range *migconfigflag {
		configs = append(configs, fmt.Sprintf("%d:%d:%s:%s", migconfig.MinSize, migconfig.MaxSize, migconfig.Zone, migconfig.Name))
	}
	return "[" + strings.Join(configs, " ") + "]"
}

// Set adds a new configuration.
func (migconfigflag *MigConfigFlag) Set(value string) error {
	tokens := strings.SplitN(value, ":", 3)
	if len(tokens) != 3 {
		return fmt.Errorf("wrong nodes configuration: %s", value)
	}
	migconfig := MigConfig{}
	if size, err := strconv.Atoi(tokens[0]); err == nil {
		if size <= 0 {
			return fmt.Errorf("min size must be >= 1")
		}
		migconfig.MinSize = size
	} else {
		return fmt.Errorf("failed to set min size: %s, expected integer", tokens[0])
	}

	if size, err := strconv.Atoi(tokens[1]); err == nil {
		if size < migconfig.MinSize {
			return fmt.Errorf("max size must be greater or equal to min size")
		}
		migconfig.MaxSize = size
	} else {
		return fmt.Errorf("failed to set max size: %s, expected integer", tokens[1])
	}

	var err error
	if migconfig.Project, migconfig.Zone, migconfig.Name, err = gceurl.ParseMigUrl(tokens[2]); err != nil {
		return fmt.Errorf("failed to parse mig url: %s got error: %v", tokens[2], err)
	}

	*migconfigflag = append(*migconfigflag, migconfig)
	return nil
}
