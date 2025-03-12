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

package kubelet

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
)

// providerRequiresNetworkingConfiguration returns whether the cloud provider
// requires special networking configuration.
func (kl *Kubelet) providerRequiresNetworkingConfiguration() bool {
	// TODO: We should have a mechanism to say whether native cloud provider
	// is used or whether we are using overlay networking. We should return
	// true for cloud providers if they implement Routes() interface and
	// we are not using overlay networking.
	if kl.cloud == nil || kl.cloud.ProviderName() != "gce" {
		return false
	}
	_, supported := kl.cloud.Routes()
	return supported
}

// updatePodCIDR updates the pod CIDR in the runtime state if it is different
// from the current CIDR. Return true if pod CIDR is actually changed.
func (kl *Kubelet) updatePodCIDR(ctx context.Context, cidr string) (bool, error) {
	kl.updatePodCIDRMux.Lock()
	defer kl.updatePodCIDRMux.Unlock()

	podCIDR := kl.runtimeState.podCIDR()

	if podCIDR == cidr {
		return false, nil
	}

	// kubelet -> generic runtime -> runtime shim -> network plugin
	// docker/non-cri implementations have a passthrough UpdatePodCIDR
	if err := kl.getRuntime().UpdatePodCIDR(ctx, cidr); err != nil {
		// If updatePodCIDR would fail, theoretically pod CIDR could not change.
		// But it is better to be on the safe side to still return true here.
		return true, fmt.Errorf("failed to update pod CIDR: %v", err)
	}
	klog.InfoS("Updating Pod CIDR", "originalPodCIDR", podCIDR, "newPodCIDR", cidr)
	kl.runtimeState.setPodCIDR(cidr)
	return true, nil
}

// GetPodDNS returns DNS settings for the pod.
// This function is defined in kubecontainer.RuntimeHelper interface so we
// have to implement it.
func (kl *Kubelet) GetPodDNS(pod *v1.Pod) (*runtimeapi.DNSConfig, error) {
	return kl.dnsConfigurer.GetPodDNS(pod)
}
