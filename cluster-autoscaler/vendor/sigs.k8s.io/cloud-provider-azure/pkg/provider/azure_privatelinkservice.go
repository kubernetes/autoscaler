/*
Copyright 2022 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	fnutil "sigs.k8s.io/cloud-provider-azure/pkg/util/collectionutil"
)

// reconcilePrivateLinkService() function makes sure a PLS is created or deleted on
// a Load Balancer frontend IP Configuration according to service spec and cluster operation
func (az *Cloud) reconcilePrivateLinkService(
	ctx context.Context,
	clusterName string,
	service *v1.Service,
	fipConfig *armnetwork.FrontendIPConfiguration,
	wantPLS bool,
) (bool /*deleted PLS*/, error) {
	logger := log.FromContextOrBackground(ctx).WithName("reconcilePrivateLinkService")
	isinternal := requiresInternalLoadBalancer(service)
	_, _, fipIPVersion := az.serviceOwnsFrontendIP(ctx, fipConfig, service)
	serviceName := getServiceName(service)
	var isIPv6 bool
	var err error
	if fipIPVersion != nil {
		isIPv6 = *fipIPVersion == armnetwork.IPVersionIPv6
	} else {
		if isIPv6, err = az.isFIPIPv6(service, fipConfig); err != nil {
			logger.Error(err, "failed to get FIP IP family", "service", serviceName)
			return false, err
		}
	}
	createPLS := wantPLS && serviceRequiresPLS(service)
	isDualStack := isServiceDualStack(service)
	if isIPv6 {
		if isDualStack || !createPLS {
			logger.V(2).Info("IPv6 is not supported for private link service, skip reconcilePrivateLinkService for service", "service", serviceName)
			return false, nil
		}
		return false, fmt.Errorf("IPv6 is not supported for private link service")
	}

	fipConfigID := fipConfig.ID
	logger.V(2).Info("Reconcile private link service for service",
		"service", serviceName,
		"LBFipConfigID", ptr.Deref(fipConfig.Name, ""),
		"wantPLS", wantPLS,
		"createPLS", createPLS)

	request := "ensure_privatelinkservice"
	if !wantPLS {
		request = "ensure_privatelinkservice_deleted"
	}
	mc := metrics.NewMetricContext("services", request, az.ResourceGroup, az.getNetworkResourceSubscriptionID(), serviceName)

	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	if createPLS {
		// Firstly, make sure it's internal service
		if !isinternal && !consts.IsK8sServiceDisableLoadBalancerFloatingIP(service) {
			return false, fmt.Errorf("reconcilePrivateLinkService for service(%s): service requiring private link service must be internal or disable floating ip", serviceName)
		}

		// Secondly, check if there is a private link service already created
		existingPLS, err := az.plsRepo.Get(ctx, az.getPLSResourceGroup(service), *fipConfigID, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "getPrivateLinkService failed", "service", serviceName, "fipConfigID", ptr.Deref(fipConfigID, ""))
			return false, err
		}

		exists := !strings.EqualFold(ptr.Deref(existingPLS.ID, ""), consts.PrivateLinkServiceNotExistID)
		if exists {
			logger.V(4).Info("found existing private link service attached", "service", serviceName, "privateLinkService", ptr.Deref(existingPLS.Name, ""))
			if !isManagedPrivateLinkSerivce(existingPLS, clusterName) {
				return false, fmt.Errorf(
					"reconcilePrivateLinkService for service(%s) failed: LB frontend(%s) already has unmanaged private link service(%s)",
					serviceName,
					ptr.Deref(fipConfigID, ""),
					ptr.Deref(existingPLS.ID, ""),
				)
			}
			// If there is an existing private link service, only owner service can update its properties
			ownerService := getPrivateLinkServiceOwner(existingPLS)
			if !strings.EqualFold(ownerService, serviceName) {
				if serviceHasAdditionalConfigs(service) {
					return false, fmt.Errorf(
						"reconcilePrivateLinkService for service(%s) failed: LB frontend(%s) already has existing private link service(%s) owned by service(%s)",
						serviceName,
						ptr.Deref(fipConfigID, ""),
						ptr.Deref(existingPLS.Name, ""),
						ownerService,
					)
				}
				logger.V(2).Info("automatically share private link service owned by another service",
					"service", serviceName,
					"privateLinkService", ptr.Deref(existingPLS.Name, ""),
					"ownerService", ownerService,
				)
				return false, nil
			}
		} else {
			existingPLS.ID = nil
			existingPLS.Location = &az.Location
			existingPLS.Properties = &armnetwork.PrivateLinkServiceProperties{}
			if az.HasExtendedLocation() {
				existingPLS.ExtendedLocation = &armnetwork.ExtendedLocation{
					Name: &az.ExtendedLocationName,
					Type: to.Ptr(getExtendedLocationTypeFromString(az.ExtendedLocationType)),
				}
			}
		}

		plsName, err := az.getPrivateLinkServiceName(existingPLS, service, fipConfig)
		if err != nil {
			return false, err
		}

		dirtyPLS, err := az.getExpectedPrivateLinkService(ctx, existingPLS, &plsName, &clusterName, service, fipConfig)
		if err != nil {
			return false, err
		}

		if dirtyPLS {
			logger.V(2).Info("updating", "service", serviceName, "pls", plsName)
			err := az.disablePLSNetworkPolicy(ctx, service)
			if err != nil {
				logger.Error(err, "Failed to disable PLS network policy", "service", serviceName, "pls", plsName)
				return false, err
			}
			existingPLS.Etag = ptr.To("")
			_, err = az.plsRepo.CreateOrUpdate(ctx, az.getPLSResourceGroup(service), *existingPLS)
			if err != nil {
				logger.Error(err, "abort backoff: pls updating", "service", serviceName, "pls", plsName)
				return false, err
			}
		}
	} else if !wantPLS {
		existingPLS, err := az.plsRepo.Get(ctx, az.getPLSResourceGroup(service), *fipConfigID, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "getPrivateLinkService failed", "service", serviceName, "LBFipConfigID", ptr.Deref(fipConfigID, ""))
			return false, err
		}

		exists := !strings.EqualFold(ptr.Deref(existingPLS.ID, ""), consts.PrivateLinkServiceNotExistID)
		if exists {
			deleteErr := az.safeDeletePLS(ctx, existingPLS, service)
			if deleteErr != nil {
				logger.Error(deleteErr, "deletePLS for frontEnd failed", "service", serviceName, "LBFipConfigID", ptr.Deref(fipConfigID, ""))
				return false, deleteErr
			}
			isOperationSucceeded = true
			logger.V(2).Info("finished", "service", serviceName)
			return true, nil // return true for successfully deleted PLS
		}
	}

	isOperationSucceeded = true
	logger.V(2).Info("finished", "service", serviceName)
	return false, nil
}

