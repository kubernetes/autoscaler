/*
Copyright 2020 The Kubernetes Authors.

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

package magnum

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/gcfg.v1"
	netutil "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/clusters"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v3/extensions/trusts"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	certutil "k8s.io/client-go/util/cert"
	klog "k8s.io/klog/v2"
)

// These Opts types are for parsing an OpenStack cloud-config file.
// The definitions are taken from cloud-provider-openstack.

// MyDuration is the encoding.TextUnmarshaler interface for time.Duration
type MyDuration struct {
	time.Duration
}

// UnmarshalText is used to convert from text to Duration
func (d *MyDuration) UnmarshalText(text []byte) error {
	res, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	d.Duration = res
	return nil
}

// LoadBalancerOpts have the options to talk to Neutron LBaaSV2 or Octavia
type LoadBalancerOpts struct {
	LBVersion            string     `gcfg:"lb-version"`          // overrides autodetection. Only support v2.
	UseOctavia           bool       `gcfg:"use-octavia"`         // uses Octavia V2 service catalog endpoint
	SubnetID             string     `gcfg:"subnet-id"`           // overrides autodetection.
	FloatingNetworkID    string     `gcfg:"floating-network-id"` // If specified, will create floating ip for loadbalancer, or do not create floating ip.
	LBMethod             string     `gcfg:"lb-method"`           // default to ROUND_ROBIN.
	LBProvider           string     `gcfg:"lb-provider"`
	CreateMonitor        bool       `gcfg:"create-monitor"`
	MonitorDelay         MyDuration `gcfg:"monitor-delay"`
	MonitorTimeout       MyDuration `gcfg:"monitor-timeout"`
	MonitorMaxRetries    uint       `gcfg:"monitor-max-retries"`
	ManageSecurityGroups bool       `gcfg:"manage-security-groups"`
	NodeSecurityGroupIDs []string   // Do not specify, get it automatically when enable manage-security-groups. TODO(FengyunPan): move it into cache
}

// BlockStorageOpts is used to talk to Cinder service
type BlockStorageOpts struct {
	BSVersion             string `gcfg:"bs-version"`        // overrides autodetection. v1 or v2. Defaults to auto
	TrustDevicePath       bool   `gcfg:"trust-device-path"` // See Issue #33128
	IgnoreVolumeAZ        bool   `gcfg:"ignore-volume-az"`
	NodeVolumeAttachLimit int    `gcfg:"node-volume-attach-limit"` // override volume attach limit for Cinder. Default is : 256
}

// RouterOpts is used for Neutron routes
type RouterOpts struct {
	RouterID string `gcfg:"router-id"` // required
}

// MetadataOpts is used for configuring how to talk to metadata service or config drive
type MetadataOpts struct {
	SearchOrder    string     `gcfg:"search-order"`
	RequestTimeout MyDuration `gcfg:"request-timeout"`
}

// Config is used to read and store information from the cloud configuration file
//
// Taken from kubernetes/pkg/cloudprovider/providers/openstack/openstack.go
// LoadBalancer, BlockStorage, Route, Metadata are not needed for the autoscaler,
// but are kept so that if a cloud-config file with those sections is provided
// then the parsing will not fail.
type Config struct {
	Global struct {
		AuthURL         string `gcfg:"auth-url"`
		Username        string
		UserID          string `gcfg:"user-id"`
		Password        string
		TenantID        string `gcfg:"tenant-id"`
		TenantName      string `gcfg:"tenant-name"`
		TrustID         string `gcfg:"trust-id"`
		DomainID        string `gcfg:"domain-id"`
		DomainName      string `gcfg:"domain-name"`
		Region          string
		CAFile          string `gcfg:"ca-file"`
		SecretName      string `gcfg:"secret-name"`
		SecretNamespace string `gcfg:"secret-namespace"`
	}
	LoadBalancer LoadBalancerOpts
	BlockStorage BlockStorageOpts
	Route        RouterOpts
	Metadata     MetadataOpts
}

func toAuthOptsExt(cfg Config) trusts.AuthOptsExt {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.Global.AuthURL,
		Username:         cfg.Global.Username,
		UserID:           cfg.Global.UserID,
		Password:         cfg.Global.Password,
		TenantID:         cfg.Global.TenantID,
		TenantName:       cfg.Global.TenantName,
		DomainID:         cfg.Global.DomainID,
		DomainName:       cfg.Global.DomainName,

		// Persistent service, so we need to be able to renew tokens.
		AllowReauth: true,
	}

	return trusts.AuthOptsExt{
		TrustID:            cfg.Global.TrustID,
		AuthOptionsBuilder: &opts,
	}
}

// readConfig parses an OpenStack cloud-config file from an io.Reader.
func readConfig(configReader io.Reader) (*Config, error) {
	var cfg Config
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			return nil, fmt.Errorf("couldn't read cloud config: %v", err)
		}
	}
	return &cfg, nil
}

// createProviderClient creates and authenticates a gophercloud provider client.
func createProviderClient(cfg *Config, opts config.AutoscalingOptions) (*gophercloud.ProviderClient, error) {
	if opts.ClusterName == "" {
		return nil, errors.New("the cluster-name parameter must be set")
	}

	authOpts := toAuthOptsExt(*cfg)

	provider, err := openstack.NewClient(cfg.Global.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("could not create openstack client: %v", err)
	}

	userAgent := gophercloud.UserAgent{}
	userAgent.Prepend(fmt.Sprintf("cluster-autoscaler/%s", version.ClusterAutoscalerVersion))
	userAgent.Prepend(fmt.Sprintf("cluster/%s", opts.ClusterName))
	provider.UserAgent = userAgent

	klog.V(5).Infof("Using user-agent %q", userAgent.Join())

	if cfg.Global.CAFile != "" {
		roots, err := certutil.NewPool(cfg.Global.CAFile)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{}
		config.RootCAs = roots
		provider.HTTPClient.Transport = netutil.SetOldTransportDefaults(&http.Transport{TLSClientConfig: config})
	}

	err = openstack.AuthenticateV3(provider, authOpts, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("could not authenticate client: %v", err)
	}

	return provider, nil
}

// createClusterClient creates a gophercloud service client for communicating with Magnum.
func createClusterClient(cfg *Config, provider *gophercloud.ProviderClient, opts config.AutoscalingOptions) (*gophercloud.ServiceClient, error) {
	clusterClient, err := openstack.NewContainerInfraV1(provider, gophercloud.EndpointOpts{Type: "container-infra", Name: "magnum", Region: cfg.Global.Region})
	if err != nil {
		return nil, fmt.Errorf("could not create container-infra client: %v", err)
	}
	return clusterClient, nil
}

// checkClusterUUID replaces a cluster name with the UUID, if the name was given in the parameters
// to the cluster autoscaler.
func checkClusterUUID(provider *gophercloud.ProviderClient, clusterClient *gophercloud.ServiceClient, opts config.AutoscalingOptions) error {
	cluster, err := clusters.Get(clusterClient, opts.ClusterName).Extract()
	if err != nil {
		return fmt.Errorf("unable to access cluster %q: %v", opts.ClusterName, err)
	}

	// Prefer to use the cluster UUID if the cluster name was given in the parameters
	if opts.ClusterName != cluster.UUID {
		klog.V(2).Infof("Using cluster UUID %q instead of name %q", cluster.UUID, opts.ClusterName)
		opts.ClusterName = cluster.UUID

		// Need to remake user-agent with UUID instead of name
		userAgent := gophercloud.UserAgent{}
		userAgent.Prepend(fmt.Sprintf("cluster-autoscaler/%s", version.ClusterAutoscalerVersion))
		userAgent.Prepend(fmt.Sprintf("cluster/%s", opts.ClusterName))
		provider.UserAgent = userAgent
		klog.V(5).Infof("Using updated user-agent %q", userAgent.Join())
	}

	return nil
}

// createHeatClient creates a gophercloud service client for communicating with Heat.
func createHeatClient(cfg *Config, provider *gophercloud.ProviderClient, opts config.AutoscalingOptions) (*gophercloud.ServiceClient, error) {
	heatClient, err := openstack.NewOrchestrationV1(provider, gophercloud.EndpointOpts{Type: "orchestration", Name: "heat", Region: cfg.Global.Region})
	if err != nil {
		return nil, fmt.Errorf("could not create orchestration client: %v", err)
	}

	return heatClient, nil
}
