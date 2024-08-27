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

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cloudprovider "k8s.io/cloud-provider"
	servicehelpers "k8s.io/cloud-provider/service/helpers"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
	"k8s.io/utils/strings/slices"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/loadbalancer"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/loadbalancer/iputil"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

var _ cloudprovider.LoadBalancer = (*Cloud)(nil)

// Since public IP is not a part of the load balancer on Azure,
// there is a chance that we could orphan public IP resources while we delete the load balancer (kubernetes/kubernetes#80571).
// We need to make sure the existence of the load balancer depends on the load balancer resource and public IP resource on Azure.
func (az *Cloud) existsPip(clusterName string, service *v1.Service) bool {
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	existsPipSingleStack := func(isIPv6 bool) bool {
		pipName, _, err := az.determinePublicIPName(clusterName, service, isIPv6)
		if err != nil {
			return false
		}
		pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
		_, existingPip, err := az.getPublicIPAddress(pipResourceGroup, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			return false
		}
		return existingPip
	}

	if v4Enabled && !existsPipSingleStack(consts.IPVersionIPv4) {
		return false
	}
	if v6Enabled && !existsPipSingleStack(consts.IPVersionIPv6) {
		return false
	}
	return true
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager.
// TODO: Break this up into different interfaces (LB, etc) when we have more than one type of service
func (az *Cloud) GetLoadBalancer(_ context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	existingLBs, err := az.ListLB(service)
	if err != nil {
		return nil, az.existsPip(clusterName, service), err
	}

	_, _, status, _, existsLb, err := az.getServiceLoadBalancer(service, clusterName, nil, false, &existingLBs)
	if err != nil || existsLb {
		return status, existsLb || az.existsPip(clusterName, service), err
	}

	flippedService := flipServiceInternalAnnotation(service)
	_, _, status, _, existsLb, err = az.getServiceLoadBalancer(flippedService, clusterName, nil, false, &existingLBs)
	if err != nil || existsLb {
		return status, existsLb || az.existsPip(clusterName, service), err
	}

	// Return exists = false only if the load balancer and the public IP are not found on Azure
	if !existsLb && !az.existsPip(clusterName, service) {
		serviceName := getServiceName(service)
		klog.V(5).Infof("getloadbalancer (cluster:%s) (service:%s) - doesn't exist", clusterName, serviceName)
		return nil, false, nil
	}

	// Return exists = true if only the public IP exists
	return nil, true, nil
}

func getPublicIPDomainNameLabel(service *v1.Service) (string, bool) {
	if labelName, found := service.Annotations[consts.ServiceAnnotationDNSLabelName]; found {
		return labelName, found
	}
	return "", false
}

// reconcileService reconcile the LoadBalancer service. It returns LoadBalancerStatus on success.
func (az *Cloud) reconcileService(_ context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	serviceName := getServiceName(service)
	resourceBaseName := az.GetLoadBalancerName(context.TODO(), "", service)
	klog.V(2).Infof("reconcileService: Start reconciling Service %q with its resource basename %q", serviceName, resourceBaseName)

	lb, err := az.reconcileLoadBalancer(clusterName, service, nodes, true /* wantLb */)
	if err != nil {
		klog.Errorf("reconcileLoadBalancer(%s) failed: %v", serviceName, err)
		return nil, err
	}

	lbStatus, lbIPsPrimaryPIPs, fipConfigs, err := az.getServiceLoadBalancerStatus(service, lb)
	if err != nil {
		klog.Errorf("getServiceLoadBalancerStatus(%s) failed: %v", serviceName, err)
		if !errors.Is(err, ErrorNotVmssInstance) {
			return nil, err
		}
	}

	serviceIPs := lbIPsPrimaryPIPs
	klog.V(2).Infof("reconcileService: reconciling security group for service %q with IPs %q, wantLb = true", serviceName, serviceIPs)
	if _, err := az.reconcileSecurityGroup(clusterName, service, ptr.Deref(lb.Name, ""), serviceIPs, true /* wantLb */); err != nil {
		klog.Errorf("reconcileSecurityGroup(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	for _, fipConfig := range fipConfigs {
		if err := az.reconcilePrivateLinkService(clusterName, service, fipConfig, true /* wantPLS */); err != nil {
			klog.Errorf("reconcilePrivateLinkService(%s) failed: %#v", serviceName, err)
			return nil, err
		}
	}

	updateService := updateServiceLoadBalancerIPs(service, lbIPsPrimaryPIPs)
	flippedService := flipServiceInternalAnnotation(updateService)
	if _, err := az.reconcileLoadBalancer(clusterName, flippedService, nil, false /* wantLb */); err != nil {
		klog.Errorf("reconcileLoadBalancer(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	// lb is not reused here because the ETAG may be changed in above operations, hence reconcilePublicIP() would get lb again from cache.
	klog.V(2).Infof("reconcileService: reconciling pip")
	if _, err := az.reconcilePublicIPs(clusterName, updateService, pointer.StringDeref(lb.Name, ""), true /* wantLb */); err != nil {
		klog.Errorf("reconcilePublicIP(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	lbName := strings.ToLower(pointer.StringDeref(lb.Name, ""))
	key := strings.ToLower(serviceName)
	if az.useMultipleStandardLoadBalancers() && isLocalService(service) {
		az.localServiceNameToServiceInfoMap.Store(key, newServiceInfo(getServiceIPFamily(service), lbName))
		// There are chances that the endpointslice changes after EnsureHostsInPool, so
		// need to check endpointslice for a second time.
		if err := az.checkAndApplyLocalServiceBackendPoolUpdates(*lb, service); err != nil {
			klog.Errorf("failed to checkAndApplyLocalServiceBackendPoolUpdates: %v", err)
			return nil, err
		}
	} else {
		az.localServiceNameToServiceInfoMap.Delete(key)
	}

	return lbStatus, nil
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager.
//
// Implementations may return a (possibly wrapped) api.RetryError to enforce
// backing off at a fixed duration. This can be used for cases like when the
// load balancer is not ready yet (e.g., it is still being provisioned) and
// polling at a fixed rate is preferred over backing off exponentially in
// order to minimize latency.
func (az *Cloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	// When a client updates the internal load balancer annotation,
	// the service may be switched from an internal LB to a public one, or vice versa.
	// Here we'll firstly ensure service do not lie in the opposite LB.

	// Serialize service reconcile process
	az.serviceReconcileLock.Lock()
	defer az.serviceReconcileLock.Unlock()

	var err error
	serviceName := getServiceName(service)
	mc := metrics.NewMetricContext("services", "ensure_loadbalancer", az.ResourceGroup, az.getNetworkResourceSubscriptionID(), serviceName)
	klog.V(5).InfoS("EnsureLoadBalancer Start", "service", serviceName, "cluster", clusterName, "service_spec", service)

	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
		klog.V(5).InfoS("EnsureLoadBalancer Finish", "service", serviceName, "cluster", clusterName, "service_spec", service, "error", err)
	}()

	lbStatus, err := az.reconcileService(ctx, clusterName, service, nodes)
	if err != nil {
		return nil, err
	}

	isOperationSucceeded = true
	return lbStatus, nil
}

func (az *Cloud) getLatestService(serviceName string, deepcopy bool) (*v1.Service, bool, error) {
	parts := strings.Split(serviceName, "/")
	ns, n := parts[0], parts[1]
	latestService, err := az.serviceLister.Services(ns).Get(n)
	switch {
	case apierrors.IsNotFound(err):
		// service absence in store means the service deletion is caught by watcher
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		if deepcopy {
			return latestService.DeepCopy(), true, nil
		}
		return latestService, true, nil
	}
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (az *Cloud) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	// Serialize service reconcile process
	az.serviceReconcileLock.Lock()
	defer az.serviceReconcileLock.Unlock()

	var err error
	serviceName := getServiceName(service)
	mc := metrics.NewMetricContext("services", "update_loadbalancer", az.ResourceGroup, az.getNetworkResourceSubscriptionID(), serviceName)
	klog.V(5).InfoS("UpdateLoadBalancer Start", "service", serviceName, "cluster", clusterName, "service_spec", service)
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
		klog.V(5).InfoS("UpdateLoadBalancer Finish", "service", serviceName, "cluster", clusterName, "service_spec", service, "error", err)
	}()

	// In case UpdateLoadBalancer gets stale service spec, retrieve the latest from lister
	service, serviceExists, err := az.getLatestService(serviceName, true)
	if err != nil {
		return fmt.Errorf("UpdateLoadBalancer: failed to get latest service %s: %w", service.Name, err)
	}
	if !serviceExists {
		isOperationSucceeded = true
		klog.V(2).Infof("UpdateLoadBalancer: skipping service %s because service is going to be deleted", serviceName)
		return nil
	}

	shouldUpdateLB, err := az.shouldUpdateLoadBalancer(clusterName, service, nodes)
	if err != nil {
		return err
	}

	if !shouldUpdateLB {
		isOperationSucceeded = true
		klog.V(2).Infof("UpdateLoadBalancer: skipping service %s because it is either being deleted or does not exist anymore", service.Name)
		return nil
	}

	_, err = az.reconcileService(ctx, clusterName, service, nodes)
	if err != nil {
		return err
	}

	isOperationSucceeded = true
	return nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (az *Cloud) EnsureLoadBalancerDeleted(_ context.Context, clusterName string, service *v1.Service) error {
	// Serialize service reconcile process
	az.serviceReconcileLock.Lock()
	defer az.serviceReconcileLock.Unlock()

	var err error
	serviceName := getServiceName(service)
	mc := metrics.NewMetricContext("services", "ensure_loadbalancer_deleted", az.ResourceGroup, az.getNetworkResourceSubscriptionID(), serviceName)
	klog.V(5).InfoS("EnsureLoadBalancerDeleted Start", "service", serviceName, "cluster", clusterName, "service_spec", service)
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
		klog.V(5).InfoS("EnsureLoadBalancerDeleted Finish", "service", serviceName, "cluster", clusterName, "service_spec", service, "error", err)
	}()

	lb, _, _, lbIPsPrimaryPIPs, _, err := az.getServiceLoadBalancer(service, clusterName, nil, false, &[]network.LoadBalancer{})
	if err != nil && !retry.HasStatusForbiddenOrIgnoredError(err) {
		return err
	}
	serviceIPsToCleanup := lbIPsPrimaryPIPs
	klog.V(2).Infof("EnsureLoadBalancerDeleted: reconciling security group for service %q with IPs %q, wantLb = false", serviceName, serviceIPsToCleanup)
	var lbName string
	if lb != nil {
		lbName = ptr.Deref(lb.Name, "")
	}
	_, err = az.reconcileSecurityGroup(clusterName, service, lbName, serviceIPsToCleanup, false /* wantLb */)
	if err != nil {
		return err
	}

	_, err = az.reconcileLoadBalancer(clusterName, service, nil, false /* wantLb */)
	if err != nil && !retry.HasStatusForbiddenOrIgnoredError(err) {
		return err
	}

	// check flipped service also
	flippedService := flipServiceInternalAnnotation(service)
	if _, err := az.reconcileLoadBalancer(clusterName, flippedService, nil, false /* wantLb */); err != nil {
		return err
	}

	if _, err = az.reconcilePublicIPs(clusterName, service, "", false /* wantLb */); err != nil {
		return err
	}

	if az.useMultipleStandardLoadBalancers() && isLocalService(service) {
		key := strings.ToLower(serviceName)
		az.localServiceNameToServiceInfoMap.Delete(key)
	}

	klog.V(2).Infof("Delete service (%s): FINISH", serviceName)
	isOperationSucceeded = true

	return nil
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (az *Cloud) GetLoadBalancerName(_ context.Context, _ string, service *v1.Service) string {
	return cloudprovider.DefaultLoadBalancerName(service)
}

func (az *Cloud) getLoadBalancerResourceGroup() string {
	if az.LoadBalancerResourceGroup != "" {
		return az.LoadBalancerResourceGroup
	}

	return az.ResourceGroup
}

// shouldChangeLoadBalancer determines if the load balancer of the service should be switched to another one
// according to the mode annotation on the service. This could be happened when the LB selection mode of an
// existing service is changed to another VMSS/VMAS.
func (az *Cloud) shouldChangeLoadBalancer(service *v1.Service, currLBName, clusterName, expectedLBName string) bool {
	// The load balancer can be changed in two cases:
	// 1. Using multiple standard load balancers.
	// 2. Migrate from multiple standard load balancers to single standard load balancer.
	if az.useStandardLoadBalancer() {
		if currLBName != expectedLBName {
			klog.V(2).Infof("shouldChangeLoadBalancer(%s, %s, %s): change the LB to another one %s", service.Name, currLBName, clusterName, expectedLBName)
			return true
		}
		return false
	}

	// basic LB
	hasMode, isAuto, vmSetName := az.getServiceLoadBalancerMode(service)

	// if no mode is given or the mode is `__auto__`, the current LB should be kept
	if !hasMode || isAuto {
		return false
	}

	lbName := trimSuffixIgnoreCase(currLBName, consts.InternalLoadBalancerNameSuffix)
	// change the LB from vmSet dedicated to primary if the vmSet becomes the primary one
	if strings.EqualFold(lbName, vmSetName) {
		if !strings.EqualFold(lbName, clusterName) &&
			strings.EqualFold(az.VMSet.GetPrimaryVMSetName(), vmSetName) {
			klog.V(2).Infof("shouldChangeLoadBalancer(%s, %s, %s): change the LB to another one", service.Name, currLBName, clusterName)
			return true
		}
		return false
	}
	if strings.EqualFold(vmSetName, az.VMSet.GetPrimaryVMSetName()) && strings.EqualFold(clusterName, lbName) {
		return false
	}

	// if the VMSS/VMAS of the current LB is different from the mode, change the LB
	// to another one
	klog.V(2).Infof("shouldChangeLoadBalancer(%s, %s, %s): change the LB to another one", service.Name, currLBName, clusterName)
	return true
}

// removeFrontendIPConfigurationFromLoadBalancer removes the given ip configs from the load balancer
// and delete the load balancer if there is no ip config on it. It returns the name of the deleted load balancer
// and it will be used in reconcileLoadBalancer to remove the load balancer from the list.
func (az *Cloud) removeFrontendIPConfigurationFromLoadBalancer(lb *network.LoadBalancer, existingLBs *[]network.LoadBalancer, fips []*network.FrontendIPConfiguration, clusterName string, service *v1.Service) (string, error) {
	if lb == nil || lb.LoadBalancerPropertiesFormat == nil || lb.FrontendIPConfigurations == nil {
		return "", nil
	}
	fipConfigs := *lb.FrontendIPConfigurations
	for i, fipConfig := range fipConfigs {
		for _, fip := range fips {
			if strings.EqualFold(pointer.StringDeref(fipConfig.Name, ""), pointer.StringDeref(fip.Name, "")) {
				fipConfigs = append(fipConfigs[:i], fipConfigs[i+1:]...)
				break
			}
		}
	}
	lb.FrontendIPConfigurations = &fipConfigs

	// also remove the corresponding rules/probes
	if lb.LoadBalancingRules != nil {
		lbRules := *lb.LoadBalancingRules
		for i := len(lbRules) - 1; i >= 0; i-- {
			for _, fip := range fips {
				if strings.Contains(pointer.StringDeref(lbRules[i].Name, ""), pointer.StringDeref(fip.Name, "")) {
					lbRules = append(lbRules[:i], lbRules[i+1:]...)
				}
			}
		}
		lb.LoadBalancingRules = &lbRules
	}
	if lb.Probes != nil {
		lbProbes := *lb.Probes
		for i := len(lbProbes) - 1; i >= 0; i-- {
			for _, fip := range fips {
				if strings.Contains(pointer.StringDeref(lbProbes[i].Name, ""), pointer.StringDeref(fip.Name, "")) {
					lbProbes = append(lbProbes[:i], lbProbes[i+1:]...)
				}
			}
		}
		lb.Probes = &lbProbes
	}

	// PLS does not support IPv6 so there will not be additional API calls.
	for _, fip := range fips {
		// clean up any private link service associated with the frontEndIPConfig
		if err := az.reconcilePrivateLinkService(clusterName, service, fip, false /* wantPLS */); err != nil {
			klog.Errorf("removeFrontendIPConfigurationFromLoadBalancer(%s, %s, %s, %s): failed to clean up PLS: %v", pointer.StringDeref(lb.Name, ""), pointer.StringDeref(fip.Name, ""), clusterName, service.Name, err)
			return "", err
		}
	}

	var deletedLBName string
	fipNames := []string{}
	for _, fip := range fips {
		fipNames = append(fipNames, pointer.StringDeref(fip.Name, ""))
	}
	logPrefix := fmt.Sprintf("removeFrontendIPConfigurationFromLoadBalancer(%s, %q, %s, %s)", pointer.StringDeref(lb.Name, ""), fipNames, clusterName, service.Name)
	if len(fipConfigs) == 0 {
		klog.V(2).Infof("%s: deleting load balancer because there is no remaining frontend IP configurations", logPrefix)
		err := az.cleanOrphanedLoadBalancer(lb, *existingLBs, service, clusterName)
		if err != nil {
			klog.Errorf("%s: failed to cleanupOrphanedLoadBalancer: %v", logPrefix, err)
			return "", err
		}
		deletedLBName = pointer.StringDeref(lb.Name, "")
	} else {
		klog.V(2).Infof("%s: updating the load balancer", logPrefix)
		err := az.CreateOrUpdateLB(service, *lb)
		if err != nil {
			klog.Errorf("%s: failed to CreateOrUpdateLB: %v", logPrefix, err)
			return "", err
		}
		_ = az.lbCache.Delete(pointer.StringDeref(lb.Name, ""))
	}
	return deletedLBName, nil
}

func (az *Cloud) cleanOrphanedLoadBalancer(lb *network.LoadBalancer, existingLBs []network.LoadBalancer, service *v1.Service, clusterName string) error {
	lbName := pointer.StringDeref(lb.Name, "")
	serviceName := getServiceName(service)
	isBackendPoolPreConfigured := az.isBackendPoolPreConfigured(service)
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	lbBackendPoolIDs := az.getBackendPoolIDs(clusterName, lbName)
	lbBackendPoolIDsToDelete := []string{}
	if v4Enabled {
		lbBackendPoolIDsToDelete = append(lbBackendPoolIDsToDelete, lbBackendPoolIDs[consts.IPVersionIPv4])
	}
	if v6Enabled {
		lbBackendPoolIDsToDelete = append(lbBackendPoolIDsToDelete, lbBackendPoolIDs[consts.IPVersionIPv6])
	}
	if isBackendPoolPreConfigured {
		klog.V(2).Infof("cleanOrphanedLoadBalancer(%s, %s, %s): ignore cleanup of dirty lb because the lb is pre-configured", lbName, serviceName, clusterName)
	} else {
		foundLB := false
		for _, existingLB := range existingLBs {
			if strings.EqualFold(pointer.StringDeref(lb.Name, ""), pointer.StringDeref(existingLB.Name, "")) {
				foundLB = true
				break
			}
		}
		if !foundLB {
			klog.V(2).Infof("cleanOrphanedLoadBalancer: the LB %s doesn't exist, will not delete it", pointer.StringDeref(lb.Name, ""))
			return nil
		}

		// When FrontendIPConfigurations is empty, we need to delete the Azure load balancer resource itself,
		// because an Azure load balancer cannot have an empty FrontendIPConfigurations collection
		klog.V(2).Infof("cleanOrphanedLoadBalancer(%s, %s, %s): deleting the LB since there are no remaining frontendIPConfigurations", lbName, serviceName, clusterName)

		// Remove backend pools from vmSets. This is required for virtual machine scale sets before removing the LB.
		vmSetName := az.mapLoadBalancerNameToVMSet(lbName, clusterName)
		if _, ok := az.VMSet.(*availabilitySet); ok {
			// do nothing for availability set
			lb.BackendAddressPools = nil
		}

		if deleteErr := az.safeDeleteLoadBalancer(*lb, clusterName, vmSetName, service); deleteErr != nil {
			klog.Warningf("cleanOrphanedLoadBalancer(%s, %s, %s): failed to DeleteLB: %v", lbName, serviceName, clusterName, deleteErr)

			rgName, vmssName, parseErr := retry.GetVMSSMetadataByRawError(deleteErr)
			if parseErr != nil {
				klog.Warningf("cleanOrphanedLoadBalancer(%s, %s, %s): failed to parse error: %v", lbName, serviceName, clusterName, parseErr)
				return deleteErr.Error()
			}
			if rgName == "" || vmssName == "" {
				klog.Warningf("cleanOrphanedLoadBalancer(%s, %s, %s): empty rgName or vmssName", lbName, serviceName, clusterName)
				return deleteErr.Error()
			}

			// if we reach here, it means the VM couldn't be deleted because it is being referenced by a VMSS
			if _, ok := az.VMSet.(*ScaleSet); !ok {
				klog.Warningf("cleanOrphanedLoadBalancer(%s, %s, %s): unexpected VMSet type, expected VMSS", lbName, serviceName, clusterName)
				return deleteErr.Error()
			}

			if !strings.EqualFold(rgName, az.ResourceGroup) {
				return fmt.Errorf("cleanOrphanedLoadBalancer(%s, %s, %s): the VMSS %s is in the resource group %s, but is referencing the LB in %s", lbName, serviceName, clusterName, vmssName, rgName, az.ResourceGroup)
			}

			vmssNamesMap := map[string]bool{vmssName: true}
			if err := az.VMSet.EnsureBackendPoolDeletedFromVMSets(vmssNamesMap, lbBackendPoolIDsToDelete); err != nil {
				klog.Errorf("cleanOrphanedLoadBalancer(%s, %s, %s): failed to EnsureBackendPoolDeletedFromVMSets: %v", lbName, serviceName, clusterName, err)
				return err
			}

			if deleteErr := az.DeleteLB(service, lbName); deleteErr != nil {
				klog.Errorf("cleanOrphanedLoadBalancer(%s, %s, %s): failed delete lb for the second time, stop retrying: %v", lbName, serviceName, clusterName, deleteErr)
				return deleteErr.Error()
			}
		}
		klog.V(10).Infof("cleanOrphanedLoadBalancer(%s, %s, %s): az.DeleteLB finished", lbName, serviceName, clusterName)
	}
	return nil
}

// safeDeleteLoadBalancer deletes the load balancer after decoupling it from the vmSet
func (az *Cloud) safeDeleteLoadBalancer(lb network.LoadBalancer, clusterName, vmSetName string, service *v1.Service) *retry.Error {
	lbBackendPoolIDs := az.getBackendPoolIDs(clusterName, pointer.StringDeref(lb.Name, ""))
	lbBackendPoolIDsToDelete := []string{}
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	if v4Enabled {
		lbBackendPoolIDsToDelete = append(lbBackendPoolIDsToDelete, lbBackendPoolIDs[consts.IPVersionIPv4])
	}
	if v6Enabled {
		lbBackendPoolIDsToDelete = append(lbBackendPoolIDsToDelete, lbBackendPoolIDs[consts.IPVersionIPv6])
	}
	if _, err := az.VMSet.EnsureBackendPoolDeleted(service, lbBackendPoolIDsToDelete, vmSetName, lb.BackendAddressPools, true); err != nil {
		return retry.NewError(false, fmt.Errorf("safeDeleteLoadBalancer: failed to EnsureBackendPoolDeleted: %w", err))
	}

	klog.V(2).Infof("safeDeleteLoadBalancer: deleting LB %s", pointer.StringDeref(lb.Name, ""))
	if rerr := az.DeleteLB(service, pointer.StringDeref(lb.Name, "")); rerr != nil {
		return rerr
	}
	_ = az.lbCache.Delete(pointer.StringDeref(lb.Name, ""))

	// Remove corresponding nodes in ActiveNodes and nodesWithCorrectLoadBalancerByPrimaryVMSet.
	for i := range az.MultipleStandardLoadBalancerConfigurations {
		if strings.EqualFold(
			trimSuffixIgnoreCase(pointer.StringDeref(lb.Name, ""), consts.InternalLoadBalancerNameSuffix),
			az.MultipleStandardLoadBalancerConfigurations[i].Name,
		) {
			for _, nodeName := range az.MultipleStandardLoadBalancerConfigurations[i].ActiveNodes.UnsortedList() {
				az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Delete(nodeName)
			}
			az.MultipleStandardLoadBalancerConfigurations[i].ActiveNodes = utilsets.NewString()
			break
		}
	}

	return nil
}

// getServiceLoadBalancer gets the loadbalancer for the service if it already exists.
// If wantLb is TRUE then -it selects a new load balancer.
// In case the selected load balancer does not exist it returns network.LoadBalancer struct
// with added metadata (such as name, location) and existsLB set to FALSE.
// By default - cluster default LB is returned.
func (az *Cloud) getServiceLoadBalancer(
	service *v1.Service,
	clusterName string,
	nodes []*v1.Node,
	wantLb bool,
	existingLBs *[]network.LoadBalancer,
) (lb *network.LoadBalancer, refreshedLBs *[]network.LoadBalancer, status *v1.LoadBalancerStatus, lbIPsPrimaryPIPs []string, exists bool, err error) {
	isInternal := requiresInternalLoadBalancer(service)
	var defaultLB *network.LoadBalancer
	primaryVMSetName := az.VMSet.GetPrimaryVMSetName()
	defaultLBName, err := az.getAzureLoadBalancerName(service, existingLBs, clusterName, primaryVMSetName, isInternal)
	if err != nil {
		return nil, nil, nil, nil, false, err
	}

	// reuse the lb list from reconcileSharedLoadBalancer to reduce the api call
	if existingLBs == nil || len(*existingLBs) == 0 {
		lbs, err := az.ListLB(service)
		if err != nil {
			return nil, nil, nil, nil, false, err
		}
		existingLBs = &lbs
	}

	// check if the service already has a load balancer
	var shouldChangeLB bool
	for i := range *existingLBs {
		existingLB := (*existingLBs)[i]

		if strings.EqualFold(*existingLB.Name, defaultLBName) {
			defaultLB = &existingLB
		}
		if isInternalLoadBalancer(&existingLB) != isInternal {
			continue
		}

		var fipConfigs []*network.FrontendIPConfiguration
		status, lbIPsPrimaryPIPs, fipConfigs, err = az.getServiceLoadBalancerStatus(service, &existingLB)
		if err != nil {
			return nil, nil, nil, nil, false, err
		}
		if status == nil {
			// service is not on this load balancer
			continue
		}
		klog.V(4).Infof("getServiceLoadBalancer(%s, %s, %v): current lb IPs: %q", service.Name, clusterName, wantLb, lbIPsPrimaryPIPs)

		// select another load balancer instead of returning
		// the current one if the change is needed
		var (
			deletedLBName string
			err           error
		)
		if wantLb && az.shouldChangeLoadBalancer(service, pointer.StringDeref(existingLB.Name, ""), clusterName, defaultLBName) {
			shouldChangeLB = true
			fipConfigNames := []string{}
			for _, fipConfig := range fipConfigs {
				fipConfigNames = append(fipConfigNames, pointer.StringDeref(fipConfig.Name, ""))
			}
			deletedLBName, err = az.removeFrontendIPConfigurationFromLoadBalancer(&existingLB, existingLBs, fipConfigs, clusterName, service)
			if err != nil {
				klog.Errorf("getServiceLoadBalancer(%s, %s, %v): failed to remove frontend IP configurations %q from load balancer: %v", service.Name, clusterName, wantLb, fipConfigNames, err)
				return nil, nil, nil, nil, false, err
			}
			if deletedLBName != "" {
				removeLBFromList(existingLBs, deletedLBName)
			}
			az.reconcileMultipleStandardLoadBalancerConfigurationStatus(
				false,
				getServiceName(service),
				pointer.StringDeref(existingLB.Name, ""),
			)

			if isLocalService(service) && az.useMultipleStandardLoadBalancers() {
				// No need for the endpoint slice informer to update the backend pool
				// for the service because the main loop will delete the old backend pool
				// and create a new one in the new load balancer.
				svcName := getServiceName(service)
				if az.backendPoolUpdater != nil {
					az.backendPoolUpdater.removeOperation(svcName)
				}

				// Remove backend pools on the previous load balancer for the local service
				if deletedLBName == "" {
					newLBs, err := az.cleanupLocalServiceBackendPool(service, nodes, existingLBs, clusterName)
					if err != nil {
						klog.Errorf("getServiceLoadBalancer(%s, %s, %v): failed to cleanup backend pool for local service: %s", service.Name, clusterName, wantLb, err.Error())
						return nil, nil, nil, nil, false, err
					}
					existingLBs = newLBs
				}
			}
			break
		}

		return &existingLB, existingLBs, status, lbIPsPrimaryPIPs, true, nil
	}

	// Service does not have a load balancer, select one.
	// Single standard load balancer doesn't need this because
	// all backends nodes should be added to same LB.
	if wantLb && !az.useStandardLoadBalancer() {
		// select new load balancer for service
		selectedLB, exists, err := az.selectLoadBalancer(clusterName, service, existingLBs, nodes)
		if err != nil {
			return nil, existingLBs, nil, nil, false, err
		}

		return selectedLB, existingLBs, status, lbIPsPrimaryPIPs, exists, err
	}

	// If the service moves to a different load balancer, return the one
	// instead of creating a new load balancer if it exists.
	if shouldChangeLB {
		for _, existingLB := range *existingLBs {
			if strings.EqualFold(pointer.StringDeref(existingLB.Name, ""), defaultLBName) {
				return &existingLB, existingLBs, status, lbIPsPrimaryPIPs, true, nil
			}
		}
	}

	// create a default LB with meta data if not present
	if defaultLB == nil {
		defaultLB = &network.LoadBalancer{
			Name:                         &defaultLBName,
			Location:                     &az.Location,
			LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{},
		}
		if az.useStandardLoadBalancer() {
			defaultLB.Sku = &network.LoadBalancerSku{
				Name: network.LoadBalancerSkuNameStandard,
			}
		}
		if az.HasExtendedLocation() {
			defaultLB.ExtendedLocation = &network.ExtendedLocation{
				Name: &az.ExtendedLocationName,
				Type: getExtendedLocationTypeFromString(az.ExtendedLocationType),
			}
		}
	}

	return defaultLB, existingLBs, nil, nil, false, nil
}

// selectLoadBalancer selects load balancer for the service in the cluster.
// The selection algorithm selects the load balancer which currently has
// the minimum lb rules. If there are multiple LBs with same number of rules,
// then selects the first one (sorted based on name).
// Note: this function is only useful for basic LB clusters.
func (az *Cloud) selectLoadBalancer(clusterName string, service *v1.Service, existingLBs *[]network.LoadBalancer, nodes []*v1.Node) (selectedLB *network.LoadBalancer, existsLb bool, err error) {
	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	klog.V(2).Infof("selectLoadBalancer for service (%s): isInternal(%v) - start", serviceName, isInternal)
	vmSetNames, err := az.VMSet.GetVMSetNames(service, nodes)
	if err != nil {
		klog.Errorf("az.selectLoadBalancer: cluster(%s) service(%s) isInternal(%t) - az.GetVMSetNames failed, err=(%v)", clusterName, serviceName, isInternal, err)
		return nil, false, err
	}
	klog.V(2).Infof("selectLoadBalancer: cluster(%s) service(%s) isInternal(%t) - vmSetNames %v", clusterName, serviceName, isInternal, *vmSetNames)

	mapExistingLBs := map[string]network.LoadBalancer{}
	for _, lb := range *existingLBs {
		mapExistingLBs[*lb.Name] = lb
	}
	selectedLBRuleCount := math.MaxInt32
	for _, currVMSetName := range *vmSetNames {
		currLBName, _ := az.getAzureLoadBalancerName(service, existingLBs, clusterName, currVMSetName, isInternal)
		lb, exists := mapExistingLBs[currLBName]
		if !exists {
			// select this LB as this is a new LB and will have minimum rules
			// create tmp lb struct to hold metadata for the new load-balancer
			var loadBalancerSKU network.LoadBalancerSkuName
			if az.useStandardLoadBalancer() {
				loadBalancerSKU = network.LoadBalancerSkuNameStandard
			} else {
				loadBalancerSKU = network.LoadBalancerSkuNameBasic
			}
			selectedLB = &network.LoadBalancer{
				Name:                         &currLBName,
				Location:                     &az.Location,
				Sku:                          &network.LoadBalancerSku{Name: loadBalancerSKU},
				LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{},
			}
			if az.HasExtendedLocation() {
				selectedLB.ExtendedLocation = &network.ExtendedLocation{
					Name: &az.ExtendedLocationName,
					Type: getExtendedLocationTypeFromString(az.ExtendedLocationType),
				}
			}

			return selectedLB, false, nil
		}

		lbRules := *lb.LoadBalancingRules
		currLBRuleCount := 0
		if lbRules != nil {
			currLBRuleCount = len(lbRules)
		}
		if currLBRuleCount < selectedLBRuleCount {
			selectedLBRuleCount = currLBRuleCount
			selectedLB = &lb
		}
	}

	if selectedLB == nil {
		err = fmt.Errorf("selectLoadBalancer: cluster(%s) service(%s) isInternal(%t) - unable to find load balancer for selected VM sets %v", clusterName, serviceName, isInternal, *vmSetNames)
		klog.Error(err)
		return nil, false, err
	}
	// validate if the selected LB has not exceeded the MaximumLoadBalancerRuleCount
	if az.Config.MaximumLoadBalancerRuleCount != 0 && selectedLBRuleCount >= az.Config.MaximumLoadBalancerRuleCount {
		err = fmt.Errorf("selectLoadBalancer: cluster(%s) service(%s) isInternal(%t) -  all available load balancers have exceeded maximum rule limit %d, vmSetNames (%v)", clusterName, serviceName, isInternal, selectedLBRuleCount, *vmSetNames)
		klog.Error(err)
		return selectedLB, existsLb, err
	}

	return selectedLB, existsLb, nil
}

// getServiceLoadBalancerStatus returns LB status for the Service.
// Before DualStack support, old logic takes the first ingress IP as non-additional one
// and the second one as additional one. With DualStack support, the second IP may be
// the IP of another IP family so the new logic returns two variables.
func (az *Cloud) getServiceLoadBalancerStatus(service *v1.Service, lb *network.LoadBalancer) (status *v1.LoadBalancerStatus, lbIPsPrimaryPIPs []string, fipConfigs []*network.FrontendIPConfiguration, err error) {
	if lb == nil {
		klog.V(10).Info("getServiceLoadBalancerStatus: lb is nil")
		return nil, nil, nil, nil
	}
	if lb.FrontendIPConfigurations == nil || len(*lb.FrontendIPConfigurations) == 0 {
		klog.V(10).Info("getServiceLoadBalancerStatus: lb.FrontendIPConfigurations is nil")
		return nil, nil, nil, nil
	}

	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	lbIngresses := []v1.LoadBalancerIngress{}
	for i := range *lb.FrontendIPConfigurations {
		ipConfiguration := (*lb.FrontendIPConfigurations)[i]
		owns, isPrimaryService, _ := az.serviceOwnsFrontendIP(ipConfiguration, service)
		if owns {
			klog.V(2).Infof("get(%s): lb(%s) - found frontend IP config, primary service: %v", serviceName, pointer.StringDeref(lb.Name, ""), isPrimaryService)

			var lbIP *string
			if isInternal {
				lbIP = ipConfiguration.PrivateIPAddress
			} else {
				if ipConfiguration.PublicIPAddress == nil {
					return nil, nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress is Nil", serviceName, *lb.Name)
				}
				pipID := ipConfiguration.PublicIPAddress.ID
				if pipID == nil {
					return nil, nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress ID is Nil", serviceName, *lb.Name)
				}
				pipName, err := getLastSegment(*pipID, "/")
				if err != nil {
					return nil, nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress Name from ID(%s)", serviceName, *lb.Name, *pipID)
				}
				pip, existsPip, err := az.getPublicIPAddress(az.getPublicIPAddressResourceGroup(service), pipName, azcache.CacheReadTypeDefault)
				if err != nil {
					return nil, nil, nil, err
				}
				if existsPip {
					lbIP = pip.IPAddress
				}
			}

			klog.V(2).Infof("getServiceLoadBalancerStatus gets ingress IP %q from frontendIPConfiguration %q for service %q", pointer.StringDeref(lbIP, ""), pointer.StringDeref(ipConfiguration.Name, ""), serviceName)

			lbIngresses = append(lbIngresses, v1.LoadBalancerIngress{IP: pointer.StringDeref(lbIP, "")})
			lbIPsPrimaryPIPs = append(lbIPsPrimaryPIPs, pointer.StringDeref(lbIP, ""))
			fipConfigs = append(fipConfigs, &ipConfiguration)
		}
	}
	if len(lbIngresses) == 0 {
		return nil, nil, nil, nil
	}

	// set additional public IPs to LoadBalancerStatus, so that kube-proxy would create their iptables rules.
	additionalIPs, err := loadbalancer.AdditionalPublicIPs(service)
	if err != nil {
		return &v1.LoadBalancerStatus{Ingress: lbIngresses}, lbIPsPrimaryPIPs, fipConfigs, err
	}
	if len(additionalIPs) > 0 {
		for _, pip := range additionalIPs {
			lbIngresses = append(lbIngresses, v1.LoadBalancerIngress{
				IP: pip.String(),
			})
		}
	}
	return &v1.LoadBalancerStatus{Ingress: lbIngresses}, lbIPsPrimaryPIPs, fipConfigs, nil
}

func (az *Cloud) determinePublicIPName(clusterName string, service *v1.Service, isIPv6 bool) (string, bool, error) {
	if name := getServicePIPName(service, isIPv6); name != "" {
		return name, true, nil
	}

	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	if id := getServicePIPPrefixID(service, isIPv6); id != "" {
		pipName, err := az.getPublicIPName(clusterName, service, isIPv6)
		return pipName, false, err
	}

	loadBalancerIP := getServiceLoadBalancerIP(service, isIPv6)

	// Assume that the service without loadBalancerIP set is a primary service.
	// If a secondary service doesn't set the loadBalancerIP, it is not allowed to share the IP.
	if len(loadBalancerIP) == 0 {
		pipName, err := az.getPublicIPName(clusterName, service, isIPv6)
		return pipName, false, err
	}

	// For the services with loadBalancerIP set, an existing public IP is required, primary
	// or secondary, or a public IP not found error would be reported.
	pip, err := az.findMatchedPIP(loadBalancerIP, "", pipResourceGroup)
	if err != nil {
		return "", false, err
	}

	if pip != nil && pip.Name != nil {
		return *pip.Name, false, nil
	}

	return "", false, fmt.Errorf("user supplied IP Address %s was not found in resource group %s", loadBalancerIP, pipResourceGroup)
}

func flipServiceInternalAnnotation(service *v1.Service) *v1.Service {
	copyService := service.DeepCopy()
	if copyService.Annotations == nil {
		copyService.Annotations = map[string]string{}
	}
	if v, ok := copyService.Annotations[consts.ServiceAnnotationLoadBalancerInternal]; ok && v == consts.TrueAnnotationValue {
		// If it is internal now, we make it external by remove the annotation
		delete(copyService.Annotations, consts.ServiceAnnotationLoadBalancerInternal)
	} else {
		// If it is external now, we make it internal
		copyService.Annotations[consts.ServiceAnnotationLoadBalancerInternal] = consts.TrueAnnotationValue
	}
	return copyService
}

func updateServiceLoadBalancerIPs(service *v1.Service, serviceIPs []string) *v1.Service {
	copyService := service.DeepCopy()
	if copyService != nil {
		for _, serviceIP := range serviceIPs {
			setServiceLoadBalancerIP(copyService, serviceIP)
		}
	}
	return copyService
}

func (az *Cloud) ensurePublicIPExists(service *v1.Service, pipName string, domainNameLabel, clusterName string, shouldPIPExisted, foundDNSLabelAnnotation, isIPv6 bool) (*network.PublicIPAddress, error) {
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	pip, existsPip, err := az.getPublicIPAddress(pipResourceGroup, pipName, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}
	serviceName := getServiceName(service)
	ipVersion := network.IPv4
	if isIPv6 {
		ipVersion = network.IPv6
	}

	var changed, owns, isUserAssignedPIP bool
	if existsPip {
		// ensure that the service tag is good for managed pips
		owns, isUserAssignedPIP = serviceOwnsPublicIP(service, &pip, clusterName)
		if owns && !isUserAssignedPIP {
			changed, err = bindServicesToPIP(&pip, []string{serviceName}, false)
			if err != nil {
				return nil, err
			}
		}

		if pip.Tags == nil {
			pip.Tags = make(map[string]*string)
		}

		// return if pip exist and dns label is the same
		if strings.EqualFold(getDomainNameLabel(&pip), domainNameLabel) {
			if existingServiceName := getServiceFromPIPDNSTags(pip.Tags); existingServiceName != "" && strings.EqualFold(existingServiceName, serviceName) {
				klog.V(6).Infof("ensurePublicIPExists for service(%s): pip(%s) - "+
					"the service is using the DNS label on the public IP", serviceName, pipName)

				var rerr *retry.Error
				if changed {
					klog.V(2).Infof("ensurePublicIPExists: updating the PIP %s for the incoming service %s", pipName, serviceName)
					err = az.CreateOrUpdatePIP(service, pipResourceGroup, pip)
					if err != nil {
						return nil, err
					}

					ctx, cancel := getContextWithCancel()
					defer cancel()
					pip, rerr = az.PublicIPAddressesClient.Get(ctx, pipResourceGroup, *pip.Name, "")
					if rerr != nil {
						return nil, rerr.Error()
					}
				}

				return &pip, nil
			}
		}

		klog.V(2).Infof("ensurePublicIPExists for service(%s): pip(%s) - updating", serviceName, pointer.StringDeref(pip.Name, ""))
		if pip.PublicIPAddressPropertiesFormat == nil {
			pip.PublicIPAddressPropertiesFormat = &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
				PublicIPAddressVersion:   ipVersion,
			}
			changed = true
		}
	} else {
		if shouldPIPExisted {
			return nil, fmt.Errorf("PublicIP from annotation azure-pip-name(-IPv6)=%s for service %s doesn't exist", pipName, serviceName)
		}

		changed = true

		pip.Name = pointer.String(pipName)
		pip.Location = pointer.String(az.Location)
		if az.HasExtendedLocation() {
			klog.V(2).Infof("Using extended location with name %s, and type %s for PIP", az.ExtendedLocationName, az.ExtendedLocationType)
			pip.ExtendedLocation = &network.ExtendedLocation{
				Name: &az.ExtendedLocationName,
				Type: getExtendedLocationTypeFromString(az.ExtendedLocationType),
			}
		}
		pip.PublicIPAddressPropertiesFormat = &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.Static,
			PublicIPAddressVersion:   ipVersion,
			IPTags:                   getServiceIPTagRequestForPublicIP(service).IPTags,
		}
		pip.Tags = map[string]*string{
			consts.ServiceTagKey:  pointer.String(""),
			consts.ClusterNameKey: &clusterName,
		}
		if _, err = bindServicesToPIP(&pip, []string{serviceName}, false); err != nil {
			return nil, err
		}

		if az.useStandardLoadBalancer() {
			pip.Sku = &network.PublicIPAddressSku{
				Name: network.PublicIPAddressSkuNameStandard,
			}

			if id := getServicePIPPrefixID(service, isIPv6); id != "" {
				pip.PublicIPPrefix = &network.SubResource{ID: pointer.String(id)}
			}

			// skip adding zone info since edge zones doesn't support multiple availability zones.
			if !az.HasExtendedLocation() {
				// only add zone information for the new standard pips
				zones, err := az.getRegionZonesBackoff(pointer.StringDeref(pip.Location, ""))
				if err != nil {
					return nil, err
				}
				if len(zones) > 0 {
					pip.Zones = &zones
				}
			}
		}
		klog.V(2).Infof("ensurePublicIPExists for service(%s): pip(%s) - creating", serviceName, *pip.Name)
	}
	if !isUserAssignedPIP && az.ensurePIPTagged(service, &pip) {
		changed = true
	}

	if foundDNSLabelAnnotation {
		updatedDNSSettings, err := reconcileDNSSettings(&pip, domainNameLabel, serviceName, pipName, isUserAssignedPIP)
		if err != nil {
			return nil, fmt.Errorf("ensurePublicIPExists for service(%s): failed to reconcileDNSSettings: %w", serviceName, err)
		}

		if updatedDNSSettings {
			changed = true
		}
	}

	// use the same family as the clusterIP as we support IPv6 single stack as well
	// as dual-stack clusters
	updatedIPSettings := az.reconcileIPSettings(&pip, service, isIPv6)
	if updatedIPSettings {
		changed = true
	}

	if changed {
		klog.V(2).Infof("CreateOrUpdatePIP(%s, %q): start", pipResourceGroup, *pip.Name)
		err = az.CreateOrUpdatePIP(service, pipResourceGroup, pip)
		if err != nil {
			klog.V(2).Infof("ensure(%s) abort backoff: pip(%s)", serviceName, *pip.Name)
			return nil, err
		}

		klog.V(10).Infof("CreateOrUpdatePIP(%s, %q): end", pipResourceGroup, *pip.Name)
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	pip, rerr := az.PublicIPAddressesClient.Get(ctx, pipResourceGroup, *pip.Name, "")
	if rerr != nil {
		return nil, rerr.Error()
	}
	return &pip, nil
}

func (az *Cloud) reconcileIPSettings(pip *network.PublicIPAddress, service *v1.Service, isIPv6 bool) bool {
	var changed bool

	serviceName := getServiceName(service)
	if isIPv6 {
		if !strings.EqualFold(string(pip.PublicIPAddressVersion), string(network.IPv6)) {
			pip.PublicIPAddressVersion = network.IPv6
			klog.V(2).Infof("service(%s): pip(%s) - should be created as IPv6", serviceName, *pip.Name)
			changed = true
		}

		if az.useStandardLoadBalancer() {
			// standard sku must have static allocation method for ipv6
			if !strings.EqualFold(string(pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod), string(network.Static)) {
				pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod = network.Static
				changed = true
			}
		} else if !strings.EqualFold(string(pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod), string(network.Dynamic)) {
			pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod = network.Dynamic
			changed = true
		}
	} else {
		if !strings.EqualFold(string(pip.PublicIPAddressVersion), string(network.IPv4)) {
			pip.PublicIPAddressVersion = network.IPv4
			klog.V(2).Infof("service(%s): pip(%s) - should be created as IPv4", serviceName, *pip.Name)
			changed = true
		}
	}

	return changed
}

func reconcileDNSSettings(
	pip *network.PublicIPAddress,
	domainNameLabel, serviceName, pipName string,
	isUserAssignedPIP bool,
) (bool, error) {
	var changed bool

	if existingServiceName := getServiceFromPIPDNSTags(pip.Tags); existingServiceName != "" && !strings.EqualFold(existingServiceName, serviceName) {
		return false, fmt.Errorf("ensurePublicIPExists for service(%s): pip(%s) - there is an existing service %s consuming the DNS label on the public IP, so the service cannot set the DNS label annotation with this value", serviceName, pipName, existingServiceName)
	}

	if len(domainNameLabel) == 0 {
		if pip.PublicIPAddressPropertiesFormat.DNSSettings != nil {
			pip.PublicIPAddressPropertiesFormat.DNSSettings = nil
			changed = true
		}
	} else {
		if pip.PublicIPAddressPropertiesFormat.DNSSettings == nil ||
			pip.PublicIPAddressPropertiesFormat.DNSSettings.DomainNameLabel == nil {
			klog.V(6).Infof("ensurePublicIPExists for service(%s): pip(%s) - no existing DNS label on the public IP, create one", serviceName, pipName)
			pip.PublicIPAddressPropertiesFormat.DNSSettings = &network.PublicIPAddressDNSSettings{
				DomainNameLabel: &domainNameLabel,
			}
			changed = true
		} else {
			existingDNSLabel := pip.PublicIPAddressPropertiesFormat.DNSSettings.DomainNameLabel
			if !strings.EqualFold(pointer.StringDeref(existingDNSLabel, ""), domainNameLabel) {
				pip.PublicIPAddressPropertiesFormat.DNSSettings.DomainNameLabel = &domainNameLabel
				changed = true
			}
		}

		if svc := getServiceFromPIPDNSTags(pip.Tags); svc == "" || !strings.EqualFold(svc, serviceName) {
			if !isUserAssignedPIP {
				pip.Tags[consts.ServiceUsingDNSKey] = &serviceName
				changed = true
			}
		}
	}

	return changed, nil
}

func getServiceFromPIPDNSTags(tags map[string]*string) string {
	v, ok := tags[consts.ServiceUsingDNSKey]
	if ok && v != nil {
		return *v
	}

	v, ok = tags[consts.LegacyServiceUsingDNSKey]
	if ok && v != nil {
		return *v
	}

	return ""
}

func deleteServicePIPDNSTags(tags *map[string]*string) {
	delete(*tags, consts.ServiceUsingDNSKey)
	delete(*tags, consts.LegacyServiceUsingDNSKey)
}

func getServiceFromPIPServiceTags(tags map[string]*string) string {
	v, ok := tags[consts.ServiceTagKey]
	if ok && v != nil {
		return *v
	}

	v, ok = tags[consts.LegacyServiceTagKey]
	if ok && v != nil {
		return *v
	}

	return ""
}

func getClusterFromPIPClusterTags(tags map[string]*string) string {
	v, ok := tags[consts.ClusterNameKey]
	if ok && v != nil {
		return *v
	}

	v, ok = tags[consts.LegacyClusterNameKey]
	if ok && v != nil {
		return *v
	}

	return ""
}

type serviceIPTagRequest struct {
	IPTagsRequestedByAnnotation bool
	IPTags                      *[]network.IPTag
}

// Get the ip tag Request for the public ip from service annotations.
func getServiceIPTagRequestForPublicIP(service *v1.Service) serviceIPTagRequest {
	if service != nil {
		if ipTagString, found := service.Annotations[consts.ServiceAnnotationIPTagsForPublicIP]; found {
			return serviceIPTagRequest{
				IPTagsRequestedByAnnotation: true,
				IPTags:                      convertIPTagMapToSlice(getIPTagMap(ipTagString)),
			}
		}
	}

	return serviceIPTagRequest{
		IPTagsRequestedByAnnotation: false,
		IPTags:                      nil,
	}
}

func getIPTagMap(ipTagString string) map[string]string {
	outputMap := make(map[string]string)
	commaDelimitedPairs := strings.Split(strings.TrimSpace(ipTagString), ",")
	for _, commaDelimitedPair := range commaDelimitedPairs {
		splitKeyValue := strings.Split(commaDelimitedPair, "=")

		// Include only valid pairs in the return value
		// Last Write wins.
		if len(splitKeyValue) == 2 {
			tagKey := strings.TrimSpace(splitKeyValue[0])
			tagValue := strings.TrimSpace(splitKeyValue[1])

			outputMap[tagKey] = tagValue
		}
	}

	return outputMap
}

func sortIPTags(ipTags *[]network.IPTag) {
	if ipTags != nil {
		sort.Slice(*ipTags, func(i, j int) bool {
			ipTag := *ipTags
			return pointer.StringDeref(ipTag[i].IPTagType, "") < pointer.StringDeref(ipTag[j].IPTagType, "") ||
				pointer.StringDeref(ipTag[i].Tag, "") < pointer.StringDeref(ipTag[j].Tag, "")
		})
	}
}

func areIPTagsEquivalent(ipTags1 *[]network.IPTag, ipTags2 *[]network.IPTag) bool {
	sortIPTags(ipTags1)
	sortIPTags(ipTags2)

	if ipTags1 == nil {
		ipTags1 = &[]network.IPTag{}
	}

	if ipTags2 == nil {
		ipTags2 = &[]network.IPTag{}
	}

	return reflect.DeepEqual(ipTags1, ipTags2)
}

func convertIPTagMapToSlice(ipTagMap map[string]string) *[]network.IPTag {
	if ipTagMap == nil {
		return nil
	}

	if len(ipTagMap) == 0 {
		return &[]network.IPTag{}
	}

	outputTags := []network.IPTag{}
	for k, v := range ipTagMap {
		ipTag := network.IPTag{
			IPTagType: pointer.String(k),
			Tag:       pointer.String(v),
		}
		outputTags = append(outputTags, ipTag)
	}

	return &outputTags
}

func getDomainNameLabel(pip *network.PublicIPAddress) string {
	if pip == nil || pip.PublicIPAddressPropertiesFormat == nil || pip.PublicIPAddressPropertiesFormat.DNSSettings == nil {
		return ""
	}
	return pointer.StringDeref(pip.PublicIPAddressPropertiesFormat.DNSSettings.DomainNameLabel, "")
}

// subnet is reused to reduce API calls when dualstack.
func (az *Cloud) isFrontendIPChanged(
	clusterName string,
	config network.FrontendIPConfiguration,
	service *v1.Service,
	lbFrontendIPConfigName string,
	subnet *network.Subnet,
) (bool, error) {
	isServiceOwnsFrontendIP, isPrimaryService, fipIPVersion := az.serviceOwnsFrontendIP(config, service)
	if isServiceOwnsFrontendIP && isPrimaryService && !strings.EqualFold(pointer.StringDeref(config.Name, ""), lbFrontendIPConfigName) {
		return true, nil
	}
	if !strings.EqualFold(pointer.StringDeref(config.Name, ""), lbFrontendIPConfigName) {
		return false, nil
	}
	pipRG := az.getPublicIPAddressResourceGroup(service)
	var isIPv6 bool
	var err error
	if fipIPVersion != "" {
		isIPv6 = fipIPVersion == network.IPv6
	} else {
		if isIPv6, err = az.isFIPIPv6(service, pipRG, &config); err != nil {
			return false, err
		}
	}
	loadBalancerIP := getServiceLoadBalancerIP(service, isIPv6)
	isInternal := requiresInternalLoadBalancer(service)
	if isInternal {
		// Judge subnet
		subnetName := getInternalSubnet(service)
		if subnetName != nil {
			if subnet == nil {
				return false, fmt.Errorf("isFrontendIPChanged: Unexpected nil subnet %q", pointer.StringDeref(subnetName, ""))
			}
			if config.Subnet != nil && !strings.EqualFold(pointer.StringDeref(config.Subnet.ID, ""), pointer.StringDeref(subnet.ID, "")) {
				return true, nil
			}
		}
		return loadBalancerIP != "" && !strings.EqualFold(loadBalancerIP, pointer.StringDeref(config.PrivateIPAddress, "")), nil
	}
	pipName, _, err := az.determinePublicIPName(clusterName, service, isIPv6)
	if err != nil {
		return false, err
	}
	pip, existsPip, err := az.getPublicIPAddress(pipRG, pipName, azcache.CacheReadTypeDefault)
	if err != nil {
		return false, err
	}
	if !existsPip {
		return true, nil
	}
	return config.PublicIPAddress != nil && !strings.EqualFold(pointer.StringDeref(pip.ID, ""), pointer.StringDeref(config.PublicIPAddress.ID, "")), nil
}

// isFrontendIPConfigUnsafeToDelete checks if a frontend IP config is safe to be deleted.
// It is safe to be deleted if and only if there is no reference from other
// loadBalancing resources, including loadBalancing rules, outbound rules, inbound NAT rules
// and inbound NAT pools.
func (az *Cloud) isFrontendIPConfigUnsafeToDelete(
	lb *network.LoadBalancer,
	service *v1.Service,
	fipConfigID *string,
) (bool, error) {
	if lb == nil || fipConfigID == nil || *fipConfigID == "" {
		return false, fmt.Errorf("isFrontendIPConfigUnsafeToDelete: incorrect parameters")
	}

	var (
		lbRules         []network.LoadBalancingRule
		outboundRules   []network.OutboundRule
		inboundNatRules []network.InboundNatRule
		inboundNatPools []network.InboundNatPool
		unsafe          bool
	)

	if lb.LoadBalancerPropertiesFormat != nil {
		if lb.LoadBalancingRules != nil {
			lbRules = *lb.LoadBalancingRules
		}
		if lb.OutboundRules != nil {
			outboundRules = *lb.OutboundRules
		}
		if lb.InboundNatRules != nil {
			inboundNatRules = *lb.InboundNatRules
		}
		if lb.InboundNatPools != nil {
			inboundNatPools = *lb.InboundNatPools
		}
	}

	// check if there are load balancing rules from other services
	// referencing this frontend IP configuration
	for _, lbRule := range lbRules {
		if lbRule.LoadBalancingRulePropertiesFormat != nil &&
			lbRule.FrontendIPConfiguration != nil &&
			lbRule.FrontendIPConfiguration.ID != nil &&
			strings.EqualFold(*lbRule.FrontendIPConfiguration.ID, *fipConfigID) {
			if !az.serviceOwnsRule(service, *lbRule.Name) {
				warningMsg := fmt.Sprintf("isFrontendIPConfigUnsafeToDelete: frontend IP configuration with ID %s on LB %s cannot be deleted because it is being referenced by load balancing rules of other services", *fipConfigID, *lb.Name)
				klog.Warning(warningMsg)
				az.Event(service, v1.EventTypeWarning, "DeletingFrontendIPConfiguration", warningMsg)
				unsafe = true
				break
			}
		}
	}

	// check if there are outbound rules
	// referencing this frontend IP configuration
	for _, outboundRule := range outboundRules {
		if outboundRule.OutboundRulePropertiesFormat != nil && outboundRule.FrontendIPConfigurations != nil {
			outboundRuleFIPConfigs := *outboundRule.FrontendIPConfigurations
			if found := findMatchedOutboundRuleFIPConfig(fipConfigID, outboundRuleFIPConfigs); found {
				warningMsg := fmt.Sprintf("isFrontendIPConfigUnsafeToDelete: frontend IP configuration with ID %s on LB %s cannot be deleted because it is being referenced by the outbound rule %s", *fipConfigID, *lb.Name, *outboundRule.Name)
				klog.Warning(warningMsg)
				az.Event(service, v1.EventTypeWarning, "DeletingFrontendIPConfiguration", warningMsg)
				unsafe = true
				break
			}
		}
	}

	// check if there are inbound NAT rules
	// referencing this frontend IP configuration
	for _, inboundNatRule := range inboundNatRules {
		if inboundNatRule.InboundNatRulePropertiesFormat != nil &&
			inboundNatRule.FrontendIPConfiguration != nil &&
			inboundNatRule.FrontendIPConfiguration.ID != nil &&
			strings.EqualFold(*inboundNatRule.FrontendIPConfiguration.ID, *fipConfigID) {
			warningMsg := fmt.Sprintf("isFrontendIPConfigUnsafeToDelete: frontend IP configuration with ID %s on LB %s cannot be deleted because it is being referenced by the inbound NAT rule %s", *fipConfigID, *lb.Name, *inboundNatRule.Name)
			klog.Warning(warningMsg)
			az.Event(service, v1.EventTypeWarning, "DeletingFrontendIPConfiguration", warningMsg)
			unsafe = true
			break
		}
	}

	// check if there are inbound NAT pools
	// referencing this frontend IP configuration
	for _, inboundNatPool := range inboundNatPools {
		if inboundNatPool.InboundNatPoolPropertiesFormat != nil &&
			inboundNatPool.FrontendIPConfiguration != nil &&
			inboundNatPool.FrontendIPConfiguration.ID != nil &&
			strings.EqualFold(*inboundNatPool.FrontendIPConfiguration.ID, *fipConfigID) {
			warningMsg := fmt.Sprintf("isFrontendIPConfigUnsafeToDelete: frontend IP configuration with ID %s on LB %s cannot be deleted because it is being referenced by the inbound NAT pool %s", *fipConfigID, *lb.Name, *inboundNatPool.Name)
			klog.Warning(warningMsg)
			az.Event(service, v1.EventTypeWarning, "DeletingFrontendIPConfiguration", warningMsg)
			unsafe = true
			break
		}
	}

	return unsafe, nil
}

func findMatchedOutboundRuleFIPConfig(fipConfigID *string, outboundRuleFIPConfigs []network.SubResource) bool {
	var found bool
	for _, config := range outboundRuleFIPConfigs {
		if config.ID != nil && strings.EqualFold(*config.ID, *fipConfigID) {
			found = true
		}
	}
	return found
}

func (az *Cloud) findFrontendIPConfigsOfService(
	fipConfigs *[]network.FrontendIPConfiguration,
	service *v1.Service,
) (map[bool]*network.FrontendIPConfiguration, error) {
	fipsOfServiceMap := map[bool]*network.FrontendIPConfiguration{}
	pipRG := az.getPublicIPAddressResourceGroup(service)
	for _, config := range *fipConfigs {
		config := config
		owns, _, fipIPVersion := az.serviceOwnsFrontendIP(config, service)
		if owns {
			var fipIsIPv6 bool
			var err error
			if fipIPVersion != "" {
				fipIsIPv6 = fipIPVersion == network.IPv6
			} else {
				if fipIsIPv6, err = az.isFIPIPv6(service, pipRG, &config); err != nil {
					return nil, err
				}
			}

			fipsOfServiceMap[fipIsIPv6] = &config
		}
	}

	return fipsOfServiceMap, nil
}

// reconcileMultipleStandardLoadBalancerConfigurations runs only once every time the
// cloud controller manager restarts or reloads itself. It checks all existing
// load balancer typed services and add service names to the ActiveServices queue
// of the corresponding load balancer configuration. It also checks if there is a configuration
// named <clustername>. If not, an error will be reported.
func (az *Cloud) reconcileMultipleStandardLoadBalancerConfigurations(
	lbs *[]network.LoadBalancer,
	service *v1.Service,
	clusterName string,
	existingLBs *[]network.LoadBalancer,
	nodes []*v1.Node,
) (err error) {
	if !az.useMultipleStandardLoadBalancers() {
		return nil
	}

	if az.multipleStandardLoadBalancerConfigurationsSynced {
		return nil
	}
	defer func() {
		if err == nil {
			az.multipleStandardLoadBalancerConfigurationsSynced = true
		}
	}()

	var found bool
	for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
		if strings.EqualFold(multiSLBConfig.Name, clusterName) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("multiple standard load balancers are enabled but no configuration named %q is found", clusterName)
	}

	svcs, err := az.KubeClient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("reconcileMultipleStandardLoadBalancerConfigurations: failed to list all load balancer services: %V", err)
		return fmt.Errorf("failed to list all load balancer services: %w", err)
	}
	rulePrefixToSVCNameMap := make(map[string]string)
	for _, svc := range svcs.Items {
		svc := svc
		if strings.EqualFold(string(svc.Spec.Type), string(v1.ServiceTypeLoadBalancer)) {
			prefix := az.GetLoadBalancerName(context.Background(), "", &svc)
			svcName := getServiceName(&svc)
			rulePrefixToSVCNameMap[strings.ToLower(prefix)] = svcName
			klog.V(2).Infof("reconcileMultipleStandardLoadBalancerConfigurations: found service %q with prefix %q", svcName, prefix)
		}
	}

	for _, existingLB := range *existingLBs {
		lbName := pointer.StringDeref(existingLB.Name, "")
		if existingLB.LoadBalancerPropertiesFormat != nil &&
			existingLB.LoadBalancingRules != nil {
			for _, rule := range *existingLB.LoadBalancingRules {
				ruleName := pointer.StringDeref(rule.Name, "")
				rulePrefix := strings.Split(ruleName, "-")[0]
				if rulePrefix == "" {
					klog.Warningf("reconcileMultipleStandardLoadBalancerConfigurations: the load balancing rule name %s is not in the correct format", ruleName)
				}
				svcName, ok := rulePrefixToSVCNameMap[strings.ToLower(rulePrefix)]
				if ok {
					klog.V(2).Infof(
						"reconcileMultipleStandardLoadBalancerConfigurations: found load balancer %q with rule %q of service %q",
						lbName, ruleName, svcName,
					)
					for i := range az.MultipleStandardLoadBalancerConfigurations {
						if strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), az.MultipleStandardLoadBalancerConfigurations[i].Name) {
							az.multipleStandardLoadBalancersActiveServicesLock.Lock()
							az.MultipleStandardLoadBalancerConfigurations[i].ActiveServices = utilsets.SafeInsert(az.MultipleStandardLoadBalancerConfigurations[i].ActiveServices, svcName)
							az.multipleStandardLoadBalancersActiveServicesLock.Unlock()
							klog.V(2).Infof("reconcileMultipleStandardLoadBalancerConfigurations: service(%s) is active on lb(%s)", svcName, lbName)
						}
					}
				}
			}
		}
	}

	return az.reconcileMultipleStandardLoadBalancerBackendNodes("", lbs, service, nodes)
}

// reconcileLoadBalancer ensures load balancer exists and the frontend ip config is setup.
// This also reconciles the Service's Ports with the LoadBalancer config.
// This entails adding rules/probes for expected Ports and removing stale rules/ports.
// nodes only used if wantLb is true
func (az *Cloud) reconcileLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node, wantLb bool) (*network.LoadBalancer, error) {
	isBackendPoolPreConfigured := az.isBackendPoolPreConfigured(service)
	serviceName := getServiceName(service)
	klog.V(2).Infof("reconcileLoadBalancer for service(%s) - wantLb(%t): started", serviceName, wantLb)

	existingLBs, err := az.ListManagedLBs(service, nodes, clusterName)
	if err != nil {
		return nil, fmt.Errorf("reconcileLoadBalancer: failed to list managed LB: %w", err)
	}

	// Delete backend pools for local service if:
	// 1. the cluster is migrating from multi-slb to single-slb,
	// 2. the service is changed from local to cluster.
	if !az.useMultipleStandardLoadBalancers() || !isLocalService(service) {
		existingLBs, err = az.cleanupLocalServiceBackendPool(service, nodes, existingLBs, clusterName)
		if err != nil {
			klog.Errorf("reconcileLoadBalancer: failed to cleanup local service backend pool for service %q, error: %s", serviceName, err.Error())
			return nil, err
		}
	}

	if err := az.reconcileMultipleStandardLoadBalancerConfigurations(existingLBs, service, clusterName, existingLBs, nodes); err != nil {
		klog.Errorf("reconcileLoadBalancer: failed to reconcile multiple standard load balancer configurations: %s", err.Error())
		return nil, err
	}

	lb, newLBs, lbStatus, _, _, err := az.getServiceLoadBalancer(service, clusterName, nodes, wantLb, existingLBs)
	if err != nil {
		klog.Errorf("reconcileLoadBalancer: failed to get load balancer for service %q, error: %v", serviceName, err)
		return nil, err
	}
	existingLBs = newLBs

	lbName := *lb.Name
	lbResourceGroup := az.getLoadBalancerResourceGroup()
	lbBackendPoolIDs := az.getBackendPoolIDsForService(service, clusterName, lbName)
	klog.V(2).Infof("reconcileLoadBalancer for service(%s): lb(%s/%s) wantLb(%t) resolved load balancer name",
		serviceName, lbResourceGroup, lbName, wantLb)
	lbFrontendIPConfigNames := az.getFrontendIPConfigNames(service)
	lbFrontendIPConfigIDs := map[bool]string{
		consts.IPVersionIPv4: az.getFrontendIPConfigID(lbName, lbFrontendIPConfigNames[consts.IPVersionIPv4]),
		consts.IPVersionIPv6: az.getFrontendIPConfigID(lbName, lbFrontendIPConfigNames[consts.IPVersionIPv6]),
	}
	dirtyLb := false

	// reconcile the load balancer's backend pool configuration.
	if wantLb {
		var (
			preConfig, backendPoolsUpdated bool
			err                            error
		)
		preConfig, backendPoolsUpdated, lb, err = az.LoadBalancerBackendPool.ReconcileBackendPools(clusterName, service, lb)
		if err != nil {
			return lb, err
		}
		if backendPoolsUpdated {
			dirtyLb = true
		}
		isBackendPoolPreConfigured = preConfig

		// If the LB is changed, refresh it to avoid etag mismatch error
		// later when create or update the LB.
		addOrUpdateLBInList(existingLBs, lb)
	}

	// reconcile the load balancer's frontend IP configurations.
	ownedFIPConfigs, toDeleteConfigs, fipChanged, err := az.reconcileFrontendIPConfigs(clusterName, service, lb, lbStatus, wantLb, lbFrontendIPConfigNames)
	if err != nil {
		return lb, err
	}
	if fipChanged {
		dirtyLb = true
	}

	// update probes/rules
	pipRG := az.getPublicIPAddressResourceGroup(service)
	for _, ownedFIPConfig := range ownedFIPConfigs {
		if ownedFIPConfig == nil {
			continue
		}
		if ownedFIPConfig.ID == nil {
			return nil, fmt.Errorf("reconcileLoadBalancer for service (%s)(%t): nil ID for frontend IP config", serviceName, wantLb)
		}

		var isIPv6 bool
		var err error
		_, _, fipIPVersion := az.serviceOwnsFrontendIP(*ownedFIPConfig, service)
		if fipIPVersion != "" {
			isIPv6 = fipIPVersion == network.IPv6
		} else {
			if isIPv6, err = az.isFIPIPv6(service, pipRG, ownedFIPConfig); err != nil {
				return nil, err
			}
		}
		lbFrontendIPConfigIDs[isIPv6] = *ownedFIPConfig.ID
	}

	var expectedProbes []network.Probe
	var expectedRules []network.LoadBalancingRule
	getExpectedLBRule := func(isIPv6 bool) error {
		expectedProbesSingleStack, expectedRulesSingleStack, err := az.getExpectedLBRules(service, lbFrontendIPConfigIDs[isIPv6], lbBackendPoolIDs[isIPv6], lbName, isIPv6)
		if err != nil {
			return err
		}
		expectedProbes = append(expectedProbes, expectedProbesSingleStack...)
		expectedRules = append(expectedRules, expectedRulesSingleStack...)
		return nil
	}
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	if wantLb && v4Enabled {
		if err = az.checkLoadBalancerResourcesConflicts(lb, lbFrontendIPConfigIDs[false], service); err != nil {
			return nil, err
		}
		if err := getExpectedLBRule(consts.IPVersionIPv4); err != nil {
			return nil, err
		}
	}
	if wantLb && v6Enabled {
		if err = az.checkLoadBalancerResourcesConflicts(lb, lbFrontendIPConfigIDs[true], service); err != nil {
			return nil, err
		}
		if err := getExpectedLBRule(consts.IPVersionIPv6); err != nil {
			return nil, err
		}
	}

	if changed := az.reconcileLBProbes(lb, service, serviceName, wantLb, expectedProbes); changed {
		dirtyLb = true
	}

	if changed := az.reconcileLBRules(lb, service, serviceName, wantLb, expectedRules); changed {
		dirtyLb = true
	}
	if changed := az.ensureLoadBalancerTagged(lb); changed {
		dirtyLb = true
	}

	// We don't care if the LB exists or not
	// We only care about if there is any change in the LB, which means dirtyLB
	// If it is not exist, and no change to that, we don't CreateOrUpdate LB
	if dirtyLb {
		if len(toDeleteConfigs) > 0 {
			for i := range toDeleteConfigs {
				fipConfigToDel := toDeleteConfigs[i]
				err := az.reconcilePrivateLinkService(clusterName, service, &fipConfigToDel, false /* wantPLS */)
				if err != nil {
					klog.Errorf(
						"reconcileLoadBalancer for service(%s): lb(%s) - failed to clean up PrivateLinkService for frontEnd(%s): %v",
						serviceName,
						lbName,
						pointer.StringDeref(fipConfigToDel.Name, ""),
						err,
					)
				}
			}
		}

		if lb.FrontendIPConfigurations == nil || len(*lb.FrontendIPConfigurations) == 0 {
			err := az.cleanOrphanedLoadBalancer(lb, *existingLBs, service, clusterName)
			if err != nil {
				klog.Errorf("reconcileLoadBalancer for service(%s): lb(%s) - failed to cleanOrphanedLoadBalancer: %v", serviceName, lbName, err)
				return nil, err
			}
		} else {
			klog.V(2).Infof("reconcileLoadBalancer: reconcileLoadBalancer for service(%s): lb(%s) - updating", serviceName, lbName)
			err := az.CreateOrUpdateLB(service, *lb)
			if err != nil {
				klog.Errorf("reconcileLoadBalancer for service(%s) abort backoff: lb(%s) - updating: %s", serviceName, lbName, err.Error())
				return nil, err
			}

			// Refresh updated lb which will be used later in other places.
			newLB, exist, err := az.getAzureLoadBalancer(lbName, azcache.CacheReadTypeDefault)
			if err != nil {
				klog.Errorf("reconcileLoadBalancer for service(%s): getAzureLoadBalancer(%s) failed: %v", serviceName, lbName, err)
				return nil, err
			}
			if !exist {
				return nil, fmt.Errorf("load balancer %q not found", lbName)
			}
			lb = newLB

			addOrUpdateLBInList(existingLBs, newLB)
		}
	}

	if wantLb && nodes != nil && !isBackendPoolPreConfigured {
		// Add the machines to the backend pool if they're not already
		vmSetName := az.mapLoadBalancerNameToVMSet(lbName, clusterName)
		// Etag would be changed when updating backend pools, so invalidate lbCache after it.
		defer func() {
			_ = az.lbCache.Delete(lbName)
		}()

		if az.useMultipleStandardLoadBalancers() {
			err := az.reconcileMultipleStandardLoadBalancerBackendNodes(lbName, existingLBs, service, nodes)
			if err != nil {
				return nil, err
			}
		}

		// Need to reconcile every managed backend pools of all managed load balancers in
		// the cluster when using multiple standard load balancers.
		// This is because there are chances for backend pools from more than one load balancers
		// change in one reconciliation loop.
		var lbToReconcile []network.LoadBalancer
		lbToReconcile = append(lbToReconcile, *lb)
		if az.useMultipleStandardLoadBalancers() {
			lbToReconcile = *existingLBs
		}
		lb, err = az.reconcileBackendPoolHosts(lb, lbToReconcile, service, nodes, clusterName, vmSetName, lbBackendPoolIDs)
		if err != nil {
			return nil, err
		}
	}

	if fipChanged {
		az.reconcileMultipleStandardLoadBalancerConfigurationStatus(wantLb, serviceName, lbName)
	}

	klog.V(2).Infof("reconcileLoadBalancer for service(%s): lb(%s) finished", serviceName, lbName)
	return lb, nil
}

