/*
Copyright 2022 The Kubernetes Authors.

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

package pdb

import (
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// RemainingPdbTracker is responsible for tracking the remaining PDBs
type RemainingPdbTracker interface {
	// SetPdbs sets PDBs of the remaining PDB tracker.
	SetPdbs(pdbs []*policyv1.PodDisruptionBudget) error
	// GetPdbs returns the current remaining PDBs.
	GetPdbs() []*policyv1.PodDisruptionBudget

	// CanRemovePods checks if the set of pods can be removed.
	// inParallel indicates if the pods can be removed in parallel. If it is false
	// then evicting pods could fail due to drain timeout.
	CanRemovePods(pods []*apiv1.Pod) (canRemove, inParallel bool, blockingPod *drain.BlockingPod)
	// RemovePods updates the remaining PDBs after pod removal.
	RemovePods(pods []*apiv1.Pod)

	// Clear resets the remaining PDB tracker to empty state.
	Clear()
}
