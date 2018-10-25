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

// Package options contains flags for initializing an autoscaler.
package options

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
)

// AutoScalerConfig configures and runs an autoscaler server
type AutoScalerConfig struct {
	//Namespace         string
	//Target            string
	//DefaultConfig     string
	//ConfigFile        string
	PollPeriod   time.Duration
	UpdatePeriod time.Duration
	Kubeconfig   string
	PrintVersion bool

	// TODO: Any value now??
	DryRun bool
}

// NewAutoScalerConfig returns a Autoscaler config
func NewAutoScalerConfig() *AutoScalerConfig {
	return &AutoScalerConfig{
		// Defaults.
		//Namespace:         os.Getenv("MY_NAMESPACE"),
		PollPeriod:   time.Second * 10,
		UpdatePeriod: time.Second * 10,
		PrintVersion: false,
		DryRun:       false,
	}
}

// AddFlags adds flags to the specified FlagSet.
func (c *AutoScalerConfig) AddFlags(fs *flag.FlagSet) {
	//fs.StringVar(&c.Target, "target", c.Target, "The target object to scale. Format: deployment/* (not case sensitive).")
	//fs.StringVar(&c.Namespace, "namespace", c.Namespace, "The Namespace of the --target. Defaults to ${MY_NAMESPACE}.")
	//fs.StringVar(&c.DefaultConfig, "default-config", c.DefaultConfig, "The default configuration (in JSON format).")
	//fs.StringVar(&c.ConfigFile, "config-file", c.ConfigFile, "A config file (in JSON format), which overrides the --default-config.")
	fs.DurationVar(&c.PollPeriod, "poll-period", c.PollPeriod, "The period to poll cluster state.")
	fs.DurationVar(&c.UpdatePeriod, "update-period", c.UpdatePeriod, "The period with which we consider applying resource changes.")
	fs.StringVar(&c.Kubeconfig, "kubeconfig", c.Kubeconfig, "Path to a kubeconfig. Only required if running out-of-cluster.")
	fs.BoolVar(&c.PrintVersion, "version", c.PrintVersion, "Print the version and exit.")
	fs.BoolVar(&c.DryRun, "dry-run", c.DryRun, "Calculate updates for a target but does not apply the update.")
}

//// InitFlags no// WordSepNormalizeFunc changes all flags that contain "_" separators
//func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
//	if strings.Contains(name, "_") {
//		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
//	}
//	return pflag.NormalizedName(name)
//}

func (c *AutoScalerConfig) InitFlags() {
	flag.Set("logtostderr", "true")
	flag.Parse()
}

// ValidateFlags validates whether flags are set up correctly
func (c *AutoScalerConfig) ValidateFlags() error {
	var errorsFound bool

	//c.Target = strings.ToLower(c.Target)
	//if !isTargetFormatValid(c.Target) {
	//	errorsFound = true
	//}
	//if c.Namespace == "" {
	//	errorsFound = true
	//	glog.Errorf("--namespace parameter not set and failed to fallback")
	//}
	//if c.DefaultConfig == "" && c.ConfigFile == "" {
	//	errorsFound = true
	//	glog.Errorf("Either --default-config or --config-file must be specified")
	//}
	if c.PollPeriod.Seconds() < 1.0 {
		errorsFound = true
		glog.Errorf("--poll-period cannot be less than 1")
	}
	if c.UpdatePeriod.Seconds() < 1.0 {
		errorsFound = true
		glog.Errorf("--update-period cannot be less than 1")
	}

	// Log all sanity check errors before returning a single error string
	if errorsFound {
		return fmt.Errorf("failed to validate config parameters")
	}
	return nil
}

//func isTargetFormatValid(target string) bool {
//	if target == "" {
//		glog.Errorf("--target parameter cannot be empty")
//		return false
//	}
//	target = strings.ToLower(target)
//
//	if strings.HasPrefix(target, "deployment/") ||
//		strings.HasPrefix(target, "daemonset/") ||
//		strings.HasPrefix(target, "replicaset/") {
//		return true
//	}
//
//	glog.Errorf("Unknown target format: must be one of deployment/*, daemonset/*, or replicaset/* (not case sensitive).")
//	return false
//}
