// Copyright 2020 Brightbox Systems Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cached

import (
	"net/http"
	"time"

	cache "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/go-cache"
	brightbox "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox"

	klog "k8s.io/klog/v2"
)

const (
	expirationTime = 5 * time.Second
	purgeTime      = 30 * time.Second
)

// Client is a cached brightbox Client
type Client struct {
	clientCache *cache.Cache
	brightbox.Client
}

// NewClient creates and returns a cached Client
func NewClient(url string, account string, httpClient *http.Client) (*Client, error) {
	cl, err := brightbox.NewClient(url, account, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{
		clientCache: cache.New(expirationTime, purgeTime),
		Client:      *cl,
	}, err
}

// Server fetches a server by id
func (c *Client) Server(identifier string) (*brightbox.Server, error) {
	if cachedServer, found := c.clientCache.Get(identifier); found {
		klog.V(4).Infof("Cache hit %q", identifier)
		return cachedServer.(*brightbox.Server), nil
	}
	server, err := c.Client.Server(identifier)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Cacheing %q", identifier)
	c.clientCache.Set(identifier, server, cache.DefaultExpiration)
	return server, nil
}

// ServerGroup fetches a server group by id
func (c *Client) ServerGroup(identifier string) (*brightbox.ServerGroup, error) {
	if cachedServerGroup, found := c.clientCache.Get(identifier); found {
		klog.V(4).Infof("Cache hit %q", identifier)
		return cachedServerGroup.(*brightbox.ServerGroup), nil
	}
	serverGroup, err := c.Client.ServerGroup(identifier)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Cacheing %q", identifier)
	c.clientCache.Set(identifier, serverGroup, cache.DefaultExpiration)
	return serverGroup, nil
}

// ConfigMap fetches a config map by id
func (c *Client) ConfigMap(identifier string) (*brightbox.ConfigMap, error) {
	if cachedConfigMap, found := c.clientCache.Get(identifier); found {
		klog.V(4).Infof("Cache hit %q", identifier)
		return cachedConfigMap.(*brightbox.ConfigMap), nil
	}
	configMap, err := c.Client.ConfigMap(identifier)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Cacheing %q", identifier)
	c.clientCache.Set(identifier, configMap, cache.DefaultExpiration)
	return configMap, nil
}

// DestroyServer removes a server by id
func (c *Client) DestroyServer(identifier string) error {
	err := c.Client.DestroyServer(identifier)
	if err == nil {
		c.clientCache.Delete(identifier)
	}
	return err
}

// DestroyServerGroup removes a server group by id
func (c *Client) DestroyServerGroup(identifier string) error {
	err := c.Client.DestroyServerGroup(identifier)
	if err == nil {
		c.clientCache.Delete(identifier)
	}
	return err
}
