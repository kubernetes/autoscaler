/*
Copyright 2024 The Kubernetes Authors.

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

// Package flags - common code for flags of all 3 VPA components

package common

import (
	"flag"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// CommonFlags contains flag definitions common to all VPA components
type CommonFlags struct {
	KubeConfig                 string
	KubeApiQps                 float64
	KubeApiBurst               float64
	EnableProfiling            bool
	VpaObjectNamespace         string
	IgnoredVpaObjectNamespaces string
}

// InitCommonFlags initializes the common flags
func InitCommonFlags() *CommonFlags {
	cf := &CommonFlags{}
	flag.StringVar(&cf.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.Float64Var(&cf.KubeApiQps, "kube-api-qps", 5.0, "QPS limit when making requests to Kubernetes apiserver")
	flag.Float64Var(&cf.KubeApiBurst, "kube-api-burst", 10.0, "QPS burst limit when making requests to Kubernetes apiserver")
	flag.BoolVar(&cf.EnableProfiling, "profiling", false, "Is debug/pprof endpoint enabled")
	flag.StringVar(&cf.VpaObjectNamespace, "vpa-object-namespace", apiv1.NamespaceAll, "Specifies the namespace to search for VPA objects. Leave empty to include all namespaces. If provided, the garbage collector will only clean this namespace.")
	flag.StringVar(&cf.IgnoredVpaObjectNamespaces, "ignored-vpa-object-namespaces", "", "A comma-separated list of namespaces to ignore when searching for VPA objects. Leave empty to avoid ignoring any namespaces. These namespaces will not be cleaned by the garbage collector.")
	return cf
}

// InitLoggingFlags initializes the logging flags
func InitLoggingFlags() {
	// Set the default log level to 4 (info)
	verbosity := flag.Lookup("v")
	if verbosity == nil {
		klog.Fatalf("Unable to find log level verbosity flag")
	}
	verbosity.Usage = "set the log level verbosity (default: 4)"
	err := flag.Set("v", "4")
	if err != nil {
		klog.Fatalf("Unable to set log level verbosity: %v", err)
	}

	// Set the default log level threshold for writing to standard error to 0 (info)
	threshold := flag.Lookup("stderrthreshold")
	if threshold == nil {
		klog.Fatalf("Unable to find log level threshold for writing to standard error flag")
	}
	threshold.Usage = "set the log level threshold for writing to standard error (default: info)"
	err = flag.Set("stderrthreshold", "0")
	if err != nil {
		klog.Fatalf("Unable to set log level threshold for writing to standard error: %v", err)
	}
}
