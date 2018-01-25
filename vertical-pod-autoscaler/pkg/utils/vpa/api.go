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

package api

import (
	"encoding/json"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	types "k8s.io/apimachinery/pkg/types"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func patchVpa(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, patches []patchRecord) (result *vpa_types.VerticalPodAutoscaler, err error) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		glog.Errorf("Cannot marshal VPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	return vpaClient.Patch(vpaName, types.JSONPatchType, bytes)
}

// InitVpaStatus inserts into VPA CRD object an empty VerticalPodAutoscalerStatus object, so later it can be filled with data.
func InitVpaStatus(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string) (result *vpa_types.VerticalPodAutoscaler, err error) {
	return patchVpa(vpaClient, vpaName, []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: vpa_types.VerticalPodAutoscalerStatus{},
	},
	})
}

// UpdateVpaRecommendation updates VPA's status/recommendation object and status/lastUpdateTime.
func UpdateVpaRecommendation(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, recommendation vpa_types.RecommendedPodResources) (result *vpa_types.VerticalPodAutoscaler, err error) {
	patches := []patchRecord{
		{
			Op:    "add",
			Path:  "/status/lastUpdateTime",
			Value: metav1.Time{time.Now()},
		},
		{
			Op:    "add",
			Path:  "/status/recommendation",
			Value: recommendation,
		},
	}
	return patchVpa(vpaClient, vpaName, patches)

}

// NewAllVpasLister returns VerticalPodAutoscalerLister configured to fetch all VPA objects.
func NewAllVpasLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.PocV1alpha1().RESTClient(), "verticalpodautoscalers", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := vpa_lister.NewVerticalPodAutoscalerLister(store)
	vpaReflector := cache.NewReflector(vpaListWatch, &vpa_types.VerticalPodAutoscaler{}, store, time.Hour)
	go vpaReflector.Run(stopChannel)
	return vpaLister
}
