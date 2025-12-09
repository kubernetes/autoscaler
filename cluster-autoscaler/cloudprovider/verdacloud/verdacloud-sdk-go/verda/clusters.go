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
	"fmt"
	"net/url"
)

// ClusterService handles cluster-related API operations.
type ClusterService struct {
	client *Client
}

// Get retrieves all clusters.
func (s *ClusterService) Get(ctx context.Context) ([]Cluster, error) {
	clusters, _, err := getRequest[[]Cluster](ctx, s.client, "/clusters")
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetByID retrieves a cluster by ID.
func (s *ClusterService) GetByID(ctx context.Context, id string) (*Cluster, error) {
	path := fmt.Sprintf("/clusters/%s", id)

	cluster, _, err := getRequest[Cluster](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// Create creates a new cluster.
func (s *ClusterService) Create(ctx context.Context, req CreateClusterRequest) (*CreateClusterResponse, error) {
	if req.LocationCode == "" {
		req.LocationCode = LocationFIN01
	}

	response, _, err := postRequest[CreateClusterResponse](ctx, s.client, "/clusters", req)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Action performs cluster lifecycle operations - currently only "discontinue" is supported
func (s *ClusterService) Action(ctx context.Context, idList any, action string) error {
	req := ClusterActionRequest{
		IDList: idList,
		Action: action,
	}

	_, _, err := putRequest[any](ctx, s.client, "/clusters", req)
	return err
}

// Discontinue discontinues a cluster.
func (s *ClusterService) Discontinue(ctx context.Context, idList any) error {
	return s.Action(ctx, idList, ClusterActionDiscontinue)
}

// GetClusterTypes retrieves available cluster types.
func (s *ClusterService) GetClusterTypes(ctx context.Context, currency string) ([]ClusterType, error) {
	path := "/cluster-types"

	if currency != "" {
		params := url.Values{}
		params.Set("currency", currency)
		path += "?" + params.Encode()
	}

	clusterTypes, _, err := getRequest[[]ClusterType](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return clusterTypes, nil
}

// GetAvailabilities retrieves cluster availabilities.
func (s *ClusterService) GetAvailabilities(ctx context.Context, locationCode string) ([]ClusterAvailability, error) {
	path := "/cluster-availability"

	if locationCode != "" {
		params := url.Values{}
		params.Set("location_code", locationCode)
		path += "?" + params.Encode()
	}

	availabilities, _, err := getRequest[[]ClusterAvailability](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return availabilities, nil
}

// CheckClusterTypeAvailability checks if a cluster type is available.
func (s *ClusterService) CheckClusterTypeAvailability(ctx context.Context, clusterType string, locationCode string) (bool, error) {
	path := fmt.Sprintf("/cluster-availability/%s", clusterType)

	if locationCode != "" {
		params := url.Values{}
		params.Set("location_code", locationCode)
		path += "?" + params.Encode()
	}

	available, _, err := getRequest[bool](ctx, s.client, path)
	if err != nil {
		return false, err
	}

	return available, nil
}

// GetImages retrieves available cluster images.
func (s *ClusterService) GetImages(ctx context.Context) ([]ClusterImage, error) {
	images, _, err := getRequest[[]ClusterImage](ctx, s.client, "/images/cluster")
	if err != nil {
		return nil, err
	}

	return images, nil
}