func (az *Cloud) reconcileBackendPoolHosts(
	currentLB *network.LoadBalancer,
	lbs []network.LoadBalancer,
	service *v1.Service,
	nodes []*v1.Node,
	clusterName, vmSetName string,
	lbBackendPoolIDs map[bool]string,
) (*network.LoadBalancer, error) {
	var res *network.LoadBalancer
	res = currentLB
	for _, lb := range lbs {
		lb := lb
		lbName := pointer.StringDeref(lb.Name, "")
		if lb.LoadBalancerPropertiesFormat != nil && lb.LoadBalancerPropertiesFormat.BackendAddressPools != nil {
			for i, backendPool := range *lb.LoadBalancerPropertiesFormat.BackendAddressPools {
				isIPv6 := isBackendPoolIPv6(pointer.StringDeref(backendPool.Name, ""))
				if strings.EqualFold(pointer.StringDeref(backendPool.Name, ""), az.getBackendPoolNameForService(service, clusterName, isIPv6)) {
					if err := az.LoadBalancerBackendPool.EnsureHostsInPool(
						service,
						nodes,
						lbBackendPoolIDs[isIPv6],
						vmSetName,
						clusterName,
						lbName,
						(*lb.LoadBalancerPropertiesFormat.BackendAddressPools)[i],
					); err != nil {
						return nil, err
					}
				}
			}
		}
		if strings.EqualFold(lbName, *currentLB.Name) {
			res = &lb
		}
	}
	return res, nil
}

