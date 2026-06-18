/*
Copyright The Kubernetes Authors.

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
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/stretchr/testify/assert"
)

// fakeCreateOrUpdatePollingHandler is a runtime.PollingHandler that completes
// successfully and returns a CreateOrUpdate response carrying the given ETag.
type fakeCreateOrUpdatePollingHandler struct {
	etag   *string
	polled bool
}

func (f *fakeCreateOrUpdatePollingHandler) Done() bool { return f.polled }

func (f *fakeCreateOrUpdatePollingHandler) Poll(_ context.Context) (*http.Response, error) {
	f.polled = true
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: http.NoBody}, nil
}

func (f *fakeCreateOrUpdatePollingHandler) Result(_ context.Context, out *armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse) error {
	out.VirtualMachineScaleSet = armcompute.VirtualMachineScaleSet{Etag: f.etag}
	return nil
}

func newTestCreateOrUpdatePoller(t *testing.T, etag *string) *runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse] {
	resp := &http.Response{StatusCode: http.StatusAccepted, Header: http.Header{}, Body: http.NoBody}
	pl := runtime.NewPipeline("test", "v0.0.0", runtime.PipelineOptions{}, nil)
	poller, err := runtime.NewPoller(resp, pl, &runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse]{
		Handler: &fakeCreateOrUpdatePollingHandler{etag: etag},
	})
	assert.NoError(t, err)
	return poller
}
