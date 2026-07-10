/*
Copyright 2024 The Kubernetes Authors.

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

package zone

// Generate mocks for the repository interface
//go:generate mockgen -destination=./mock_repo.go -package=zone -copyright_file ../../../hack/boilerplate/boilerplate.generatego.txt -source=repo.go Repository

import (
	"context"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/providerclient"
	fnutil "sigs.k8s.io/cloud-provider-azure/pkg/util/collectionutil"
)

type Repository interface {
	ListZones(ctx context.Context) (map[string][]string, error)
}

type repo struct {
	client providerclient.Interface
}

func NewRepo(client providerclient.Interface) (Repository, error) {
	return &repo{client: client}, nil
}

func (r *repo) ListZones(ctx context.Context) (map[string][]string, error) {
	zones, err := r.client.GetVirtualMachineSupportedZones(ctx)
	if err != nil {
		return nil, err
	}

	var rv = make(map[string][]string, len(zones))
	for region, z := range zones {
		rv[region] = fnutil.Map(func(s *string) string { return *s }, z)
	}

	return rv, nil
}
