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

package apis

import _ "embed" // This is by design.

// The following block embeds the cluster-autoscaler Custom Resource Definition (CRD) YAML
// files statically into the compiled Go binary.
//
// WHY THIS EXISTS:
// Downstream consumers (such as external controllers or testing environments) that import
// this module often need to load and install these CRD schemas onto a target Kubernetes cluster
// (for example, in integration/E2E test suites). To avoid manual copy-pasting of static YAML
// assets from the upstream repository, these definitions are bundled directly into the client package.
var (
	// CapacityBufferCRD holds the embedded YAML bytes for the CapacityBuffer CRD.
	//go:embed config/crd/autoscaling.x-k8s.io_capacitybuffers.yaml
	CapacityBufferCRD []byte

	// CapacityQuotaCRD holds the embedded YAML bytes for the CapacityQuota CRD.
	//go:embed config/crd/autoscaling.x-k8s.io_capacityquotas.yaml
	CapacityQuotaCRD []byte

	// ProvisioningRequestCRD holds the embedded YAML bytes for the ProvisioningRequest CRD.
	//go:embed config/crd/autoscaling.x-k8s.io_provisioningrequests.yaml
	ProvisioningRequestCRD []byte
)