func (az *Cloud) getPLSResourceGroup(service *v1.Service) string {
	if resourceGroup, found := service.Annotations[consts.ServiceAnnotationPLSResourceGroup]; found {
		resourceGroupName := strings.TrimSpace(resourceGroup)
		if len(resourceGroupName) > 0 {
			return resourceGroupName
		}
	}

	return az.PrivateLinkServiceResourceGroup
}

func (az *Cloud) disablePLSNetworkPolicy(ctx context.Context, service *v1.Service) error {
	serviceName := getServiceName(service)
	subnetName := getPLSSubnetName(service)
	if subnetName == nil {
		subnetName = &az.SubnetName
	}

	rg := az.VnetResourceGroup
	if rg == "" {
		rg = az.ResourceGroup
	}

	subnet, err := az.subnetRepo.Get(ctx, rg, az.VnetName, *subnetName)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr != nil && respErr.StatusCode == http.StatusNotFound {
				return fmt.Errorf("disablePLSNetworkPolicy: failed to get private link service subnet(%s) for service(%s)", *subnetName, serviceName)
			}
		}
		return err
	}
	if subnet.Properties == nil {
		subnet.Properties = &armnetwork.SubnetPropertiesFormat{}
	}

	// Policy already disabled
	if subnet.Properties.PrivateLinkServiceNetworkPolicies != nil && *subnet.Properties.PrivateLinkServiceNetworkPolicies == armnetwork.VirtualNetworkPrivateLinkServiceNetworkPoliciesDisabled {
		return nil
	}

	subnet.Properties.PrivateLinkServiceNetworkPolicies = to.Ptr(armnetwork.VirtualNetworkPrivateLinkServiceNetworkPoliciesDisabled)
	err = az.subnetRepo.CreateOrUpdate(ctx, rg, az.VnetName, *subnetName, *subnet)
	if err != nil {
		return err
	}
	return nil
}

