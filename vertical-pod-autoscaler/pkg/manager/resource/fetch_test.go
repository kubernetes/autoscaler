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

package resource

import (
	"log"
	"testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestGetControlAndUpdateMode(t *testing.T) {
	labels := map[string]string{
		resourcesControlKey: `open`,
		podUpdateModeKey:    string(vpa_types.UpdateModeAuto),
	}
	fetch := &fetcherObject{defaultUpdateMode: vpa_types.UpdateModeInitial}
	openVal, mode, ok := fetch.getControlAndUpdateMode(labels)
	if openVal != `open` || mode != vpa_types.UpdateModeAuto || !ok {
		log.Println(openVal, mode, ok)
		t.FailNow()
	}

	delete(labels, podUpdateModeKey)
	openVal, mode, ok = fetch.getControlAndUpdateMode(labels)
	if openVal != `open` || mode != vpa_types.UpdateModeInitial || !ok {
		log.Println(openVal, mode, ok)
		t.FailNow()
	}

	delete(labels, resourcesControlKey)
	openVal, mode, ok = fetch.getControlAndUpdateMode(labels)
	if ok {
		log.Println(openVal, mode, ok)
		t.FailNow()
	}
}
