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
	"context"
	"errors"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	klog "k8s.io/klog/v2"
)

const (
	// StatusConfigMapName is the name of ConfigMap with status.
	StatusConfigMapName = "cluster-autoscaler-status"
	// ConfigMapLastUpdatedKey is the name of annotation informing about status ConfigMap last update.
	ConfigMapLastUpdatedKey = "cluster-autoscaler.kubernetes.io/last-updated"
	// ConfigMapLastUpdateFormat it the timestamp format used for last update annotation in status ConfigMap
	ConfigMapLastUpdateFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
)

// LogEventRecorder records events on some top-level object, to give user (without access to logs) a view of most important CA actions.
type LogEventRecorder struct {
	recorder     record.EventRecorder
	statusObject runtime.Object
	active       bool
}

// Event records an event on underlying object. This does nothing if the underlying object is not set.
func (ler *LogEventRecorder) Event(eventtype, reason, message string) {
	if ler.active && ler.statusObject != nil {
		ler.recorder.Event(ler.statusObject, eventtype, reason, message)
	}
}

// Eventf records an event on underlying object. This does nothing if the underlying object is not set.
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
// ConfigMap if it doesn't exist. If logRecorder is passed and configmap update is successful
// logRecorder's internal reference will be updated.
func WriteStatusConfigMap(kubeClient kube_client.Interface, namespace string, msg string, logRecorder *LogEventRecorder) (*apiv1.ConfigMap, error) {
	statusUpdateTime := time.Now().Format(ConfigMapLastUpdateFormat)
	statusMsg := fmt.Sprintf("Cluster-autoscaler status at %s:\n%v", statusUpdateTime, msg)
	var configMap *apiv1.ConfigMap
	var getStatusError, writeStatusError error
	var errMsg string
	maps := kubeClient.CoreV1().ConfigMaps(namespace)
	configMap, getStatusError = maps.Get(context.TODO(), StatusConfigMapName, metav1.GetOptions{})
	if getStatusError == nil {
		configMap.Data["status"] = statusMsg
		if configMap.ObjectMeta.Annotations == nil {
			configMap.ObjectMeta.Annotations = make(map[string]string)
		}
		configMap.ObjectMeta.Annotations[ConfigMapLastUpdatedKey] = statusUpdateTime
		configMap, writeStatusError = maps.Update(context.TODO(), configMap, metav1.UpdateOptions{})
	} else if kube_errors.IsNotFound(getStatusError) {
		configMap = &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      StatusConfigMapName,
				Annotations: map[string]string{
					ConfigMapLastUpdatedKey: statusUpdateTime,
				},
			},
			Data: map[string]string{
				"status": statusMsg,
			},
		}
		configMap, writeStatusError = maps.Create(context.TODO(), configMap, metav1.CreateOptions{})
	} else {
		errMsg = fmt.Sprintf("Failed to retrieve status configmap for update: %v", getStatusError)
	}
	if writeStatusError != nil {
		errMsg = fmt.Sprintf("Failed to write status configmap: %v", writeStatusError)
	}
	if errMsg != "" {
		klog.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	klog.V(8).Infof("Successfully wrote status configmap with body \"%v\"", statusMsg)
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
	err := maps.Delete(context.TODO(), StatusConfigMapName, metav1.DeleteOptions{})
	if err != nil {
		klog.Error("Failed to delete status configmap")
	}
	return err
}
