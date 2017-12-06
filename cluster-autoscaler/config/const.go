/*
Copyright 2017 The Kubernetes Authors.

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

const (
	// Defaults.

	// DefaultMaxClusterCores is the default maximum number of cores in the cluster.
	DefaultMaxClusterCores = 5000 * 64
	// DefaultMaxClusterMemory is the default maximum number of gigabytes of memory in cluster.
	DefaultMaxClusterMemory = 5000 * 64 * 20

	// Useful universal constants.

	// Gigabyte is 2^30 bytes.
	Gigabyte = 1024 * 1024 * 1024
)
