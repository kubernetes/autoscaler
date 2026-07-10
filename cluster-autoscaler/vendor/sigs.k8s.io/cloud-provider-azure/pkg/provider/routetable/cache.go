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

package routetable

import (
	"context"
	"time"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/routetableclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
)

const (
	DefaultCacheTTL = 120 * time.Second
)

func NewCache(
	client routetableclient.Interface,
	resourceGroup string,
	cacheTTL time.Duration,
	disableAPICallCache bool,
) (cache.Resource, error) {
	getter := func(ctx context.Context, key string) (interface{}, error) {
		rt, err := client.Get(ctx, resourceGroup, key)
		found, err := errutils.CheckResourceExistsFromAzcoreError(err)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, nil
		}

		return rt, nil
	}

	if cacheTTL == 0 {
		cacheTTL = DefaultCacheTTL
	}
	return cache.NewTimedCache(cacheTTL, getter, disableAPICallCache)
}