// addOrUpdateLBInList adds or updates the given lb in the list
func addOrUpdateLBInList(lbs *[]network.LoadBalancer, targetLB *network.LoadBalancer) {
	for i, lb := range *lbs {
		if strings.EqualFold(pointer.StringDeref(lb.Name, ""), pointer.StringDeref(targetLB.Name, "")) {
			(*lbs)[i] = *targetLB
			return
		}
	}
	*lbs = append(*lbs, *targetLB)
}

// removeLBFromList removes the given lb from the list
func removeLBFromList(lbs *[]network.LoadBalancer, lbName string) {
	if lbs != nil {
		for i := len(*lbs) - 1; i >= 0; i-- {
			if strings.EqualFold(pointer.StringDeref((*lbs)[i].Name, ""), lbName) {
				*lbs = append((*lbs)[:i], (*lbs)[i+1:]...)
				break
			}
		}
	}
}

// removeNodeFromLBConfig searches for the occurrence of the given node in the lb configs and removes it
func (az *Cloud) removeNodeFromLBConfig(nodeNameToLBConfigIDXMap map[string]int, nodeName string) {
	if idx, ok := nodeNameToLBConfigIDXMap[nodeName]; ok {
		currentLBConfigName := az.MultipleStandardLoadBalancerConfigurations[idx].Name
		klog.V(4).Infof("reconcileMultipleStandardLoadBalancerBackendNodes: remove node(%s) on lb(%s)", nodeName, currentLBConfigName)
		az.multipleStandardLoadBalancersActiveNodesLock.Lock()
		az.MultipleStandardLoadBalancerConfigurations[idx].ActiveNodes.Delete(strings.ToLower(nodeName))
		az.multipleStandardLoadBalancersActiveNodesLock.Unlock()
	}
}

