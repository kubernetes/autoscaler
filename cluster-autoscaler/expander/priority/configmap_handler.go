/*
Copyright 2016 The Kubernetes Authors.

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

package priority

import (
	"errors"
	"fmt"

	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	api "k8s.io/kubernetes/pkg/apis/core"
)

const (
	// PriorityConfigMapName defines a name of the ConfigMap used to store priority expander configuration
	PriorityConfigMapName = "cluster-autoscaler-priority-expander"
	// ConfigMapKey defines the key used in the ConfigMap to configure priorities
	ConfigMapKey = "priorities"
)

// InitPriorityConfigMap initializes ConfigMap with priority expander configurations. It checks if the map exists
// and has the correct top level key. If it doesn't, it returns error or Exits. If the value is found,
// the current value is fetched and a Watcher is started to watch for changes. It returns the current value of
// the config map, the channel with value updates and an error.
func InitPriorityConfigMap(maps v1.ConfigMapInterface, namespace string) (string, <-chan watch.Event, error) {
	errMsg := ""
	priorities := ""

	configMap, getStatusError := maps.Get(PriorityConfigMapName, metav1.GetOptions{})
	if getStatusError == nil {
		priorities = configMap.Data[ConfigMapKey]
	} else if kube_errors.IsNotFound(getStatusError) {
		errMsg = fmt.Sprintf("Priority expander config map %s/%s not found. You have to create it before starting cluster-autoscaler "+
			"with priority expander.", namespace, PriorityConfigMapName)
	} else {
		errMsg = fmt.Sprintf("Failed to retrieve priority expander configmap %s/%s: %v", namespace, PriorityConfigMapName,
			getStatusError)
	}
	if errMsg != "" {
		return "", nil, errors.New(errMsg)
	}

	watcher, err := maps.Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, PriorityConfigMapName).String(),
		Watch:         true,
	})
	if err != nil {
		errMsg = fmt.Sprintf("Error when starting a watcher for changes of the priority expander configmap %s/%s: %v",
			namespace, PriorityConfigMapName, err)
		klog.Errorf(errMsg)
		return "", nil, errors.New(errMsg)
	}

	return priorities, watcher.ResultChan(), nil
}
