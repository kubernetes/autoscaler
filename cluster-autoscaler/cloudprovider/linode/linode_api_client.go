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

	"golang.org/x/oauth2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
	"k8s.io/autoscaler/cluster-autoscaler/version"
)

const (
	userAgent = "kubernetes/cluster-autoscaler/" + version.ClusterAutoscalerVersion
)

// linodeAPIClient is the interface used to call linode API
type linodeAPIClient interface {
	ListLKEClusterPools(ctx context.Context, clusterID int, opts *linodego.ListOptions) ([]linodego.LKEClusterPool, error)
	CreateLKEClusterPool(ctx context.Context, clusterID int, createOpts linodego.LKEClusterPoolCreateOptions) (*linodego.LKEClusterPool, error)
	DeleteLKEClusterPool(ctx context.Context, clusterID int, id int) error
}

// buildLinodeAPIClient returns the struct ready to perform calls to linode API
func buildLinodeAPIClient(linodeToken string) linodeAPIClient {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: linodeToken})
	oauth2Client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	client := linodego.NewClient(oauth2Client)
	client.SetUserAgent(userAgent)
	return &client
}