// removeDeletedNodesFromLoadBalancerConfigurations removes the deleted nodes
// that do not exist in nodes list from the load balancer configurations
func (az *Cloud) removeDeletedNodesFromLoadBalancerConfigurations(nodes []*v1.Node) map[string]int {
	nodeNamesSet := utilsets.NewString()
	for _, node := range nodes {
		nodeNamesSet.Insert(node.Name)
	}

	az.multipleStandardLoadBalancersActiveNodesLock.Lock()
	defer az.multipleStandardLoadBalancersActiveNodesLock.Unlock()

	// Remove the nodes from the load balancer configurations if they are not in the node list.
	nodeNameToLBConfigIDXMap := make(map[string]int)
	for i, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
		if multiSLBConfig.ActiveNodes != nil {
			for _, nodeName := range multiSLBConfig.ActiveNodes.UnsortedList() {
				if nodeNamesSet.Has(nodeName) {
					nodeNameToLBConfigIDXMap[nodeName] = i
				} else {
					klog.V(4).Infof("reconcileMultipleStandardLoadBalancerBackendNodes: node(%s) is gone, remove it from lb(%s)", nodeName, multiSLBConfig.Name)
					az.MultipleStandardLoadBalancerConfigurations[i].ActiveNodes.Delete(nodeName)
				}
			}
		}
	}

	return nodeNameToLBConfigIDXMap
}

// accommodateNodesByPrimaryVMSet decides which load balancer configuration the node should be added to by primary vmSet
func (az *Cloud) accommodateNodesByPrimaryVMSet(
	lbName string,
	lbs *[]network.LoadBalancer,
	nodes []*v1.Node,
	nodeNameToLBConfigIDXMap map[string]int,
) error {
	for _, node := range nodes {
		if _, ok := az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Load(strings.ToLower(node.Name)); ok {
			continue
		}

		// TODO(niqi): reduce the API calls for VMAS and standalone VMs
		vmSetName, err := az.VMSet.GetNodeVMSetName(node)
		if err != nil {
			klog.Errorf("accommodateNodesByPrimaryVMSet: failed to get vmSetName for node(%s): %s", node.Name, err.Error())
			return err
		}
		for i := range az.MultipleStandardLoadBalancerConfigurations {
			multiSLBConfig := az.MultipleStandardLoadBalancerConfigurations[i]
			if strings.EqualFold(multiSLBConfig.PrimaryVMSet, vmSetName) {
				foundPrimaryLB := isLBInList(lbs, multiSLBConfig.Name)
				if !foundPrimaryLB && !strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), multiSLBConfig.Name) {
					klog.V(4).Infof("accommodateNodesByPrimaryVMSet: node(%s) should be on lb(%s) because of primary vmSet (%s), but the lb is not found and will not be created this time, will ignore the primaryVMSet", node.Name, multiSLBConfig.Name, vmSetName)
					continue
				}

				az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Store(strings.ToLower(node.Name), struct{}{})
				if !multiSLBConfig.ActiveNodes.Has(node.Name) {
					klog.V(4).Infof("accommodateNodesByPrimaryVMSet: node(%s) should be on lb(%s) because of primary vmSet (%s)", node.Name, multiSLBConfig.Name, vmSetName)

					az.removeNodeFromLBConfig(nodeNameToLBConfigIDXMap, node.Name)

					az.multipleStandardLoadBalancersActiveNodesLock.Lock()
					az.MultipleStandardLoadBalancerConfigurations[i].ActiveNodes = utilsets.SafeInsert(az.MultipleStandardLoadBalancerConfigurations[i].ActiveNodes, node.Name)
					az.multipleStandardLoadBalancersActiveNodesLock.Unlock()
				}
				break
			}
		}
	}

	return nil
}

// accommodateNodesByNodeSelector decides which load balancer configuration the node should be added to by node selector
func (az *Cloud) accommodateNodesByNodeSelector(
	lbName string,
	lbs *[]network.LoadBalancer,
	service *v1.Service,
	nodes []*v1.Node,
	nodeNameToLBConfigIDXMap map[string]int,
) error {
	for _, node := range nodes {
		// Skip nodes that have been matched with a load balancer
		// by primary vmSet.
		if _, ok := az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Load(strings.ToLower(node.Name)); ok {
			continue
		}

		// If the vmSet of the node does not match any load balancer,
		// pick all load balancers whose node selector matches the node.
		var eligibleLBsIDX []int
		for i, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
			if multiSLBConfig.NodeSelector != nil &&
				(len(multiSLBConfig.NodeSelector.MatchLabels) > 0 || len(multiSLBConfig.NodeSelector.MatchExpressions) > 0) {
				nodeSelector, err := metav1.LabelSelectorAsSelector(multiSLBConfig.NodeSelector)
				if err != nil {
					klog.Errorf("accommodateNodesByNodeSelector: failed to parse nodeSelector for lb(%s): %s", multiSLBConfig.Name, err.Error())
					return err
				}
				if nodeSelector.Matches(labels.Set(node.Labels)) {
					klog.V(4).Infof("accommodateNodesByNodeSelector: lb(%s) matches node(%s) labels", multiSLBConfig.Name, node.Name)
					found := isLBInList(lbs, multiSLBConfig.Name)
					if !found && !strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), multiSLBConfig.Name) {
						klog.V(4).Infof("accommodateNodesByNodeSelector: but the lb is not found and will not be created this time, will ignore this load balancer")
						continue
					}
					eligibleLBsIDX = append(eligibleLBsIDX, i)
				}
			}
		}
		// If no load balancer is matched, all load balancers without node selector are eligible.
		if len(eligibleLBsIDX) == 0 {
			for i, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
				if multiSLBConfig.NodeSelector == nil {
					eligibleLBsIDX = append(eligibleLBsIDX, i)
				}
			}
		}
		// Check if the valid load balancer exists or will exist
		// after the reconciliation.
		for i := len(eligibleLBsIDX) - 1; i >= 0; i-- {
			multiSLBConfig := az.MultipleStandardLoadBalancerConfigurations[eligibleLBsIDX[i]]
			found := isLBInList(lbs, multiSLBConfig.Name)
			if !found && !strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), multiSLBConfig.Name) {
				klog.V(4).Infof("accommodateNodesByNodeSelector: the load balancer %s is a valid placement target for node %s, but the lb is not found and will not be created this time, ignore this load balancer", multiSLBConfig.Name, node.Name)
				eligibleLBsIDX = append(eligibleLBsIDX[:i], eligibleLBsIDX[i+1:]...)
			}
		}
		if idx, ok := nodeNameToLBConfigIDXMap[node.Name]; ok {
			if IntInSlice(idx, eligibleLBsIDX) {
				klog.V(4).Infof("accommodateNodesByNodeSelector: node(%s) is already on the eligible lb(%s)", node.Name, az.MultipleStandardLoadBalancerConfigurations[idx].Name)
				continue
			}
		}

		// Pick one with the fewest nodes among all eligible load balancers.
		minNodesIDX := -1
		minNodes := math.MaxInt32
		az.multipleStandardLoadBalancersActiveNodesLock.Lock()
		for _, idx := range eligibleLBsIDX {
			multiSLBConfig := az.MultipleStandardLoadBalancerConfigurations[idx]
			if multiSLBConfig.ActiveNodes.Len() < minNodes {
				minNodes = multiSLBConfig.ActiveNodes.Len()
				minNodesIDX = idx
			}
		}
		az.multipleStandardLoadBalancersActiveNodesLock.Unlock()

		if idx, ok := nodeNameToLBConfigIDXMap[node.Name]; ok && idx != minNodesIDX {
			az.removeNodeFromLBConfig(nodeNameToLBConfigIDXMap, node.Name)
		}

		// Emit a warning for the orphaned node.
		if minNodesIDX == -1 {
			warningMsg := fmt.Sprintf("failed to find a lb for node %s", node.Name)
			az.Event(service, v1.EventTypeWarning, "FailedToFindLoadBalancerForNode", warningMsg)
			continue
		}

		klog.V(4).Infof("accommodateNodesByNodeSelector: node(%s) should be on lb(%s) it is the eligible LB with fewest number of nodes", node.Name, az.MultipleStandardLoadBalancerConfigurations[minNodesIDX].Name)
		az.multipleStandardLoadBalancersActiveNodesLock.Lock()
		az.MultipleStandardLoadBalancerConfigurations[minNodesIDX].ActiveNodes = utilsets.SafeInsert(az.MultipleStandardLoadBalancerConfigurations[minNodesIDX].ActiveNodes, node.Name)
		az.multipleStandardLoadBalancersActiveNodesLock.Unlock()
	}

	return nil
}

// isLBInList checks if the lb is in the list by multipleStandardLoadBalancerConfig name
func isLBInList(lbs *[]network.LoadBalancer, lbConfigName string) bool {
	if lbs != nil {
		for _, lb := range *lbs {
			if strings.EqualFold(trimSuffixIgnoreCase(pointer.StringDeref(lb.Name, ""), consts.InternalLoadBalancerNameSuffix), lbConfigName) {
				return true
			}
		}
	}
	return false
}

