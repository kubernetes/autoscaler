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

package magnum

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v3/extensions/trusts"
)

const (
	clusterStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
	clusterStatusUpdateComplete   = "UPDATE_COMPLETE"
	clusterStatusUpdateFailed     = "UPDATE_FAILED"

	waitForStatusTimeStep       = 30 * time.Second
	waitForUpdateStatusTimeout  = 2 * time.Minute
	waitForCompleteStatusTimout = 10 * time.Minute

	// Could move to property of magnumManager implementations if needed
	scaleToZeroSupported = false

	// Time that the goroutine that first acquires clusterUpdateMutex
	// in deleteNodes should wait for other synchronous calls to deleteNodes.
	deleteNodesBatchingDelay = 2 * time.Second
)

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

// NodeRef stores the name, machineID and providerID of a node.
type NodeRef struct {
	Name       string
	MachineID  string
	ProviderID string
	IPs        []string
}
