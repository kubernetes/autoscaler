/*
Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.

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

// Copied from cacheUtil.WaitForCacheSync to use wait.PollUntil() method instead of wait.PollImmediateUntil()
// method in the waitForCacheSync() function

package cache

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	// syncedPollPeriod controls how often you look at the status of your sync funcs
	syncedPollPeriod = 100 * time.Millisecond
)

// WaitForCacheSync waits for caches to populate.  It returns true if it was successful, false
// if the controller should shutdown
// callers should prefer WaitForNamedCacheSync()
func WaitForCacheSync(stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
	err := wait.PollUntil(syncedPollPeriod,
		func() (bool, error) {
			for _, syncFunc := range cacheSyncs {
				if !syncFunc() {
					return false, nil
				}
			}
			return true, nil
		},
		stopCh)
	if err != nil {
		klog.V(2).Infof("stop requested")
		return false
	}

	klog.V(4).Infof("caches populated")
	return true
}