// reconcileMultipleStandardLoadBalancerBackendNodes makes sure the arrangement of nodes
// across load balancer configurations is expected. This is used in two places:
// 1. Every time the cloud provide restarts.
// 2. Every time we ensure hosts in pool.
// It consists of two parts. First we put corresponding nodes to the load balancers
// whose primary vmSet matches the node. Then we put the rest of the nodes to the
// most eligible load balancers according to the node selector and the number of
// nodes currently in the load balancer.
// For availability set (no cache) amd vmss flex (with cache) clusters,
// a list call will be introduced every time we
// try to get the vmSet of a node. This is acceptable because of two reasons:
// 1. In AKS, we don't support multiple availability sets in a cluster so the
// cluster scale is small. For self-managed clusters, it is not recommended to
// use multiple standard load balancers with availability sets.
// 2. We only check nodes that are not matched by primary vmSet before we ensure
// hosts in pool. So the number API calls is under control.
func (az *Cloud) reconcileMultipleStandardLoadBalancerBackendNodes(
	lbName string,
	lbs *[]network.LoadBalancer,
	service *v1.Service,
	nodes []*v1.Node,
) error {
	// Remove the nodes from the load balancer configurations if they are not in the node list.
	nodeNameToLBConfigIDXMap := az.removeDeletedNodesFromLoadBalancerConfigurations(nodes)

	err := az.accommodateNodesByPrimaryVMSet(lbName, lbs, nodes, nodeNameToLBConfigIDXMap)
	if err != nil {
		return err
	}

	err = az.accommodateNodesByNodeSelector(lbName, lbs, service, nodes, nodeNameToLBConfigIDXMap)
	if err != nil {
		return err
	}

	return nil
}

func (az *Cloud) reconcileMultipleStandardLoadBalancerConfigurationStatus(wantLb bool, svcName, lbName string) {
	lbName = trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix)
	for i := range az.MultipleStandardLoadBalancerConfigurations {
		if strings.EqualFold(lbName, az.MultipleStandardLoadBalancerConfigurations[i].Name) {
			az.multipleStandardLoadBalancersActiveServicesLock.Lock()

			if wantLb {
				klog.V(4).Infof("reconcileMultipleStandardLoadBalancerConfigurationStatus: service(%s) is active on lb(%s)", svcName, lbName)
				az.MultipleStandardLoadBalancerConfigurations[i].ActiveServices = utilsets.SafeInsert(az.MultipleStandardLoadBalancerConfigurations[i].ActiveServices, svcName)
			} else {
				klog.V(4).Infof("reconcileMultipleStandardLoadBalancerConfigurationStatus: service(%s) is not active on lb(%s) any more", svcName, lbName)
				az.MultipleStandardLoadBalancerConfigurations[i].ActiveServices.Delete(svcName)
			}
			az.multipleStandardLoadBalancersActiveServicesLock.Unlock()
			break
		}
	}
}

func (az *Cloud) reconcileLBProbes(lb *network.LoadBalancer, service *v1.Service, serviceName string, wantLb bool, expectedProbes []network.Probe) bool {
	expectedProbes, _ = az.keepSharedProbe(service, *lb, expectedProbes, wantLb)

	// remove unwanted probes
	dirtyProbes := false
	var updatedProbes []network.Probe
	if lb.Probes != nil {
		updatedProbes = *lb.Probes
	}
	for i := len(updatedProbes) - 1; i >= 0; i-- {
		existingProbe := updatedProbes[i]
		if az.serviceOwnsRule(service, *existingProbe.Name) {
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb probe(%s) - considering evicting", serviceName, wantLb, *existingProbe.Name)
			keepProbe := false
			if findProbe(expectedProbes, existingProbe) {
				klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb probe(%s) - keeping", serviceName, wantLb, *existingProbe.Name)
				keepProbe = true
			}
			if !keepProbe {
				updatedProbes = append(updatedProbes[:i], updatedProbes[i+1:]...)
				klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb probe(%s) - dropping", serviceName, wantLb, *existingProbe.Name)
				dirtyProbes = true
			}
		}
	}
	// add missing, wanted probes
	for _, expectedProbe := range expectedProbes {
		foundProbe := false
		if findProbe(updatedProbes, expectedProbe) {
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb probe(%s) - already exists", serviceName, wantLb, *expectedProbe.Name)
			foundProbe = true
		}
		if !foundProbe {
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb probe(%s) - adding", serviceName, wantLb, *expectedProbe.Name)
			updatedProbes = append(updatedProbes, expectedProbe)
			dirtyProbes = true
		}
	}
	if dirtyProbes {
		probesJSON, _ := json.Marshal(expectedProbes)
		klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb probes updated: %s", serviceName, wantLb, string(probesJSON))
		lb.Probes = &updatedProbes
	}
	return dirtyProbes
}

func (az *Cloud) reconcileLBRules(lb *network.LoadBalancer, service *v1.Service, serviceName string, wantLb bool, expectedRules []network.LoadBalancingRule) bool {
	// update rules
	dirtyRules := false
	var updatedRules []network.LoadBalancingRule
	if lb.LoadBalancingRules != nil {
		updatedRules = *lb.LoadBalancingRules
	}

	// update rules: remove unwanted
	for i := len(updatedRules) - 1; i >= 0; i-- {
		existingRule := updatedRules[i]
		if az.serviceOwnsRule(service, *existingRule.Name) {
			keepRule := false
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb rule(%s) - considering evicting", serviceName, wantLb, *existingRule.Name)
			if findRule(expectedRules, existingRule, wantLb) {
				klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb rule(%s) - keeping", serviceName, wantLb, *existingRule.Name)
				keepRule = true
			}
			if !keepRule {
				klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb rule(%s) - dropping", serviceName, wantLb, *existingRule.Name)
				updatedRules = append(updatedRules[:i], updatedRules[i+1:]...)
				dirtyRules = true
			}
		}
	}
	// update rules: add needed
	for _, expectedRule := range expectedRules {
		foundRule := false
		if findRule(updatedRules, expectedRule, wantLb) {
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb rule(%s) - already exists", serviceName, wantLb, *expectedRule.Name)
			foundRule = true
		}
		if !foundRule {
			klog.V(10).Infof("reconcileLoadBalancer for service (%s)(%t): lb rule(%s) adding", serviceName, wantLb, *expectedRule.Name)
			updatedRules = append(updatedRules, expectedRule)
			dirtyRules = true
		}
	}
	if dirtyRules {
		ruleJSON, _ := json.Marshal(expectedRules)
		klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb rules updated: %s", serviceName, wantLb, string(ruleJSON))
		lb.LoadBalancingRules = &updatedRules
	}
	return dirtyRules
}

func (az *Cloud) reconcileFrontendIPConfigs(clusterName string,
	service *v1.Service,
	lb *network.LoadBalancer,
	status *v1.LoadBalancerStatus,
	wantLb bool,
	lbFrontendIPConfigNames map[bool]string) ([]*network.FrontendIPConfiguration, []network.FrontendIPConfiguration, bool, error) {
	var err error
	lbName := *lb.Name
	serviceName := getServiceName(service)
	isInternal := requiresInternalLoadBalancer(service)
	dirtyConfigs := false
	var newConfigs []network.FrontendIPConfiguration
	var toDeleteConfigs []network.FrontendIPConfiguration
	if lb.FrontendIPConfigurations != nil {
		newConfigs = *lb.FrontendIPConfigurations
	}

	var ownedFIPConfigs []*network.FrontendIPConfiguration
	if !wantLb {
		for i := len(newConfigs) - 1; i >= 0; i-- {
			config := newConfigs[i]
			isServiceOwnsFrontendIP, _, _ := az.serviceOwnsFrontendIP(config, service)
			if isServiceOwnsFrontendIP {
				unsafe, err := az.isFrontendIPConfigUnsafeToDelete(lb, service, config.ID)
				if err != nil {
					return nil, toDeleteConfigs, false, err
				}

				// If the frontend IP configuration is not being referenced by:
				// 1. loadBalancing rules of other services with different ports;
				// 2. outbound rules;
				// 3. inbound NAT rules;
				// 4. inbound NAT pools,
				// do the deletion, or skip it.
				if !unsafe {
					var configNameToBeDeleted string
					if newConfigs[i].Name != nil {
						configNameToBeDeleted = *newConfigs[i].Name
						klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb frontendconfig(%s) - dropping", serviceName, wantLb, configNameToBeDeleted)
					} else {
						klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): nil name of lb frontendconfig", serviceName, wantLb)
					}

					toDeleteConfigs = append(toDeleteConfigs, newConfigs[i])
					newConfigs = append(newConfigs[:i], newConfigs[i+1:]...)
					dirtyConfigs = true
				}
			}
		}
	} else {
		var (
			previousZone *[]string
			isFipChanged bool
			subnet       network.Subnet
			existsSubnet bool
		)

		if isInternal {
			subnetName := getInternalSubnet(service)
			if subnetName == nil {
				subnetName = &az.SubnetName
			}
			subnet, existsSubnet, err = az.getSubnet(az.VnetName, *subnetName)
			if err != nil {
				return nil, toDeleteConfigs, false, err
			}
			if !existsSubnet {
				return nil, toDeleteConfigs, false, fmt.Errorf("ensure(%s): lb(%s) - failed to get subnet: %s/%s", serviceName, lbName, az.VnetName, *subnetName)
			}
		}

		pipRG := az.getPublicIPAddressResourceGroup(service)

		for i := len(newConfigs) - 1; i >= 0; i-- {
			config := newConfigs[i]
			isServiceOwnsFrontendIP, _, fipIPVersion := az.serviceOwnsFrontendIP(config, service)
			if !isServiceOwnsFrontendIP {
				klog.V(4).Infof("reconcileFrontendIPConfigs for service (%s): the frontend IP configuration %s does not belong to the service", serviceName, pointer.StringDeref(config.Name, ""))
				continue
			}
			klog.V(4).Infof("reconcileFrontendIPConfigs for service (%s): checking owned frontend IP configuration %s", serviceName, pointer.StringDeref(config.Name, ""))
			var isIPv6 bool
			var err error
			if fipIPVersion != "" {
				isIPv6 = fipIPVersion == network.IPv6
			} else {
				if isIPv6, err = az.isFIPIPv6(service, pipRG, &config); err != nil {
					return nil, toDeleteConfigs, false, err
				}
			}

			isFipChanged, err = az.isFrontendIPChanged(clusterName, config, service, lbFrontendIPConfigNames[isIPv6], &subnet)
			if err != nil {
				return nil, toDeleteConfigs, false, err
			}
			if isFipChanged {
				klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb frontendconfig(%s) - dropping", serviceName, wantLb, *config.Name)
				toDeleteConfigs = append(toDeleteConfigs, newConfigs[i])
				newConfigs = append(newConfigs[:i], newConfigs[i+1:]...)
				dirtyConfigs = true
				previousZone = config.Zones
			}
		}

		ownedFIPConfigMap, err := az.findFrontendIPConfigsOfService(&newConfigs, service)
		if err != nil {
			return nil, toDeleteConfigs, false, err
		}
		for _, config := range ownedFIPConfigMap {
			ownedFIPConfigs = append(ownedFIPConfigs, config)
		}

		addNewFIPOfService := func(isIPv6 bool) error {
			klog.V(4).Infof("ensure(%s): lb(%s) - creating a new frontend IP config %q (isIPv6=%t)",
				serviceName, lbName, lbFrontendIPConfigNames[isIPv6], isIPv6)

			// construct FrontendIPConfigurationPropertiesFormat
			var fipConfigurationProperties *network.FrontendIPConfigurationPropertiesFormat
			if isInternal {
				configProperties := network.FrontendIPConfigurationPropertiesFormat{
					Subnet: &subnet,
				}

				if isIPv6 {
					configProperties.PrivateIPAddressVersion = network.IPv6
				}

				loadBalancerIP := getServiceLoadBalancerIP(service, isIPv6)
				privateIP := ""
				ingressIPInSubnet := func(ingresses []v1.LoadBalancerIngress) bool {
					for _, ingress := range ingresses {
						ingressIP := ingress.IP
						if (net.ParseIP(ingressIP).To4() == nil) == isIPv6 && ipInSubnet(ingressIP, &subnet) {
							privateIP = ingressIP
							break
						}
					}
					return privateIP != ""
				}
				if loadBalancerIP != "" {
					klog.V(4).Infof("reconcileFrontendIPConfigs for service (%s): use loadBalancerIP %q from Service spec", serviceName, loadBalancerIP)
					configProperties.PrivateIPAllocationMethod = network.Static
					configProperties.PrivateIPAddress = &loadBalancerIP
				} else if status != nil && len(status.Ingress) > 0 && ingressIPInSubnet(status.Ingress) {
					klog.V(4).Infof("reconcileFrontendIPConfigs for service (%s): keep the original private IP %s", serviceName, privateIP)
					configProperties.PrivateIPAllocationMethod = network.Static
					configProperties.PrivateIPAddress = pointer.String(privateIP)
				} else {
					// We'll need to call GetLoadBalancer later to retrieve allocated IP.
					klog.V(4).Infof("reconcileFrontendIPConfigs for service (%s): dynamically allocate the private IP", serviceName)
					configProperties.PrivateIPAllocationMethod = network.Dynamic
				}

				fipConfigurationProperties = &configProperties
			} else {
				pipName, shouldPIPExisted, err := az.determinePublicIPName(clusterName, service, isIPv6)
				if err != nil {
					return err
				}
				domainNameLabel, found := getPublicIPDomainNameLabel(service)
				pip, err := az.ensurePublicIPExists(service, pipName, domainNameLabel, clusterName, shouldPIPExisted, found, isIPv6)
				if err != nil {
					return err
				}
				fipConfigurationProperties = &network.FrontendIPConfigurationPropertiesFormat{
					PublicIPAddress: &network.PublicIPAddress{ID: pip.ID},
				}
			}

			newConfig := network.FrontendIPConfiguration{
				Name:                                    pointer.String(lbFrontendIPConfigNames[isIPv6]),
				ID:                                      pointer.String(fmt.Sprintf(consts.FrontendIPConfigIDTemplate, az.getNetworkResourceSubscriptionID(), az.ResourceGroup, pointer.StringDeref(lb.Name, ""), lbFrontendIPConfigNames[isIPv6])),
				FrontendIPConfigurationPropertiesFormat: fipConfigurationProperties,
			}

			if isInternal {
				if err := az.getFrontendZones(&newConfig, previousZone, isFipChanged, serviceName, lbFrontendIPConfigNames[isIPv6]); err != nil {
					klog.Errorf("reconcileLoadBalancer for service (%s)(%t): failed to getFrontendZones: %s", serviceName, wantLb, err.Error())
					return err
				}
			}
			newConfigs = append(newConfigs, newConfig)
			klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb frontendconfig(%s) - adding", serviceName, wantLb, lbFrontendIPConfigNames[isIPv6])
			dirtyConfigs = true
			return nil
		}

		v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
		if v4Enabled && ownedFIPConfigMap[false] == nil {
			if err := addNewFIPOfService(false); err != nil {
				return nil, toDeleteConfigs, false, err
			}
		}
		if v6Enabled && ownedFIPConfigMap[true] == nil {
			if err := addNewFIPOfService(true); err != nil {
				return nil, toDeleteConfigs, false, err
			}
		}
	}

	if dirtyConfigs {
		lb.FrontendIPConfigurations = &newConfigs
	}

	return ownedFIPConfigs, toDeleteConfigs, dirtyConfigs, err
}

func (az *Cloud) getFrontendZones(
	fipConfig *network.FrontendIPConfiguration,
	previousZone *[]string,
	isFipChanged bool,
	serviceName, lbFrontendIPConfigName string) error {
	if !isFipChanged { // fetch zone information from API for new frontends
		// only add zone information for new internal frontend IP configurations for standard load balancer not deployed to an edge zone.
		location := az.Location
		zones, err := az.getRegionZonesBackoff(location)
		if err != nil {
			return err
		}
		if az.useStandardLoadBalancer() && len(zones) > 0 && !az.HasExtendedLocation() {
			fipConfig.Zones = &zones
		}
	} else {
		if previousZone == nil { // keep the existing zone information for existing frontends
			klog.V(2).Infof("getFrontendZones for service (%s): lb frontendconfig(%s): setting zone to nil", serviceName, lbFrontendIPConfigName)
		} else {
			zoneStr := strings.Join(*previousZone, ",")
			klog.V(2).Infof("getFrontendZones for service (%s): lb frontendconfig(%s): setting zone to %s", serviceName, lbFrontendIPConfigName, zoneStr)
		}
		fipConfig.Zones = previousZone
	}
	return nil
}

// checkLoadBalancerResourcesConflicts checks if the service is consuming
// ports which conflict with the existing loadBalancer resources,
// including inbound NAT rule, inbound NAT pools and loadBalancing rules
func (az *Cloud) checkLoadBalancerResourcesConflicts(
	lb *network.LoadBalancer,
	frontendIPConfigID string,
	service *v1.Service,
) error {
	if service.Spec.Ports == nil {
		return nil
	}
	ports := service.Spec.Ports

	for _, port := range ports {
		if lb.LoadBalancingRules != nil {
			for _, rule := range *lb.LoadBalancingRules {
				if lbRuleConflictsWithPort(rule, frontendIPConfigID, port) {
					// ignore self-owned rules for unit test
					if rule.Name != nil && az.serviceOwnsRule(service, *rule.Name) {
						continue
					}
					return fmt.Errorf("checkLoadBalancerResourcesConflicts: service port %s is trying to "+
						"consume the port %d which is being referenced by an existing loadBalancing rule %s with "+
						"the same protocol %s and frontend IP config with ID %s",
						port.Name,
						*rule.FrontendPort,
						*rule.Name,
						rule.Protocol,
						*rule.FrontendIPConfiguration.ID)
				}
			}
		}

		if lb.InboundNatRules != nil {
			for _, inboundNatRule := range *lb.InboundNatRules {
				if inboundNatRuleConflictsWithPort(inboundNatRule, frontendIPConfigID, port) {
					return fmt.Errorf("checkLoadBalancerResourcesConflicts: service port %s is trying to "+
						"consume the port %d which is being referenced by an existing inbound NAT rule %s with "+
						"the same protocol %s and frontend IP config with ID %s",
						port.Name,
						*inboundNatRule.FrontendPort,
						*inboundNatRule.Name,
						inboundNatRule.Protocol,
						*inboundNatRule.FrontendIPConfiguration.ID)
				}
			}
		}

		if lb.InboundNatPools != nil {
			for _, pool := range *lb.InboundNatPools {
				if inboundNatPoolConflictsWithPort(pool, frontendIPConfigID, port) {
					return fmt.Errorf("checkLoadBalancerResourcesConflicts: service port %s is trying to "+
						"consume the port %d which is being in the range (%d-%d) of an existing "+
						"inbound NAT pool %s with the same protocol %s and frontend IP config with ID %s",
						port.Name,
						port.Port,
						*pool.FrontendPortRangeStart,
						*pool.FrontendPortRangeEnd,
						*pool.Name,
						pool.Protocol,
						*pool.FrontendIPConfiguration.ID)
				}
			}
		}
	}

	return nil
}

func inboundNatPoolConflictsWithPort(pool network.InboundNatPool, frontendIPConfigID string, port v1.ServicePort) bool {
	return pool.InboundNatPoolPropertiesFormat != nil &&
		pool.FrontendIPConfiguration != nil &&
		pool.FrontendIPConfiguration.ID != nil &&
		strings.EqualFold(*pool.FrontendIPConfiguration.ID, frontendIPConfigID) &&
		strings.EqualFold(string(pool.Protocol), string(port.Protocol)) &&
		pool.FrontendPortRangeStart != nil &&
		pool.FrontendPortRangeEnd != nil &&
		*pool.FrontendPortRangeStart <= port.Port &&
		*pool.FrontendPortRangeEnd >= port.Port
}

func inboundNatRuleConflictsWithPort(inboundNatRule network.InboundNatRule, frontendIPConfigID string, port v1.ServicePort) bool {
	return inboundNatRule.InboundNatRulePropertiesFormat != nil &&
		inboundNatRule.FrontendIPConfiguration != nil &&
		inboundNatRule.FrontendIPConfiguration.ID != nil &&
		strings.EqualFold(*inboundNatRule.FrontendIPConfiguration.ID, frontendIPConfigID) &&
		strings.EqualFold(string(inboundNatRule.Protocol), string(port.Protocol)) &&
		inboundNatRule.FrontendPort != nil &&
		*inboundNatRule.FrontendPort == port.Port
}

func lbRuleConflictsWithPort(rule network.LoadBalancingRule, frontendIPConfigID string, port v1.ServicePort) bool {
	return rule.LoadBalancingRulePropertiesFormat != nil &&
		rule.FrontendIPConfiguration != nil &&
		rule.FrontendIPConfiguration.ID != nil &&
		strings.EqualFold(*rule.FrontendIPConfiguration.ID, frontendIPConfigID) &&
		strings.EqualFold(string(rule.Protocol), string(port.Protocol)) &&
		rule.FrontendPort != nil &&
		*rule.FrontendPort == port.Port
}

