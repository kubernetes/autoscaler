/*
Copyright 2016 The Kubernetes Authors.

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

package linode

import (
	"context"
	"net/http"
	"time"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	klog "k8s.io/klog/v2"
)

const (
	userAgent = "kubernetes/cluster-autoscaler/" + version.ClusterAutoscalerVersion
)

// linodeAPIClient is the interface used to call linode API
type linodeAPIClient interface {
	ListLKEClusterPools(ctx context.Context, clusterID int, opts *linodego.ListOptions) ([]linodego.LKEClusterPool, error)
	UpdateLKEClusterPool(ctx context.Context, clusterID, id int, updateOpts linodego.LKEClusterPoolUpdateOptions) (*linodego.LKEClusterPool, error)
	DeleteLKEClusterPoolNode(ctx context.Context, poolID int, id string) error
}

// buildLinodeAPIClient returns the struct ready to perform calls to linode API
func buildLinodeAPIClient(baseURL, apiVersion, token string) linodeAPIClient {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	client := linodego.NewClient(oauth2Client)
	client.SetUserAgent(userAgent)
	if baseURL != "" {
		klog.V(4).Infof("using baseURL %q for Linode client", baseURL)
		client.SetBaseURL(baseURL)
	}
	if apiVersion != "" {
		klog.V(4).Infof("using apiVersion %q for Linode client", apiVersion)
		client.SetAPIVersion(apiVersion)
	}
	return &client
}
