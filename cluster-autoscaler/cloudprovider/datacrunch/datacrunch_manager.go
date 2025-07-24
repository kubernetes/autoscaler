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

package datacrunch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
)

// datacrunchManager handles Datacrunch communication and data caching of
// node groups
type datacrunchManager struct {
	client           *datacrunchclient.Client
	nodeGroups       map[string]*datacrunchNodeGroup
	apiCallContext   context.Context
	clusterConfig    *ClusterConfig
	cachedServerType *serverTypeCache
	cachedServers    *serversCache
}

// ClusterConfig holds the configuration for all the nodepools
type ClusterConfig struct {
	NodeConfigs map[string]*NodeConfig `json:"node_configs"`
}

// InstanceOption is the option for the instance type
type InstanceOption string

const (
	// InstanceOptionPreferSpot is the option to prefer spot instances
	InstanceOptionPreferSpot InstanceOption = "prefer_spot"
	// InstanceOptionPreferOnDemand is the option to prefer on-demand instances
	InstanceOptionPreferOnDemand InstanceOption = "prefer_on_demand"
	// InstanceOptionSpotOnly is the option to only use spot instances
	InstanceOptionSpotOnly InstanceOption = "spot_only"
	// InstanceOptionOnDemandOnly is the option to only use on-demand instances
	InstanceOptionOnDemandOnly InstanceOption = "on_demand_only"
)

// PricingOption is the option for the pricing type
type PricingOption string

const (
	// PricingOptionFixed is the option to use fixed pricing
	PricingOptionFixed PricingOption = "fixed"
	// PricingOptionDynamic is the option to use dynamic pricing
	PricingOptionDynamic PricingOption = "dynamic"
)

// NodeConfig holds the configuration for a single nodepool
type NodeConfig struct {
	Taints     []apiv1.Taint     `json:"taints"`
	Labels     map[string]string `json:"labels"`
	ImageType  string            `json:"image_type"`
	DiskSizeGB int               `json:"disk_size_gb"`
	// useful in MiG scenarios when MiG is configured as part of a startup script
	OverrideNumGPUs *int `json:"override_num_gpus"`
	// base64 encoded startup script. Takes precedence over StartupScriptFetchUrl.
	StartupScriptBase64 string         `json:"startup_script_base64"`
	InstanceOption      InstanceOption `json:"instance_option"`
	PricingOption       *PricingOption `json:"pricing_option,omitempty"`
	SSHKeyIDs           []string       `json:"ssh_key_ids"`
}

func newManager() (*datacrunchManager, error) {
	token := os.Getenv("DATACRUNCH_CLIENT_ID")
	secret := os.Getenv("DATACRUNCH_CLIENT_SECRET")
	if token == "" || secret == "" {
		return nil, errors.New("`DATACRUNCH_CLIENT_ID` and `DATACRUNCH_CLIENT_SECRET` must be specified")
	}

	client := datacrunchclient.NewClient(token, secret)

	ctx := context.Background()

	clusterConfigBase64 := os.Getenv("DATACRUNCH_CLUSTER_CONFIG")
	clusterConfigBaseJSON := os.Getenv("DATACRUNCH_CLUSTER_CONFIG_JSON")
	clusterConfigFile := os.Getenv("DATACRUNCH_CLUSTER_CONFIG_FILE")

	if clusterConfigBase64 == "" && clusterConfigFile == "" && clusterConfigBaseJSON == "" {
		return nil, errors.New("one of `DATACRUNCH_CLUSTER_CONFIG`, `DATACRUNCH_CLUSTER_CONFIG_FILE` or `DATACRUNCH_CLUSTER_CONFIG_JSON` must be specified")
	}
	var clusterConfig = &ClusterConfig{}

	var clusterConfigJsonData []byte
	var readErr error
	if clusterConfigBase64 != "" {
		clusterConfigJsonData, readErr = base64.StdEncoding.DecodeString(clusterConfigBase64)
		if readErr != nil {
			return nil, fmt.Errorf("failed to parse cluster config error: %s", readErr)
		}
	} else if clusterConfigFile != "" {
		clusterConfigJsonData, readErr = os.ReadFile(clusterConfigFile)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read cluster config file: %s", readErr)
		}
	} else if clusterConfigBaseJSON != "" {
		clusterConfigJsonData = []byte(clusterConfigBaseJSON)
	}

	unmarshalErr := json.Unmarshal(clusterConfigJsonData, &clusterConfig)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster config JSON: %s", unmarshalErr)
	}

	m := &datacrunchManager{
		client:           client,
		nodeGroups:       make(map[string]*datacrunchNodeGroup),
		apiCallContext:   ctx,
		clusterConfig:    clusterConfig,
		cachedServerType: newServerTypeCache(ctx, client),
		cachedServers:    newServersCache(ctx, client),
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *datacrunchManager) Refresh() error {
	return nil
}

