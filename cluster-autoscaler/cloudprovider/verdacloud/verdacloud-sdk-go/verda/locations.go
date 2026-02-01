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

import "context"

// LocationService handles location-related API operations.
type LocationService struct {
	client *Client
}

// Get retrieves all available locations.
func (s *LocationService) Get(ctx context.Context) ([]Location, error) {
	locations, _, err := getRequest[[]Location](ctx, s.client, "/locations")
	if err != nil {
		return nil, err
	}
	return locations, nil
}
