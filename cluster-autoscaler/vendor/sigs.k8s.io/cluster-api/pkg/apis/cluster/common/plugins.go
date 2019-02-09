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

package common

import (
	"sync"

	"github.com/pkg/errors"
	"k8s.io/klog"
)

var (
	providersMutex sync.Mutex
	providers      = make(map[string]interface{})
)

// RegisterClusterProvisioner registers a ClusterProvisioner by name.  This
// is expected to happen during app startup.
func RegisterClusterProvisioner(name string, provisioner interface{}) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	if _, found := providers[name]; found {
		klog.Fatalf("Cluster provisioner %q was registered twice", name)
	}
	klog.V(1).Infof("Registered cluster provisioner %q", name)
	providers[name] = provisioner
}

func ClusterProvisioner(name string) (interface{}, error) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	provisioner, found := providers[name]
	if !found {
		return nil, errors.Errorf("unable to find provisioner for %s", name)
	}
	return provisioner, nil
}