func (m *datacrunchManager) allServers(nodeGroup string) ([]*datacrunchclient.Instance, error) {
	servers, err := m.cachedServers.getServersByNodeGroupName(nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for datacrunch: %v", err)
	}
	return servers, nil
}

func (m *datacrunchManager) deleteByNode(node *apiv1.Node) error {
	instance, err := m.serverForNode(node)
	if err != nil {
		return fmt.Errorf("failed to delete node %s error: %v", node.Name, err)
	}
	if instance == nil {
		return fmt.Errorf("failed to delete node %s instance not found", node.Name)
	}
	return m.deleteServer(instance)
}

func (m *datacrunchManager) deleteServer(instance *datacrunchclient.Instance) error {
	req := datacrunchclient.InstanceActionRequest{
		Action: "delete",
		ID:     instance.ID,
	}

	klog.V(4).Infof("deleting server %s", instance.ID)

	err := m.client.PerformInstanceAction(req)
	if err != nil {
		return fmt.Errorf("failed to delete server %s error: %v", instance.ID, err)
	}

	// Wait for instance deletion, then cleanup detached volumes so we don't run into quota issues
	// NOTE: Not sure if we even need to wait here, someone from datacrunch need to confirm this.
	go func() {
		klog.V(4).Infof("deleting volumes for server %s", instance.ID)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		errorCount := 0
		maxErrorCount := 3

		for range ticker.C {
			volumes, err := m.client.ListVolumesInTrash()
			if err != nil {
				klog.Errorf("failed to list volumes in trash. error: %v", err)
				errorCount++
				if errorCount > maxErrorCount {
					ticker.Stop()
					return
				}
				continue
			}

			for _, volume := range volumes {
				if volume.Name == instance.Hostname {
					klog.V(4).Infof("found detached volume for instance %s, deleting volume %s", instance.Hostname, volume.ID)
					err := m.client.DeleteVolume(volume.ID, true)
					if err != nil {
						klog.Errorf("failed to delete volume %s. error: %v", volume.ID, err)
						errorCount++
						if errorCount > maxErrorCount {
							ticker.Stop()
							return
						}
						continue
					}
					ticker.Stop()
					return
				}
			}

			klog.Warningf("no volumes found for instance %s", instance.ID)
			ticker.Stop()
			return

		}
	}()

	return nil
}

func (m *datacrunchManager) validProviderID(providerID string) bool {
	return strings.HasPrefix(providerID, providerIDPrefix)
}

func (m *datacrunchManager) serverForNode(node *apiv1.Node) (*datacrunchclient.Instance, error) {
	var nodeIdOrName string
	if node.Spec.ProviderID != "" {
		if !m.validProviderID(node.Spec.ProviderID) {
			// This cluster-autoscaler provider only handles DataCrunch instances.
			// Any other provider ID prefix is invalid, and we return no instance. Returning an error here breaks hybrid
			// clusters with nodes from multiple providers.
			return nil, nil
		}
		nodeIdOrName = strings.TrimPrefix(node.Spec.ProviderID, providerIDPrefix)
	} else {
		nodeIdOrName = node.Name
	}

	instance, err := m.cachedServers.getServer(nodeIdOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance for node %s error: %v", node.Name, err)
	}
	return instance, nil
}
