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
