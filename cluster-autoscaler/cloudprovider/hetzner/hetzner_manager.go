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

package hetzner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"net"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/klog/v2"
)

var (
	httpClient = &http.Client{
		Transport: instrumentedRoundTripper(),
	}
)

// hetznerManager handles Hetzner communication and data caching of
// node groups
type hetznerManager struct {
	client           *hcloud.Client
	nodeGroups       map[string]*hetznerNodeGroup
	apiCallContext   context.Context
	clusterConfig    *ClusterConfig
	sshKey           *hcloud.SSHKey
	network          *hcloud.Network
	firewall         *hcloud.Firewall
	createTimeout    time.Duration
	publicIPv4       bool
	publicIPv6       bool
	cachedServerType *serverTypeCache
	cachedServers    *serversCache
	ipReserver       *ipReserver
	subnet           *net.IPNet
}

// ClusterConfig holds the configuration for all the nodepools
type ClusterConfig struct {
	ImagesForArch    ImageList
	NodeConfigs      map[string]*NodeConfig
	IsUsingNewFormat bool
	LegacyConfig     LegacyConfig
}

// ImageList holds the image id/names for the different architectures
type ImageList struct {
	Arm64 string
	Amd64 string
}

// NodeConfig holds the configuration for a single nodepool
type NodeConfig struct {
	CloudInit      string
	PlacementGroup string
	Taints         []apiv1.Taint
	Labels         map[string]string
}

// LegacyConfig holds the configuration in the legacy format
type LegacyConfig struct {
	CloudInit string
	ImageName string
}

func newManager() (*hetznerManager, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, errors.New("`HCLOUD_TOKEN` is not specified")
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(token),
		hcloud.WithHTTPClient(httpClient),
		hcloud.WithApplication("cluster-autoscaler", version.ClusterAutoscalerVersion),
		hcloud.WithPollOpts(hcloud.PollOpts{
			BackoffFunc: hcloud.ExponentialBackoff(2, 500*time.Millisecond),
		}),
		hcloud.WithDebugWriter(&debugWriter{}),
	}

	endpoint := os.Getenv("HCLOUD_ENDPOINT")
	if endpoint != "" {
		opts = append(opts, hcloud.WithEndpoint(endpoint))
	}

	client := hcloud.NewClient(opts...)

	ctx := context.Background()
	var err error

	clusterConfigBase64 := os.Getenv("HCLOUD_CLUSTER_CONFIG")
	clusterConfigFile := os.Getenv("HCLOUD_CLUSTER_CONFIG_FILE")
	cloudInitBase64 := os.Getenv("HCLOUD_CLOUD_INIT")

	if clusterConfigBase64 == "" && cloudInitBase64 == "" && clusterConfigFile == "" {
		return nil, errors.New("neither `HCLOUD_CLUSTER_CONFIG`, `HCLOUD_CLOUD_INIT` nor `HCLOUD_CLUSTER_CONFIG_FILE` is specified")
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
	}

	if clusterConfigJsonData != nil {
		unmarshalErr := json.Unmarshal(clusterConfigJsonData, &clusterConfig)
		if unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal cluster config JSON: %s", unmarshalErr)
		}
		clusterConfig.IsUsingNewFormat = true
	} else {
		cloudInit, decErr := base64.StdEncoding.DecodeString(cloudInitBase64)
		if decErr != nil {
			return nil, fmt.Errorf("failed to parse cloud init error: %s", decErr)
		}

		imageName := os.Getenv("HCLOUD_IMAGE")
		if imageName == "" {
			imageName = "ubuntu-20.04"
		}

		clusterConfig.LegacyConfig.CloudInit = string(cloudInit)
		clusterConfig.LegacyConfig.ImageName = imageName
	}

	publicIPv4 := true
	publicIPv4Str := os.Getenv("HCLOUD_PUBLIC_IPV4")
	if publicIPv4Str != "" {
		publicIPv4, err = strconv.ParseBool(publicIPv4Str)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HCLOUD_PUBLIC_IPV4: %s", err)
		}
	}

	publicIPv6 := true
	publicIPv6Str := os.Getenv("HCLOUD_PUBLIC_IPV6")
	if publicIPv6Str != "" {
		publicIPv6, err = strconv.ParseBool(publicIPv6Str)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HCLOUD_PUBLIC_IPV6: %s", err)
		}
	}

	var sshKey *hcloud.SSHKey
	sshKeyIdOrName := os.Getenv("HCLOUD_SSH_KEY")
	if sshKeyIdOrName != "" {
		sshKey, _, err = client.SSHKey.Get(ctx, sshKeyIdOrName)
		if err != nil {
			return nil, fmt.Errorf("failed to get ssh key error: %s", err)
		}
	}

	var network *hcloud.Network
	networkIdOrName := os.Getenv("HCLOUD_NETWORK")
	if networkIdOrName != "" {
		network, _, err = client.Network.Get(ctx, networkIdOrName)
		if err != nil {
			return nil, fmt.Errorf("failed to get network error: %s", err)
		}

	}

	createTimeout := serverCreateTimeoutDefault
	v, err := strconv.Atoi(os.Getenv("HCLOUD_SERVER_CREATION_TIMEOUT"))
	if err == nil && v != 0 {
		createTimeout = time.Duration(v) * time.Minute
	}

	var firewall *hcloud.Firewall
	firewallIdOrName := os.Getenv("HCLOUD_FIREWALL")
	if firewallIdOrName != "" {
		firewall, _, err = client.Firewall.Get(ctx, firewallIdOrName)
		if err != nil {
			return nil, fmt.Errorf("failed to get firewall error: %s", err)
		}
	}

    serversCache := newServersCache(ctx, client)

    var subnet *net.IPNet
    subnetStr := os.Getenv("HCLOUD_SUBNET")
    if subnetStr != "" && network != nil {
        _, subnet, err = net.ParseCIDR(subnetStr)
        if err != nil {
            return nil, fmt.Errorf("failed to parse HCLOUD_SUBNET: %s", err)
        }
        // Validate that the subnet is part of the network
        found := false
        for _, nSubnet := range network.Subnets {
            _, nSubnetRange, _ := net.ParseCIDR(nSubnet.IPRange.String())
            if nSubnetRange.String() == subnet.String() {
                found = true
                break
            }
        }
        if !found {
            return nil, fmt.Errorf("HCLOUD_SUBNET %s is not part of the specified HCLOUD_NETWORK %s", subnetStr, network.Name)
        }
    }
    //_, subnet, _ = net.ParseCIDR("10.0.95.128/25")
	m := &hetznerManager{
		client:           client,
		nodeGroups:       make(map[string]*hetznerNodeGroup),
		sshKey:           sshKey,
		network:          network,
		firewall:         firewall,
		createTimeout:    createTimeout,
		apiCallContext:   ctx,
		publicIPv4:       publicIPv4,
		publicIPv6:       publicIPv6,
		clusterConfig:    clusterConfig,
		cachedServerType: newServerTypeCache(ctx, client),
		cachedServers:    serversCache,
		subnet: subnet,
		ipReserver: newIPReserver(ctx, client, serversCache),
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *hetznerManager) Refresh() error {
	return nil
}