func (az *Cloud) safeDeletePLS(ctx context.Context, pls *armnetwork.PrivateLinkService, service *v1.Service) error {
	logger := log.FromContextOrBackground(ctx).WithName("safeDeletePLS")
	if pls == nil {
		return nil
	}

	peConns := pls.Properties.PrivateEndpointConnections
	for _, peConn := range peConns {
		logger.V(2).Info("deleting PEConnection", "PEConnection", ptr.Deref(peConn.Name, ""))
		err := az.plsRepo.DeletePEConnection(ctx, az.getPLSResourceGroup(service), ptr.Deref(pls.Name, ""), ptr.Deref(peConn.Name, ""))
		if err != nil {
			return err
		}
	}

	resourceGroup := az.getPLSResourceGroup(service)
	plsName := ptr.Deref(pls.Name, "")
	lbFrontendID := ptr.Deref((pls.Properties.LoadBalancerFrontendIPConfigurations)[0].ID, "")
	rerr := az.plsRepo.Delete(ctx, resourceGroup, plsName, lbFrontendID)
	if rerr != nil {
		return rerr
	}
	logger.V(2).Info("finished", "privateLinkService", ptr.Deref(pls.Name, ""))
	return nil
}

// getPrivateLinkServiceName() returns the name of private link service, or any error
func (az *Cloud) getPrivateLinkServiceName(
	existingPLS *armnetwork.PrivateLinkService,
	service *v1.Service,
	fipConfig *armnetwork.FrontendIPConfiguration,
) (string, error) {
	existingName := existingPLS.Name
	serviceName := getServiceName(service)

	if nameFromService, found := service.Annotations[consts.ServiceAnnotationPLSName]; found {
		nameFromService = strings.TrimSpace(nameFromService)
		if existingName != nil && !strings.EqualFold(ptr.Deref(existingName, ""), nameFromService) {
			return "", fmt.Errorf(
				"getPrivateLinkServiceName(%s) failed: cannot change existing private link service name (%s) to (%s)",
				serviceName,
				ptr.Deref(existingName, ""),
				nameFromService,
			)
		}
		return nameFromService, nil
	}

	if existingName != nil {
		return ptr.Deref(existingName, ""), nil
	}

	// default PLS name: pls-<frontendIPConfigName>
	return fmt.Sprintf("%s-%s", "pls", *fipConfig.Name), nil
}

