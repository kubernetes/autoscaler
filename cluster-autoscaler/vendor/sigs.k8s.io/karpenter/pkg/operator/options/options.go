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

package options

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/samber/lo"
	cliflag "k8s.io/component-base/cli/flag"

	"sigs.k8s.io/karpenter/pkg/utils/env"
)

type PreferencePolicy string

const (
	PreferencePolicyIgnore  PreferencePolicy = "Ignore"
	PreferencePolicyRespect PreferencePolicy = "Respect"
)

type MinValuesPolicy string

const (
	MinValuesPolicyStrict     MinValuesPolicy = "Strict"
	MinValuesPolicyBestEffort MinValuesPolicy = "BestEffort"
)

var (
	validLogLevels          = []string{"", "debug", "info", "error"}
	validPreferencePolicies = []PreferencePolicy{PreferencePolicyIgnore, PreferencePolicyRespect}

	Injectables = []Injectable{&Options{}}
)

type optionsKey struct{}

type FeatureGates struct {
	inputStr string

	NodeRepair              bool
	ReservedCapacity        bool
	SpotToSpotConsolidation bool
	NodeOverlay             bool
	StaticCapacity          bool
	CapacityBuffer          bool
}

// Options contains all CLI flags / env vars for karpenter-core. It adheres to the options.Injectable interface.
type Options struct {
	ServiceName                      string
	MetricsPort                      int
	HealthProbePort                  int
	KubeClientQPS                    int
	KubeClientBurst                  int
	EnableProfiling                  bool
	DisableControllerWarmup          bool
	DisableLeaderElection            bool
	DisableClusterStateObservability bool
	LeaderElectionName               string
	LeaderElectionNamespace          string
	MemoryLimit                      int64
	CPURequests                      int64
	LogLevel                         string
	LogOutputPaths                   string
	LogErrorOutputPaths              string
	BatchMaxDuration                 time.Duration
	BatchIdleDuration                time.Duration
	preferencePolicyRaw              string
	PreferencePolicy                 PreferencePolicy
	minValuesPolicyRaw               string
	MinValuesPolicy                  MinValuesPolicy
	IgnoreDRARequests                bool // NOTE: This flag will be removed once formal DRA support is GA in Karpenter.
	FeatureGates                     FeatureGates
}

type FlagSet struct {
	*flag.FlagSet
}

// BoolVarWithEnv defines a bool flag with a specified name, default value, usage string, and fallback environment
// variable.
func (fs *FlagSet) BoolVarWithEnv(p *bool, name string, envVar string, val bool, usage string) {
	*p = env.WithDefaultBool(envVar, val)
	fs.BoolFunc(name, usage, func(val string) error {
		if val != "true" && val != "false" {
			return fmt.Errorf("%q is not a valid value, must be true or false", val)
		}
		*p = (val) == "true"
		return nil
	})
}

func (o *Options) AddFlags(fs *FlagSet) {
	fs.StringVar(&o.ServiceName, "karpenter-service", env.WithDefaultString("KARPENTER_SERVICE", ""), "The Karpenter Service name for the dynamic webhook certificate")
	fs.IntVar(&o.MetricsPort, "metrics-port", env.WithDefaultInt("METRICS_PORT", 8080), "The port the metric endpoint binds to for operating metrics about the controller itself")
	fs.IntVar(&o.HealthProbePort, "health-probe-port", env.WithDefaultInt("HEALTH_PROBE_PORT", 8081), "The port the health probe endpoint binds to for reporting controller health")
	fs.IntVar(&o.KubeClientQPS, "kube-client-qps", env.WithDefaultInt("KUBE_CLIENT_QPS", 200), "The smoothed rate of qps to kube-apiserver")
	fs.IntVar(&o.KubeClientBurst, "kube-client-burst", env.WithDefaultInt("KUBE_CLIENT_BURST", 300), "The maximum allowed burst of queries to the kube-apiserver")
	fs.BoolVarWithEnv(&o.EnableProfiling, "enable-profiling", "ENABLE_PROFILING", false, "Enable the profiling on the metric endpoint")
	fs.BoolVarWithEnv(&o.DisableControllerWarmup, "disable-controller-warmup", "DISABLE_CONTROLLER_WARMUP", true, "Disable controller warmup which starts controller sources before leader election is won. Controller warmup pre-populates caches and improves leader failover time.")
	fs.BoolVarWithEnv(&o.DisableLeaderElection, "disable-leader-election", "DISABLE_LEADER_ELECTION", false, "Disable the leader election client before executing the main loop. Disable when running replicated components for high availability is not desired.")
	fs.BoolVarWithEnv(&o.DisableClusterStateObservability, "disable-cluster-state-observability", "DISABLE_CLUSTER_STATE_OBSERVABILITY", false, "Disable cluster state metrics and events")
	fs.StringVar(&o.LeaderElectionName, "leader-election-name", env.WithDefaultString("LEADER_ELECTION_NAME", "karpenter-leader-election"), "Leader election name to create and monitor the lease if running outside the cluster")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", env.WithDefaultString("LEADER_ELECTION_NAMESPACE", ""), "Leader election namespace to create and monitor the lease if running outside the cluster")
	fs.Int64Var(&o.MemoryLimit, "memory-limit", env.WithDefaultInt64("MEMORY_LIMIT", -1), "Memory limit on the container running the controller. The GC soft memory limit is set to 90% of this value.")
	fs.Int64Var(&o.CPURequests, "cpu-requests", env.WithDefaultInt64("CPU_REQUESTS", 1000), "CPU requests in millicores on the container running the controller.")
	fs.StringVar(&o.LogLevel, "log-level", env.WithDefaultString("LOG_LEVEL", "info"), "Log verbosity level. Can be one of 'debug', 'info', or 'error'")
	fs.StringVar(&o.LogOutputPaths, "log-output-paths", env.WithDefaultString("LOG_OUTPUT_PATHS", "stdout"), "Optional comma separated paths for directing log output")
	fs.StringVar(&o.LogErrorOutputPaths, "log-error-output-paths", env.WithDefaultString("LOG_ERROR_OUTPUT_PATHS", "stderr"), "Optional comma separated paths for logging error output")
	fs.DurationVar(&o.BatchMaxDuration, "batch-max-duration", env.WithDefaultDuration("BATCH_MAX_DURATION", 10*time.Second), "The maximum length of a batch window. The longer this is, the more pods we can consider for provisioning at one time which usually results in fewer but larger nodes.")
	fs.DurationVar(&o.BatchIdleDuration, "batch-idle-duration", env.WithDefaultDuration("BATCH_IDLE_DURATION", time.Second), "The maximum amount of time with no new pending pods that if exceeded ends the current batching window. If pods arrive faster than this time, the batching window will be extended up to the maxDuration. If they arrive slower, the pods will be batched separately.")
	fs.StringVar(&o.preferencePolicyRaw, "preference-policy", env.WithDefaultString("PREFERENCE_POLICY", string(PreferencePolicyRespect)), "How the Karpenter scheduler should treat preferences. Preferences include preferredDuringSchedulingIgnoreDuringExecution node and pod affinities/anti-affinities and ScheduleAnyways topologySpreadConstraints. Can be one of 'Ignore' and 'Respect'")
	fs.StringVar(&o.minValuesPolicyRaw, "min-values-policy", env.WithDefaultString("MIN_VALUES_POLICY", string(MinValuesPolicyStrict)), "Min values policy for scheduling. Options include 'Strict' for existing behavior where min values are strictly enforced or 'BestEffort' where Karpenter relaxes min values when it isn't satisfied.")
	fs.BoolVarWithEnv(&o.IgnoreDRARequests, "ignore-dra-requests", "IGNORE_DRA_REQUESTS", true, "When set, Karpenter will ignore pods' DRA requests during scheduling simulations. NOTE: This flag will be removed once formal DRA support is GA in Karpenter.")
	fs.StringVar(&o.FeatureGates.inputStr, "feature-gates", env.WithDefaultString("FEATURE_GATES", "NodeRepair=false,ReservedCapacity=true,SpotToSpotConsolidation=false,NodeOverlay=false,StaticCapacity=false,CapacityBuffer=false"), "Optional features can be enabled / disabled using feature gates. Current options are: NodeRepair, ReservedCapacity, SpotToSpotConsolidation, NodeOverlay, StaticCapacity, and CapacityBuffer.")
}

