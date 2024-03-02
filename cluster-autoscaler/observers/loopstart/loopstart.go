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

package loopstart

// Observer interface is used to store object that needed to be refreshed in each CA loop.
// It returns error and a bool value whether the loop should be skipped.
type Observer interface {
	Refresh()
}

// ObserversList interface is used to store objects that needed to be refreshed in each CA loop.
type ObserversList struct {
	observers []Observer
}

// Refresh refresh observers each CA loop.
func (l *ObserversList) Refresh() {
	for _, observer := range l.observers {
		observer.Refresh()
	}
}

// NewObserversList return new ObserversList.
func NewObserversList(observers []Observer) *ObserversList {
	return &ObserversList{observers}
}