// getExpectedPrivateLinkService builds expected PLS object from service spec
func (az *Cloud) getExpectedPrivateLinkService(
	ctx context.Context,
	existingPLS *armnetwork.PrivateLinkService,
	plsName *string,
	clusterName *string,
	service *v1.Service,
	fipConfig *armnetwork.FrontendIPConfiguration,
) (dirtyPLS bool, err error) {
	dirtyPLS = false

	if existingPLS == nil {
		return false, fmt.Errorf("getExpectedPrivateLinkService: existingPLS is nil (unexpected)")
	}

	// This only happens for an empty
	if existingPLS.Name == nil || !strings.EqualFold(*existingPLS.Name, *plsName) {
		existingPLS.Name = plsName
		dirtyPLS = true
	}

	// Set failed PLS as dirty so that provision can be retried
	if existingPLS.Properties.ProvisioningState != nil && *existingPLS.Properties.ProvisioningState == armnetwork.ProvisioningStateFailed {
		dirtyPLS = true
	}

	// Shadow copy properties to avoid changing pls cache
	plsProperties := *existingPLS.Properties
	existingPLS.Properties = &plsProperties

	// Set LBFrontendIpConfiguration
	if existingPLS.Properties.LoadBalancerFrontendIPConfigurations == nil {
		existingPLS.Properties.LoadBalancerFrontendIPConfigurations = []*armnetwork.FrontendIPConfiguration{{ID: fipConfig.ID}}
		dirtyPLS = true
	}

	changed, err := az.reconcilePLSIpConfigs(ctx, existingPLS, service)
	if err != nil {
		return false, err
	}
	if changed {
		dirtyPLS = true
	}

	if reconcilePLSEnableProxyProtocol(existingPLS, service) {
		dirtyPLS = true
	}

	if reconcilePLSFqdn(existingPLS, service) {
		dirtyPLS = true
	}

	changed, err = reconcilePLSVisibility(existingPLS, service)
	if err != nil {
		return false, err
	}
	if changed {
		dirtyPLS = true
	}

	if az.reconcilePLSTags(existingPLS, clusterName, service) {
		dirtyPLS = true
	}

	return dirtyPLS, nil
}

