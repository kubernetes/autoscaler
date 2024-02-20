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

package localssdsize

// LocalSSDSizeProvider contains methods to calculate local ssd disk size for GCE based on some input parameters (e.g. machine type name)
type LocalSSDSizeProvider interface {
	// Computes local ssd disk size in GiB based on machine type name
	SSDSizeInGiB(string) uint64
}

// localSSDDiskSizeInGiB is the size of each local SSD in GiB
// (cf. https://cloud.google.com/compute/docs/disks/local-ssd)
const localSSDDiskSizeInGiB = uint64(375)

// SimpleLocalSSDProvider implements LocalSSDSizeProvider
// It always returns a constant size
type SimpleLocalSSDProvider struct {
	ssdDiskSize uint64
}

// NewSimpleLocalSSDProvider creates an instance of SimpleLocalSSDProvider with `LocalSSDDiskSizeInGiB` as the disk size and returns a pointer to it
func NewSimpleLocalSSDProvider() LocalSSDSizeProvider {
	return &SimpleLocalSSDProvider{
		ssdDiskSize: localSSDDiskSizeInGiB,
	}
}

// SSDSizeInGiB Returns a constant disk size in GiB
// First parameter is not used and added to conform to the interface `LocalSSDSizeProvider`
func (lsp *SimpleLocalSSDProvider) SSDSizeInGiB(_ string) uint64 {
	return lsp.ssdDiskSize
}