// buildLBRules
// for following sku: basic loadbalancer vs standard load balancer
// for following scenario: internal vs external
func (az *Cloud) getExpectedLBRules(
	service *v1.Service,
	lbFrontendIPConfigID string,
	lbBackendPoolID string,
	lbName string,
	isIPv6 bool) ([]network.Probe, []network.LoadBalancingRule, error) {

	var expectedRules []network.LoadBalancingRule
	var expectedProbes []network.Probe

	// support podPresence health check when External Traffic Policy is local
	// take precedence over user defined probe configuration
	// healthcheck proxy server serves http requests
	// https://github.com/kubernetes/kubernetes/blob/7c013c3f64db33cf19f38bb2fc8d9182e42b0b7b/pkg/proxy/healthcheck/service_health.go#L236
	var nodeEndpointHealthprobe *network.Probe
	var nodeEndpointHealthprobeAdded bool
	if servicehelpers.NeedsHealthCheck(service) && !(consts.IsPLSEnabled(service.Annotations) && consts.IsPLSProxyProtocolEnabled(service.Annotations)) {
		podPresencePath, podPresencePort := servicehelpers.GetServiceHealthCheckPathPort(service)
		lbRuleName := az.getLoadBalancerRuleName(service, v1.ProtocolTCP, podPresencePort, isIPv6)
		probeInterval, numberOfProbes, err := az.getHealthProbeConfigProbeIntervalAndNumOfProbe(service, podPresencePort)
		if err != nil {
			return nil, nil, err
		}
		nodeEndpointHealthprobe = &network.Probe{
			Name: &lbRuleName,
			ProbePropertiesFormat: &network.ProbePropertiesFormat{
				RequestPath:       pointer.String(podPresencePath),
				Protocol:          network.ProbeProtocolHTTP,
				Port:              pointer.Int32(podPresencePort),
				IntervalInSeconds: probeInterval,
				ProbeThreshold:    numberOfProbes,
			},
		}
	}

	var useSharedProbe bool
	if az.useSharedLoadBalancerHealthProbeMode() &&
		!strings.EqualFold(string(service.Spec.ExternalTrafficPolicy), string(v1.ServiceExternalTrafficPolicyLocal)) {
		nodeEndpointHealthprobe = az.buildClusterServiceSharedProbe()
		useSharedProbe = true
	}

	// In HA mode, lb forward traffic of all port to backend
	// HA mode is only supported on standard loadbalancer SKU in internal mode
	if consts.IsK8sServiceUsingInternalLoadBalancer(service) &&
		az.useStandardLoadBalancer() &&
		consts.IsK8sServiceHasHAModeEnabled(service) {

		lbRuleName := az.getloadbalancerHAmodeRuleName(service, isIPv6)
		klog.V(2).Infof("getExpectedLBRules lb name (%s) rule name (%s)", lbName, lbRuleName)

		props, err := az.getExpectedHAModeLoadBalancingRuleProperties(service, lbFrontendIPConfigID, lbBackendPoolID)
		if err != nil {
			return nil, nil, fmt.Errorf("error generate lb rule for ha mod loadbalancer. err: %w", err)
		}
		//Here we need to find one health probe rule for the HA lb rule.
		if nodeEndpointHealthprobe == nil {
			// use user customized health probe rule if any
			for _, port := range service.Spec.Ports {
				portprobe, err := az.buildHealthProbeRulesForPort(service, port, lbRuleName, nil, false)
				if err != nil {
					klog.V(2).ErrorS(err, "error occurred when buildHealthProbeRulesForPort", "service", service.Name, "namespace", service.Namespace,
						"rule-name", lbRuleName, "port", port.Port)
					//ignore error because we only need one correct rule
				}
				if portprobe != nil {
					props.Probe = &network.SubResource{
						ID: pointer.String(az.getLoadBalancerProbeID(lbName, *portprobe.Name)),
					}
					expectedProbes = append(expectedProbes, *portprobe)
					break
				}
			}
		} else {
			props.Probe = &network.SubResource{
				ID: pointer.String(az.getLoadBalancerProbeID(lbName, *nodeEndpointHealthprobe.Name)),
			}
			expectedProbes = append(expectedProbes, *nodeEndpointHealthprobe)
		}

		expectedRules = append(expectedRules, network.LoadBalancingRule{
			Name:                              &lbRuleName,
			LoadBalancingRulePropertiesFormat: props,
		})
		// end of HA mode handling
	} else {
		// generate lb rule for each port defined in svc object

		for _, port := range service.Spec.Ports {
			lbRuleName := az.getLoadBalancerRuleName(service, port.Protocol, port.Port, isIPv6)
			klog.V(2).Infof("getExpectedLBRules lb name (%s) rule name (%s)", lbName, lbRuleName)
			isNoLBRuleRequired, err := consts.IsLBRuleOnK8sServicePortDisabled(service.Annotations, port.Port)
			if err != nil {
				err := fmt.Errorf("failed to parse annotation %s: %w", consts.BuildAnnotationKeyForPort(port.Port, consts.PortAnnotationNoLBRule), err)
				klog.V(2).ErrorS(err, "error occurred when getExpectedLoadBalancingRulePropertiesForPort", "service", service.Name, "namespace", service.Namespace,
					"rule-name", lbRuleName, "port", port.Port)
			}
			if isNoLBRuleRequired {
				klog.V(2).Infof("getExpectedLBRules lb name (%s) rule name (%s) no lb rule required", lbName, lbRuleName)
				continue
			}
			if port.Protocol == v1.ProtocolSCTP && !(az.useStandardLoadBalancer() && consts.IsK8sServiceUsingInternalLoadBalancer(service)) {
				return expectedProbes, expectedRules, fmt.Errorf("SCTP is only supported on standard loadbalancer in internal mode")
			}

			transportProto, _, _, err := getProtocolsFromKubernetesProtocol(port.Protocol)
			if err != nil {
				return expectedProbes, expectedRules, fmt.Errorf("failed to parse transport protocol: %w", err)
			}
			props, err := az.getExpectedLoadBalancingRulePropertiesForPort(service, lbFrontendIPConfigID, lbBackendPoolID, port, *transportProto)
			if err != nil {
				return expectedProbes, expectedRules, fmt.Errorf("error generate lb rule for ha mod loadbalancer. err: %w", err)
			}

			isNoHealthProbeRule, err := consts.IsHealthProbeRuleOnK8sServicePortDisabled(service.Annotations, port.Port)
			if err != nil {
				err := fmt.Errorf("failed to parse annotation %s: %w", consts.BuildAnnotationKeyForPort(port.Port, consts.PortAnnotationNoHealthProbeRule), err)
				klog.V(2).ErrorS(err, "error occurred when buildHealthProbeRulesForPort", "service", service.Name, "namespace", service.Namespace,
					"rule-name", lbRuleName, "port", port.Port)
			}
			if !isNoHealthProbeRule {
				portprobe, err := az.buildHealthProbeRulesForPort(service, port, lbRuleName, nodeEndpointHealthprobe, useSharedProbe)
				if err != nil {
					klog.V(2).ErrorS(err, "error occurred when buildHealthProbeRulesForPort", "service", service.Name, "namespace", service.Namespace,
						"rule-name", lbRuleName, "port", port.Port)
					return expectedProbes, expectedRules, err
				}
				if portprobe != nil {
					props.Probe = &network.SubResource{
						ID: pointer.String(az.getLoadBalancerProbeID(lbName, *portprobe.Name)),
					}
					expectedProbes = append(expectedProbes, *portprobe)
				} else if nodeEndpointHealthprobe != nil {
					props.Probe = &network.SubResource{
						ID: pointer.String(az.getLoadBalancerProbeID(lbName, *nodeEndpointHealthprobe.Name)),
					}
					if !nodeEndpointHealthprobeAdded {
						expectedProbes = append(expectedProbes, *nodeEndpointHealthprobe)
						nodeEndpointHealthprobeAdded = true
					}
				}
			}
			if consts.IsK8sServiceDisableLoadBalancerFloatingIP(service) {
				props.BackendPort = pointer.Int32(port.NodePort)
				props.EnableFloatingIP = pointer.Bool(false)
			}
			expectedRules = append(expectedRules, network.LoadBalancingRule{
				Name:                              &lbRuleName,
				LoadBalancingRulePropertiesFormat: props,
			})
		}
	}

	return expectedProbes, expectedRules, nil
}

// getDefaultLoadBalancingRulePropertiesFormat returns the loadbalancing rule for one port
func (az *Cloud) getExpectedLoadBalancingRulePropertiesForPort(
	service *v1.Service,
	lbFrontendIPConfigID string,
	lbBackendPoolID string, servicePort v1.ServicePort, transportProto network.TransportProtocol) (*network.LoadBalancingRulePropertiesFormat, error) {
	var err error

	loadDistribution := network.LoadDistributionDefault
	if service.Spec.SessionAffinity == v1.ServiceAffinityClientIP {
		loadDistribution = network.LoadDistributionSourceIP
	}

	var lbIdleTimeout *int32
	if lbIdleTimeout, err = consts.Getint32ValueFromK8sSvcAnnotation(service.Annotations, consts.ServiceAnnotationLoadBalancerIdleTimeout, func(val *int32) error {
		const (
			min = 4
			max = 100
		)
		if *val < min || *val > max {
			return fmt.Errorf("idle timeout value must be a whole number representing minutes between %d and %d, actual value: %d", min, max, *val)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error parsing idle timeout key: %s, err: %w", consts.ServiceAnnotationLoadBalancerIdleTimeout, err)
	} else if lbIdleTimeout == nil {
		lbIdleTimeout = pointer.Int32(4)
	}

	props := &network.LoadBalancingRulePropertiesFormat{
		Protocol:            transportProto,
		FrontendPort:        pointer.Int32(servicePort.Port),
		BackendPort:         pointer.Int32(servicePort.Port),
		DisableOutboundSnat: pointer.Bool(az.disableLoadBalancerOutboundSNAT()),
		EnableFloatingIP:    pointer.Bool(true),
		LoadDistribution:    loadDistribution,
		FrontendIPConfiguration: &network.SubResource{
			ID: pointer.String(lbFrontendIPConfigID),
		},
		BackendAddressPool: &network.SubResource{
			ID: pointer.String(lbBackendPoolID),
		},
		IdleTimeoutInMinutes: lbIdleTimeout,
	}
	if strings.EqualFold(string(transportProto), string(network.TransportProtocolTCP)) && az.useStandardLoadBalancer() {
		props.EnableTCPReset = pointer.Bool(!consts.IsTCPResetDisabled(service.Annotations))
	}

	// Azure ILB does not support secondary IPs as floating IPs on the LB. Therefore, floating IP needs to be turned
	// off and the rule should point to the nodeIP:nodePort.
	if consts.IsK8sServiceUsingInternalLoadBalancer(service) && isBackendPoolIPv6(lbBackendPoolID) {
		props.BackendPort = pointer.Int32(servicePort.NodePort)
		props.EnableFloatingIP = pointer.Bool(false)
	}
	return props, nil
}

// getExpectedHAModeLoadBalancingRuleProperties build load balancing rule for lb in HA mode
func (az *Cloud) getExpectedHAModeLoadBalancingRuleProperties(
	service *v1.Service,
	lbFrontendIPConfigID string,
	lbBackendPoolID string) (*network.LoadBalancingRulePropertiesFormat, error) {
	props, err := az.getExpectedLoadBalancingRulePropertiesForPort(service, lbFrontendIPConfigID, lbBackendPoolID, v1.ServicePort{}, network.TransportProtocolAll)
	if err != nil {
		return nil, fmt.Errorf("error generate lb rule for ha mod loadbalancer. err: %w", err)
	}
	props.EnableTCPReset = pointer.Bool(!consts.IsTCPResetDisabled(service.Annotations))

	return props, nil
}

// This reconciles the Network Security Group similar to how the LB is reconciled.
// This entails adding required, missing SecurityRules and removing stale rules.
func (az *Cloud) reconcileSecurityGroup(
	clusterName string, service *v1.Service,
	lbName string, lbIPs []string,
	wantLb bool,
) (*network.SecurityGroup, error) {
	logger := klog.Background().WithName("reconcileSecurityGroup").
		WithValues("cluster", clusterName).
		WithValues("service", getServiceName(service)).
		WithValues("load-balancer", lbName).
		WithValues("delete-lb", !wantLb)
	logger.V(2).Info("Starting")

	if wantLb && len(lbIPs) == 0 {
		return nil, fmt.Errorf("no load balancer IP for setting up security rules for service %s", service.Name)
	}

	additionalIPs, err := loadbalancer.AdditionalPublicIPs(service)
	if wantLb && err != nil {
		return nil, fmt.Errorf("unable to get additional public IPs: %w", err)
	}

	var accessControl *loadbalancer.AccessControl
	{
		sg, err := az.getSecurityGroup(azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}

		var opts []loadbalancer.AccessControlOption
		if !wantLb {
			// When deleting LB, we don't need to validate the annotation
			opts = append(opts, loadbalancer.SkipAnnotationValidation())
		}
		accessControl, err = loadbalancer.NewAccessControl(service, &sg, opts...)
		if err != nil {
			logger.Error(err, "Failed to parse access control configuration for service")
			return nil, err
		}
		// - use both annotation `service.beta.kubernetes.io/azure-allowed-service-tags` and `spec.loadBalancerSourceRanges`
		//   WARNING: This issue has been around for a while, and we shouldn’t mess with the existing settings.
		if len(accessControl.SourceRanges) > 0 && len(accessControl.AllowedServiceTags) > 0 {
			// Suggesting to use aks custom annotation instead of spec.loadBalancerSourceRanges
			logger.V(2).Info(
				"Service is using both of spec.loadBalancerSourceRanges and annotation service.beta.kubernetes.io/azure-allowed-service-tags",
			)
			az.Event(service, v1.EventTypeWarning, "ConflictConfiguration", fmt.Sprintf(
				"Please use annotation %s instead of spec.loadBalancerSourceRanges while using %s annotation at the same time.",
				consts.ServiceAnnotationAllowedIPRanges, consts.ServiceAnnotationAllowedServiceTags,
			))
		}

		if len(accessControl.InvalidRanges) > 0 {
			az.Event(service, v1.EventTypeWarning, "InvalidConfiguration", fmt.Sprintf(
				"Found invalid LoadBalancerSourceRanges %v, ignoring and adding a default DenyAll rule in security group.",
				accessControl.InvalidRanges,
			))
		}
	}

	var (
		disableFloatingIP                                = consts.IsK8sServiceDisableLoadBalancerFloatingIP(service)
		lbIPAddresses, _                                 = iputil.ParseAddresses(lbIPs)
		lbIPv4Addresses, lbIPv6Addresses                 = iputil.GroupAddressesByFamily(lbIPAddresses)
		additionalIPv4Addresses, additionalIPv6Addresses = iputil.GroupAddressesByFamily(additionalIPs)
		backendIPv4Addresses, backendIPv6Addresses       []netip.Addr
	)
	{
		// Get backend node IPs
		lb, lbFound, err := az.getAzureLoadBalancer(lbName, azcache.CacheReadTypeDefault)
		{
			if err != nil {
				return nil, err
			}
			if wantLb && !lbFound {
				logger.Error(err, "Failed to get load balancer")
				return nil, fmt.Errorf("unable to get lb %s", lbName)
			}
		}
		var backendIPv4List, backendIPv6List []string
		if lbFound {
			backendIPv4List, backendIPv6List = az.LoadBalancerBackendPool.GetBackendPrivateIPs(clusterName, service, lb)
		}
		backendIPv4Addresses, _ = iputil.ParseAddresses(backendIPv4List)
		backendIPv6Addresses, _ = iputil.ParseAddresses(backendIPv6List)
	}

	{
		// Disassociate all IPs from the security group
		dstIPv4Addresses := append(lbIPv4Addresses, backendIPv4Addresses...)
		dstIPv4Addresses = append(dstIPv4Addresses, additionalIPv4Addresses...)

		dstIPv6Addresses := append(lbIPv6Addresses, backendIPv6Addresses...)
		dstIPv6Addresses = append(dstIPv6Addresses, additionalIPv6Addresses...)

		accessControl.CleanSecurityGroup(dstIPv4Addresses, dstIPv6Addresses)
	}

	if wantLb {
		var (
			dstIPv4Addresses = additionalIPv4Addresses
			dstIPv6Addresses = additionalIPv6Addresses
		)

		if disableFloatingIP {
			// use the backend node IPs
			dstIPv4Addresses = append(dstIPv4Addresses, backendIPv4Addresses...)
			dstIPv6Addresses = append(dstIPv6Addresses, backendIPv6Addresses...)
		} else {
			// use the LoadBalancer IPs
			dstIPv4Addresses = append(dstIPv4Addresses, lbIPv4Addresses...)
			dstIPv6Addresses = append(dstIPv6Addresses, lbIPv6Addresses...)
		}

		err := accessControl.PatchSecurityGroup(dstIPv4Addresses, dstIPv6Addresses)
		if err != nil {
			logger.Error(err, "Failed to patch security group")
			return nil, err
		}
	}

	rv, updated, err := accessControl.SecurityGroup()
	if err != nil {
		err = fmt.Errorf("unable to apply access control configuration to security group: %w", err)
		logger.Error(err, "Failed to get security group after patching")
		return nil, err
	}
	if az.ensureSecurityGroupTagged(rv) {
		updated = true
	}

	if updated {
		logger.V(2).Info("Preparing to update security group")
		logger.V(10).Info("CreateOrUpdateSecurityGroup begin")
		err := az.CreateOrUpdateSecurityGroup(*rv)
		if err != nil {
			logger.Error(err, "Failed to update security group")
			return nil, err
		}
		logger.V(10).Info("CreateOrUpdateSecurityGroup end")
		_ = az.nsgCache.Delete(pointer.StringDeref(rv.Name, ""))
	}
	return rv, nil
}

func (az *Cloud) shouldUpdateLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) (bool, error) {
	existingManagedLBs, err := az.ListManagedLBs(service, nodes, clusterName)
	if err != nil {
		return false, fmt.Errorf("shouldUpdateLoadBalancer: failed to list managed load balancers: %w", err)
	}

	_, _, _, _, existsLb, _ := az.getServiceLoadBalancer(service, clusterName, nodes, false, existingManagedLBs)
	return existsLb && service.ObjectMeta.DeletionTimestamp == nil && service.Spec.Type == v1.ServiceTypeLoadBalancer, nil
}

func allowsConsolidation(rule network.SecurityRule) bool {
	return strings.HasPrefix(pointer.StringDeref(rule.Name, ""), "shared")
}

// Determine if we should release existing owned public IPs
func shouldReleaseExistingOwnedPublicIP(existingPip *network.PublicIPAddress, lbShouldExist, lbIsInternal, isUserAssignedPIP bool, desiredPipName string, ipTagRequest serviceIPTagRequest) bool {
	// skip deleting user created pip
	if isUserAssignedPIP {
		return false
	}

	// Latch some variables for readability purposes.
	pipName := *(*existingPip).Name

	// Assume the current IP Tags are empty by default unless properties specify otherwise.
	currentIPTags := &[]network.IPTag{}
	pipPropertiesFormat := (*existingPip).PublicIPAddressPropertiesFormat
	if pipPropertiesFormat != nil {
		currentIPTags = (*pipPropertiesFormat).IPTags
	}

	// Check whether the public IP is being referenced by other service.
	// The owned public IP can be released only when there is not other service using it.
	if serviceTag := getServiceFromPIPServiceTags(existingPip.Tags); serviceTag != "" {
		// case 1: there is at least one reference when deleting the PIP
		if !lbShouldExist && len(parsePIPServiceTag(&serviceTag)) > 0 {
			return false
		}

		// case 2: there is at least one reference from other service
		if lbShouldExist && len(parsePIPServiceTag(&serviceTag)) > 1 {
			return false
		}
	}

	// Release the ip under the following criteria -
	// #1 - If we don't actually want a load balancer,
	return !lbShouldExist ||
		// #2 - If the load balancer is internal, and thus doesn't require public exposure
		lbIsInternal ||
		// #3 - If the name of this public ip does not match the desired name,
		// NOTICE: For IPv6 Service created with CCM v1.27.1, the created PIP has IPv6 suffix.
		// We need to recreate such PIP and current logic to delete needs no change.
		(pipName != desiredPipName) ||
		// #4 If the service annotations have specified the ip tags that the public ip must have, but they do not match the ip tags of the existing instance
		(ipTagRequest.IPTagsRequestedByAnnotation && !areIPTagsEquivalent(currentIPTags, ipTagRequest.IPTags))
}

// ensurePIPTagged ensures the public IP of the service is tagged as configured
func (az *Cloud) ensurePIPTagged(service *v1.Service, pip *network.PublicIPAddress) bool {
	configTags := parseTags(az.Tags, az.TagsMap)
	annotationTags := make(map[string]*string)
	if _, ok := service.Annotations[consts.ServiceAnnotationAzurePIPTags]; ok {
		annotationTags = parseTags(service.Annotations[consts.ServiceAnnotationAzurePIPTags], map[string]string{})
	}

	for k, v := range annotationTags {
		found, key := findKeyInMapCaseInsensitive(configTags, k)
		if !found {
			configTags[k] = v
		} else if !strings.EqualFold(pointer.StringDeref(v, ""), pointer.StringDeref(configTags[key], "")) {
			configTags[key] = v
		}
	}

	// include the cluster name and service names tags when comparing
	var clusterName, serviceNames, serviceNameUsingDNS *string
	if v := getClusterFromPIPClusterTags(pip.Tags); v != "" {
		clusterName = &v
	}
	if v := getServiceFromPIPServiceTags(pip.Tags); v != "" {
		serviceNames = &v
	}
	if v := getServiceFromPIPDNSTags(pip.Tags); v != "" {
		serviceNameUsingDNS = &v
	}
	if clusterName != nil {
		configTags[consts.ClusterNameKey] = clusterName
	}
	if serviceNames != nil {
		configTags[consts.ServiceTagKey] = serviceNames
	}
	if serviceNameUsingDNS != nil {
		configTags[consts.ServiceUsingDNSKey] = serviceNameUsingDNS
	}

	tags, changed := az.reconcileTags(pip.Tags, configTags)
	pip.Tags = tags

	return changed
}

// reconcilePublicIPs reconciles the PublicIP resources similar to how the LB is reconciled.
func (az *Cloud) reconcilePublicIPs(clusterName string, service *v1.Service, lbName string, wantLb bool) ([]*network.PublicIPAddress, error) {
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)

	reconciledPIPs := []*network.PublicIPAddress{}
	pips, err := az.listPIP(pipResourceGroup, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}

	pipsV4, pipsV6 := []network.PublicIPAddress{}, []network.PublicIPAddress{}
	for _, pip := range pips {
		if pip.PublicIPAddressPropertiesFormat == nil || pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion == "" ||
			pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion == network.IPv4 {
			pipsV4 = append(pipsV4, pip)
		} else {
			pipsV6 = append(pipsV6, pip)
		}
	}

	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	if v4Enabled {
		reconciledPIP, err := az.reconcilePublicIP(pipsV4, clusterName, service, lbName, wantLb, false)
		if err != nil {
			return reconciledPIPs, err
		}
		if reconciledPIP != nil {
			reconciledPIPs = append(reconciledPIPs, reconciledPIP)
		}
	}
	if v6Enabled {
		reconciledPIP, err := az.reconcilePublicIP(pipsV6, clusterName, service, lbName, wantLb, true)
		if err != nil {
			return reconciledPIPs, err
		}
		if reconciledPIP != nil {
			reconciledPIPs = append(reconciledPIPs, reconciledPIP)
		}
	}
	return reconciledPIPs, nil
}

