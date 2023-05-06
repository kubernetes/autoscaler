/*
Copyright 2023 The Kubernetes Authors.

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

package scaledowncandidates

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
)

// Observer is an observer of scale down candidates
type Observer interface {
	// UpdateScaleDownCandidates updates scale down candidates.
	UpdateScaleDownCandidates([]*apiv1.Node, time.Time)
}

// ObserversList is a slice of observers of scale down candidates
type ObserversList struct {
	observers []Observer
}

// Register adds new observer to the list.
func (l *ObserversList) Register(o Observer) {
	l.observers = append(l.observers, o)
}

// Update updates scale down candidates for each observer.
func (l *ObserversList) Update(nodes []*apiv1.Node, now time.Time) {
	for _, observer := range l.observers {
		observer.UpdateScaleDownCandidates(nodes, now)
	}
}

// NewObserversList return empty list of observers.
func NewObserversList() *ObserversList {
	return &ObserversList{}
}
