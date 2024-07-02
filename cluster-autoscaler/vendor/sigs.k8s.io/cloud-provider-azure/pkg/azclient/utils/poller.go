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

package utils

import (
	"context"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

func NewPollerWrapper[ResponseType interface{}](poller *runtime.Poller[ResponseType], err error) *PollerWrapper[ResponseType] {
	return &PollerWrapper[ResponseType]{
		poller: poller,
		err:    err,
	}
}

type PollerWrapper[ResponseType interface{}] struct {
	poller *runtime.Poller[ResponseType]
	err    error
}

// Poller is the poller to be used for polling.
// assume that the poller will ends
func (handler *PollerWrapper[ResponseType]) WaitforPollerResp(ctx context.Context) (result *ResponseType, err error) {
	if handler.err != nil {
		return nil, handler.err
	}
	if handler.poller == nil {
		return nil, errors.New("poller is nil")
	}
	resp, err := handler.poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: time.Second * 15,
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