// reconcilePublicIP reconciles the PublicIP resources similar to how the LB is reconciled with the specified IP family.
func (az *Cloud) reconcilePublicIP(pips []network.PublicIPAddress, clusterName string, service *v1.Service, lbName string, wantLb, isIPv6 bool) (*network.PublicIPAddress, error) {
	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	serviceIPTagRequest := getServiceIPTagRequestForPublicIP(service)
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)

	var (
		lb               *network.LoadBalancer
		desiredPipName   string
		err              error
		shouldPIPExisted bool
	)

	if !isInternal && wantLb {
		desiredPipName, shouldPIPExisted, err = az.determinePublicIPName(clusterName, service, isIPv6)
		if err != nil {
			return nil, err
		}
	}

	if lbName != "" {
		lb, _, err = az.getAzureLoadBalancer(lbName, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}
	}

	discoveredDesiredPublicIP, pipsToBeDeleted, deletedDesiredPublicIP, pipsToBeUpdated, err := az.getPublicIPUpdates(
		clusterName, service, pips, wantLb, isInternal, desiredPipName, serviceName, serviceIPTagRequest, shouldPIPExisted, isIPv6)
	if err != nil {
		return nil, err
	}

	var deleteFuncs, updateFuncs []func() error
	for _, pip := range pipsToBeUpdated {
		pipCopy := *pip
		updateFuncs = append(updateFuncs, func() error {
			klog.V(2).Infof("reconcilePublicIP for service(%s): pip(%s), isIPv6(%v) - updating", serviceName, *pip.Name, isIPv6)
			return az.CreateOrUpdatePIP(service, pipResourceGroup, pipCopy)
		})
	}
	errs := utilerrors.AggregateGoroutines(updateFuncs...)
	if errs != nil {
		return nil, utilerrors.Flatten(errs)
	}

	for _, pip := range pipsToBeDeleted {
		pipCopy := *pip
		deleteFuncs = append(deleteFuncs, func() error {
			klog.V(2).Infof("reconcilePublicIP for service(%s): pip(%s), isIPv6(%v) - deleting", serviceName, *pip.Name, isIPv6)
			return az.safeDeletePublicIP(service, pipResourceGroup, &pipCopy, lb)
		})
	}
	errs = utilerrors.AggregateGoroutines(deleteFuncs...)
	if errs != nil {
		return nil, utilerrors.Flatten(errs)
	}

	if !isInternal && wantLb {
		// Confirm desired public ip resource exists
		var pip *network.PublicIPAddress
		domainNameLabel, found := getPublicIPDomainNameLabel(service)
		errorIfPublicIPDoesNotExist := shouldPIPExisted && discoveredDesiredPublicIP && !deletedDesiredPublicIP
		if pip, err = az.ensurePublicIPExists(service, desiredPipName, domainNameLabel, clusterName, errorIfPublicIPDoesNotExist, found, isIPv6); err != nil {
			return nil, err
		}
		return pip, nil
	}
	return nil, nil
}

// getPublicIPUpdates handles one IP family only according to isIPv6 and PIP IP version.
func (az *Cloud) getPublicIPUpdates(
	clusterName string,
	service *v1.Service,
	pips []network.PublicIPAddress,
	wantLb bool,
	isInternal bool,
	desiredPipName string,
	serviceName string,
	serviceIPTagRequest serviceIPTagRequest,
	serviceAnnotationRequestsNamedPublicIP,
	isIPv6 bool,
) (bool, []*network.PublicIPAddress, bool, []*network.PublicIPAddress, error) {
	var (
		err                       error
		discoveredDesiredPublicIP bool
		deletedDesiredPublicIP    bool
		pipsToBeDeleted           []*network.PublicIPAddress
		pipsToBeUpdated           []*network.PublicIPAddress
	)
	for i := range pips {
		pip := pips[i]
		if pip.PublicIPAddressPropertiesFormat != nil && pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion != "" {
			if (pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion == network.IPv4 && isIPv6) ||
				(pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion == network.IPv6 && !isIPv6) {
				continue
			}
		}

		if pip.Name == nil {
			return false, nil, false, nil, fmt.Errorf("PIP name is empty: %v", pip)
		}
		pipName := *pip.Name

		// If we've been told to use a specific public ip by the client, let's track whether or not it actually existed
		// when we inspect the set in Azure.
		discoveredDesiredPublicIP = discoveredDesiredPublicIP || wantLb && !isInternal && pipName == desiredPipName

		// Now, let's perform additional analysis to determine if we should release the public ips we have found.
		// We can only let them go if (a) they are owned by this service and (b) they meet the criteria for deletion.
		owns, isUserAssignedPIP := serviceOwnsPublicIP(service, &pip, clusterName)
		if owns {
			var dirtyPIP, toBeDeleted bool
			if !wantLb && !isUserAssignedPIP {
				klog.V(2).Infof("reconcilePublicIP for service(%s): unbinding the service from pip %s", serviceName, *pip.Name)
				if err = unbindServiceFromPIP(&pip, service, serviceName, clusterName, isUserAssignedPIP); err != nil {
					return false, nil, false, nil, err
				}
				dirtyPIP = true
			}
			if !isUserAssignedPIP {
				changed := az.ensurePIPTagged(service, &pip)
				if changed {
					dirtyPIP = true
				}
			}
			if shouldReleaseExistingOwnedPublicIP(&pip, wantLb, isInternal, isUserAssignedPIP, desiredPipName, serviceIPTagRequest) {
				// Then, release the public ip
				pipsToBeDeleted = append(pipsToBeDeleted, &pip)

				// Flag if we deleted the desired public ip
				deletedDesiredPublicIP = deletedDesiredPublicIP || pipName == desiredPipName

				// An aside: It would be unusual, but possible, for us to delete a public ip referred to explicitly by name
				// in Service annotations (which is usually reserved for non-service-owned externals), if that IP is tagged as
				// having been owned by a particular Kubernetes cluster.

				// If the pip is going to be deleted, we do not need to update it
				toBeDeleted = true
			}

			// Update tags of PIP only instead of deleting it.
			if !toBeDeleted && dirtyPIP {
				pipsToBeUpdated = append(pipsToBeUpdated, &pip)
			}
		}
	}

	if !isInternal && serviceAnnotationRequestsNamedPublicIP && !discoveredDesiredPublicIP && wantLb {
		return false, nil, false, nil, fmt.Errorf("reconcilePublicIP for service(%s): pip(%s) not found", serviceName, desiredPipName)
	}
	return discoveredDesiredPublicIP, pipsToBeDeleted, deletedDesiredPublicIP, pipsToBeUpdated, err
}

// safeDeletePublicIP deletes public IP by removing its reference first.
func (az *Cloud) safeDeletePublicIP(service *v1.Service, pipResourceGroup string, pip *network.PublicIPAddress, lb *network.LoadBalancer) error {
	// Remove references if pip.IPConfiguration is not nil.
	if pip.PublicIPAddressPropertiesFormat != nil &&
		pip.PublicIPAddressPropertiesFormat.IPConfiguration != nil {
		// Fetch latest pip to check if the pip in the cache is stale.
		// In some cases the public IP to be deleted is still referencing
		// the frontend IP config on the LB. This is because the pip is
		// stored in the cache and is not up-to-date.
		latestPIP, ok, err := az.getPublicIPAddress(pipResourceGroup, *pip.Name, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			klog.Errorf("safeDeletePublicIP: failed to get latest public IP %s/%s: %s", pipResourceGroup, *pip.Name, err.Error())
			return err
		}
		if ok && latestPIP.PublicIPAddressPropertiesFormat != nil &&
			latestPIP.PublicIPAddressPropertiesFormat.IPConfiguration != nil &&
			lb != nil && lb.LoadBalancerPropertiesFormat != nil &&
			lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations != nil {
			referencedLBRules := []network.SubResource{}
			frontendIPConfigUpdated := false
			loadBalancerRuleUpdated := false

			// Check whether there are still frontend IP configurations referring to it.
			ipConfigurationID := pointer.StringDeref(pip.PublicIPAddressPropertiesFormat.IPConfiguration.ID, "")
			if ipConfigurationID != "" {
				lbFrontendIPConfigs := *lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations
				for i := len(lbFrontendIPConfigs) - 1; i >= 0; i-- {
					config := lbFrontendIPConfigs[i]
					if strings.EqualFold(ipConfigurationID, pointer.StringDeref(config.ID, "")) {
						if config.FrontendIPConfigurationPropertiesFormat != nil &&
							config.FrontendIPConfigurationPropertiesFormat.LoadBalancingRules != nil {
							referencedLBRules = *config.FrontendIPConfigurationPropertiesFormat.LoadBalancingRules
						}

						frontendIPConfigUpdated = true
						lbFrontendIPConfigs = append(lbFrontendIPConfigs[:i], lbFrontendIPConfigs[i+1:]...)
						break
					}
				}

				if frontendIPConfigUpdated {
					lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations = &lbFrontendIPConfigs
				}
			}

			// Check whether there are still load balancer rules referring to it.
			if len(referencedLBRules) > 0 {
				referencedLBRuleIDs := utilsets.NewString()
				for _, refer := range referencedLBRules {
					referencedLBRuleIDs.Insert(pointer.StringDeref(refer.ID, ""))
				}

				if lb.LoadBalancerPropertiesFormat.LoadBalancingRules != nil {
					lbRules := *lb.LoadBalancerPropertiesFormat.LoadBalancingRules
					for i := len(lbRules) - 1; i >= 0; i-- {
						ruleID := pointer.StringDeref(lbRules[i].ID, "")
						if ruleID != "" && referencedLBRuleIDs.Has(ruleID) {
							loadBalancerRuleUpdated = true
							lbRules = append(lbRules[:i], lbRules[i+1:]...)
						}
					}

					if loadBalancerRuleUpdated {
						lb.LoadBalancerPropertiesFormat.LoadBalancingRules = &lbRules
					}
				}
			}

			// Update load balancer when frontendIPConfigUpdated or loadBalancerRuleUpdated.
			if frontendIPConfigUpdated || loadBalancerRuleUpdated {
				err := az.CreateOrUpdateLB(service, *lb)
				if err != nil {
					klog.Errorf("safeDeletePublicIP for service(%s) failed with error: %v", getServiceName(service), err)
					return err
				}
			}
		}
	}

	pipName := pointer.StringDeref(pip.Name, "")
	klog.V(10).Infof("DeletePublicIP(%s, %q): start", pipResourceGroup, pipName)
	err := az.DeletePublicIP(service, pipResourceGroup, pipName)
	if err != nil {
		return err
	}
	klog.V(10).Infof("DeletePublicIP(%s, %q): end", pipResourceGroup, pipName)

	return nil
}

func findRule(rules []network.LoadBalancingRule, rule network.LoadBalancingRule, wantLB bool) bool {
	for _, existingRule := range rules {
		if strings.EqualFold(pointer.StringDeref(existingRule.Name, ""), pointer.StringDeref(rule.Name, "")) &&
			equalLoadBalancingRulePropertiesFormat(existingRule.LoadBalancingRulePropertiesFormat, rule.LoadBalancingRulePropertiesFormat, wantLB) {
			return true
		}
	}
	return false
}

// equalLoadBalancingRulePropertiesFormat checks whether the provided LoadBalancingRulePropertiesFormat are equal.
// Note: only fields used in reconcileLoadBalancer are considered.
// s: existing, t: target
func equalLoadBalancingRulePropertiesFormat(s *network.LoadBalancingRulePropertiesFormat, t *network.LoadBalancingRulePropertiesFormat, wantLB bool) bool {
	if s == nil || t == nil {
		return false
	}

	properties := reflect.DeepEqual(s.Protocol, t.Protocol)
	if !properties {
		return false
	}

	if reflect.DeepEqual(s.Protocol, network.TransportProtocolTCP) {
		properties = properties && reflect.DeepEqual(pointer.BoolDeref(s.EnableTCPReset, false), pointer.BoolDeref(t.EnableTCPReset, false))
	}

	properties = properties && equalSubResource(s.FrontendIPConfiguration, t.FrontendIPConfiguration) &&
		equalSubResource(s.BackendAddressPool, t.BackendAddressPool) &&
		reflect.DeepEqual(s.LoadDistribution, t.LoadDistribution) &&
		reflect.DeepEqual(s.FrontendPort, t.FrontendPort) &&
		reflect.DeepEqual(s.BackendPort, t.BackendPort) &&
		equalSubResource(s.Probe, t.Probe) &&
		reflect.DeepEqual(s.EnableFloatingIP, t.EnableFloatingIP) &&
		reflect.DeepEqual(pointer.BoolDeref(s.DisableOutboundSnat, false), pointer.BoolDeref(t.DisableOutboundSnat, false))

	if wantLB && s.IdleTimeoutInMinutes != nil && t.IdleTimeoutInMinutes != nil {
		return properties && reflect.DeepEqual(s.IdleTimeoutInMinutes, t.IdleTimeoutInMinutes)
	}
	return properties
}

func equalSubResource(s *network.SubResource, t *network.SubResource) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	return strings.EqualFold(pointer.StringDeref(s.ID, ""), pointer.StringDeref(t.ID, ""))
}

// This compares rule's Name, Protocol, SourcePortRange, DestinationPortRange, SourceAddressPrefix, Access, and Direction.
// Note that it compares rule's DestinationAddressPrefix only when it's not consolidated rule as such rule does not have DestinationAddressPrefix defined.
// We intentionally do not compare DestinationAddressPrefixes in consolidated case because reconcileSecurityRule has to consider the two rules equal,
// despite different DestinationAddressPrefixes, in order to give it a chance to consolidate the two rules.
func findSecurityRule(rules []network.SecurityRule, rule network.SecurityRule) bool {
	for _, existingRule := range rules {
		if !strings.EqualFold(pointer.StringDeref(existingRule.Name, ""), pointer.StringDeref(rule.Name, "")) {
			continue
		}
		if !strings.EqualFold(string(existingRule.Protocol), string(rule.Protocol)) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.SourcePortRange, ""), pointer.StringDeref(rule.SourcePortRange, "")) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.DestinationPortRange, ""), pointer.StringDeref(rule.DestinationPortRange, "")) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.SourceAddressPrefix, ""), pointer.StringDeref(rule.SourceAddressPrefix, "")) {
			continue
		}
		if !allowsConsolidation(existingRule) && !allowsConsolidation(rule) {
			if !strings.EqualFold(pointer.StringDeref(existingRule.DestinationAddressPrefix, ""), pointer.StringDeref(rule.DestinationAddressPrefix, "")) {
				continue
			}
			if !slices.Equal(stringSlice(existingRule.DestinationAddressPrefixes), stringSlice(rule.DestinationAddressPrefixes)) {
				continue
			}
		}
		if !strings.EqualFold(string(existingRule.Access), string(rule.Access)) {
			continue
		}
		if !strings.EqualFold(string(existingRule.Direction), string(rule.Direction)) {
			continue
		}
		return true
	}
	return false
}

func (az *Cloud) getPublicIPAddressResourceGroup(service *v1.Service) string {
	if resourceGroup, found := service.Annotations[consts.ServiceAnnotationLoadBalancerResourceGroup]; found {
		resourceGroupName := strings.TrimSpace(resourceGroup)
		if len(resourceGroupName) > 0 {
			return resourceGroupName
		}
	}

	return az.ResourceGroup
}

func (az *Cloud) isBackendPoolPreConfigured(service *v1.Service) bool {
	preConfigured := false
	isInternal := requiresInternalLoadBalancer(service)

	if az.PreConfiguredBackendPoolLoadBalancerTypes == consts.PreConfiguredBackendPoolLoadBalancerTypesAll {
		preConfigured = true
	}
	if (az.PreConfiguredBackendPoolLoadBalancerTypes == consts.PreConfiguredBackendPoolLoadBalancerTypesInternal) && isInternal {
		preConfigured = true
	}
	if (az.PreConfiguredBackendPoolLoadBalancerTypes == consts.PreConfiguredBackendPoolLoadBalancerTypesExternal) && !isInternal {
		preConfigured = true
	}

	return preConfigured
}

// Check if service requires an internal load balancer.
func requiresInternalLoadBalancer(service *v1.Service) bool {
	if l, found := service.Annotations[consts.ServiceAnnotationLoadBalancerInternal]; found {
		return l == consts.TrueAnnotationValue
	}

	return false
}

func getInternalSubnet(service *v1.Service) *string {
	if requiresInternalLoadBalancer(service) {
		if l, found := service.Annotations[consts.ServiceAnnotationLoadBalancerInternalSubnet]; found && strings.TrimSpace(l) != "" {
			return &l
		}
	}

	return nil
}

func ipInSubnet(ip string, subnet *network.Subnet) bool {
	if subnet == nil || subnet.SubnetPropertiesFormat == nil {
		return false
	}
	netIP, err := netip.ParseAddr(ip)
	if err != nil {
		klog.Errorf("ipInSubnet: failed to parse ip %s: %v", netIP, err)
		return false
	}
	cidrs := make([]string, 0)
	if subnet.AddressPrefix != nil {
		cidrs = append(cidrs, *subnet.AddressPrefix)
	}
	if subnet.AddressPrefixes != nil {
		cidrs = append(cidrs, *subnet.AddressPrefixes...)
	}
	for _, cidr := range cidrs {
		network, err := netip.ParsePrefix(cidr)
		if err != nil {
			klog.Errorf("ipInSubnet: failed to parse ip cidr %s: %v", cidr, err)
			continue
		}
		if network.Contains(netIP) {
			return true
		}
	}
	return false
}

// getServiceLoadBalancerMode parses the mode value.
// if the value is __auto__ it returns isAuto = TRUE.
// if anything else it returns the unique VM set names after trimming spaces.
func (az *Cloud) getServiceLoadBalancerMode(service *v1.Service) (bool, bool, string) {
	mode, hasMode := service.Annotations[consts.ServiceAnnotationLoadBalancerMode]
	if az.useStandardLoadBalancer() && hasMode {
		klog.Warningf("single standard load balancer doesn't work with annotation %q, would ignore it", consts.ServiceAnnotationLoadBalancerMode)
	}
	mode = strings.TrimSpace(mode)
	isAuto := strings.EqualFold(mode, consts.ServiceAnnotationLoadBalancerAutoModeValue)

	return hasMode, isAuto, mode
}

func useSharedSecurityRule(service *v1.Service) bool {
	if l, ok := service.Annotations[consts.ServiceAnnotationSharedSecurityRule]; ok {
		return l == consts.TrueAnnotationValue
	}

	return false
}

// serviceOwnsPublicIP checks if the service owns the pip and if the pip is user-created.
// The pip is user-created if and only if there is no service tags.
// The service owns the pip if:
// 1. The serviceName is included in the service tags of a system-created pip.
// 2. The service LoadBalancerIP matches the IP address of a user-created pip.
func serviceOwnsPublicIP(service *v1.Service, pip *network.PublicIPAddress, clusterName string) (bool, bool) {
	if service == nil || pip == nil {
		klog.Warningf("serviceOwnsPublicIP: nil service or public IP")
		return false, false
	}

	if pip.PublicIPAddressPropertiesFormat == nil || pointer.StringDeref(pip.IPAddress, "") == "" {
		klog.Warningf("serviceOwnsPublicIP: empty pip.IPAddress")
		return false, false
	}

	serviceName := getServiceName(service)

	isIPv6 := pip.PublicIPAddressVersion == network.IPv6
	if pip.Tags != nil {
		serviceTag := getServiceFromPIPServiceTags(pip.Tags)
		clusterTag := getClusterFromPIPClusterTags(pip.Tags)

		// if there is no service tag on the pip, it is user-created pip
		if serviceTag == "" {
			return isServiceSelectPIP(service, pip, isIPv6), true
		}

		// if there is service tag on the pip, it is system-created pip
		if isSVCNameInPIPTag(serviceTag, serviceName) {
			// Backward compatible for clusters upgraded from old releases.
			// In such case, only "service" tag is set.
			if clusterTag == "" {
				return true, false
			}

			// If cluster name tag is set, then return true if it matches.
			return strings.EqualFold(clusterTag, clusterName), false
		}

		// if the service is not included in the tags of the system-created pip, check the ip address
		// or pip name, this could happen for secondary services
		return isServiceSelectPIP(service, pip, isIPv6), false
	}

	// if the pip has no tags, it should be user-created
	return isServiceSelectPIP(service, pip, isIPv6), true
}

func isServiceLoadBalancerIPMatchesPIP(service *v1.Service, pip *network.PublicIPAddress, isIPV6 bool) bool {
	return strings.EqualFold(pointer.StringDeref(pip.IPAddress, ""), getServiceLoadBalancerIP(service, isIPV6))
}

func isServicePIPNameMatchesPIP(service *v1.Service, pip *network.PublicIPAddress, isIPV6 bool) bool {
	return strings.EqualFold(pointer.StringDeref(pip.Name, ""), getServicePIPName(service, isIPV6))
}

func isServiceSelectPIP(service *v1.Service, pip *network.PublicIPAddress, isIPV6 bool) bool {
	return isServiceLoadBalancerIPMatchesPIP(service, pip, isIPV6) || isServicePIPNameMatchesPIP(service, pip, isIPV6)
}

