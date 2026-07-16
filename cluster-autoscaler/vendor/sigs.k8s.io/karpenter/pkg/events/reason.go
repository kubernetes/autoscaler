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

package events

// Reasons of events controllers emit
const (
	// disruption
	DisruptionBlocked          = "DisruptionBlocked"
	DisruptionLaunching        = "DisruptionLaunching"
	DisruptionTerminating      = "DisruptionTerminating"
	DisruptionWaitingReadiness = "DisruptionWaitingReadiness"
	Unconsolidatable           = "Unconsolidatable"
	ConsolidationCandidate     = "ConsolidationCandidate"
	ConsolidationRejected      = "ConsolidationRejected"
	ConsolidationApproved      = "ConsolidationApproved"

	// provisioning/scheduling
	FailedScheduling          = "FailedScheduling"
	NoCompatibleInstanceTypes = "NoCompatibleInstanceTypes"
	Nominated                 = "Nominated"

	// node/health
	NodeRepairBlocked = "NodeRepairBlocked"

	// node/termination/terminator
	Disrupted                      = "Disrupted"
	Evicted                        = "Evicted"
	FailedDraining                 = "FailedDraining"
	TerminationGracePeriodExpiring = "TerminationGracePeriodExpiring"
	TerminationFailed              = "FailedTermination"

	// nodeclaim/consistency
	FailedConsistencyCheck = "FailedConsistencyCheck"

	// nodeclaim/lifecycle
	InsufficientCapacityError = "InsufficientCapacityError"
	UnregisteredTaintMissing  = "UnregisteredTaintMissing"
	NodeClassNotReady         = "NodeClassNotReady"
)
