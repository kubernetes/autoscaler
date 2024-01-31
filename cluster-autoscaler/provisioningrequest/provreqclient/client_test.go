/*
Copyright 2023 The Kubernetes Authors.

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

package provreqclient

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

func TestFetchPodTemplates(t *testing.T) {
	pr1 := ProvisioningRequestWrapperForTesting("namespace", "name-1")
	pr2 := ProvisioningRequestWrapperForTesting("namespace", "name-2")
	mockProvisioningRequests := []*provreqwrapper.ProvisioningRequest{pr1, pr2}

	ctx := context.Background()
	c := NewFakeProvisioningRequestClient(ctx, t, mockProvisioningRequests...)
	got, err := c.FetchPodTemplates(pr1.V1Beta1())
	if err != nil {
		t.Errorf("provisioningRequestClient.ProvisioningRequests() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("provisioningRequestClient.ProvisioningRequests() got: %v, want 1 element", err)
	}
	if diff := cmp.Diff(pr1.PodTemplates(), got); diff != "" {
		t.Errorf("Template mismatch, diff (-want +got):\n%s", diff)
	}
}