func isSVCNameInPIPTag(tag, svcName string) bool {
	svcNames := parsePIPServiceTag(&tag)

	for _, name := range svcNames {
		if strings.EqualFold(name, svcName) {
			return true
		}
	}

	return false
}

func parsePIPServiceTag(serviceTag *string) []string {
	if serviceTag == nil || len(*serviceTag) == 0 {
		return []string{}
	}

	serviceNames := strings.FieldsFunc(*serviceTag, func(r rune) bool {
		return r == ','
	})
	for i, name := range serviceNames {
		serviceNames[i] = strings.TrimSpace(name)
	}

	return serviceNames
}

// bindServicesToPIP add the incoming service name to the PIP's tag
// parameters: public IP address to be updated and incoming service names
// return values:
// 1. a bool flag to indicate if there is a new service added
// 2. an error when the pip is nil
// example:
// "ns1/svc1" + ["ns1/svc1", "ns2/svc2"] = "ns1/svc1,ns2/svc2"
func bindServicesToPIP(pip *network.PublicIPAddress, incomingServiceNames []string, replace bool) (bool, error) {
	if pip == nil {
		return false, fmt.Errorf("nil public IP")
	}

	if pip.Tags == nil {
		pip.Tags = map[string]*string{consts.ServiceTagKey: pointer.String("")}
	}

	serviceTagValue := pointer.String(getServiceFromPIPServiceTags(pip.Tags))
	serviceTagValueSet := make(map[string]struct{})
	existingServiceNames := parsePIPServiceTag(serviceTagValue)
	addedNew := false

	// replace is used when unbinding the service from PIP so addedNew remains false all the time
	if replace {
		serviceTagValue = pointer.String(strings.Join(incomingServiceNames, ","))
		pip.Tags[consts.ServiceTagKey] = serviceTagValue

		return false, nil
	}

	for _, name := range existingServiceNames {
		if _, ok := serviceTagValueSet[name]; !ok {
			serviceTagValueSet[name] = struct{}{}
		}
	}

	for _, serviceName := range incomingServiceNames {
		if serviceTagValue == nil || *serviceTagValue == "" {
			serviceTagValue = pointer.String(serviceName)
			addedNew = true
		} else {
			// detect duplicates
			if _, ok := serviceTagValueSet[serviceName]; !ok {
				*serviceTagValue += fmt.Sprintf(",%s", serviceName)
				addedNew = true
			} else {
				klog.V(10).Infof("service %s has been bound to the pip already", serviceName)
			}
		}
	}
	pip.Tags[consts.ServiceTagKey] = serviceTagValue

	return addedNew, nil
}

func unbindServiceFromPIP(pip *network.PublicIPAddress, _ *v1.Service,
	serviceName, _ string, isUserAssignedPIP bool) error {
	if pip == nil || pip.Tags == nil {
		return fmt.Errorf("nil public IP or tags")
	}

	if existingServiceName := getServiceFromPIPDNSTags(pip.Tags); existingServiceName != "" && strings.EqualFold(existingServiceName, serviceName) {
		deleteServicePIPDNSTags(&pip.Tags)
	}
	if isUserAssignedPIP {
		return nil
	}

	// skip removing tags for user assigned pips
	serviceTagValue := pointer.String(getServiceFromPIPServiceTags(pip.Tags))
	existingServiceNames := parsePIPServiceTag(serviceTagValue)
	var found bool
	for i := len(existingServiceNames) - 1; i >= 0; i-- {
		if strings.EqualFold(existingServiceNames[i], serviceName) {
			existingServiceNames = append(existingServiceNames[:i], existingServiceNames[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		klog.Warningf("cannot find the service %s in the corresponding PIP", serviceName)
	}

	_, err := bindServicesToPIP(pip, existingServiceNames, true)
	return err
}

// ensureLoadBalancerTagged ensures every load balancer in the resource group is tagged as configured
func (az *Cloud) ensureLoadBalancerTagged(lb *network.LoadBalancer) bool {
	if az.Tags == "" && (az.TagsMap == nil || len(az.TagsMap) == 0) {
		return false
	}
	tags := parseTags(az.Tags, az.TagsMap)
	if lb.Tags == nil {
		lb.Tags = make(map[string]*string)
	}

	tags, changed := az.reconcileTags(lb.Tags, tags)
	lb.Tags = tags

	return changed
}

// ensureSecurityGroupTagged ensures the security group is tagged as configured
func (az *Cloud) ensureSecurityGroupTagged(sg *network.SecurityGroup) bool {
	if az.Tags == "" && (az.TagsMap == nil || len(az.TagsMap) == 0) {
		return false
	}
	tags := parseTags(az.Tags, az.TagsMap)
	if sg.Tags == nil {
		sg.Tags = make(map[string]*string)
	}

	tags, changed := az.reconcileTags(sg.Tags, tags)
	sg.Tags = tags

	return changed
}

// For a load balancer, all frontend ip should reference either a subnet or publicIpAddress.
// Thus Azure do not allow mixed type (public and internal) load balancer.
// So we'd have a separate name for internal load balancer.
// This would be the name for Azure LoadBalancer resource.
func (az *Cloud) getAzureLoadBalancerName(
	service *v1.Service,
	existingLBs *[]network.LoadBalancer,
	clusterName, vmSetName string,
	isInternal bool,
) (string, error) {
	if az.LoadBalancerName != "" {
		clusterName = az.LoadBalancerName
	}
	lbNamePrefix := vmSetName
	// The LB name prefix is set to the name of the cluster when:
	// 1. the LB belongs to the primary agent pool.
	// 2. using the single SLB.
	if strings.EqualFold(vmSetName, az.VMSet.GetPrimaryVMSetName()) || az.useSingleStandardLoadBalancer() {
		lbNamePrefix = clusterName
	}

	// For multiple standard load balancers scenario:
	// 1. Filter out the eligible load balancers.
	// 2. Choose the most eligible load balancer.
	if az.useMultipleStandardLoadBalancers() {
		eligibleLBs, err := az.getEligibleLoadBalancersForService(service)
		if err != nil {
			return "", err
		}

		currentLBName := az.getServiceCurrentLoadBalancerName(service)
		lbNamePrefix = getMostEligibleLBForService(currentLBName, eligibleLBs, existingLBs)
	}

	if isInternal {
		return fmt.Sprintf("%s%s", lbNamePrefix, consts.InternalLoadBalancerNameSuffix), nil
	}
	return lbNamePrefix, nil
}

func getMostEligibleLBForService(
	currentLBName string,
	eligibleLBs []string,
	existingLBs *[]network.LoadBalancer,
) string {
	// 1. If the LB is eligible and being used, choose it.
	if StringInSlice(currentLBName, eligibleLBs) {
		klog.V(4).Infof("getMostEligibleLBForService: choose %s as it is eligible and being used", currentLBName)
		return currentLBName
	}

	// 2. If the LB is eligible and not created yet, choose it because it has the fewest rules.
	for _, eligibleLB := range eligibleLBs {
		var found bool
		if existingLBs != nil {
			for _, existingLB := range *existingLBs {
				if strings.EqualFold(trimSuffixIgnoreCase(pointer.StringDeref(existingLB.Name, ""), consts.InternalLoadBalancerNameSuffix), eligibleLB) {
					found = true
					break
				}
			}
		}
		if !found {
			klog.V(4).Infof("getMostEligibleLBForService: choose %s as it is eligible and not existing", eligibleLB)
			return eligibleLB
		}
	}

	// 3. If all eligible LBs are existing, choose the one with the fewest rules.
	var expectedLBName string
	ruleCount := 301
	if existingLBs != nil {
		for _, existingLB := range *existingLBs {
			if StringInSlice(trimSuffixIgnoreCase(pointer.StringDeref(existingLB.Name, ""), consts.InternalLoadBalancerNameSuffix), eligibleLBs) {
				if existingLB.LoadBalancerPropertiesFormat != nil &&
					existingLB.LoadBalancingRules != nil {
					if len(*existingLB.LoadBalancingRules) < ruleCount {
						ruleCount = len(*existingLB.LoadBalancingRules)
						expectedLBName = pointer.StringDeref(existingLB.Name, "")
					}
				}
			}
		}
	}

	if expectedLBName != "" {
		klog.V(4).Infof("getMostEligibleLBForService: choose %s with fewest %d rules", expectedLBName, ruleCount)
	}

	return trimSuffixIgnoreCase(expectedLBName, consts.InternalLoadBalancerNameSuffix)
}

func (az *Cloud) getServiceCurrentLoadBalancerName(service *v1.Service) string {
	for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
		if az.isLoadBalancerInUseByService(service, multiSLBConfig) {
			return multiSLBConfig.Name
		}
	}
	return ""
}

// getEligibleLoadBalancersForService filter out the eligible load balancers for the service.
// It follows four kinds of constraints:
// 1. Service annotation `service.beta.kubernetes.io/azure-load-balancer-configurations: lb1,lb2`.
// 2. AllowServicePlacement flag. Default to true, if set to false, the new services will not be put onto the LB.
// But the existing services that is using the LB will not be affected.
// 3. ServiceLabelSelector. The service will be put onto the LB only if the service has the labels specified in the selector.
// 4. ServiceNamespaceSelector. The service will be put onto the LB only if the service is in the namespaces specified in the selector.
// 5. If there is no label/namespace selector on the LB, it can be a valid placement target if and only if the service has no other choice.
func (az *Cloud) getEligibleLoadBalancersForService(service *v1.Service) ([]string, error) {
	var (
		eligibleLBs               []MultipleStandardLoadBalancerConfiguration
		eligibleLBNames           []string
		lbSelectedByAnnotation    []string
		lbFailedLabelSelector     []string
		lbFailedNamespaceSelector []string
		lbFailedPlacementFlag     []string
	)

	logger := klog.Background().
		WithName("getEligibleLoadBalancersForService").
		WithValues("service", service.Name)

	// 1. Service selects LBs defined in the annotation.
	// If there is no annotation given, it selects all LBs.
	lbsFromAnnotation := consts.GetLoadBalancerConfigurationsNames(service)
	if len(lbsFromAnnotation) > 0 {
		lbNamesSet := utilsets.NewString(lbsFromAnnotation...)
		for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
			if lbNamesSet.Has(multiSLBConfig.Name) {
				logger.V(4).Info("selects the load balancer by annotation",
					"load balancer configuration name", multiSLBConfig.Name)
				eligibleLBs = append(eligibleLBs, multiSLBConfig)
				lbSelectedByAnnotation = append(lbSelectedByAnnotation, multiSLBConfig.Name)
			}
		}
		if len(lbSelectedByAnnotation) == 0 {
			return nil, fmt.Errorf("service %q selects %d load balancers by annotation, but none of them is defined in cloud provider configuration", service.Name, len(lbsFromAnnotation))
		}
	} else {
		logger.V(4).Info("the service does not select any load balancer by annotation, all load balancers are eligible")
		eligibleLBs = append(eligibleLBs, az.MultipleStandardLoadBalancerConfigurations...)
		for _, eligibleLB := range eligibleLBs {
			lbSelectedByAnnotation = append(lbSelectedByAnnotation, eligibleLB.Name)
		}
	}

	var selectorMatched bool
	for i := len(eligibleLBs) - 1; i >= 0; i-- {
		eligibleLB := eligibleLBs[i]

		// 2. If the LB does not allow service placement, it is not eligible,
		// unless the service is already using the LB.
		if !pointer.BoolDeref(eligibleLB.AllowServicePlacement, true) {
			if az.isLoadBalancerInUseByService(service, eligibleLB) {
				logger.V(4).Info("although the load balancer has AllowServicePlacement=false, service is allowed to be placed on load balancer because it is using the load balancer",
					"load balancer configuration name", eligibleLB.Name)
			} else {
				logger.V(4).Info("the load balancer has AllowServicePlacement=false, service is not allowed to be placed on load balancer",
					"load balancer configuration name", eligibleLB.Name)
				eligibleLBs = append(eligibleLBs[:i], eligibleLBs[i+1:]...)
				lbFailedPlacementFlag = append(lbFailedPlacementFlag, eligibleLB.Name)
				continue
			}
		}

		// 3. Check the service label selector. The service can be migrated from one LB to another LB
		// if the service does not match the selector of the LB that it is currently using.
		if eligibleLB.ServiceLabelSelector != nil {
			serviceLabelSelector, err := metav1.LabelSelectorAsSelector(eligibleLB.ServiceLabelSelector)
			if err != nil {
				logger.Error(err, "failed to parse label selector",
					"label selector", eligibleLB.ServiceLabelSelector.String(),
					"load balancer configuration name", eligibleLB.Name)
				return []string{}, err
			}
			if !serviceLabelSelector.Matches(labels.Set(service.Labels)) {
				logger.V(2).Info("service does not match the label selector",
					"label selector", eligibleLB.ServiceLabelSelector.String(),
					"load balancer configuration name", eligibleLB.Name)
				eligibleLBs = append(eligibleLBs[:i], eligibleLBs[i+1:]...)
				lbFailedLabelSelector = append(lbFailedLabelSelector, eligibleLB.Name)
				continue
			}
			logger.V(4).Info("service matches the label selector",
				"label selector", eligibleLB.ServiceLabelSelector.String(),
				"load balancer configuration name", eligibleLB.Name)
			selectorMatched = true
		}

		// 4. Check the service namespace selector. The service can be migrated from one LB to another LB
		// if the service does not match the selector of the LB that it is currently using.
		if eligibleLB.ServiceNamespaceSelector != nil {
			serviceNamespaceSelector, err := metav1.LabelSelectorAsSelector(eligibleLB.ServiceNamespaceSelector)
			if err != nil {
				logger.Error(err, "failed to parse namespace selector",
					"namespace selector", eligibleLB.ServiceNamespaceSelector.String(),
					"load balancer configuration name", eligibleLB.Name)
				return []string{}, err
			}
			ns, err := az.KubeClient.CoreV1().Namespaces().Get(context.Background(), service.Namespace, metav1.GetOptions{})
			if err != nil {
				logger.Error(err, "failed to get namespace",
					"namespace", service.Namespace,
					"load balancer configuration name", eligibleLB.Name)
				return []string{}, err
			}
			if !serviceNamespaceSelector.Matches(labels.Set(ns.Labels)) {
				logger.V(2).Info("namespace does not match the namespace selector",
					"namespace", service.Namespace,
					"namespace selector", eligibleLB.ServiceNamespaceSelector.String(),
					"load balancer configuration name", eligibleLB.Name)
				eligibleLBs = append(eligibleLBs[:i], eligibleLBs[i+1:]...)
				lbFailedNamespaceSelector = append(lbFailedNamespaceSelector, eligibleLB.Name)
				continue
			}
			logger.V(4).Info("namespace matches the namespace selector",
				"namespace", service.Namespace,
				"namespace selector", eligibleLB.ServiceNamespaceSelector.String(),
				"load balancer configuration name", eligibleLB.Name)
			selectorMatched = true
		}
	}

	serviceName := getServiceName(service)
	if len(eligibleLBs) == 0 {
		return []string{}, fmt.Errorf(
			"service %q selects %d load balancers (%s), but %d of them (%s) have AllowServicePlacement set to false and the service is not using any of them, %d of them (%s) do not match the service label selector, and %d of them (%s) do not match the service namespace selector",
			serviceName,
			len(lbSelectedByAnnotation),
			strings.Join(lbSelectedByAnnotation, ", "),
			len(lbFailedPlacementFlag),
			strings.Join(lbFailedPlacementFlag, ", "),
			len(lbFailedLabelSelector),
			strings.Join(lbFailedLabelSelector, ", "),
			len(lbFailedNamespaceSelector),
			strings.Join(lbFailedNamespaceSelector, ", "),
		)
	}

	if selectorMatched {
		for i := len(eligibleLBs) - 1; i >= 0; i-- {
			eligibleLB := eligibleLBs[i]
			if eligibleLB.ServiceLabelSelector == nil && eligibleLB.ServiceNamespaceSelector == nil {
				logger.V(6).Info("service matches at least one label/namespace selector of the load balancer, so it should not be placed on the load balancer that does not have any label/namespace selector",
					"load balancer configuration name", eligibleLB.Name)
				eligibleLBs = append(eligibleLBs[:i], eligibleLBs[i+1:]...)
			}
		}
	} else {
		logger.V(4).Info("no load balancer that has label/namespace selector matches the service, so the service can be placed on the load balancers that do not have label/namespace selector")
	}

	for _, eligibleLB := range eligibleLBs {
		eligibleLBNames = append(eligibleLBNames, eligibleLB.Name)
	}

	return eligibleLBNames, nil
}

func (az *Cloud) isLoadBalancerInUseByService(service *v1.Service, lbConfig MultipleStandardLoadBalancerConfiguration) bool {
	az.multipleStandardLoadBalancersActiveServicesLock.Lock()
	defer az.multipleStandardLoadBalancersActiveServicesLock.Unlock()

	serviceName := getServiceName(service)
	return lbConfig.ActiveServices.Has(serviceName)
}

// There are two cases when a service owns the frontend IP config:
// 1. The primary service, which means the frontend IP config is created after the creation of the service.
// This means the name of the config can be tracked by the service UID.
// 2. The secondary services must have their loadBalancer IP set if they want to share the same config as the primary
// service. Hence, it can be tracked by the loadBalancer IP.
// If the IP version is not empty, which means it is the secondary Service, it returns IP version of the Service FIP.
func (az *Cloud) serviceOwnsFrontendIP(fip network.FrontendIPConfiguration, service *v1.Service) (bool, bool, network.IPVersion) {
	var isPrimaryService bool
	baseName := az.GetLoadBalancerName(context.TODO(), "", service)
	if strings.HasPrefix(pointer.StringDeref(fip.Name, ""), baseName) {
		klog.V(6).Infof("serviceOwnsFrontendIP: found primary service %s of the frontend IP config %s", service.Name, *fip.Name)
		isPrimaryService = true
		return true, isPrimaryService, ""
	}

	loadBalancerIPs := getServiceLoadBalancerIPs(service)
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	var pipNames []string
	if len(loadBalancerIPs) == 0 {
		if !requiresInternalLoadBalancer(service) {
			pipNames = getServicePIPNames(service)
			for _, pipName := range pipNames {
				if pipName != "" {
					pip, err := az.findMatchedPIP("", pipName, pipResourceGroup)
					if err != nil {
						klog.Warningf("serviceOwnsFrontendIP: unexpected error when finding match public IP of the service %s with name %s: %v", service.Name, pipName, err)
						return false, isPrimaryService, ""
					}
					if publicIPOwnsFrontendIP(service, &fip, pip) {
						return true, isPrimaryService, pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion
					}
				}
			}
		}
		// it is a must that the secondary services set the loadBalancer IP or pip name
		return false, isPrimaryService, ""
	}

	// for external secondary service the public IP address should be checked
	if !requiresInternalLoadBalancer(service) {
		for _, loadBalancerIP := range loadBalancerIPs {
			pip, err := az.findMatchedPIP(loadBalancerIP, "", pipResourceGroup)
			if err != nil {
				klog.Warningf("serviceOwnsFrontendIP: unexpected error when finding match public IP of the service %s with loadBalancerIP %s: %v", service.Name, loadBalancerIP, err)
				return false, isPrimaryService, ""
			}

			if publicIPOwnsFrontendIP(service, &fip, pip) {
				return true, isPrimaryService, pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion
			}
			klog.V(6).Infof("serviceOwnsFrontendIP: the public IP with ID %s is being referenced by other service with public IP address %s "+
				"OR it is of incorrect IP version", *pip.ID, *pip.IPAddress)
		}

		return false, isPrimaryService, ""
	}

	// for internal secondary service the private IP address on the frontend IP config should be checked
	if fip.PrivateIPAddress == nil {
		return false, isPrimaryService, ""
	}
	privateIPAddrVersion := network.IPv4
	if net.ParseIP(*fip.PrivateIPAddress).To4() == nil {
		privateIPAddrVersion = network.IPv6
	}

	privateIPEquals := false
	for _, loadBalancerIP := range loadBalancerIPs {
		if strings.EqualFold(*fip.PrivateIPAddress, loadBalancerIP) {
			privateIPEquals = true
			break
		}
	}
	return privateIPEquals, isPrimaryService, privateIPAddrVersion
}

func (az *Cloud) getFrontendIPConfigNames(service *v1.Service) map[bool]string {
	isDualStack := isServiceDualStack(service)
	defaultLBFrontendIPConfigName := az.getDefaultFrontendIPConfigName(service)
	return map[bool]string{
		consts.IPVersionIPv4: getResourceByIPFamily(defaultLBFrontendIPConfigName, isDualStack, consts.IPVersionIPv4),
		consts.IPVersionIPv6: getResourceByIPFamily(defaultLBFrontendIPConfigName, isDualStack, consts.IPVersionIPv6),
	}
}

func (az *Cloud) getDefaultFrontendIPConfigName(service *v1.Service) string {
	baseName := az.GetLoadBalancerName(context.TODO(), "", service)
	subnetName := getInternalSubnet(service)
	if subnetName != nil {
		ipcName := fmt.Sprintf("%s-%s", baseName, *subnetName)

		// Azure lb front end configuration name must not exceed 80 characters
		maxLength := consts.FrontendIPConfigNameMaxLength - consts.IPFamilySuffixLength
		if len(ipcName) > maxLength {
			ipcName = ipcName[:maxLength]
			// Cutting the string may result in char like "-" as the string end.
			// If the last char is not a letter or '_', replace it with "_".
			if !unicode.IsLetter(rune(ipcName[len(ipcName)-1:][0])) && ipcName[len(ipcName)-1:] != "_" {
				ipcName = ipcName[:len(ipcName)-1] + "_"
			}
		}
		return ipcName
	}
	return baseName
}
