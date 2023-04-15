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

package hcloud

// Architecture specifies the architecture of the CPU.
type Architecture string

const (
	// ArchitectureX86 is the architecture for Intel/AMD x86 CPUs.
	ArchitectureX86 Architecture = "x86"

	// ArchitectureARM is the architecture for ARM CPUs.
	ArchitectureARM Architecture = "arm"
)
