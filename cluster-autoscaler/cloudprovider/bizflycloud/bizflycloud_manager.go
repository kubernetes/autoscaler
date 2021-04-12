/*
Copyright 2021 The Kubernetes Authors.

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

package bizflycloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud/gobizfly"
	klog "k8s.io/klog/v2"
)

const (
	// ProviderName specifies the name for the Bizfly provider
	ProviderName  string = "bizflycloud"
	defaultRegion string = "HN"
	authPassword  string = "password"
	authAppCred   string = "application_credential"
	defaultApiUrl string = "https://manage.bizflycloud.vn"

	bizflyCloudAuthMethod      string = "BIZFLYCLOUD_AUTH_METHOD"
	bizflyCloudEmailEnvName    string = "BIZFLYCLOUD_EMAIL"
	bizflyCloudPasswordEnvName string = "BIZFLYCLOUD_PASSWORD"
	bizflyCloudRegionEnvName   string = "BIZFLYCLOUD_REGION"
	bizflyCloudAppCredID       string = "BIZFLYCLOUD_APP_CREDENTIAL_ID"
	bizflyCloudAppCredSecret   string = "BIZFLYCLOUD_APP_CREDENTIAL_SECRET"
	bizflyCloudApiUrl          string = "BIZFLYCLOUD_API_URL"
	bizflyCloudTenantID        string = "BIZFLYCLOUD_TENANT_ID"
	clusterName                string = "CLUSTER_NAME"
)

type nodeGroupClient interface {
	// Get lists all the cluster information in a Kubernetes cluster to lists all the worker pools .
	Get(ctx context.Context, id string) (*gobizfly.FullCluster, error)

	// GetClusterWorkerPool get the details of an existing worker pool.
	GetClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) (*gobizfly.WorkerPoolWithNodes, error)

	// UpdateClusterWorkerPool updates the details of an existing worker pool.
	UpdateClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string, uwp *gobizfly.UpdateWorkerPoolRequest) error

	// DeleteClusterWorkerPoolNode deletes a specific node in a worker pool.
	DeleteClusterWorkerPoolNode(ctx context.Context, clusterUID string, PoolID string, NodeID string) error
}

// Manager handles Bizflycloud communication and data caching of
// node groups (worker pools in BKE)
type Manager struct {
	client     nodeGroupClient
	clusterID  string
	nodeGroups []*NodeGroup
}

// Config is the configuration of the Bizflycloud cloud provider (just for test)
type Config struct {
	ClusterID string `json:"cluster_id"`
	Token     string `json:"token"`
	URL       string `json:"url"`
}

func newManager(configReader io.Reader) (*Manager, error) {
	//newManager will authenticate directly with BKE to build manager
	authMethod := os.Getenv(bizflyCloudAuthMethod)
	username := os.Getenv(bizflyCloudEmailEnvName)
	password := os.Getenv(bizflyCloudPasswordEnvName)
	region := os.Getenv(bizflyCloudRegionEnvName)
	appCredId := os.Getenv(bizflyCloudAppCredID)
	appCredSecret := os.Getenv(bizflyCloudAppCredSecret)
	apiUrl := os.Getenv(bizflyCloudApiUrl)
	tenantId := os.Getenv(bizflyCloudTenantID)
	clusterID := os.Getenv(clusterName)

	switch authMethod {
	case authPassword:
		{
			if username == "" {
				return nil, errors.New("you have to provide username variable")
			}
			if password == "" {
				return nil, errors.New("you have to provide password variable")
			}
		}
	case authAppCred:
		{
			if appCredId == "" {
				return nil, errors.New("you have to provide application credential ID")
			}
			if appCredSecret == "" {
				return nil, errors.New("you have to provide application credential secret")
			}
		}
	}

	if region == "" {
		region = defaultRegion
	}

	if apiUrl == "" {
		apiUrl = defaultApiUrl
	}

	bizflyClient, err := gobizfly.NewClient(gobizfly.WithTenantName(username), gobizfly.WithAPIUrl(apiUrl), gobizfly.WithTenantID(tenantId), gobizfly.WithRegionName(region))
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Bizflycloud client: %s", err)
	}

	token, err := bizflyClient.Token.Create(
		context.Background(),
		&gobizfly.TokenCreateRequest{
			AuthMethod:    authMethod,
			Username:      username,
			Password:      password,
			AppCredID:     appCredId,
			AppCredSecret: appCredSecret})

	if err != nil {
		return nil, fmt.Errorf("cannot create token: %w", err)
	}

	bizflyClient.SetKeystoneToken(token.KeystoneToken)
	m := &Manager{
		client:     bizflyClient.KubernetesEngine,
		clusterID:  clusterID,
		nodeGroups: make([]*NodeGroup, 0),
	}
	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	ctx := context.Background()
	nodePools, err := m.client.Get(ctx, m.clusterID)
	if err != nil {
		return err
	}
	var group []*NodeGroup
	for _, nodePool := range nodePools.WorkerPools {
		if !nodePool.EnableAutoScaling {
			continue
		}
		poolNode, err := m.client.GetClusterWorkerPool(ctx, m.clusterID, nodePool.UID)
		if err != nil {
			return err
		}
		klog.V(4).Infof("adding worker pool: %q name: %s min: %d max: %d desire: %d",
			nodePool.UID, nodePool.Name, nodePool.MinSize, nodePool.MaxSize, nodePool.DesiredSize)

		group = append(group, &NodeGroup{
			id:        nodePool.UID,
			clusterID: m.clusterID,
			client:    m.client,
			nodePool:  poolNode,
			minSize:   nodePool.MinSize,
			maxSize:   nodePool.MaxSize,
		})
	}
	if len(group) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no worker pools are configured")
	}

	m.nodeGroups = group
	return nil
}