func (m *hetznerManager) allServers(nodeGroup string) ([]*hcloud.Server, error) {
	servers, err := m.cachedServers.getServersByNodeGroupName(nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for hcloud: %v", err)
	}

	return servers, nil
}

func (m *hetznerManager) deleteByNode(node *apiv1.Node) error {
	server, err := m.serverForNode(node)
	if err != nil {
		return fmt.Errorf("failed to delete node %s error: %v", node.Name, err)
	}

	if server == nil {
		return fmt.Errorf("failed to delete node %s server not found", node.Name)
	}

	return m.deleteServer(server)
}

func (m *hetznerManager) assignIP(server *hcloud.Server, ip net.IP) error {
    // Basic validation
    if server == nil || ip == nil || m.network == nil {
        return fmt.Errorf("invalid parameters: server=%v, ip=%v, network=%v", server != nil, ip != nil, m.network != nil)
    }

    ctx := m.apiCallContext
    klog.Infof("Assigning static IP %s to server %s", ip.String(), server.Name)

    // Check if server is already attached to the correct network
    isAttached := false
    for _, privNet := range server.PrivateNet {
        if privNet.Network != nil && privNet.Network.ID == m.network.ID {
            isAttached = true
            break
        }
    }

    // Detach from network if already attached
    if isAttached {
        klog.V(1).Infof("Detaching server %s from network %s", server.Name, m.network.Name)

        detachAction, _, err := m.client.Server.DetachFromNetwork(ctx, server, hcloud.ServerDetachFromNetworkOpts{
            Network: m.network,
        })
        if err != nil {
            return fmt.Errorf("failed to detach from network: %w", err)
        }

        if err := m.client.Action.WaitFor(ctx, detachAction); err != nil {
            return fmt.Errorf("waiting for network detachment failed: %w", err)
        }
    }

    // Attach to network with static IP
    klog.V(1).Infof("Connecting server %s to network %s with IP %s", server.Name, m.network.Name, ip.String())

    attachAction, _, err := m.client.Server.AttachToNetwork(ctx, server, hcloud.ServerAttachToNetworkOpts{
        Network: m.network,
        IP:      ip,
    })
    if err != nil {
        return fmt.Errorf("failed to attach to network: %w", err)
    }

    if err := m.client.Action.WaitFor(ctx, attachAction); err != nil {
        return fmt.Errorf("waiting for network attachment failed: %w", err)
    }

    klog.Infof("Server %s successfully connected with IP %s", server.Name, ip.String())
    return nil
}