// reconcile Private link service's IP configurations
func (az *Cloud) reconcilePLSIpConfigs(
	ctx context.Context,
	existingPLS *armnetwork.PrivateLinkService,
	service *v1.Service,
) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("reconcilePLSIpConfigs")
	changed := false
	serviceName := getServiceName(service)

	subnetName := getPLSSubnetName(service)
	if subnetName == nil {
		subnetName = &az.SubnetName
	}
	rg := az.VnetResourceGroup
	if rg == "" {
		rg = az.ResourceGroup
	}
	subnet, err := az.subnetRepo.Get(ctx, rg, az.VnetName, *subnetName)
	if err != nil {
		var runtimError *azcore.ResponseError
		if errors.As(err, &runtimError) {
			if runtimError != nil && runtimError.StatusCode == http.StatusNotFound {
				return false, fmt.Errorf("checkAndUpdatePLSIPConfigs: failed to get private link service subnet(%s) for service(%s)", *subnetName, serviceName)
			}
		}
		return false, err
	}

	ipConfigCount, err := getPLSIPConfigCount(service)
	if err != nil {
		return false, err
	}

	staticIps, primaryIP, err := getPLSStaticIPs(service)
	if err != nil {
		return false, err
	}

	if int(ipConfigCount) < len(staticIps) {
		return false, fmt.Errorf("checkAndUpdatePLSIPConfigs: ipConfigCount(%d) must be no smaller than number of static IPs specified(%d)", ipConfigCount, len(staticIps))
	}

	if existingPLS.Properties.IPConfigurations == nil {
		existingPLS.Properties.IPConfigurations = []*armnetwork.PrivateLinkServiceIPConfiguration{}
		changed = true
	}

	if int32(len(existingPLS.Properties.IPConfigurations)) != ipConfigCount {
		changed = true
	}

	existingStaticIps := make([]string, 0)
	for _, ipConfig := range existingPLS.Properties.IPConfigurations {
		if !strings.EqualFold(ptr.Deref(subnet.ID, ""), ptr.Deref(ipConfig.Properties.Subnet.ID, "")) {
			changed = true
		}
		if *ipConfig.Properties.PrivateIPAllocationMethod == armnetwork.IPAllocationMethodStatic {
			logger.V(10).Info("Found static IP", "ip", ptr.Deref(ipConfig.Properties.PrivateIPAddress, ""))
			if _, found := staticIps[ptr.Deref(ipConfig.Properties.PrivateIPAddress, "")]; !found {
				changed = true
			}
			existingStaticIps = append(existingStaticIps, ptr.Deref(ipConfig.Properties.PrivateIPAddress, ""))
		}
		if *ipConfig.Properties.Primary {
			if *ipConfig.Properties.PrivateIPAllocationMethod == armnetwork.IPAllocationMethodStatic {
				if !strings.EqualFold(primaryIP, ptr.Deref(ipConfig.Properties.PrivateIPAddress, "")) {
					changed = true
				}
			} else {
				// Dynamic
				if primaryIP != "" {
					changed = true
				}
			}
		}
	}
	if len(existingStaticIps) != len(staticIps) {
		changed = true
	}

	if changed {
		getFrontendIPConfigName := func(suffix string) (string, error) {
			// frontend ipConfig name length cannot exceed 80
			maxPrefixLen := consts.FrontendIPConfigNameMaxLength - len(suffix)
			if maxPrefixLen <= 0 {
				return "", fmt.Errorf("reconcilePLSIpConfigs: frontend ipConfig suffix %s is too long (not likely to happen)", suffix)
			}
			prefix := fmt.Sprintf("%s-%s", ptr.Deref(subnet.Name, ""), ptr.Deref(existingPLS.Name, ""))
			if len(prefix) > maxPrefixLen {
				prefix = prefix[:maxPrefixLen]
			}
			return prefix + suffix, nil
		}

		var ipConfigs []*armnetwork.PrivateLinkServiceIPConfiguration
		for k := range staticIps {
			ip := k
			isPrimary := strings.EqualFold(ip, primaryIP)
			suffix := fmt.Sprintf("-static-%s", ip)
			configName, err := getFrontendIPConfigName(suffix)
			if err != nil {
				return false, err
			}
			ipConfigs = append(ipConfigs, &armnetwork.PrivateLinkServiceIPConfiguration{
				Name: &configName,
				Properties: &armnetwork.PrivateLinkServiceIPConfigurationProperties{
					PrivateIPAddress:          &ip,
					PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
					Subnet: &armnetwork.Subnet{
						ID: subnet.ID,
					},
					Primary:                 &isPrimary,
					PrivateIPAddressVersion: to.Ptr(armnetwork.IPVersionIPv4),
				},
			})
		}
		for i := 0; i < int(ipConfigCount)-len(staticIps); i++ {
			isPrimary := primaryIP == "" && i == 0
			suffix := fmt.Sprintf("-dynamic-%d", i)
			configName, err := getFrontendIPConfigName(suffix)
			if err != nil {
				return false, err
			}
			ipConfigs = append(ipConfigs, &armnetwork.PrivateLinkServiceIPConfiguration{
				Name: &configName,
				Properties: &armnetwork.PrivateLinkServiceIPConfigurationProperties{
					PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
					Subnet: &armnetwork.Subnet{
						ID: subnet.ID,
					},
					Primary:                 &isPrimary,
					PrivateIPAddressVersion: to.Ptr(armnetwork.IPVersionIPv4),
				},
			})
		}
		existingPLS.Properties.IPConfigurations = ipConfigs
	}
	return changed, nil
}

func serviceRequiresPLS(service *v1.Service) bool {
	return getBoolValueFromServiceAnnotations(service, consts.ServiceAnnotationPLSCreation)
}

func reconcilePLSEnableProxyProtocol(
	existingPLS *armnetwork.PrivateLinkService,
	service *v1.Service,
) bool {
	changed := false
	enableProxyProtocol := getBoolValueFromServiceAnnotations(service, consts.ServiceAnnotationPLSProxyProtocol)
	if enableProxyProtocol && (existingPLS.Properties.EnableProxyProtocol == nil || !*existingPLS.Properties.EnableProxyProtocol) {
		changed = true
	} else if !enableProxyProtocol && (existingPLS.Properties.EnableProxyProtocol != nil && *existingPLS.Properties.EnableProxyProtocol) {
		changed = true
	}
	if changed {
		existingPLS.Properties.EnableProxyProtocol = &enableProxyProtocol
	}
	return changed
}

