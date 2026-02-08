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
)

// LongTermService handles long-term contract API operations.
type LongTermService struct {
	client *Client
}

// GetInstancePeriods retrieves available long-term periods for instances.
func (s *LongTermService) GetInstancePeriods(ctx context.Context) ([]LongTermPeriod, error) {
	path := "/long-term/periods/instances"

	periods, _, err := getRequest[[]LongTermPeriod](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return periods, nil
}

func (s *LongTermService) GetPeriods(ctx context.Context) ([]LongTermPeriod, error) {
	path := "/long-term/periods"

	periods, _, err := getRequest[[]LongTermPeriod](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return periods, nil
}

func (s *LongTermService) GetClusterPeriods(ctx context.Context) ([]LongTermPeriod, error) {
	path := "/long-term/periods/clusters"

	periods, _, err := getRequest[[]LongTermPeriod](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return periods, nil
}
