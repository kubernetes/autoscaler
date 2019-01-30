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

package units

const (
	// GB - GigaByte size (10^9)
	GB = 1000 * 1000 * 1000
	// GiB - GibiByte size (2^30)
	GiB = 1024 * 1024 * 1024
	// MB - MegaByte size (10^6)
	MB = 1000 * 1000
	// MiB - MebiByte size (2^20)
	MiB = 1024 * 1024
)
