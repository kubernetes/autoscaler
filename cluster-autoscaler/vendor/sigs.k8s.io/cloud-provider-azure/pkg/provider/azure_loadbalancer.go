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
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-02-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	cloudprovider "k8s.io/cloud-provider"
	servicehelpers "k8s.io/cloud-provider/service/helpers"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// GetLoadBalancer returns whether the specified load balancer and its components exist, and
// if so, what its status is.
func (az *Cloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	// Since public IP is not a part of the load balancer on Azure,
	// there is a chance that we could orphan public IP resources while we delete the load blanacer (kubernetes/kubernetes#80571).
	// We need to make sure the existence of the load balancer depends on the load balancer resource and public IP resource on Azure.
	existsPip := func() bool {
		pipName, _, err := az.determinePublicIPName(clusterName, service)
		if err != nil {
			return false
		}
		pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
		_, existsPip, err := az.getPublicIPAddress(pipResourceGroup, pipName)
		if err != nil {
			return false
		}
		return existsPip
	}()

	_, status, existsLb, err := az.getServiceLoadBalancer(service, clusterName, nil, false, []network.LoadBalancer{})
	if err != nil {
		return nil, existsPip, err
	}

	// Return exists = false only if the load balancer and the public IP are not found on Azure
	if !existsLb && !existsPip {
		serviceName := getServiceName(service)
		klog.V(5).Infof("getloadbalancer (cluster:%s) (service:%s) - doesn't exist", clusterName, serviceName)
		return nil, false, nil
	}

	// Return exists = true if either the load balancer or the public IP (or both) exists
	return status, true, nil
}

