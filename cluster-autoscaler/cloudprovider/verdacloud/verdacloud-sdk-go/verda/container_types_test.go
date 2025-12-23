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

func TestContainerTypesService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get container types with default currency", func(t *testing.T) {
		ctx := context.Background()
		containerTypes, err := client.ContainerTypes.Get(ctx, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(containerTypes) == 0 {
			t.Error("expected at least one container type")
		}

		// Verify first container type has expected fields
		if len(containerTypes) > 0 {
			ct := containerTypes[0]
			if ct.ID == "" {
				t.Error("expected container type to have an ID")
			}
			if ct.Model == "" {
				t.Error("expected container type to have a Model")
			}
			if ct.Name == "" {
				t.Error("expected container type to have a Name")
			}
			if ct.InstanceType == "" {
				t.Error("expected container type to have an InstanceType")
			}
			if ct.Currency == "" {
				t.Error("expected container type to have a Currency")
			}
			if ct.Manufacturer == "" {
				t.Error("expected container type to have a Manufacturer")
			}
		}
	})

	t.Run("get container types with USD currency", func(t *testing.T) {
		ctx := context.Background()
		containerTypes, err := client.ContainerTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(containerTypes) == 0 {
			t.Error("expected at least one container type")
		}
	})

	t.Run("get container types with EUR currency", func(t *testing.T) {
		ctx := context.Background()
		containerTypes, err := client.ContainerTypes.Get(ctx, "eur")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(containerTypes) == 0 {
			t.Error("expected at least one container type")
		}
	})

	t.Run("verify flexible float for prices", func(t *testing.T) {
		ctx := context.Background()
		containerTypes, err := client.ContainerTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(containerTypes) > 0 {
			ct := containerTypes[0]
			// Verify FlexibleFloat conversion works
			serverlessPrice := ct.ServerlessPrice.Float64()
			if serverlessPrice <= 0 {
				t.Error("expected serverless price to be positive")
			}

			spotPrice := ct.ServerlessSpotPrice.Float64()
			if spotPrice <= 0 {
				t.Error("expected spot price to be positive")
			}
		}
	})
}
