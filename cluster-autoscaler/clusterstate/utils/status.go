/*
Copyright 2017 The Kubernetes Authors.

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

package utils

import (
	"errors"
	"fmt"
	"time"

	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"

	"github.com/golang/glog"
)

const (
	// StatusConfigMapName is the name of ConfigMap with status.
	StatusConfigMapName = "cluster-autoscaler-status"
	// ConfigMapLastUpdatedKey is the name of annotation informing about status ConfigMap last update.
	ConfigMapLastUpdatedKey = "cluster-autoscaler.kubernetes.io/last-updated"
)

// LogEventRecorder records events on some top-level object, to give user (without access to logs) a view of most important CA actions.
type LogEventRecorder struct {
	recorder     record.EventRecorder
	statusObject runtime.Object
	active       bool
}

// Event record an event on underlying object. This does nothing if the underlying object is not set.
func (ler *LogEventRecorder) Event(eventtype, reason, message string) {
	if ler.active && ler.statusObject != nil {
		ler.recorder.Event(ler.statusObject, eventtype, reason, message)
	}
}

// Eventf record an event on underlying object. This does nothing if the underlying object is not set.
func (ler *LogEventRecorder) Eventf(eventtype, reason, message string, args ...interface{}) {
	if ler.active && ler.statusObject != nil {
		ler.recorder.Eventf(ler.statusObject, eventtype, reason, message, args...)
	}
}

// NewStatusMapRecorder creates a LogEventRecorder creating events on status configmap.
// If the configmap doesn't exist it will be created (with 'Initializing' status).
// If active == false the map will not be created and no events will be recorded.
func NewStatusMapRecorder(kubeClient kube_client.Interface, namespace string, recorder record.EventRecorder, active bool) (*LogEventRecorder, error) {
	var mapObj runtime.Object
	var err error
	if active {
		mapObj, err = WriteStatusConfigMap(kubeClient, namespace, "Initializing", nil)
		if err != nil {
			return nil, errors.New("Failed to init status ConfigMap")
		}
	}
	return &LogEventRecorder{
		recorder:     recorder,
		statusObject: mapObj,
		active:       active,
	}, nil
}

// WriteStatusConfigMap writes updates status ConfigMap with a given message or creates a new
// ConfigMap if it doesn't exist. If logRecorder is passed and configmap update is successfull
// logRecorder's internal reference will be updated.
func WriteStatusConfigMap(kubeClient kube_client.Interface, namespace string, msg string, logRecorder *LogEventRecorder) (*apiv1.ConfigMap, error) {
	statusUpdateTime := time.Now()
	statusMsg := fmt.Sprintf("Cluster-autoscaler status at %v:\n%v", statusUpdateTime, msg)
	var configMap *apiv1.ConfigMap
	var getStatusError, writeStatusError error
	var errMsg string
	maps := kubeClient.CoreV1().ConfigMaps(namespace)
	configMap, getStatusError = maps.Get(StatusConfigMapName, metav1.GetOptions{})
	if getStatusError == nil {
		configMap.Data["status"] = statusMsg
		if configMap.ObjectMeta.Annotations == nil {
			configMap.ObjectMeta.Annotations = make(map[string]string)
		}
		configMap.ObjectMeta.Annotations[ConfigMapLastUpdatedKey] = fmt.Sprintf("%v", statusUpdateTime)
		configMap, writeStatusError = maps.Update(configMap)
	} else if kube_errors.IsNotFound(getStatusError) {
		configMap = &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      StatusConfigMapName,
				Annotations: map[string]string{
					ConfigMapLastUpdatedKey: fmt.Sprintf("%v", statusUpdateTime),
				},
			},
			Data: map[string]string{
				"status": statusMsg,
			},
		}
		configMap, writeStatusError = maps.Create(configMap)
	} else {
		errMsg = fmt.Sprintf("Failed to retrieve status configmap for update: %v", getStatusError)
	}
	if writeStatusError != nil {
		errMsg = fmt.Sprintf("Failed to write status configmap: %v", writeStatusError)
	}
	if errMsg != "" {
		glog.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	glog.V(8).Infof("Succesfully wrote status configmap with body \"%v\"", statusMsg)
	// Having this as a side-effect is somewhat ugly
	// But it makes error handling easier, as we get a free retry each loop
	if logRecorder != nil {
		logRecorder.statusObject = configMap
	}
	return configMap, nil
}

// DeleteStatusConfigMap deletes status configmap
func DeleteStatusConfigMap(kubeClient kube_client.Interface, namespace string) error {
	maps := kubeClient.CoreV1().ConfigMaps(namespace)
	err := maps.Delete(StatusConfigMapName, &metav1.DeleteOptions{})
	if err != nil {
		glog.Error("Failed to delete status configmap")
	}
	return err
}