func getPublicIPDomainNameLabel(service *v1.Service) (string, bool) {
	if labelName, found := service.Annotations[consts.ServiceAnnotationDNSLabelName]; found {
		return labelName, found
	}
	return "", false
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
func (az *Cloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	// When a client updates the internal load balancer annotation,
	// the service may be switched from an internal LB to a public one, or vise versa.
	// Here we'll firstly ensure service do not lie in the opposite LB.
	serviceName := getServiceName(service)
	klog.V(5).Infof("ensureloadbalancer(%s): START clusterName=%q, service: %v", serviceName, clusterName, service)

	mc := metrics.NewMetricContext("services", "ensure_loadbalancer", az.ResourceGroup, az.SubscriptionID, serviceName)
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	lb, err := az.reconcileLoadBalancer(clusterName, service, nodes, true /* wantLb */)
	if err != nil {
		klog.Errorf("reconcileLoadBalancer(%s) failed: %v", serviceName, err)
		return nil, err
	}

	lbStatus, _, err := az.getServiceLoadBalancerStatus(service, lb)
	if err != nil {
		klog.Errorf("getServiceLoadBalancerStatus(%s) failed: %v", serviceName, err)
		return nil, err
	}

	var serviceIP *string
	if lbStatus != nil && len(lbStatus.Ingress) > 0 {
		serviceIP = &lbStatus.Ingress[0].IP
	}
	klog.V(2).Infof("EnsureLoadBalancer: reconciling security group for service %q with IP %q, wantLb = true", serviceName, logSafe(serviceIP))
	if _, err := az.reconcileSecurityGroup(clusterName, service, serviceIP, true /* wantLb */); err != nil {
		klog.Errorf("reconcileSecurityGroup(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	updateService := updateServiceLoadBalancerIP(service, to.String(serviceIP))
	flippedService := flipServiceInternalAnnotation(updateService)
	if _, err := az.reconcileLoadBalancer(clusterName, flippedService, nil, false /* wantLb */); err != nil {
		klog.Errorf("reconcileLoadBalancer(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	// lb is not reused here because the ETAG may be changed in above operations, hence reconcilePublicIP() would get lb again from cache.
	klog.V(2).Infof("EnsureLoadBalancer: reconciling pip")
	if _, err := az.reconcilePublicIP(clusterName, updateService, to.String(lb.Name), true /* wantLb */); err != nil {
		klog.Errorf("reconcilePublicIP(%s) failed: %#v", serviceName, err)
		return nil, err
	}

	isOperationSucceeded = true
	return lbStatus, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
func (az *Cloud) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	shouldUpdateLB, err := az.shouldUpdateLoadBalancer(clusterName, service, nodes)
	if err != nil {
		return err
	}

	if !shouldUpdateLB {
		klog.V(2).Infof("UpdateLoadBalancer: skipping service %s because it is either being deleted or does not exist anymore", service.Name)
		return nil
	}

	_, err = az.EnsureLoadBalancer(ctx, clusterName, service, nodes)
	return err
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
func (az *Cloud) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	klog.V(5).Infof("Delete service (%s): START clusterName=%q", serviceName, clusterName)

	mc := metrics.NewMetricContext("services", "ensure_loadbalancer_deleted", az.ResourceGroup, az.SubscriptionID, serviceName)
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	serviceIPToCleanup, err := az.findServiceIPAddress(ctx, clusterName, service, isInternal)
	if err != nil && !retry.HasStatusForbiddenOrIgnoredError(err) {
		return err
	}

	klog.V(2).Infof("EnsureLoadBalancerDeleted: reconciling security group for service %q with IP %q, wantLb = false", serviceName, serviceIPToCleanup)
	if _, err := az.reconcileSecurityGroup(clusterName, service, &serviceIPToCleanup, false /* wantLb */); err != nil {
		return err
	}

	if _, err := az.reconcileLoadBalancer(clusterName, service, nil, false /* wantLb */); err != nil && !retry.HasStatusForbiddenOrIgnoredError(err) {
		return err
	}

	if _, err := az.reconcilePublicIP(clusterName, service, "", false /* wantLb */); err != nil {
		return err
	}

	klog.V(2).Infof("Delete service (%s): FINISH", serviceName)
	isOperationSucceeded = true

	return nil
}

// GetLoadBalancerName returns the LoadBalancer name.
func (az *Cloud) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
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
func (az *Cloud) shouldChangeLoadBalancer(service *v1.Service, currLBName, clusterName string) bool {
	hasMode, isAuto, vmSetName := az.getServiceLoadBalancerMode(service)

	// if no mode is given or the mode is `__auto__`, the current LB should be kept
	if !hasMode || isAuto {
		return false
	}

	// if using the single standard load balancer, the current LB should be kept
	useSingleSLB := az.useStandardLoadBalancer() && !az.EnableMultipleStandardLoadBalancers
	if useSingleSLB {
		return false
	}

	// if the current LB is what we want, keep it
	lbName := strings.TrimSuffix(currLBName, consts.InternalLoadBalancerNameSuffix)
	if strings.EqualFold(lbName, vmSetName) {
		return false
	}
	if strings.EqualFold(vmSetName, az.VMSet.GetPrimaryVMSetName()) && strings.EqualFold(clusterName, lbName) {
		return false
	}

	// if the vmSet selected by the annotation is sharing the primary slb, and the service
	// has been associated to the primary slb, keep it
	useMultipleSLBs := az.useStandardLoadBalancer() && az.EnableMultipleStandardLoadBalancers
	if useMultipleSLBs &&
		az.getVMSetNamesSharingPrimarySLB().Has(strings.ToLower(vmSetName)) &&
		strings.EqualFold(lbName, clusterName) {
		return false
	}

	// if the VMSS/VMAS of the current LB is different from the mode, change the LB
	// to another one
	klog.V(2).Infof("shouldChangeLoadBalancer(%s, %s, %s): change the LB to another one", service.Name, currLBName, clusterName)
	return true
}

func (az *Cloud) removeFrontendIPConfigurationFromLoadBalancer(lb *network.LoadBalancer, existingLBs []network.LoadBalancer, fip *network.FrontendIPConfiguration, clusterName string, service *v1.Service) error {
	if lb == nil || lb.LoadBalancerPropertiesFormat == nil || lb.FrontendIPConfigurations == nil {
		return nil
	}
	fipConfigs := *lb.FrontendIPConfigurations
	for i, fipConfig := range fipConfigs {
		if strings.EqualFold(to.String(fipConfig.Name), to.String(fip.Name)) {
			fipConfigs = append(fipConfigs[:i], fipConfigs[i+1:]...)
			break
		}
	}
	lb.FrontendIPConfigurations = &fipConfigs

	// also remove the corresponding rules/probes
	if lb.LoadBalancingRules != nil {
		lbRules := *lb.LoadBalancingRules
		for i := len(lbRules) - 1; i >= 0; i-- {
			if strings.Contains(to.String(lbRules[i].Name), to.String(fip.Name)) {
				lbRules = append(lbRules[:i], lbRules[i+1:]...)
			}
		}
		lb.LoadBalancingRules = &lbRules
	}
	if lb.Probes != nil {
		lbProbes := *lb.Probes
		for i := len(lbProbes) - 1; i >= 0; i-- {
			if strings.Contains(to.String(lbProbes[i].Name), to.String(fip.Name)) {
				lbProbes = append(lbProbes[:i], lbProbes[i+1:]...)
			}
		}
		lb.Probes = &lbProbes
	}

	if len(fipConfigs) == 0 {
		klog.V(2).Infof("removeFrontendIPConfigurationFromLoadBalancer(%s, %s, %s, %s): deleting load balancer because there is no remaining frontend IP configurations", to.String(lb.Name), to.String(fip.Name), clusterName, service.Name)
		err := az.cleanOrphanedLoadBalancer(lb, existingLBs, service, clusterName)
		if err != nil {
			klog.Errorf("removeFrontendIPConfigurationFromLoadBalancer(%s, %s, %s, %s): failed to cleanupOrphanedLoadBalancer: %v", to.String(lb.Name), to.String(fip.Name), clusterName, service.Name, err)
			return err
		}
	} else {
		klog.V(2).Infof("removeFrontendIPConfigurationFromLoadBalancer(%s, %s, %s, %s): updating the load balancer", to.String(lb.Name), to.String(fip.Name), clusterName, service.Name)
		err := az.CreateOrUpdateLB(service, *lb)
		if err != nil {
			klog.Errorf("removeFrontendIPConfigurationFromLoadBalancer(%s, %s, %s, %s): failed to CreateOrUpdateLB: %v", to.String(lb.Name), to.String(fip.Name), clusterName, service.Name, err)
			return err
		}
		_ = az.lbCache.Delete(to.String(lb.Name))
	}
	return nil
}

func (az *Cloud) cleanOrphanedLoadBalancer(lb *network.LoadBalancer, existingLBs []network.LoadBalancer, service *v1.Service, clusterName string) error {
	lbName := to.String(lb.Name)
	serviceName := getServiceName(service)
	isBackendPoolPreConfigured := az.isBackendPoolPreConfigured(service)
	lbResourceGroup := az.getLoadBalancerResourceGroup()
	lbBackendPoolName := getBackendPoolName(clusterName, service)
	lbBackendPoolID := az.getBackendPoolID(lbName, lbResourceGroup, lbBackendPoolName)
	if isBackendPoolPreConfigured {
		klog.V(2).Infof("cleanOrphanedLoadBalancer(%s, %s, %s): ignore cleanup of dirty lb because the lb is pre-configured", lbName, serviceName, clusterName)
	} else {
		foundLB := false
		for _, existingLB := range existingLBs {
			if strings.EqualFold(to.String(lb.Name), to.String(existingLB.Name)) {
				foundLB = true
				break
			}
		}
		if !foundLB {
			klog.V(2).Infof("cleanOrphanedLoadBalancer: the LB %s doesn't exist, will not delete it", to.String(lb.Name))
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

		deleteErr := az.safeDeleteLoadBalancer(*lb, clusterName, vmSetName, service)
		if deleteErr != nil {
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
			err := az.VMSet.EnsureBackendPoolDeletedFromVMSets(vmssNamesMap, lbBackendPoolID)
			if err != nil {
				klog.Errorf("cleanOrphanedLoadBalancer(%s, %s, %s): failed to EnsureBackendPoolDeletedFromVMSets: %v", lbName, serviceName, clusterName, err)
				return err
			}

			deleteErr := az.DeleteLB(service, lbName)
			if deleteErr != nil {
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
	if strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration) {
		lbBackendPoolID := az.getBackendPoolID(to.String(lb.Name), az.getLoadBalancerResourceGroup(), getBackendPoolName(clusterName, service))
		err := az.VMSet.EnsureBackendPoolDeleted(service, lbBackendPoolID, vmSetName, lb.BackendAddressPools, true)
		if err != nil {
			return retry.NewError(false, fmt.Errorf("safeDeleteLoadBalancer: failed to EnsureBackendPoolDeleted: %w", err))
		}
	}

	klog.V(2).Infof("safeDeleteLoadBalancer: deleting LB %s because the corresponding vmSet is supposed to be in the primary SLB", to.String(lb.Name))
	rerr := az.DeleteLB(service, to.String(lb.Name))
	if rerr != nil {
		return rerr
	}
	_ = az.lbCache.Delete(to.String(lb.Name))

	return nil
}

func extractBackendIPConfigurationIDsFromLB(lb network.LoadBalancer, lbBackendPoolName string) []string {
	result := make([]string, 0)
	if lb.LoadBalancerPropertiesFormat != nil &&
		lb.BackendAddressPools != nil {
		for i := 0; i < len(*lb.BackendAddressPools); i++ {
			backendPool := (*lb.BackendAddressPools)[i]
			if strings.EqualFold(to.String(backendPool.Name), lbBackendPoolName) {
				if backendPool.BackendAddressPoolPropertiesFormat != nil &&
					backendPool.BackendIPConfigurations != nil {
					for _, ipConfiguration := range *backendPool.BackendIPConfigurations {
						if ipConfiguration.ID != nil {
							result = append(result, to.String(ipConfiguration.ID))
						}
					}
				}
			}
		}
	}

	return result
}

// reconcileSharedLoadBalancer deletes the dedicated SLBs of the non-primary vmSets. There are
// two scenarios where this operation is needed:
// 1. Using multiple slbs and the vmSet is supposed to share the primary slb.
// 2. When migrating from multiple slbs to single slb mode.
// It also ensures those vmSets are joint the backend pools of the primary SLBs.
// It runs only once everytime the cloud controller manager restarts.
func (az *Cloud) reconcileSharedLoadBalancer(service *v1.Service, clusterName string, nodes []*v1.Node) ([]network.LoadBalancer, error) {
	var (
		primarySLBs, existingLBs []network.LoadBalancer
		changed                  bool
		err                      error
	)

	existingLBs, err = az.ListManagedLBs(service, nodes, clusterName)
	if err != nil {
		return nil, fmt.Errorf("reconcileSharedLoadBalancer: failed to list managed LB: %w", err)
	}

	// only run once since the controller manager rebooted
	if az.isSharedLoadBalancerSynced {
		return existingLBs, nil
	}
	defer func() {
		if err == nil {
			az.isSharedLoadBalancerSynced = true
		}
	}()

	// skip if the cluster is using basic LB
	if !az.useStandardLoadBalancer() {
		return existingLBs, nil
	}

	lbBackendPoolName := getBackendPoolName(clusterName, service)
	lbNamesToBeDeleted := sets.NewString()
	// 1: delete unwanted LBs
	for _, lb := range existingLBs {
		lbNamePrefix := strings.TrimSuffix(to.String(lb.Name), consts.InternalLoadBalancerNameSuffix)

		// skip the internal or external primary load balancer
		if strings.EqualFold(lbNamePrefix, clusterName) {
			primarySLBs = append(primarySLBs, lb)
			continue
		}

		// skip if the multiple slbs mode is enabled and
		// the vmSet is supposed to have dedicated SLBs
		vmSetName := strings.ToLower(az.mapLoadBalancerNameToVMSet(to.String(lb.Name), clusterName))
		if az.EnableMultipleStandardLoadBalancers && !az.getVMSetNamesSharingPrimarySLB().Has(vmSetName) {
			continue
		}

		// For non-primary load balancer, the lb name is the name of the VMSet.
		// If the VMSet name is in az.NodePoolsWithoutDedicatedSLB, we should
		// decouple the VMSet from the lb and delete the lb. Then adding the VMSet
		// to the backend pool of the primary slb.
		rerr := az.safeDeleteLoadBalancer(lb, clusterName, vmSetName, service)
		if rerr != nil {
			return nil, rerr.Error()
		}

		// remove the deleted lb from the list and construct a new primary
		// lb, so that getServiceLoadBalancer doesn't have to call list api again
		lbNamesToBeDeleted.Insert(strings.ToLower(to.String(lb.Name)))
		changed = true
	}

	if !changed {
		klog.V(4).Infof("reconcileSharedLoadBalancer: no changes made, return now")
		return existingLBs, nil
	}

	vmSetsToBeMovedToPrimarySLB := sets.NewString()
	ipConfigIDsToBeAddedToPrimarySLB := sets.NewString()
	// 2: add nodes to the backend pool of the primary SLBs
	for i := len(existingLBs) - 1; i >= 0; i-- {
		lb := existingLBs[i]
		if !lbNamesToBeDeleted.Has(strings.ToLower(to.String(lb.Name))) {
			continue
		}

		vmSetName := strings.ToLower(az.mapLoadBalancerNameToVMSet(to.String(lb.Name), clusterName))
		vmSetsToBeMovedToPrimarySLB.Insert(vmSetName)
		isInternalLB := strings.HasSuffix(to.String(lb.Name), consts.InternalLoadBalancerNameSuffix)
		primarySLBName := clusterName
		if isInternalLB {
			primarySLBName = fmt.Sprintf("%s%s", clusterName, consts.InternalLoadBalancerNameSuffix)
		}
		primaryLBBackendPoolID := az.getBackendPoolID(primarySLBName, az.getLoadBalancerResourceGroup(), getBackendPoolName(clusterName, service))

		klog.V(2).Infof("reconcileSharedLoadBalancer: binding the vmSet %s to the backend pool %s", vmSetName, primaryLBBackendPoolID)
		if strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration) {
			err = az.VMSet.EnsureHostsInPool(service, nodes, primaryLBBackendPoolID, vmSetName)
			if err != nil {
				return nil, fmt.Errorf("reconcileSharedLoadBalancer: failed to EnsureHostsInPool: %w", err)
			}

			for _, id := range extractBackendIPConfigurationIDsFromLB(lb, lbBackendPoolName) {
				ipConfigIDsToBeAddedToPrimarySLB.Insert(id)
			}
		}

		// remove the deleted LB from the list
		existingLBs = append(existingLBs[:i], existingLBs[i+1:]...)
	}

	for _, primarySLB := range primarySLBs {
		if primarySLB.LoadBalancerPropertiesFormat != nil &&
			primarySLB.BackendAddressPools != nil {
			for i := 0; i < len(*primarySLB.BackendAddressPools); i++ {
				if strings.EqualFold(to.String((*primarySLB.BackendAddressPools)[i].Name), lbBackendPoolName) {
					if az.isLBBackendPoolTypeNodeIPConfig() {
						backendPoolIPConfigs := (*primarySLB.BackendAddressPools)[i].BackendIPConfigurations
						for _, id := range ipConfigIDsToBeAddedToPrimarySLB.List() {
							*backendPoolIPConfigs = append(*backendPoolIPConfigs, network.InterfaceIPConfiguration{
								ID: to.StringPtr(id),
							})
						}
					} else if az.isLBBackendPoolTypeNodeIP() {
						backendPool := (*primarySLB.BackendAddressPools)[i]
						if backendPool.LoadBalancerBackendAddresses == nil {
							lbBackendPoolAddresses := make([]network.LoadBalancerBackendAddress, 0)
							backendPool.LoadBalancerBackendAddresses = &lbBackendPoolAddresses
						}

						if err := az.LoadBalancerBackendPool.EnsureHostsInPool(service, nodes, "", "", clusterName, to.String(primarySLB.Name), backendPool); err != nil {
							return nil, fmt.Errorf("reconcileSharedLoadBalancer: failed to EnsureHostsInPool: %w", err)
						}

						(*primarySLB.BackendAddressPools)[i] = backendPool
					}

					break
				}
			}
		}
	}

	for i, existingLB := range existingLBs {
		for _, primarySLB := range primarySLBs {
			if strings.EqualFold(to.String(existingLB.Name), to.String(primarySLB.Name)) {
				// Proactively disable the etag to prevent etag mismatch error when putting lb later.
				// This could happen because when we remove the hosts from the lb, the nrp
				// would put the lb to remove the backend references as well.
				primarySLB.Etag = nil

				existingLBs[i] = primarySLB
			}
		}
	}

	return existingLBs, nil
}

// getServiceLoadBalancer gets the loadbalancer for the service if it already exists.
// If wantLb is TRUE then -it selects a new load balancer.
// In case the selected load balancer does not exist it returns network.LoadBalancer struct
// with added metadata (such as name, location) and existsLB set to FALSE.
// By default - cluster default LB is returned.
func (az *Cloud) getServiceLoadBalancer(service *v1.Service, clusterName string, nodes []*v1.Node, wantLb bool, existingLBs []network.LoadBalancer) (lb *network.LoadBalancer, status *v1.LoadBalancerStatus, exists bool, err error) {
	isInternal := requiresInternalLoadBalancer(service)
	var defaultLB *network.LoadBalancer
	primaryVMSetName := az.VMSet.GetPrimaryVMSetName()
	defaultLBName := az.getAzureLoadBalancerName(clusterName, primaryVMSetName, isInternal)
	useMultipleSLBs := az.useStandardLoadBalancer() && az.EnableMultipleStandardLoadBalancers

	// reuse the lb list from reconcileSharedLoadBalancer to reduce the api call
	if len(existingLBs) == 0 {
		existingLBs, err = az.ListLB(service)
		if err != nil {
			return nil, nil, false, err
		}
	}

	// check if the service already has a load balancer
	for i := range existingLBs {
		existingLB := existingLBs[i]
		existingLBNamePrefix := strings.TrimSuffix(to.String(existingLB.Name), consts.InternalLoadBalancerNameSuffix)

		// for the primary standard load balancer (internal or external), when enabled multiple slbs
		if strings.EqualFold(existingLBNamePrefix, clusterName) && useMultipleSLBs {
			shouldRemoveVMSetFromSLB := func(vmSetName string) bool {
				// not removing the vmSet from the primary SLB
				// if it is supposed to share the primary SLB.
				if az.getVMSetNamesSharingPrimarySLB().Has(strings.ToLower(vmSetName)) {
					return false
				}

				// removing the vmSet from the primary SLB if
				// it is not the primary vmSet. There are two situations:
				// 1. when migrating from single SLB to multiple SLBs, we
				// need to remove all non-primary vmSets from the primary SLB;
				// 2. when migrating from shared mode to dedicated SLB, we
				// need to remove the specific vmSet from the primary SLB.
				return !strings.EqualFold(vmSetName, primaryVMSetName) && vmSetName != ""
			}
			cleanedLB, err := az.LoadBalancerBackendPool.CleanupVMSetFromBackendPoolByCondition(&existingLB, service, nodes, clusterName, shouldRemoveVMSetFromSLB)
			if err != nil {
				return nil, nil, false, err
			}
			existingLB = *cleanedLB
			existingLBs[i] = *cleanedLB
		}
		if strings.EqualFold(*existingLB.Name, defaultLBName) {
			defaultLB = &existingLB
		}
		if isInternalLoadBalancer(&existingLB) != isInternal {
			continue
		}
		status, fipConfig, err := az.getServiceLoadBalancerStatus(service, &existingLB)
		if err != nil {
			return nil, nil, false, err
		}
		if status == nil {
			// service is not on this load balancer
			continue
		}

		// select another load balancer instead of returning
		// the current one if the change is needed
		if wantLb && az.shouldChangeLoadBalancer(service, to.String(existingLB.Name), clusterName) {
			if err := az.removeFrontendIPConfigurationFromLoadBalancer(&existingLB, existingLBs, fipConfig, clusterName, service); err != nil {
				klog.Errorf("getServiceLoadBalancer(%s, %s, %v): failed to remove frontend IP configuration from load balancer: %v", service.Name, clusterName, wantLb, err)
				return nil, nil, false, err
			}
			break
		}

		return &existingLB, status, true, nil
	}

	// Service does not have a load balancer, select one.
	// Single standard load balancer doesn't need this because
	// all backends nodes should be added to same LB.
	useSingleSLB := az.useStandardLoadBalancer() && !az.EnableMultipleStandardLoadBalancers
	if wantLb && !useSingleSLB {
		// select new load balancer for service
		selectedLB, exists, err := az.selectLoadBalancer(clusterName, service, &existingLBs, nodes)
		if err != nil {
			return nil, nil, false, err
		}

		return selectedLB, nil, exists, err
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

	return defaultLB, nil, false, nil
}

// selectLoadBalancer selects load balancer for the service in the cluster.
// The selection algorithm selects the load balancer which currently has
// the minimum lb rules. If there are multiple LBs with same number of rules,
// then selects the first one (sorted based on name).
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
		currLBName := az.getAzureLoadBalancerName(clusterName, currVMSetName, isInternal)
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

func (az *Cloud) getServiceLoadBalancerStatus(service *v1.Service, lb *network.LoadBalancer) (status *v1.LoadBalancerStatus, fipConfig *network.FrontendIPConfiguration, err error) {
	if lb == nil {
		klog.V(10).Info("getServiceLoadBalancerStatus: lb is nil")
		return nil, nil, nil
	}
	if lb.FrontendIPConfigurations == nil || *lb.FrontendIPConfigurations == nil {
		klog.V(10).Info("getServiceLoadBalancerStatus: lb.FrontendIPConfigurations is nil")
		return nil, nil, nil
	}
	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	for _, ipConfiguration := range *lb.FrontendIPConfigurations {
		owns, isPrimaryService, err := az.serviceOwnsFrontendIP(ipConfiguration, service)
		if err != nil {
			return nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to filter frontend IP configs with error: %w", serviceName, to.String(lb.Name), err)
		}
		if owns {
			klog.V(2).Infof("get(%s): lb(%s) - found frontend IP config, primary service: %v", serviceName, to.String(lb.Name), isPrimaryService)

			var lbIP *string
			if isInternal {
				lbIP = ipConfiguration.PrivateIPAddress
			} else {
				if ipConfiguration.PublicIPAddress == nil {
					return nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress is Nil", serviceName, *lb.Name)
				}
				pipID := ipConfiguration.PublicIPAddress.ID
				if pipID == nil {
					return nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress ID is Nil", serviceName, *lb.Name)
				}
				pipName, err := getLastSegment(*pipID, "/")
				if err != nil {
					return nil, nil, fmt.Errorf("get(%s): lb(%s) - failed to get LB PublicIPAddress Name from ID(%s)", serviceName, *lb.Name, *pipID)
				}
				pip, existsPip, err := az.getPublicIPAddress(az.getPublicIPAddressResourceGroup(service), pipName)
				if err != nil {
					return nil, nil, err
				}
				if existsPip {
					lbIP = pip.IPAddress
				}
			}

			klog.V(2).Infof("getServiceLoadBalancerStatus gets ingress IP %q from frontendIPConfiguration %q for service %q", to.String(lbIP), to.String(ipConfiguration.Name), serviceName)

			// set additional public IPs to LoadBalancerStatus, so that kube-proxy would create their iptables rules.
			lbIngress := []v1.LoadBalancerIngress{{IP: to.String(lbIP)}}
			additionalIPs, err := getServiceAdditionalPublicIPs(service)
			if err != nil {
				return &v1.LoadBalancerStatus{Ingress: lbIngress}, &ipConfiguration, err
			}
			if len(additionalIPs) > 0 {
				for _, pip := range additionalIPs {
					lbIngress = append(lbIngress, v1.LoadBalancerIngress{
						IP: pip,
					})
				}
			}

			return &v1.LoadBalancerStatus{Ingress: lbIngress}, &ipConfiguration, nil
		}
	}

	return nil, nil, nil
}

func (az *Cloud) determinePublicIPName(clusterName string, service *v1.Service) (string, bool, error) {
	var shouldPIPExisted bool
	if name, found := service.Annotations[consts.ServiceAnnotationPIPName]; found && name != "" {
		shouldPIPExisted = true
		return name, shouldPIPExisted, nil
	}

	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	loadBalancerIP := service.Spec.LoadBalancerIP

	// Assume that the service without loadBalancerIP set is a primary service.
	// If a secondary service doesn't set the loadBalancerIP, it is not allowed to share the IP.
	if len(loadBalancerIP) == 0 {
		return az.getPublicIPName(clusterName, service), shouldPIPExisted, nil
	}

	// For the services with loadBalancerIP set, an existing public IP is required, primary
	// or secondary, or a public IP not found error would be reported.
	pip, err := az.findMatchedPIPByLoadBalancerIP(service, loadBalancerIP, pipResourceGroup)
	if err != nil {
		return "", shouldPIPExisted, err
	}

	if pip != nil && pip.Name != nil {
		return *pip.Name, shouldPIPExisted, nil
	}

	return "", shouldPIPExisted, fmt.Errorf("user supplied IP Address %s was not found in resource group %s", loadBalancerIP, pipResourceGroup)
}

func (az *Cloud) findMatchedPIPByLoadBalancerIP(service *v1.Service, loadBalancerIP, pipResourceGroup string) (*network.PublicIPAddress, error) {
	pips, err := az.ListPIP(service, pipResourceGroup)
	if err != nil {
		return nil, err
	}

	for _, pip := range pips {
		if pip.PublicIPAddressPropertiesFormat.IPAddress != nil &&
			*pip.PublicIPAddressPropertiesFormat.IPAddress == loadBalancerIP {
			return &pip, nil
		}
	}

	return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: cannot find public IP with IP address %s in resource group %s", loadBalancerIP, pipResourceGroup)
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

func updateServiceLoadBalancerIP(service *v1.Service, serviceIP string) *v1.Service {
	copyService := service.DeepCopy()
	if len(serviceIP) > 0 && copyService != nil {
		copyService.Spec.LoadBalancerIP = serviceIP
	}
	return copyService
}

func (az *Cloud) findServiceIPAddress(ctx context.Context, clusterName string, service *v1.Service, isInternalLb bool) (string, error) {
	if len(service.Spec.LoadBalancerIP) > 0 {
		return service.Spec.LoadBalancerIP, nil
	}

	if len(service.Status.LoadBalancer.Ingress) > 0 && len(service.Status.LoadBalancer.Ingress[0].IP) > 0 {
		return service.Status.LoadBalancer.Ingress[0].IP, nil
	}

	_, lbStatus, existsLb, err := az.getServiceLoadBalancer(service, clusterName, nil, false, []network.LoadBalancer{})
	if err != nil {
		return "", err
	}
	if !existsLb {
		klog.V(2).Infof("Expected to find an IP address for service %s but did not. Assuming it has been removed", service.Name)
		return "", nil
	}
	if len(lbStatus.Ingress) < 1 {
		klog.V(2).Infof("Expected to find an IP address for service %s but it had no ingresses. Assuming it has been removed", service.Name)
		return "", nil
	}

	return lbStatus.Ingress[0].IP, nil
}

func (az *Cloud) ensurePublicIPExists(service *v1.Service, pipName string, domainNameLabel, clusterName string, shouldPIPExisted, foundDNSLabelAnnotation bool) (*network.PublicIPAddress, error) {
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	pip, existsPip, err := az.getPublicIPAddress(pipResourceGroup, pipName)
	if err != nil {
		return nil, err
	}

	serviceName := getServiceName(service)

	var changed bool
	if existsPip {
		// ensure that the service tag is good for managed pips
		owns, isUserAssignedPIP := serviceOwnsPublicIP(service, &pip, clusterName)
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

		klog.V(2).Infof("ensurePublicIPExists for service(%s): pip(%s) - updating", serviceName, to.String(pip.Name))
		if pip.PublicIPAddressPropertiesFormat == nil {
			pip.PublicIPAddressPropertiesFormat = &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.IPAllocationMethodStatic,
			}
			changed = true
		}
	} else {
		if shouldPIPExisted {
			return nil, fmt.Errorf("PublicIP from annotation azure-pip-name=%s for service %s doesn't exist", pipName, serviceName)
		}

		changed = true

		pip.Name = to.StringPtr(pipName)
		pip.Location = to.StringPtr(az.Location)
		if az.HasExtendedLocation() {
			klog.V(2).Infof("Using extended location with name %s, and type %s for PIP", az.ExtendedLocationName, az.ExtendedLocationType)
			pip.ExtendedLocation = &network.ExtendedLocation{
				Name: &az.ExtendedLocationName,
				Type: getExtendedLocationTypeFromString(az.ExtendedLocationType),
			}
		}
		pip.PublicIPAddressPropertiesFormat = &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.IPAllocationMethodStatic,
			IPTags:                   getServiceIPTagRequestForPublicIP(service).IPTags,
		}
		pip.Tags = map[string]*string{
			consts.ServiceTagKey:  to.StringPtr(""),
			consts.ClusterNameKey: &clusterName,
		}
		if _, err = bindServicesToPIP(&pip, []string{serviceName}, false); err != nil {
			return nil, err
		}

		if az.useStandardLoadBalancer() {
			pip.Sku = &network.PublicIPAddressSku{
				Name: network.PublicIPAddressSkuNameStandard,
			}

			// skip adding zone info since edge zones doesn't support multiple availability zones.
			if !az.HasExtendedLocation() {
				// only add zone information for the new standard pips
				zones, err := az.getRegionZonesBackoff(to.String(pip.Location))
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

	if foundDNSLabelAnnotation {
		updatedDNSSettings, err := reconcileDNSSettings(&pip, domainNameLabel, serviceName, pipName)
		if err != nil {
			return nil, fmt.Errorf("ensurePublicIPExists for service(%s): failed to reconcileDNSSettings: %w", serviceName, err)
		}

		if updatedDNSSettings {
			changed = true
		}
	}

	// use the same family as the clusterIP as we support IPv6 single stack as well
	// as dual-stack clusters
	updatedIPSettings := az.reconcileIPSettings(&pip, service)
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

func (az *Cloud) reconcileIPSettings(pip *network.PublicIPAddress, service *v1.Service) bool {
	var changed bool

	serviceName := getServiceName(service)
	ipv6 := utilnet.IsIPv6String(service.Spec.ClusterIP)
	if ipv6 {
		klog.V(2).Infof("service(%s): pip(%s) - creating as ipv6 for clusterIP:%v", serviceName, *pip.Name, service.Spec.ClusterIP)

		if !strings.EqualFold(string(pip.PublicIPAddressVersion), string(network.IPVersionIPv6)) {
			pip.PublicIPAddressVersion = network.IPVersionIPv6
			changed = true
		}

		if az.useStandardLoadBalancer() {
			// standard sku must have static allocation method for ipv6
			if !strings.EqualFold(string(pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod), string(network.IPAllocationMethodStatic)) {
				pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod = network.IPAllocationMethodStatic
				changed = true
			}
		} else if !strings.EqualFold(string(pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod), string(network.IPAllocationMethodDynamic)) {
			pip.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod = network.IPAllocationMethodDynamic
			changed = true
		}
	} else {
		klog.V(2).Infof("service(%s): pip(%s) - creating as ipv4 for clusterIP:%v", serviceName, *pip.Name, service.Spec.ClusterIP)

		if !strings.EqualFold(string(pip.PublicIPAddressVersion), string(network.IPVersionIPv6)) {
			pip.PublicIPAddressVersion = network.IPVersionIPv4
			changed = true
		}
	}

	return changed
}

func reconcileDNSSettings(pip *network.PublicIPAddress, domainNameLabel, serviceName, pipName string) (bool, error) {
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
			if !strings.EqualFold(to.String(existingDNSLabel), domainNameLabel) {
				return false, fmt.Errorf("ensurePublicIPExists for service(%s): pip(%s) - there is an existing DNS label %s on the public IP", serviceName, pipName, *existingDNSLabel)
			}
		}

		if svc := getServiceFromPIPDNSTags(pip.Tags); svc == "" || !strings.EqualFold(svc, serviceName) {
			pip.Tags[consts.ServiceUsingDNSKey] = &serviceName
			changed = true
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
			return to.String(ipTag[i].IPTagType) < to.String(ipTag[j].IPTagType) ||
				to.String(ipTag[i].Tag) < to.String(ipTag[j].Tag)
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
			IPTagType: to.StringPtr(k),
			Tag:       to.StringPtr(v),
		}
		outputTags = append(outputTags, ipTag)
	}

	return &outputTags
}

func getDomainNameLabel(pip *network.PublicIPAddress) string {
	if pip == nil || pip.PublicIPAddressPropertiesFormat == nil || pip.PublicIPAddressPropertiesFormat.DNSSettings == nil {
		return ""
	}
	return to.String(pip.PublicIPAddressPropertiesFormat.DNSSettings.DomainNameLabel)
}

func getIdleTimeout(s *v1.Service) (*int32, error) {
	const (
		min = 4
		max = 30
	)

	val, ok := s.Annotations[consts.ServiceAnnotationLoadBalancerIdleTimeout]
	if !ok {
		// Return a nil here as this will set the value to the azure default
		return nil, nil
	}

	errInvalidTimeout := fmt.Errorf("idle timeout value must be a whole number representing minutes between %d and %d", min, max)
	toInt, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error parsing idle timeout value: %w: %v", err, errInvalidTimeout)
	}
	to32 := int32(toInt)

	if to32 < min || to32 > max {
		return nil, errInvalidTimeout
	}
	return &to32, nil
}

func (az *Cloud) isFrontendIPChanged(clusterName string, config network.FrontendIPConfiguration, service *v1.Service, lbFrontendIPConfigName string) (bool, error) {
	isServiceOwnsFrontendIP, isPrimaryService, err := az.serviceOwnsFrontendIP(config, service)
	if err != nil {
		return false, err
	}
	if isServiceOwnsFrontendIP && isPrimaryService && !strings.EqualFold(to.String(config.Name), lbFrontendIPConfigName) {
		return true, nil
	}
	if !strings.EqualFold(to.String(config.Name), lbFrontendIPConfigName) {
		return false, nil
	}
	loadBalancerIP := service.Spec.LoadBalancerIP
	isInternal := requiresInternalLoadBalancer(service)
	if isInternal {
		// Judge subnet
		subnetName := subnet(service)
		if subnetName != nil {
			subnet, existsSubnet, err := az.getSubnet(az.VnetName, *subnetName)
			if err != nil {
				return false, err
			}
			if !existsSubnet {
				return false, fmt.Errorf("failed to get subnet")
			}
			if config.Subnet != nil && !strings.EqualFold(to.String(config.Subnet.Name), to.String(subnet.Name)) {
				return true, nil
			}
		}
		if loadBalancerIP == "" {
			return config.PrivateIPAllocationMethod == network.IPAllocationMethodStatic, nil
		}
		return config.PrivateIPAllocationMethod != network.IPAllocationMethodStatic || !strings.EqualFold(loadBalancerIP, to.String(config.PrivateIPAddress)), nil
	}
	pipName, _, err := az.determinePublicIPName(clusterName, service)
	if err != nil {
		return false, err
	}
	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)
	pip, existsPip, err := az.getPublicIPAddress(pipResourceGroup, pipName)
	if err != nil {
		return false, err
	}
	if !existsPip {
		return true, nil
	}
	return config.PublicIPAddress != nil && !strings.EqualFold(to.String(pip.ID), to.String(config.PublicIPAddress.ID)), nil
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

func (az *Cloud) findFrontendIPConfigOfService(
	fipConfigs *[]network.FrontendIPConfiguration,
	service *v1.Service,
) (*network.FrontendIPConfiguration, bool, error) {
	for _, config := range *fipConfigs {
		owns, isPrimaryService, err := az.serviceOwnsFrontendIP(config, service)
		if err != nil {
			return nil, false, err
		}
		if owns {
			return &config, isPrimaryService, nil
		}
	}

	return nil, false, nil
}

// reconcileLoadBalancer ensures load balancer exists and the frontend ip config is setup.
// This also reconciles the Service's Ports  with the LoadBalancer config.
// This entails adding rules/probes for expected Ports and removing stale rules/ports.
// nodes only used if wantLb is true
func (az *Cloud) reconcileLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node, wantLb bool) (*network.LoadBalancer, error) {
	isInternal := requiresInternalLoadBalancer(service)
	isBackendPoolPreConfigured := az.isBackendPoolPreConfigured(service)
	serviceName := getServiceName(service)
	klog.V(2).Infof("reconcileLoadBalancer for service(%s) - wantLb(%t): started", serviceName, wantLb)

	existingLBs, err := az.reconcileSharedLoadBalancer(service, clusterName, nodes)
	if err != nil {
		klog.Errorf("reconcileLoadBalancer: failed to reconcile shared load balancer: %v", err)
		return nil, err
	}

	lb, _, _, err := az.getServiceLoadBalancer(service, clusterName, nodes, wantLb, existingLBs)
	if err != nil {
		klog.Errorf("reconcileLoadBalancer: failed to get load balancer for service %q, error: %v", serviceName, err)
		return nil, err
	}
	lbName := *lb.Name
	lbResourceGroup := az.getLoadBalancerResourceGroup()
	lbBackendPoolID := az.getBackendPoolID(lbName, az.getLoadBalancerResourceGroup(), getBackendPoolName(clusterName, service))
	klog.V(2).Infof("reconcileLoadBalancer for service(%s): lb(%s/%s) wantLb(%t) resolved load balancer name", serviceName, lbResourceGroup, lbName, wantLb)
	defaultLBFrontendIPConfigName := az.getDefaultFrontendIPConfigName(service)
	defaultLBFrontendIPConfigID := az.getFrontendIPConfigID(lbName, lbResourceGroup, defaultLBFrontendIPConfigName)
	dirtyLb := false

	lbIdleTimeout, err := getIdleTimeout(service)
	if wantLb && err != nil {
		return nil, err
	}

	// reconcile the load balancer's backend pool configuration.
	if wantLb {
		preConfig, changed, err := az.LoadBalancerBackendPool.ReconcileBackendPools(clusterName, service, lb)
		if err != nil {
			return lb, err
		}
		if changed {
			dirtyLb = true
		}
		isBackendPoolPreConfigured = preConfig
	}

	// reconcile the load balancer's frontend IP configurations.
	ownedFIPConfig, changed, err := az.reconcileFrontendIPConfigs(clusterName, service, lb, wantLb, defaultLBFrontendIPConfigName)
	if err != nil {
		return lb, err
	}
	if changed {
		dirtyLb = true
	}

	// update probes/rules
	if ownedFIPConfig != nil {
		if ownedFIPConfig.ID != nil {
			defaultLBFrontendIPConfigID = *ownedFIPConfig.ID
		} else {
			return nil, fmt.Errorf("reconcileLoadBalancer for service (%s)(%t): nil ID for frontend IP config", serviceName, wantLb)
		}
	}

	if wantLb {
		err = az.checkLoadBalancerResourcesConflicts(lb, defaultLBFrontendIPConfigID, service)
		if err != nil {
			return nil, err
		}
	}

	expectedProbes, expectedRules, err := az.getExpectedLBRules(service, wantLb, defaultLBFrontendIPConfigID, lbBackendPoolID, lbName, lbIdleTimeout)
	if err != nil {
		return nil, err
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
		if lb.FrontendIPConfigurations == nil || len(*lb.FrontendIPConfigurations) == 0 {
			err := az.cleanOrphanedLoadBalancer(lb, existingLBs, service, clusterName)
			if err != nil {
				klog.Errorf("reconcileLoadBalancer for service(%s): lb(%s) - failed to cleanOrphanedLoadBalancer: %v", serviceName, lbName, err)
				return nil, err
			}
		} else {
			klog.V(2).Infof("reconcileLoadBalancer: reconcileLoadBalancer for service(%s): lb(%s) - updating", serviceName, lbName)
			err := az.CreateOrUpdateLB(service, *lb)
			if err != nil {
				klog.Errorf("reconcileLoadBalancer for service(%s) abort backoff: lb(%s) - updating", serviceName, lbName)
				return nil, err
			}

			if isInternal {
				// Refresh updated lb which will be used later in other places.
				newLB, exist, err := az.getAzureLoadBalancer(lbName, azcache.CacheReadTypeDefault)
				if err != nil {
					klog.Errorf("reconcileLoadBalancer for service(%s): getAzureLoadBalancer(%s) failed: %v", serviceName, lbName, err)
					return nil, err
				}
				if !exist {
					return nil, fmt.Errorf("load balancer %q not found", lbName)
				}
				lb = &newLB
			}
		}
	}

	if wantLb && nodes != nil && !isBackendPoolPreConfigured {
		// Add the machines to the backend pool if they're not already
		vmSetName := az.mapLoadBalancerNameToVMSet(lbName, clusterName)
		// Etag would be changed when updating backend pools, so invalidate lbCache after it.
		defer func() {
			_ = az.lbCache.Delete(lbName)
		}()

		if lb.LoadBalancerPropertiesFormat != nil && lb.BackendAddressPools != nil {
			backendPools := *lb.BackendAddressPools
			for _, backendPool := range backendPools {
				if strings.EqualFold(to.String(backendPool.Name), getBackendPoolName(clusterName, service)) {
					if err := az.LoadBalancerBackendPool.EnsureHostsInPool(service, nodes, lbBackendPoolID, vmSetName, clusterName, lbName, backendPool); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	klog.V(2).Infof("reconcileLoadBalancer for service(%s): lb(%s) finished", serviceName, lbName)
	return lb, nil
}

func (az *Cloud) reconcileLBProbes(lb *network.LoadBalancer, service *v1.Service, serviceName string, wantLb bool, expectedProbes []network.Probe) bool {
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

func (az *Cloud) reconcileFrontendIPConfigs(clusterName string, service *v1.Service, lb *network.LoadBalancer, wantLb bool, defaultLBFrontendIPConfigName string) (*network.FrontendIPConfiguration, bool, error) {
	var err error
	lbName := *lb.Name
	serviceName := getServiceName(service)
	isInternal := requiresInternalLoadBalancer(service)
	dirtyConfigs := false
	var newConfigs []network.FrontendIPConfiguration
	if lb.FrontendIPConfigurations != nil {
		newConfigs = *lb.FrontendIPConfigurations
	}

	var ownedFIPConfig *network.FrontendIPConfiguration
	if !wantLb {
		for i := len(newConfigs) - 1; i >= 0; i-- {
			config := newConfigs[i]
			isServiceOwnsFrontendIP, _, err := az.serviceOwnsFrontendIP(config, service)
			if err != nil {
				return nil, false, err
			}
			if isServiceOwnsFrontendIP {
				unsafe, err := az.isFrontendIPConfigUnsafeToDelete(lb, service, config.ID)
				if err != nil {
					return nil, false, err
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

					newConfigs = append(newConfigs[:i], newConfigs[i+1:]...)
					dirtyConfigs = true
				}
			}
		}
	} else {
		for i := len(newConfigs) - 1; i >= 0; i-- {
			config := newConfigs[i]
			isFipChanged, err := az.isFrontendIPChanged(clusterName, config, service, defaultLBFrontendIPConfigName)
			if err != nil {
				return nil, false, err
			}
			if isFipChanged {
				klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb frontendconfig(%s) - dropping", serviceName, wantLb, *config.Name)
				newConfigs = append(newConfigs[:i], newConfigs[i+1:]...)
				dirtyConfigs = true
			}
		}

		ownedFIPConfig, _, err = az.findFrontendIPConfigOfService(&newConfigs, service)
		if err != nil {
			return nil, false, err
		}

		if ownedFIPConfig == nil {
			klog.V(4).Infof("ensure(%s): lb(%s) - creating a new frontend IP config", serviceName, lbName)

			// construct FrontendIPConfigurationPropertiesFormat
			var fipConfigurationProperties *network.FrontendIPConfigurationPropertiesFormat
			if isInternal {
				subnetName := subnet(service)
				if subnetName == nil {
					subnetName = &az.SubnetName
				}
				subnet, existsSubnet, err := az.getSubnet(az.VnetName, *subnetName)
				if err != nil {
					return nil, false, err
				}

				if !existsSubnet {
					return nil, false, fmt.Errorf("ensure(%s): lb(%s) - failed to get subnet: %s/%s", serviceName, lbName, az.VnetName, az.SubnetName)
				}

				configProperties := network.FrontendIPConfigurationPropertiesFormat{
					Subnet: &subnet,
				}

				if utilnet.IsIPv6String(service.Spec.ClusterIP) {
					configProperties.PrivateIPAddressVersion = network.IPVersionIPv6
				}

				loadBalancerIP := service.Spec.LoadBalancerIP
				if loadBalancerIP != "" {
					configProperties.PrivateIPAllocationMethod = network.IPAllocationMethodStatic
					configProperties.PrivateIPAddress = &loadBalancerIP
				} else {
					// We'll need to call GetLoadBalancer later to retrieve allocated IP.
					configProperties.PrivateIPAllocationMethod = network.IPAllocationMethodDynamic
				}

				fipConfigurationProperties = &configProperties
			} else {
				pipName, shouldPIPExisted, err := az.determinePublicIPName(clusterName, service)
				if err != nil {
					return nil, false, err
				}
				domainNameLabel, found := getPublicIPDomainNameLabel(service)
				pip, err := az.ensurePublicIPExists(service, pipName, domainNameLabel, clusterName, shouldPIPExisted, found)
				if err != nil {
					return nil, false, err
				}
				fipConfigurationProperties = &network.FrontendIPConfigurationPropertiesFormat{
					PublicIPAddress: &network.PublicIPAddress{ID: pip.ID},
				}
			}

			newConfig := network.FrontendIPConfiguration{
				Name:                                    to.StringPtr(defaultLBFrontendIPConfigName),
				ID:                                      to.StringPtr(fmt.Sprintf(consts.FrontendIPConfigIDTemplate, az.SubscriptionID, az.ResourceGroup, *lb.Name, defaultLBFrontendIPConfigName)),
				FrontendIPConfigurationPropertiesFormat: fipConfigurationProperties,
			}

			// only add zone information for new internal frontend IP configurations for standard load balancer not deployed to an edge zone.
			location := az.Location
			zones, err := az.getRegionZonesBackoff(location)
			if err != nil {
				return nil, false, err
			}
			if isInternal && az.useStandardLoadBalancer() && len(zones) > 0 && !az.HasExtendedLocation() {
				newConfig.Zones = &zones
			}
			newConfigs = append(newConfigs, newConfig)
			klog.V(2).Infof("reconcileLoadBalancer for service (%s)(%t): lb frontendconfig(%s) - adding", serviceName, wantLb, defaultLBFrontendIPConfigName)
			dirtyConfigs = true
		}
	}

	if dirtyConfigs {
		lb.FrontendIPConfigurations = &newConfigs
	}

	return ownedFIPConfig, dirtyConfigs, err
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

func parseHealthProbeProtocolAndPath(service *v1.Service) (string, string) {
	var protocol, path string
	if v, ok := service.Annotations[consts.ServiceAnnotationLoadBalancerHealthProbeProtocol]; ok {
		protocol = v
	} else {
		return protocol, path
	}
	// ignore the request path if using TCP
	if strings.EqualFold(protocol, string(network.ProbeProtocolHTTP)) ||
		strings.EqualFold(protocol, string(network.ProbeProtocolHTTPS)) {
		if v, ok := service.Annotations[consts.ServiceAnnotationLoadBalancerHealthProbeRequestPath]; ok {
			path = v
		}
	}
	return protocol, path
}

func (az *Cloud) getExpectedLBRules(
	service *v1.Service,
	wantLb bool,
	lbFrontendIPConfigID string,
	lbBackendPoolID string,
	lbName string,
	lbIdleTimeout *int32) ([]network.Probe, []network.LoadBalancingRule, error) {

	var ports []v1.ServicePort
	if wantLb {
		ports = service.Spec.Ports
	} else {
		ports = []v1.ServicePort{}
	}

	var enableTCPReset *bool
	if az.useStandardLoadBalancer() {
		enableTCPReset = to.BoolPtr(true)
	}

	var expectedProbes []network.Probe
	var expectedRules []network.LoadBalancingRule
	highAvailabilityPortsEnabled := false
	for _, port := range ports {
		if !requiresInternalLoadBalancer(service) && port.Protocol == v1.ProtocolSCTP {
			return nil, nil, fmt.Errorf("SCTP is only supported on internal LoadBalancer")
		}

		if highAvailabilityPortsEnabled {
			// Since the port is always 0 when enabling HA, only one rule should be configured.
			break
		}

		lbRuleName := az.getLoadBalancerRuleName(service, port.Protocol, port.Port)
		klog.V(2).Infof("getExpectedLBRules lb name (%s) rule name (%s)", lbName, lbRuleName)

		transportProto, _, probeProto, err := getProtocolsFromKubernetesProtocol(port.Protocol)
		if err != nil {
			return expectedProbes, expectedRules, err
		}

		probeProtocol, requestPath := parseHealthProbeProtocolAndPath(service)
		if servicehelpers.NeedsHealthCheck(service) {
			podPresencePath, podPresencePort := servicehelpers.GetServiceHealthCheckPathPort(service)
			if probeProtocol == "" {
				probeProtocol = string(network.ProbeProtocolHTTP)
			}

			needRequestPath := strings.EqualFold(probeProtocol, string(network.ProbeProtocolHTTP)) || strings.EqualFold(probeProtocol, string(network.ProbeProtocolHTTPS))
			if requestPath == "" && needRequestPath {
				requestPath = podPresencePath
			}

			expectedProbes = append(expectedProbes, network.Probe{
				Name: &lbRuleName,
				ProbePropertiesFormat: &network.ProbePropertiesFormat{
					RequestPath:       to.StringPtr(requestPath),
					Protocol:          network.ProbeProtocol(probeProtocol),
					Port:              to.Int32Ptr(podPresencePort),
					IntervalInSeconds: to.Int32Ptr(5),
					NumberOfProbes:    to.Int32Ptr(2),
				},
			})
		} else if port.Protocol != v1.ProtocolUDP && port.Protocol != v1.ProtocolSCTP {
			// we only add the expected probe if we're doing TCP
			if probeProtocol == "" {
				probeProtocol = string(*probeProto)
			}
			var actualPath *string
			if !strings.EqualFold(probeProtocol, string(network.ProbeProtocolTCP)) {
				if requestPath != "" {
					actualPath = to.StringPtr(requestPath)
				} else {
					actualPath = to.StringPtr("/healthz")
				}
			}
			expectedProbes = append(expectedProbes, network.Probe{
				Name: &lbRuleName,
				ProbePropertiesFormat: &network.ProbePropertiesFormat{
					Protocol:          network.ProbeProtocol(probeProtocol),
					RequestPath:       actualPath,
					Port:              to.Int32Ptr(port.NodePort),
					IntervalInSeconds: to.Int32Ptr(5),
					NumberOfProbes:    to.Int32Ptr(2),
				},
			})
		}

		loadDistribution := network.LoadDistributionDefault
		if service.Spec.SessionAffinity == v1.ServiceAffinityClientIP {
			loadDistribution = network.LoadDistributionSourceIP
		}

		expectedRule := network.LoadBalancingRule{
			Name: &lbRuleName,
			LoadBalancingRulePropertiesFormat: &network.LoadBalancingRulePropertiesFormat{
				Protocol: *transportProto,
				FrontendIPConfiguration: &network.SubResource{
					ID: to.StringPtr(lbFrontendIPConfigID),
				},
				BackendAddressPool: &network.SubResource{
					ID: to.StringPtr(lbBackendPoolID),
				},
				LoadDistribution:    loadDistribution,
				FrontendPort:        to.Int32Ptr(port.Port),
				BackendPort:         to.Int32Ptr(port.Port),
				DisableOutboundSnat: to.BoolPtr(az.disableLoadBalancerOutboundSNAT()),
				EnableTCPReset:      enableTCPReset,
				EnableFloatingIP:    to.BoolPtr(true),
			},
		}

		if port.Protocol == v1.ProtocolTCP {
			expectedRule.LoadBalancingRulePropertiesFormat.IdleTimeoutInMinutes = lbIdleTimeout
		}

		if requiresInternalLoadBalancer(service) &&
			strings.EqualFold(az.LoadBalancerSku, consts.LoadBalancerSkuStandard) &&
			(strings.EqualFold(service.Annotations[consts.ServiceAnnotationLoadBalancerEnableHighAvailabilityPorts], consts.TrueAnnotationValue) || port.Protocol == v1.ProtocolSCTP) {
			expectedRule.FrontendPort = to.Int32Ptr(0)
			expectedRule.BackendPort = to.Int32Ptr(0)
			expectedRule.Protocol = network.TransportProtocolAll
			highAvailabilityPortsEnabled = true
		}

		// we didn't construct the probe objects for UDP or SCTP because they're not allowed on Azure.
		// However, when externalTrafficPolicy is Local, Kubernetes HTTP health check would be used for probing.
		if servicehelpers.NeedsHealthCheck(service) || (port.Protocol != v1.ProtocolUDP && port.Protocol != v1.ProtocolSCTP) {
			expectedRule.Probe = &network.SubResource{
				ID: to.StringPtr(az.getLoadBalancerProbeID(lbName, az.getLoadBalancerResourceGroup(), lbRuleName)),
			}
		}

		expectedRules = append(expectedRules, expectedRule)
	}

	return expectedProbes, expectedRules, nil
}

// This reconciles the Network Security Group similar to how the LB is reconciled.
// This entails adding required, missing SecurityRules and removing stale rules.
func (az *Cloud) reconcileSecurityGroup(clusterName string, service *v1.Service, lbIP *string, wantLb bool) (*network.SecurityGroup, error) {
	serviceName := getServiceName(service)
	klog.V(5).Infof("reconcileSecurityGroup(%s): START clusterName=%q", serviceName, clusterName)

	ports := service.Spec.Ports
	if ports == nil {
		if useSharedSecurityRule(service) {
			klog.V(2).Infof("Attempting to reconcile security group for service %s, but service uses shared rule and we don't know which port it's for", service.Name)
			return nil, fmt.Errorf("no port info for reconciling shared rule for service %s", service.Name)
		}
		ports = []v1.ServicePort{}
	}

	sg, err := az.getSecurityGroup(azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}

	destinationIPAddress := ""
	if wantLb && lbIP == nil {
		return nil, fmt.Errorf("no load balancer IP for setting up security rules for service %s", service.Name)
	}
	if lbIP != nil {
		destinationIPAddress = *lbIP
	}

	if destinationIPAddress == "" {
		destinationIPAddress = "*"
	}

	additionalIPs, err := getServiceAdditionalPublicIPs(service)
	if err != nil {
		return nil, fmt.Errorf("unable to get additional public IPs, error=%v", err)
	}

	destinationIPAddresses := []string{destinationIPAddress}
	if destinationIPAddress != "*" {
		destinationIPAddresses = append(destinationIPAddresses, additionalIPs...)
	}

	sourceRanges, err := servicehelpers.GetLoadBalancerSourceRanges(service)
	if err != nil {
		return nil, err
	}
	serviceTags := getServiceTags(service)
	if len(serviceTags) != 0 {
		delete(sourceRanges, consts.DefaultLoadBalancerSourceRanges)
	}

	var sourceAddressPrefixes []string
	if (sourceRanges == nil || servicehelpers.IsAllowAll(sourceRanges)) && len(serviceTags) == 0 {
		if !requiresInternalLoadBalancer(service) {
			sourceAddressPrefixes = []string{"Internet"}
		}
	} else {
		for _, ip := range sourceRanges {
			sourceAddressPrefixes = append(sourceAddressPrefixes, ip.String())
		}
		sourceAddressPrefixes = append(sourceAddressPrefixes, serviceTags...)
	}

	expectedSecurityRules, err := az.getExpectedSecurityRules(wantLb, ports, sourceAddressPrefixes, service, destinationIPAddresses, sourceRanges)
	if err != nil {
		return nil, err
	}

	// update security rules
	dirtySg, updatedRules, err := az.reconcileSecurityRules(sg, service, serviceName, wantLb, expectedSecurityRules, ports, sourceAddressPrefixes, destinationIPAddresses)
	if err != nil {
		return nil, err
	}

	changed := az.ensureSecurityGroupTagged(&sg)
	if changed {
		dirtySg = true
	}

	if dirtySg {
		sg.SecurityRules = &updatedRules
		klog.V(2).Infof("reconcileSecurityGroup for service(%s): sg(%s) - updating", serviceName, *sg.Name)
		klog.V(10).Infof("CreateOrUpdateSecurityGroup(%q): start", *sg.Name)
		err := az.CreateOrUpdateSecurityGroup(sg)
		if err != nil {
			klog.V(2).Infof("ensure(%s) abort backoff: sg(%s) - updating", serviceName, *sg.Name)
			return nil, err
		}
		klog.V(10).Infof("CreateOrUpdateSecurityGroup(%q): end", *sg.Name)
		_ = az.nsgCache.Delete(to.String(sg.Name))
	}
	return &sg, nil
}

func (az *Cloud) reconcileSecurityRules(sg network.SecurityGroup, service *v1.Service, serviceName string, wantLb bool, expectedSecurityRules []network.SecurityRule, ports []v1.ServicePort, sourceAddressPrefixes []string, destinationIPAddresses []string) (bool, []network.SecurityRule, error) {
	dirtySg := false
	var updatedRules []network.SecurityRule
	if sg.SecurityGroupPropertiesFormat != nil && sg.SecurityGroupPropertiesFormat.SecurityRules != nil {
		updatedRules = *sg.SecurityGroupPropertiesFormat.SecurityRules
	}

	for _, r := range updatedRules {
		klog.V(10).Infof("Existing security rule while processing %s: %s:%s -> %s:%s", service.Name, logSafe(r.SourceAddressPrefix), logSafe(r.SourcePortRange), logSafeCollection(r.DestinationAddressPrefix, r.DestinationAddressPrefixes), logSafe(r.DestinationPortRange))
	}

	// update security rules: remove unwanted rules that belong privately
	// to this service
	for i := len(updatedRules) - 1; i >= 0; i-- {
		existingRule := updatedRules[i]
		if az.serviceOwnsRule(service, *existingRule.Name) {
			klog.V(10).Infof("reconcile(%s)(%t): sg rule(%s) - considering evicting", serviceName, wantLb, *existingRule.Name)
			keepRule := false
			if findSecurityRule(expectedSecurityRules, existingRule) {
				klog.V(10).Infof("reconcile(%s)(%t): sg rule(%s) - keeping", serviceName, wantLb, *existingRule.Name)
				keepRule = true
			}
			if !keepRule {
				klog.V(10).Infof("reconcile(%s)(%t): sg rule(%s) - dropping", serviceName, wantLb, *existingRule.Name)
				updatedRules = append(updatedRules[:i], updatedRules[i+1:]...)
				dirtySg = true
			}
		}
	}

	// update security rules: if the service uses a shared rule and is being deleted,
	// then remove it from the shared rule
	if useSharedSecurityRule(service) && !wantLb {
		for _, port := range ports {
			for _, sourceAddressPrefix := range sourceAddressPrefixes {
				sharedRuleName := az.getSecurityRuleName(service, port, sourceAddressPrefix)
				sharedIndex, sharedRule, sharedRuleFound := findSecurityRuleByName(updatedRules, sharedRuleName)
				if !sharedRuleFound {
					klog.V(4).Infof("Didn't find shared rule %s for service %s", sharedRuleName, service.Name)
					continue
				}
				if sharedRule.DestinationAddressPrefixes == nil {
					klog.V(4).Infof("Didn't find DestinationAddressPrefixes in shared rule for service %s", service.Name)
					continue
				}
				existingPrefixes := *sharedRule.DestinationAddressPrefixes
				for _, destinationIPAddress := range destinationIPAddresses {
					addressIndex, found := findIndex(existingPrefixes, destinationIPAddress)
					if !found {
						klog.Warningf("Didn't find destination address %v in shared rule %s for service %s", destinationIPAddress, sharedRuleName, service.Name)
						continue
					}
					if len(existingPrefixes) == 1 {
						updatedRules = append(updatedRules[:sharedIndex], updatedRules[sharedIndex+1:]...)
					} else {
						newDestinations := append(existingPrefixes[:addressIndex], existingPrefixes[addressIndex+1:]...)
						sharedRule.DestinationAddressPrefixes = &newDestinations
						updatedRules[sharedIndex] = sharedRule
					}
					dirtySg = true
				}

			}
		}
	}

	// update security rules: prepare rules for consolidation
	for index, rule := range updatedRules {
		if allowsConsolidation(rule) {
			updatedRules[index] = makeConsolidatable(rule)
		}
	}
	for index, rule := range expectedSecurityRules {
		if allowsConsolidation(rule) {
			expectedSecurityRules[index] = makeConsolidatable(rule)
		}
	}
	// update security rules: add needed
	for _, expectedRule := range expectedSecurityRules {
		foundRule := false
		if findSecurityRule(updatedRules, expectedRule) {
			klog.V(10).Infof("reconcile(%s)(%t): sg rule(%s) - already exists", serviceName, wantLb, *expectedRule.Name)
			foundRule = true
		}
		if foundRule && allowsConsolidation(expectedRule) {
			index, _ := findConsolidationCandidate(updatedRules, expectedRule)
			updatedRules[index] = consolidate(updatedRules[index], expectedRule)
			dirtySg = true
		}
		if !foundRule {
			klog.V(10).Infof("reconcile(%s)(%t): sg rule(%s) - adding", serviceName, wantLb, *expectedRule.Name)

			nextAvailablePriority, err := getNextAvailablePriority(updatedRules)
			if err != nil {
				return false, nil, err
			}

			expectedRule.Priority = to.Int32Ptr(nextAvailablePriority)
			updatedRules = append(updatedRules, expectedRule)
			dirtySg = true
		}
	}

	for _, r := range updatedRules {
		klog.V(10).Infof("Updated security rule while processing %s: %s:%s -> %s:%s", service.Name, logSafe(r.SourceAddressPrefix), logSafe(r.SourcePortRange), logSafeCollection(r.DestinationAddressPrefix, r.DestinationAddressPrefixes), logSafe(r.DestinationPortRange))
	}
	return dirtySg, updatedRules, nil
}

func (az *Cloud) getExpectedSecurityRules(wantLb bool, ports []v1.ServicePort, sourceAddressPrefixes []string, service *v1.Service, destinationIPAddresses []string, sourceRanges utilnet.IPNetSet) ([]network.SecurityRule, error) {
	expectedSecurityRules := []network.SecurityRule{}

	if wantLb {
		expectedSecurityRules = make([]network.SecurityRule, len(ports)*len(sourceAddressPrefixes))

		for i, port := range ports {
			_, securityProto, _, err := getProtocolsFromKubernetesProtocol(port.Protocol)
			if err != nil {
				return nil, err
			}
			for j := range sourceAddressPrefixes {
				ix := i*len(sourceAddressPrefixes) + j
				securityRuleName := az.getSecurityRuleName(service, port, sourceAddressPrefixes[j])
				nsgRule := network.SecurityRule{
					Name: to.StringPtr(securityRuleName),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Protocol:             *securityProto,
						SourcePortRange:      to.StringPtr("*"),
						DestinationPortRange: to.StringPtr(strconv.Itoa(int(port.Port))),
						SourceAddressPrefix:  to.StringPtr(sourceAddressPrefixes[j]),
						Access:               network.SecurityRuleAccessAllow,
						Direction:            network.SecurityRuleDirectionInbound,
					},
				}
				if len(destinationIPAddresses) == 1 {
					// continue to use DestinationAddressPrefix to avoid NSG updates for existing rules.
					nsgRule.DestinationAddressPrefix = to.StringPtr(destinationIPAddresses[0])
				} else {
					nsgRule.DestinationAddressPrefixes = to.StringSlicePtr(destinationIPAddresses)
				}
				expectedSecurityRules[ix] = nsgRule
			}
		}

		shouldAddDenyRule := false
		if len(sourceRanges) > 0 && !servicehelpers.IsAllowAll(sourceRanges) {
			if v, ok := service.Annotations[consts.ServiceAnnotationDenyAllExceptLoadBalancerSourceRanges]; ok && strings.EqualFold(v, consts.TrueAnnotationValue) {
				shouldAddDenyRule = true
			}
		}
		if shouldAddDenyRule {
			for _, port := range ports {
				_, securityProto, _, err := getProtocolsFromKubernetesProtocol(port.Protocol)
				if err != nil {
					return nil, err
				}
				securityRuleName := az.getSecurityRuleName(service, port, "deny_all")
				nsgRule := network.SecurityRule{
					Name: to.StringPtr(securityRuleName),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Protocol:             *securityProto,
						SourcePortRange:      to.StringPtr("*"),
						DestinationPortRange: to.StringPtr(strconv.Itoa(int(port.Port))),
						SourceAddressPrefix:  to.StringPtr("*"),
						Access:               network.SecurityRuleAccessDeny,
						Direction:            network.SecurityRuleDirectionInbound,
					},
				}
				if len(destinationIPAddresses) == 1 {
					// continue to use DestinationAddressPrefix to avoid NSG updates for existing rules.
					nsgRule.DestinationAddressPrefix = to.StringPtr(destinationIPAddresses[0])
				} else {
					nsgRule.DestinationAddressPrefixes = to.StringSlicePtr(destinationIPAddresses)
				}
				expectedSecurityRules = append(expectedSecurityRules, nsgRule)
			}
		}
	}

	for _, r := range expectedSecurityRules {
		klog.V(10).Infof("Expecting security rule for %s: %s:%s -> %v %v :%s", service.Name, to.String(r.SourceAddressPrefix), to.String(r.SourcePortRange), to.String(r.DestinationAddressPrefix), to.StringSlice(r.DestinationAddressPrefixes), to.String(r.DestinationPortRange))
	}
	return expectedSecurityRules, nil
}

func (az *Cloud) shouldUpdateLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) (bool, error) {
	existingManagedLBs, err := az.ListManagedLBs(service, nodes, clusterName)
	if err != nil {
		return false, fmt.Errorf("shouldUpdateLoadBalancer: failed to list managed load balancers: %w", err)
	}

	_, _, existsLb, _ := az.getServiceLoadBalancer(service, clusterName, nodes, false, existingManagedLBs)
	return existsLb && service.ObjectMeta.DeletionTimestamp == nil, nil
}

func logSafe(s *string) string {
	if s == nil {
		return "(nil)"
	}
	return *s
}

func logSafeCollection(s *string, strs *[]string) string {
	if s == nil {
		if strs == nil {
			return "(nil)"
		}
		return "[" + strings.Join(*strs, ",") + "]"
	}
	return *s
}

func findSecurityRuleByName(rules []network.SecurityRule, ruleName string) (int, network.SecurityRule, bool) {
	for index, rule := range rules {
		if rule.Name != nil && strings.EqualFold(*rule.Name, ruleName) {
			return index, rule, true
		}
	}
	return 0, network.SecurityRule{}, false
}

func findIndex(strs []string, s string) (int, bool) {
	for index, str := range strs {
		if strings.EqualFold(str, s) {
			return index, true
		}
	}
	return 0, false
}

func allowsConsolidation(rule network.SecurityRule) bool {
	return strings.HasPrefix(to.String(rule.Name), "shared")
}

func findConsolidationCandidate(rules []network.SecurityRule, rule network.SecurityRule) (int, bool) {
	for index, r := range rules {
		if allowsConsolidation(r) {
			if strings.EqualFold(to.String(r.Name), to.String(rule.Name)) {
				return index, true
			}
		}
	}

	return 0, false
}

func makeConsolidatable(rule network.SecurityRule) network.SecurityRule {
	return network.SecurityRule{
		Name: rule.Name,
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Priority:                   rule.Priority,
			Protocol:                   rule.Protocol,
			SourcePortRange:            rule.SourcePortRange,
			SourcePortRanges:           rule.SourcePortRanges,
			DestinationPortRange:       rule.DestinationPortRange,
			DestinationPortRanges:      rule.DestinationPortRanges,
			SourceAddressPrefix:        rule.SourceAddressPrefix,
			SourceAddressPrefixes:      rule.SourceAddressPrefixes,
			DestinationAddressPrefixes: collectionOrSingle(rule.DestinationAddressPrefixes, rule.DestinationAddressPrefix),
			Access:                     rule.Access,
			Direction:                  rule.Direction,
		},
	}
}

func consolidate(existingRule network.SecurityRule, newRule network.SecurityRule) network.SecurityRule {
	destinations := appendElements(existingRule.SecurityRulePropertiesFormat.DestinationAddressPrefixes, newRule.DestinationAddressPrefix, newRule.DestinationAddressPrefixes)
	destinations = deduplicate(destinations) // there are transient conditions during controller startup where it tries to add a service that is already added

	return network.SecurityRule{
		Name: existingRule.Name,
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Priority:                   existingRule.Priority,
			Protocol:                   existingRule.Protocol,
			SourcePortRange:            existingRule.SourcePortRange,
			SourcePortRanges:           existingRule.SourcePortRanges,
			DestinationPortRange:       existingRule.DestinationPortRange,
			DestinationPortRanges:      existingRule.DestinationPortRanges,
			SourceAddressPrefix:        existingRule.SourceAddressPrefix,
			SourceAddressPrefixes:      existingRule.SourceAddressPrefixes,
			DestinationAddressPrefixes: destinations,
			Access:                     existingRule.Access,
			Direction:                  existingRule.Direction,
		},
	}
}

func collectionOrSingle(collection *[]string, s *string) *[]string {
	if collection != nil && len(*collection) > 0 {
		return collection
	}
	if s == nil {
		return &[]string{}
	}
	return &[]string{*s}
}

func appendElements(collection *[]string, appendString *string, appendStrings *[]string) *[]string {
	newCollection := []string{}

	if collection != nil {
		newCollection = append(newCollection, *collection...)
	}
	if appendString != nil {
		newCollection = append(newCollection, *appendString)
	}
	if appendStrings != nil {
		newCollection = append(newCollection, *appendStrings...)
	}

	return &newCollection
}

func deduplicate(collection *[]string) *[]string {
	if collection == nil {
		return nil
	}

	seen := map[string]bool{}
	result := make([]string, 0, len(*collection))

	for _, v := range *collection {
		if seen[v] {
			// skip this element
		} else {
			seen[v] = true
			result = append(result, v)
		}
	}

	return &result
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
		} else if !strings.EqualFold(to.String(v), to.String(configTags[key])) {
			configTags[key] = v
		}
	}

	// include the cluster name and service names tags when comparing
	var clusterName, serviceNames *string
	if v := getClusterFromPIPClusterTags(pip.Tags); v != "" {
		clusterName = &v
	}
	if v := getServiceFromPIPServiceTags(pip.Tags); v != "" {
		serviceNames = &v
	}
	if clusterName != nil {
		configTags[consts.ClusterNameKey] = clusterName
	}
	if serviceNames != nil {
		configTags[consts.ServiceTagKey] = serviceNames
	}

	tags, changed := az.reconcileTags(pip.Tags, configTags)
	pip.Tags = tags

	return changed
}

// This reconciles the PublicIP resources similar to how the LB is reconciled.
func (az *Cloud) reconcilePublicIP(clusterName string, service *v1.Service, lbName string, wantLb bool) (*network.PublicIPAddress, error) {
	isInternal := requiresInternalLoadBalancer(service)
	serviceName := getServiceName(service)
	serviceIPTagRequest := getServiceIPTagRequestForPublicIP(service)

	var (
		lb               *network.LoadBalancer
		desiredPipName   string
		err              error
		shouldPIPExisted bool
	)

	if !isInternal && wantLb {
		desiredPipName, shouldPIPExisted, err = az.determinePublicIPName(clusterName, service)
		if err != nil {
			return nil, err
		}
	}

	if lbName != "" {
		loadBalancer, _, err := az.getAzureLoadBalancer(lbName, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}
		lb = &loadBalancer
	}

	pipResourceGroup := az.getPublicIPAddressResourceGroup(service)

	pips, err := az.ListPIP(service, pipResourceGroup)
	if err != nil {
		return nil, err
	}

	discoveredDesiredPublicIP, pipsToBeDeleted, deletedDesiredPublicIP, pipsToBeUpdated, err := az.getPublicIPUpdates(clusterName, service, pips, wantLb, isInternal, desiredPipName, serviceName, serviceIPTagRequest, shouldPIPExisted)
	if err != nil {
		return nil, err
	}

	var deleteFuncs, updateFuncs []func() error
	for _, pip := range pipsToBeUpdated {
		pipCopy := *pip
		updateFuncs = append(updateFuncs, func() error {
			klog.V(2).Infof("reconcilePublicIP for service(%s): pip(%s) - updating", serviceName, *pip.Name)
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
			klog.V(2).Infof("reconcilePublicIP for service(%s): pip(%s) - deleting", serviceName, *pip.Name)
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
		if pip, err = az.ensurePublicIPExists(service, desiredPipName, domainNameLabel, clusterName, errorIfPublicIPDoesNotExist, found); err != nil {
			return nil, err
		}
		return pip, nil
	}
	return nil, nil
}

func (az *Cloud) getPublicIPUpdates(clusterName string, service *v1.Service, pips []network.PublicIPAddress, wantLb bool, isInternal bool, desiredPipName string, serviceName string, serviceIPTagRequest serviceIPTagRequest, serviceAnnotationRequestsNamedPublicIP bool) (bool, []*network.PublicIPAddress, bool, []*network.PublicIPAddress, error) {
	var (
		err                       error
		discoveredDesiredPublicIP bool
		deletedDesiredPublicIP    bool
		pipsToBeDeleted           []*network.PublicIPAddress
		pipsToBeUpdated           []*network.PublicIPAddress
	)
	for i := range pips {
		pip := pips[i]
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
				err = unbindServiceFromPIP(&pip, service, serviceName, clusterName)
				if err != nil {
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
		pip.PublicIPAddressPropertiesFormat.IPConfiguration != nil &&
		lb != nil && lb.LoadBalancerPropertiesFormat != nil &&
		lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations != nil {
		referencedLBRules := []network.SubResource{}
		frontendIPConfigUpdated := false
		loadBalancerRuleUpdated := false

		// Check whether there are still frontend IP configurations referring to it.
		ipConfigurationID := to.String(pip.PublicIPAddressPropertiesFormat.IPConfiguration.ID)
		if ipConfigurationID != "" {
			lbFrontendIPConfigs := *lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations
			for i := len(lbFrontendIPConfigs) - 1; i >= 0; i-- {
				config := lbFrontendIPConfigs[i]
				if strings.EqualFold(ipConfigurationID, to.String(config.ID)) {
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
			referencedLBRuleIDs := sets.NewString()
			for _, refer := range referencedLBRules {
				referencedLBRuleIDs.Insert(to.String(refer.ID))
			}

			if lb.LoadBalancerPropertiesFormat.LoadBalancingRules != nil {
				lbRules := *lb.LoadBalancerPropertiesFormat.LoadBalancingRules
				for i := len(lbRules) - 1; i >= 0; i-- {
					ruleID := to.String(lbRules[i].ID)
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

	pipName := to.String(pip.Name)
	klog.V(10).Infof("DeletePublicIP(%s, %q): start", pipResourceGroup, pipName)
	err := az.DeletePublicIP(service, pipResourceGroup, pipName)
	if err != nil {
		return err
	}
	klog.V(10).Infof("DeletePublicIP(%s, %q): end", pipResourceGroup, pipName)

	return nil
}

func findProbe(probes []network.Probe, probe network.Probe) bool {
	for _, existingProbe := range probes {
		if strings.EqualFold(to.String(existingProbe.Name), to.String(probe.Name)) && to.Int32(existingProbe.Port) == to.Int32(probe.Port) {
			return true
		}
	}
	return false
}

func findRule(rules []network.LoadBalancingRule, rule network.LoadBalancingRule, wantLB bool) bool {
	for _, existingRule := range rules {
		if strings.EqualFold(to.String(existingRule.Name), to.String(rule.Name)) &&
			equalLoadBalancingRulePropertiesFormat(existingRule.LoadBalancingRulePropertiesFormat, rule.LoadBalancingRulePropertiesFormat, wantLB) {
			return true
		}
	}
	return false
}

// equalLoadBalancingRulePropertiesFormat checks whether the provided LoadBalancingRulePropertiesFormat are equal.
// Note: only fields used in reconcileLoadBalancer are considered.
func equalLoadBalancingRulePropertiesFormat(s *network.LoadBalancingRulePropertiesFormat, t *network.LoadBalancingRulePropertiesFormat, wantLB bool) bool {
	if s == nil || t == nil {
		return false
	}

	properties := reflect.DeepEqual(s.Protocol, t.Protocol) &&
		reflect.DeepEqual(s.FrontendIPConfiguration, t.FrontendIPConfiguration) &&
		reflect.DeepEqual(s.BackendAddressPool, t.BackendAddressPool) &&
		reflect.DeepEqual(s.LoadDistribution, t.LoadDistribution) &&
		reflect.DeepEqual(s.FrontendPort, t.FrontendPort) &&
		reflect.DeepEqual(s.BackendPort, t.BackendPort) &&
		reflect.DeepEqual(s.EnableFloatingIP, t.EnableFloatingIP) &&
		reflect.DeepEqual(to.Bool(s.EnableTCPReset), to.Bool(t.EnableTCPReset)) &&
		reflect.DeepEqual(to.Bool(s.DisableOutboundSnat), to.Bool(t.DisableOutboundSnat))

	if wantLB && s.IdleTimeoutInMinutes != nil && t.IdleTimeoutInMinutes != nil {
		return properties && reflect.DeepEqual(s.IdleTimeoutInMinutes, t.IdleTimeoutInMinutes)
	}
	return properties
}

// This compares rule's Name, Protocol, SourcePortRange, DestinationPortRange, SourceAddressPrefix, Access, and Direction.
// Note that it compares rule's DestinationAddressPrefix only when it's not consolidated rule as such rule does not have DestinationAddressPrefix defined.
// We intentionally do not compare DestinationAddressPrefixes in consolidated case because reconcileSecurityRule has to consider the two rules equal,
// despite different DestinationAddressPrefixes, in order to give it a chance to consolidate the two rules.
func findSecurityRule(rules []network.SecurityRule, rule network.SecurityRule) bool {
	for _, existingRule := range rules {
		if !strings.EqualFold(to.String(existingRule.Name), to.String(rule.Name)) {
			continue
		}
		if !strings.EqualFold(string(existingRule.Protocol), string(rule.Protocol)) {
			continue
		}
		if !strings.EqualFold(to.String(existingRule.SourcePortRange), to.String(rule.SourcePortRange)) {
			continue
		}
		if !strings.EqualFold(to.String(existingRule.DestinationPortRange), to.String(rule.DestinationPortRange)) {
			continue
		}
		if !strings.EqualFold(to.String(existingRule.SourceAddressPrefix), to.String(rule.SourceAddressPrefix)) {
			continue
		}
		if !allowsConsolidation(existingRule) && !allowsConsolidation(rule) {
			if !strings.EqualFold(to.String(existingRule.DestinationAddressPrefix), to.String(rule.DestinationAddressPrefix)) {
				continue
			}
			if !reflect.DeepEqual(to.StringSlice(existingRule.DestinationAddressPrefixes), to.StringSlice(rule.DestinationAddressPrefixes)) {
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

func subnet(service *v1.Service) *string {
	if requiresInternalLoadBalancer(service) {
		if l, found := service.Annotations[consts.ServiceAnnotationLoadBalancerInternalSubnet]; found && strings.TrimSpace(l) != "" {
			return &l
		}
	}

	return nil
}

// getServiceLoadBalancerMode parses the mode value.
// if the value is __auto__ it returns isAuto = TRUE.
// if anything else it returns the unique VM set names after trimming spaces.
func (az *Cloud) getServiceLoadBalancerMode(service *v1.Service) (bool, bool, string) {
	mode, hasMode := service.Annotations[consts.ServiceAnnotationLoadBalancerMode]
	useSingleSLB := az.useStandardLoadBalancer() && !az.EnableMultipleStandardLoadBalancers
	if useSingleSLB && hasMode {
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

func getServiceTags(service *v1.Service) []string {
	if service == nil {
		return nil
	}

	if serviceTags, found := service.Annotations[consts.ServiceAnnotationAllowedServiceTag]; found {
		result := []string{}
		tags := strings.Split(strings.TrimSpace(serviceTags), ",")
		for _, tag := range tags {
			serviceTag := strings.TrimSpace(tag)
			if serviceTag != "" {
				result = append(result, serviceTag)
			}
		}

		return result
	}

	return nil
}

// serviceOwnsPublicIP checks if the service owns the pip and if the pip is user-created.
// The pip is user-created if and only if there is no service tags.
// The service owns the pip if:
// 1. The serviceName is included in the service tags of a system-created pip.
// 2. The service.Spec.LoadBalancerIP matches the IP address of a user-created pip.
func serviceOwnsPublicIP(service *v1.Service, pip *network.PublicIPAddress, clusterName string) (bool, bool) {
	if service == nil || pip == nil {
		klog.Warningf("serviceOwnsPublicIP: nil service or public IP")
		return false, false
	}

	if pip.PublicIPAddressPropertiesFormat == nil || to.String(pip.IPAddress) == "" {
		klog.Warningf("serviceOwnsPublicIP: empty pip.IPAddress")
		return false, false
	}

	serviceName := getServiceName(service)

	if pip.Tags != nil {
		serviceTag := getServiceFromPIPServiceTags(pip.Tags)
		clusterTag := getClusterFromPIPClusterTags(pip.Tags)

		// if there is no service tag on the pip, it is user-created pip
		if serviceTag == "" {
			return strings.EqualFold(to.String(pip.IPAddress), service.Spec.LoadBalancerIP), true
		}

		// if there is service tag on the pip, it is system-created pip
		if isSVCNameInPIPTag(serviceTag, serviceName) {
			// Backward compatible for clusters upgraded from old releases.
			// In such case, only "service" tag is set.
			if clusterTag == "" {
				return true, false
			}

			// If cluster name tag is set, then return true if it matches.
			if clusterTag == clusterName {
				return true, false
			}
		} else {
			// if the service is not included in te tags of the system-created pip, check the ip address
			// this could happen for secondary services
			return strings.EqualFold(to.String(pip.IPAddress), service.Spec.LoadBalancerIP), false
		}
	}

	return false, false
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
		pip.Tags = map[string]*string{consts.ServiceTagKey: to.StringPtr("")}
	}

	serviceTagValue := to.StringPtr(getServiceFromPIPServiceTags(pip.Tags))
	serviceTagValueSet := make(map[string]struct{})
	existingServiceNames := parsePIPServiceTag(serviceTagValue)
	addedNew := false

	// replace is used when unbinding the service from PIP so addedNew remains false all the time
	if replace {
		serviceTagValue = to.StringPtr(strings.Join(incomingServiceNames, ","))
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
			serviceTagValue = to.StringPtr(serviceName)
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

func unbindServiceFromPIP(pip *network.PublicIPAddress, service *v1.Service, serviceName, clusterName string) error {
	if pip == nil || pip.Tags == nil {
		return fmt.Errorf("nil public IP or tags")
	}

	// skip removing tags for user assigned pips
	serviceTagValue := to.StringPtr(getServiceFromPIPServiceTags(pip.Tags))
	existingServiceNames := parsePIPServiceTag(serviceTagValue)
	var found bool
	for i := len(existingServiceNames) - 1; i >= 0; i-- {
		if strings.EqualFold(existingServiceNames[i], serviceName) {
			existingServiceNames = append(existingServiceNames[:i], existingServiceNames[i+1:]...)
			found = true
		}
	}
	if !found {
		klog.Warningf("cannot find the service %s in the corresponding PIP", serviceName)
	}

	_, err := bindServicesToPIP(pip, existingServiceNames, true)
	if err != nil {
		return err
	}

	if existingServiceName := getServiceFromPIPDNSTags(pip.Tags); existingServiceName != "" && strings.EqualFold(existingServiceName, serviceName) {
		pip.Tags[consts.ServiceUsingDNSKey] = to.StringPtr("")
	}

	return nil
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