func (o *Options) Parse(fs *FlagSet, args ...string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		return fmt.Errorf("parsing flags, %w", err)
	}
	if !lo.Contains(validLogLevels, o.LogLevel) {
		return fmt.Errorf("validating cli flags / env vars, invalid LOG_LEVEL %q", o.LogLevel)
	}
	if !lo.Contains(validPreferencePolicies, PreferencePolicy(o.preferencePolicyRaw)) {
		return fmt.Errorf("validating cli flags / env vars, invalid PREFERENCE_POLICY %q", o.preferencePolicyRaw)
	}
	if !lo.Contains([]MinValuesPolicy{MinValuesPolicyStrict, MinValuesPolicyBestEffort}, MinValuesPolicy(o.minValuesPolicyRaw)) {
		return fmt.Errorf("validating cli flags / env vars, invalid MIN_VALUES_POLICY %q", o.minValuesPolicyRaw)
	}
	if o.CPURequests <= 0 {
		o.CPURequests = 1000
	}
	gates, err := ParseFeatureGates(o.FeatureGates.inputStr)
	if err != nil {
		return fmt.Errorf("parsing feature gates, %w", err)
	}
	o.FeatureGates = gates
	o.PreferencePolicy = PreferencePolicy(o.preferencePolicyRaw)
	o.MinValuesPolicy = MinValuesPolicy(o.minValuesPolicyRaw)
	return nil
}

func (o *Options) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, o)
}

func DefaultFeatureGates() FeatureGates {
	return FeatureGates{
		NodeRepair:              false,
		ReservedCapacity:        true,
		SpotToSpotConsolidation: false,
		NodeOverlay:             false,
		StaticCapacity:          false,
		CapacityBuffer:          false,
	}
}

func ParseFeatureGates(gateStr string) (FeatureGates, error) {
	gateMap := map[string]bool{}
	gates := DefaultFeatureGates()

	// Parses feature gates with the upstream mechanism. This is meant to be used with flag directly but this enables
	// simple merging with environment vars.
	if err := cliflag.NewMapStringBool(&gateMap).Set(gateStr); err != nil {
		return gates, err
	}
	if val, ok := gateMap["NodeRepair"]; ok {
		gates.NodeRepair = val
	}
	if val, ok := gateMap["SpotToSpotConsolidation"]; ok {
		gates.SpotToSpotConsolidation = val
	}
	if val, ok := gateMap["ReservedCapacity"]; ok {
		gates.ReservedCapacity = val
	}
	if val, ok := gateMap["NodeOverlay"]; ok {
		gates.NodeOverlay = val
	}
	if val, ok := gateMap["StaticCapacity"]; ok {
		gates.StaticCapacity = val
	}
	if val, ok := gateMap["CapacityBuffer"]; ok {
		gates.CapacityBuffer = val
	}

	return gates, nil
}

func ToContext(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

func FromContext(ctx context.Context) *Options {
	retval := ctx.Value(optionsKey{})
	if retval == nil {
		// This is a developer error if this happens, so we should panic
		panic("options doesn't exist in context")
	}
	return retval.(*Options)
}