func reconcilePLSFqdn(
	existingPLS *armnetwork.PrivateLinkService,
	service *v1.Service,
) bool {
	changed := false
	fqdns := getPLSFqdns(service)
	if existingPLS.Properties.Fqdns == nil {
		if len(fqdns) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(fqdns, fnutil.Map(func(s *string) string {
		return *s
	}, existingPLS.Properties.Fqdns)) {
		changed = true
	}

	if changed {
		existingPLS.Properties.Fqdns = fnutil.Map(func(s string) *string { return &s }, fqdns)
	}
	return changed
}

func reconcilePLSVisibility(
	existingPLS *armnetwork.PrivateLinkService,
	service *v1.Service,
) (bool, error) {
	changed := false
	visibilitySubs, _ := getPLSVisibility(service)
	autoApprovalSubs := getPLSAutoApproval(service)

	if existingPLS.Properties.Visibility == nil || existingPLS.Properties.Visibility.Subscriptions == nil {
		if len(visibilitySubs) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(visibilitySubs, fnutil.Map(func(s *string) string {
		return *s
	}, existingPLS.Properties.Visibility.Subscriptions)) {
		changed = true
	}

	if existingPLS.Properties.AutoApproval == nil || existingPLS.Properties.AutoApproval.Subscriptions == nil {
		if len(autoApprovalSubs) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(autoApprovalSubs, fnutil.Map(func(s *string) string {
		return *s
	}, existingPLS.Properties.AutoApproval.Subscriptions)) {
		changed = true
	}

	if changed {
		existingPLS.Properties.Visibility = &armnetwork.PrivateLinkServicePropertiesVisibility{
			Subscriptions: to.SliceOfPtrs(visibilitySubs...),
		}
		existingPLS.Properties.AutoApproval = &armnetwork.PrivateLinkServicePropertiesAutoApproval{
			Subscriptions: to.SliceOfPtrs(autoApprovalSubs...),
		}
	}
	return changed, nil
}

func (az *Cloud) reconcilePLSTags(
	existingPLS *armnetwork.PrivateLinkService,
	clusterName *string,
	service *v1.Service,
) bool {
	configTags := parseTags(az.Tags, az.TagsMap)
	serviceName := getServiceName(service)

	if existingPLS.Tags == nil {
		existingPLS.Tags = make(map[string]*string)
	}

	tags := existingPLS.Tags

	// include the cluster name and service name tags when comparing
	if v, ok := tags[consts.ClusterNameTagKey]; ok && v != nil {
		configTags[consts.ClusterNameTagKey] = v
	} else {
		configTags[consts.ClusterNameTagKey] = clusterName
	}
	if v, ok := tags[consts.OwnerServiceTagKey]; ok && v != nil {
		configTags[consts.OwnerServiceTagKey] = v
	} else {
		configTags[consts.OwnerServiceTagKey] = &serviceName
	}

	tags, changed := az.reconcileTags(existingPLS.Tags, configTags)
	existingPLS.Tags = tags

	return changed
}

func getPLSSubnetName(service *v1.Service) *string {
	if l, found := service.Annotations[consts.ServiceAnnotationPLSIpConfigurationSubnet]; found && strings.TrimSpace(l) != "" {
		return &l
	}

	if requiresInternalLoadBalancer(service) {
		if l, found := service.Annotations[consts.ServiceAnnotationLoadBalancerInternalSubnet]; found && strings.TrimSpace(l) != "" {
			return &l
		}
	}

	return nil
}

func getPLSIPConfigCount(service *v1.Service) (int32, error) {
	ipConfigCnt, err := consts.Getint32ValueFromK8sSvcAnnotation(
		service.Annotations,
		consts.ServiceAnnotationPLSIpConfigurationIPAddressCount,
		func(val *int32) error {
			const (
				MinimumNumOfIPConfig = 1
				MaximumNumOfIPConfig = 8
			)
			if *val < MinimumNumOfIPConfig {
				return fmt.Errorf("minimum number of private link service ipConfig is %d, %d provided", MinimumNumOfIPConfig, *val)
			}
			if *val > MaximumNumOfIPConfig {
				return fmt.Errorf("maximum number of private link service ipConfig is %d, %d provided", MaximumNumOfIPConfig, *val)
			}
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	if ipConfigCnt != nil {
		return *ipConfigCnt, nil
	}
	return consts.PLSDefaultNumOfIPConfig, nil
}

func getPLSFqdns(service *v1.Service) []string {
	fqdns := make([]string, 0)
	if v, ok := service.Annotations[consts.ServiceAnnotationPLSFqdns]; ok {
		fqdnList := strings.Split(strings.TrimSpace(v), " ")
		for _, fqdn := range fqdnList {
			fqdn = strings.TrimSpace(fqdn)
			if fqdn == "" {
				continue
			}
			fqdns = append(fqdns, fqdn)
		}
	}
	return fqdns
}

func getPLSVisibility(service *v1.Service) ([]string, bool) {
	visibilityList := make([]string, 0)
	if val, ok := service.Annotations[consts.ServiceAnnotationPLSVisibility]; ok {
		visibilities := strings.Split(strings.TrimSpace(val), " ")
		for _, vis := range visibilities {
			vis = strings.TrimSpace(vis)
			if vis == "" {
				continue
			}
			if vis == "*" {
				visibilityList = []string{"*"}
				return visibilityList, true
			}
			visibilityList = append(visibilityList, vis)
		}
	}
	return visibilityList, false
}

func getPLSAutoApproval(service *v1.Service) []string {
	autoApprovalList := make([]string, 0)
	if val, ok := service.Annotations[consts.ServiceAnnotationPLSAutoApproval]; ok {
		autoApprovals := strings.Split(strings.TrimSpace(val), " ")
		for _, autoApp := range autoApprovals {
			autoApp = strings.TrimSpace(autoApp)
			if autoApp == "" {
				continue
			}
			autoApprovalList = append(autoApprovalList, autoApp)
		}
	}
	return autoApprovalList
}

func getPLSStaticIPs(service *v1.Service) (map[string]bool, string, error) {
	result := make(map[string]bool)
	primaryIP := ""
	if val, ok := service.Annotations[consts.ServiceAnnotationPLSIpConfigurationIPAddress]; ok {
		ips := strings.Split(strings.TrimSpace(val), " ")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue // skip empty string
			}

			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				return nil, "", fmt.Errorf("getPLSStaticIPs: %s is not a valid IP address", ip)
			}

			if parsedIP.To4() == nil {
				return nil, "", fmt.Errorf("getPLSStaticIPs: private link service ip config only supports IPv4, %s provided", ip)
			}

			result[ip] = true
			if primaryIP == "" {
				primaryIP = ip
			}
		}
	}

	return result, primaryIP, nil
}

func isManagedPrivateLinkSerivce(existingPLS *armnetwork.PrivateLinkService, clusterName string) bool {
	tags := existingPLS.Tags
	v, ok := tags[consts.ClusterNameTagKey]
	return ok && v != nil && strings.EqualFold(strings.TrimSpace(*v), clusterName)
}

// find owner service for an existing private link service from its tags
func getPrivateLinkServiceOwner(existingPLS *armnetwork.PrivateLinkService) string {
	tags := existingPLS.Tags
	v, ok := tags[consts.OwnerServiceTagKey]
	if ok && v != nil {
		return *v
	}
	return ""
}

// Return true if service has private link service config annotations
func serviceHasAdditionalConfigs(service *v1.Service) bool {
	tagKeyList := []string{
		consts.ServiceAnnotationPLSName,
		consts.ServiceAnnotationPLSIpConfigurationSubnet,
		consts.ServiceAnnotationPLSIpConfigurationIPAddressCount,
		consts.ServiceAnnotationPLSIpConfigurationIPAddress,
		consts.ServiceAnnotationPLSFqdns,
		consts.ServiceAnnotationPLSProxyProtocol,
		consts.ServiceAnnotationPLSVisibility,
		consts.ServiceAnnotationPLSAutoApproval}
	for _, k := range tagKeyList {
		if _, found := service.Annotations[k]; found {
			return true
		}
	}
	return false
}
