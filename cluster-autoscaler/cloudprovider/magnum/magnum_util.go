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

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/trusts"
	provider_os "k8s.io/kubernetes/pkg/cloudprovider/providers/openstack"
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

func toAuthOptsExt(cfg provider_os.Config) trusts.AuthOptsExt {
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
