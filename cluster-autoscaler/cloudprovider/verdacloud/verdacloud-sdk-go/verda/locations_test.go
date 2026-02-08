/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"context"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestLocationsService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all locations", func(t *testing.T) {
		ctx := context.Background()
		locations, err := client.Locations.Get(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(locations) == 0 {
			t.Error("expected at least one location")
		}

		location := locations[0]
		if location.Code != LocationFIN01 {
			t.Errorf("expected location code '%s', got '%s'", LocationFIN01, location.Code)
		}

		if location.Name == "" {
			t.Error("expected location name to not be empty")
		}

		if location.CountryCode == "" {
			t.Error("expected country code to not be empty")
		}
	})
}
