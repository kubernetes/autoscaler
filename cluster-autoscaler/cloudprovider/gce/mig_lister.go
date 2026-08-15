/*
Copyright 2021 The Kubernetes Authors.

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

package gce

// MigLister is a mig listing interface
type MigLister interface {
	// GetMigs returns the list of migs
	GetMigs() []Mig
	// HandleMigIssue handles an issue with a given mig
	HandleMigIssue(migRef GceRef, err error)
}

type migLister struct {
	cache *GceCache
}

// NewMigLister returns an instance of migLister
func NewMigLister(cache *GceCache) *migLister {
	return &migLister{
		cache: cache,
	}
}

// GetMigs returns the list of migs
func (l *migLister) GetMigs() []Mig {
	return l.cache.GetMigs()
}

// HandleMigIssue handles an issue with a given mig
func (l *migLister) HandleMigIssue(_ GceRef, _ error) {
}
