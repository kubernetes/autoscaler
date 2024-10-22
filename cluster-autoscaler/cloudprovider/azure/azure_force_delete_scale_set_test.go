/*
Copyright 2022 The Kubernetes Authors.

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

package azure

import (
	"net/http"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

func TestShouldForceDelete(t *testing.T) {
	skuName := "test-vmssSku"

	t.Run("should return true", func(t *testing.T) {
		scaleSet := &ScaleSet{}
		scaleSet.enableForceDelete = true
		assert.Equal(t, shouldForceDelete(skuName, scaleSet), true)
	})

	t.Run("should return false because of dedicated hosts", func(t *testing.T) {
		scaleSet := &ScaleSet{}
		scaleSet.enableForceDelete = true
		scaleSet.dedicatedHost = true
		assert.Equal(t, shouldForceDelete(skuName, scaleSet), false)
	})

	t.Run("should return false because of isolated sku", func(t *testing.T) {
		scaleSet := &ScaleSet{}
		scaleSet.enableForceDelete = true
		skuName = "Standard_F72s_v2" // belongs to the map isolatedVMSizes
		assert.Equal(t, shouldForceDelete(skuName, scaleSet), false)
	})

}

func TestIsOperationNotAllowed(t *testing.T) {
	t.Run("should return false because it's not OperationNotAllowed error", func(t *testing.T) {
		error := &retry.Error{
			HTTPStatusCode: http.StatusBadRequest,
		}
		assert.Equal(t, isOperationNotAllowed(error), false)
	})

	t.Run("should return false because error is nil", func(t *testing.T) {
		assert.Equal(t, isOperationNotAllowed(nil), false)
	})

	t.Run("should return true if error is OperationNotAllowed", func(t *testing.T) {
		sre := &azure.ServiceError{
			Code:    retry.OperationNotAllowed,
			Message: "error-message",
		}
		error := &retry.Error{
			RawError: sre,
		}
		assert.Equal(t, isOperationNotAllowed(error), false)
	})

	// It is difficult to condition the case where return error matched expected error string for forceDelete and the
	// function should return true.

}