func (m *hetznerManager) createServer(ctx context.Context, opts hcloud.ServerCreateOpts) (*hcloud.Server, error) {
    // Initialize labels if not present
    if opts.Labels == nil {
        opts.Labels = make(map[string]string)
    }

    // Prepare IP reservation and assignment
    var ipToAssign net.IP
    if m.network != nil && m.subnet != nil {
        // Create server initially stopped for network configuration
        startAfterCreate := false
        opts.StartAfterCreate = &startAfterCreate

        if opts.Networks != nil {
            // remove network from opts if it was set
            klog.Warningf("Network is set in ServerCreateOpts, but will be ignored for IP assignment: %s", opts.Networks[0].Name)
            opts.Networks = []*hcloud.Network{}
        }

        // Reserve IP address
        var err error
        if ipToAssign, err = m.ipReserver.reserveNewIP(m.subnet); err != nil {
            return nil, fmt.Errorf("IP reservation failed: %w", err)
        }

        // Store reserved IP as label
        opts.Labels[m.ipReserver.getReservedIPLabelName()] = ipToAssign.String()
    }

    // Create server
    serverCreateResult, _, err := m.client.Server.Create(ctx, opts)
    if err != nil {
        if ipToAssign != nil {
            m.ipReserver.removeReservedIP(ipToAssign)
        }
        return nil, fmt.Errorf("Server creation of type %s in region %s failed: %w", opts.ServerType.Name, opts.Location.Name, err)
    }

    server := serverCreateResult.Server
    actions := append(serverCreateResult.NextActions, serverCreateResult.Action)

    // Wait for creation actions to complete
    if err = m.client.Action.WaitFor(ctx, actions...); err != nil {
        if ipToAssign != nil {
            m.ipReserver.removeReservedIP(ipToAssign)
        }
        return server, fmt.Errorf("Waiting for server actions for %s failed: %w", server.Name, err)
    }

    // Assign IP if one was reserved
    if ipToAssign != nil {
        if err = m.assignIP(server, ipToAssign); err != nil {
            m.ipReserver.removeReservedIP(ipToAssign)
            return server, fmt.Errorf("IP assignment for server %s failed: %w", server.Name, err)
        }
    }

    // Start the server if it wasn't started automatically
    if opts.StartAfterCreate != nil && !*opts.StartAfterCreate {
        powerOnAction, _, err := m.client.Server.Poweron(ctx, server)
        if err != nil {
            return server, fmt.Errorf("Powering on server %s failed: %v", server.Name, err)
        }
        if err = m.client.Action.WaitFor(ctx, powerOnAction); err != nil {
            return server, fmt.Errorf("Waiting for power-on for server %s failed: %v", server.Name, err)
        }
    }

    return server, nil
}

func (m *hetznerManager) deleteServer(server *hcloud.Server) error {
    if server != nil {
        reservedIP, reservedIPExists := m.ipReserver.getReservedIPFromLabel(server)
        if reservedIPExists {
            m.ipReserver.removeReservedIP(reservedIP)
        }
    }
	_, _, err := m.client.Server.DeleteWithResult(m.apiCallContext, server)
	return err
}

func (m *hetznerManager) validProviderID(providerID string) bool {
	return strings.HasPrefix(providerID, providerIDPrefix)
}

func (m *hetznerManager) serverForNode(node *apiv1.Node) (*hcloud.Server, error) {
	var nodeIdOrName string
	if node.Spec.ProviderID != "" {
		if !m.validProviderID(node.Spec.ProviderID) {
			// This cluster-autoscaler provider only handles Hetzner Cloud servers.
			// Any other provider ID prefix is invalid, and we return no server. Returning an error here breaks hybrid
			// clusters with nodes from Hetzner Cloud & Robot (or other providers).
			return nil, nil
		}
		nodeIdOrName = strings.TrimPrefix(node.Spec.ProviderID, providerIDPrefix)
	} else {
		nodeIdOrName = node.Name
	}

	server, err := m.cachedServers.getServer(nodeIdOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for node %s error: %v", node.Name, err)
	}
	return server, nil
}
