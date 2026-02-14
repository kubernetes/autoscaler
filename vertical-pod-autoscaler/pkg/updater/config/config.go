/*
Copyright The Kubernetes Authors.

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
	"flag"
	"os"
	"time"

	"github.com/spf13/pflag"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

// UpdaterConfig holds all configuration for the admission controller component
type UpdaterConfig struct {
	// Common flags
	CommonFlags *common.CommonFlags

	UpdaterInterval              time.Duration
	MinReplicas                  int
	EvictionToleranceFraction    float64
	EvictionRateLimit            float64
	EvictionRateBurst            int
	Namespace                    string
	Address                      string
	UseAdmissionControllerStatus bool
	InPlaceSkipDisruptionBudget  bool

	DefaultUpdateThreshold     float64
	PodLifetimeUpdateThreshold time.Duration
	EvictAfterOOMThreshold     time.Duration
}

// DefaultUpdaterConfig returns a UpdaterConfig with default values
func DefaultUpdaterConfig() *UpdaterConfig {
	return &UpdaterConfig{
		CommonFlags:                  common.DefaultCommonConfig(),
		UpdaterInterval:              1 * time.Minute,
		MinReplicas:                  2,
		EvictionToleranceFraction:    0.5,
		EvictionRateLimit:            -1,
		EvictionRateBurst:            1,
		Namespace:                    os.Getenv("NAMESPACE"),
		Address:                      ":8943",
		UseAdmissionControllerStatus: true,
		InPlaceSkipDisruptionBudget:  false,

		DefaultUpdateThreshold:     0.1,
		PodLifetimeUpdateThreshold: time.Hour * 12,
		EvictAfterOOMThreshold:     10 * time.Minute,
	}
}

// InitUpdaterFlags initializes flags for the updater component
func InitUpdaterFlags() *UpdaterConfig {
	config := DefaultUpdaterConfig()
	config.CommonFlags = common.InitCommonFlags()

	flag.DurationVar(&config.UpdaterInterval, "updater-interval", config.UpdaterInterval, "How often updater should run")
	flag.IntVar(&config.MinReplicas, "min-replicas", config.MinReplicas, "Minimum number of replicas to perform update")
	flag.Float64Var(&config.EvictionToleranceFraction, "eviction-tolerance", config.EvictionToleranceFraction, "Fraction of replica count that can be evicted for update, if more than one pod can be evicted.")
	flag.Float64Var(&config.EvictionRateLimit, "eviction-rate-limit", config.EvictionRateLimit, "Number of pods that can be evicted per seconds. A rate limit set to 0 or -1 will disable the rate limiter.")
	flag.IntVar(&config.EvictionRateBurst, "eviction-rate-burst", config.EvictionRateBurst, "Burst of pods that can be evicted.")
	flag.StringVar(&config.Address, "address", config.Address, "The address to expose Prometheus metrics.")
	flag.BoolVar(&config.UseAdmissionControllerStatus, "use-admission-controller-status", config.UseAdmissionControllerStatus, "If true, updater will only evict pods when admission controller status is valid.")
	flag.BoolVar(&config.InPlaceSkipDisruptionBudget, "in-place-skip-disruption-budget", config.InPlaceSkipDisruptionBudget, "[ALPHA] If true, VPA updater skips disruption budget checks for in-place pod updates when all containers have NotRequired resize policy (or no policy defined) for both CPU and memory resources. Disruption budgets are still respected when any container has RestartContainer resize policy for any resource.")

	flag.Float64Var(&config.DefaultUpdateThreshold, "pod-update-threshold", config.DefaultUpdateThreshold, "Ignore updates that have priority lower than the value of this flag")
	flag.DurationVar(&config.PodLifetimeUpdateThreshold, "in-recommendation-bounds-eviction-lifetime-threshold", config.PodLifetimeUpdateThreshold, "Pods that live for at least that long can be evicted even if their request is within the [MinRecommended...MaxRecommended] range")
	flag.DurationVar(&config.EvictAfterOOMThreshold, "evict-after-oom-threshold", config.EvictAfterOOMThreshold, `The default duration to evict pods that have OOMed in less than evict-after-oom-threshold since start.`)

	// These need to happen last. kube_flag.InitFlags() synchronizes and parses
	// flags from the flag package to pflag, so feature gates must be added to
	// pflag before InitFlags() is called.
	klog.InitFlags(nil)
	common.InitLoggingFlags()
	features.MutableFeatureGate.AddFlag(pflag.CommandLine)
	kube_flag.InitFlags()

	ValidateUpdaterConfig(config)

	return config
}

// ValidateUpdaterConfig performs validation of the updater flags
func ValidateUpdaterConfig(config *UpdaterConfig) {
	common.ValidateCommonConfig(config.CommonFlags)
}
