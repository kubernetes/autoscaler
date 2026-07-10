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

package lockmap

import "sync"

// LockMap used to lock on entries
type LockMap struct {
	sync.Mutex
	mutexMap map[string]*sync.Mutex
}

// NewLockMap returns a new lock map
func NewLockMap() *LockMap {
	return &LockMap{
		mutexMap: make(map[string]*sync.Mutex),
	}
}

// LockEntry acquires a lock associated with the specific entry
func (lm *LockMap) LockEntry(entry string) {
	lm.Lock()
	// check if entry does not exists, then add entry
	mutex, exists := lm.mutexMap[entry]
	if !exists {
		mutex = &sync.Mutex{}
		lm.mutexMap[entry] = mutex
	}
	lm.Unlock()
	mutex.Lock()
}

// UnlockEntry release the lock associated with the specific entry
func (lm *LockMap) UnlockEntry(entry string) {
	lm.Lock()
	defer lm.Unlock()

	mutex, exists := lm.mutexMap[entry]
	if !exists {
		return
	}
	mutex.Unlock()
}
