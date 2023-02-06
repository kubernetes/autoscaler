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
	"fmt"
	"net"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// reconcilePrivateLinkService() function makes sure a PLS is created or deleted on
// a Load Balancer frontend IP Configuration according to service spec and cluster operation
func (az *Cloud) reconcilePrivateLinkService(
	clusterName string,
	service *v1.Service,
	fipConfig *network.FrontendIPConfiguration,
	wantPLS bool,
) error {
	createPLS := wantPLS && serviceRequiresPLS(service)
	serviceName := getServiceName(service)
	fipConfigID := fipConfig.ID
	klog.V(2).Infof("reconcilePrivateLinkService for service(%s) - LB fipConfigID(%s) - wantPLS(%t) - createPLS(%t)", serviceName, to.String(fipConfig.Name), wantPLS, createPLS)

	if createPLS {
		// Firstly, make sure it's internal service
		if !requiresInternalLoadBalancer(service) && !consts.IsK8sServiceDisableLoadBalancerFloatingIP(service) {
			return fmt.Errorf("reconcilePrivateLinkService for service(%s): service requiring private link service must be internal or disable floating ip", serviceName)
		}

		// Secondly, check if there is a private link service already created
		existingPLS, err := az.getPrivateLinkService(fipConfigID, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("reconcilePrivateLinkService for service(%s): getPrivateLinkService(%s) failed: %v", serviceName, to.String(fipConfigID), err)
			return err
		}

		exists := !strings.EqualFold(to.String(existingPLS.ID), consts.PrivateLinkServiceNotExistID)
		if exists {
			klog.V(4).Infof("reconcilePrivateLinkService for service(%s): found existing private link service attached(%s)", serviceName, to.String(existingPLS.Name))
			if !isManagedPrivateLinkSerivce(&existingPLS, clusterName) {
				return fmt.Errorf(
					"reconcilePrivateLinkService for service(%s) failed: LB frontend(%s) already has unmanaged private link service(%s)",
					serviceName,
					to.String(fipConfigID),
					to.String(existingPLS.ID),
				)
			}
			// If there is an existing private link service, only owner service can update its properties
			ownerService := getPrivateLinkServiceOwner(&existingPLS)
			if !strings.EqualFold(ownerService, serviceName) {
				if serviceHasAdditionalConfigs(service) {
					return fmt.Errorf(
						"reconcilePrivateLinkService for service(%s) failed: LB frontend(%s) already has existing private link service(%s) owned by service(%s)",
						serviceName,
						to.String(fipConfigID),
						to.String(existingPLS.Name),
						ownerService,
					)
				}
				klog.V(2).Infof(
					"reconcilePrivateLinkService for service(%s): automatically share private link service(%s) owned by service(%s)",
					serviceName,
					to.String(existingPLS.Name),
					ownerService,
				)
				return nil
			}
		} else {
			existingPLS.ID = nil
			existingPLS.Location = &az.Location
			existingPLS.PrivateLinkServiceProperties = &network.PrivateLinkServiceProperties{}
			if az.HasExtendedLocation() {
				existingPLS.ExtendedLocation = &network.ExtendedLocation{
					Name: &az.ExtendedLocationName,
					Type: getExtendedLocationTypeFromString(az.ExtendedLocationType),
				}
			}
		}

		plsName, err := az.getPrivateLinkServiceName(&existingPLS, service, fipConfig)
		if err != nil {
			return err
		}

		dirtyPLS, err := az.getExpectedPrivateLinkService(&existingPLS, &plsName, &clusterName, service, fipConfig)
		if err != nil {
			return err
		}

		if dirtyPLS {
			klog.V(2).Infof("reconcilePrivateLinkService for service(%s): pls(%s) - updating", serviceName, plsName)
			err := az.disablePLSNetworkPolicy(service)
			if err != nil {
				klog.Errorf("reconcilePrivateLinkService for service(%s) disable PLS network policy failed for pls(%s): %v", serviceName, plsName, err.Error())
				return err
			}
			existingPLS.Etag = to.StringPtr("")
			err = az.CreateOrUpdatePLS(service, existingPLS)
			if err != nil {
				klog.Errorf("reconcilePrivateLinkService for service(%s) abort backoff: pls(%s) - updating: %s", serviceName, plsName, err.Error())
				return err
			}
		}
	} else if !wantPLS {
		existingPLS, err := az.getPrivateLinkService(fipConfigID, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("reconcilePrivateLinkService for service(%s): getPrivateLinkService(%s) failed: %v", serviceName, to.String(fipConfigID), err)
			return err
		}

		exists := !strings.EqualFold(to.String(existingPLS.ID), consts.PrivateLinkServiceNotExistID)
		if exists {
			deleteErr := az.safeDeletePLS(&existingPLS, service)
			if deleteErr != nil {
				klog.Errorf("reconcilePrivateLinkService for service(%s): deletePLS for frontEnd(%s) failed: %v", serviceName, to.String(fipConfigID), err)
				return deleteErr.Error()
			}
		}
	}

	klog.V(2).Infof("reconcilePrivateLinkService for service(%s) finished", serviceName)
	return nil
}

func (az *Cloud) disablePLSNetworkPolicy(service *v1.Service) error {
	serviceName := getServiceName(service)
	subnetName := getPLSSubnetName(service)
	if subnetName == nil {
		subnetName = &az.SubnetName
	}

	subnet, existsSubnet, err := az.getSubnet(az.VnetName, *subnetName)
	if err != nil {
		return err
	}
	if !existsSubnet {
		return fmt.Errorf("disablePLSNetworkPolicy: failed to get private link service subnet(%s) for service(%s)", *subnetName, serviceName)
	}

	// Policy already disabled
	if subnet.PrivateLinkServiceNetworkPolicies == network.VirtualNetworkPrivateLinkServiceNetworkPoliciesDisabled {
		return nil
	}

	subnet.PrivateLinkServiceNetworkPolicies = network.VirtualNetworkPrivateLinkServiceNetworkPoliciesDisabled
	err = az.CreateOrUpdateSubnet(service, subnet)
	if err != nil {
		return err
	}
	return nil
}

func (az *Cloud) safeDeletePLS(pls *network.PrivateLinkService, service *v1.Service) *retry.Error {
	if pls == nil {
		return nil
	}

	peConns := pls.PrivateEndpointConnections
	if peConns != nil {
		for _, peConn := range *peConns {
			klog.V(2).Infof("deletePLS: deleting PEConnection %s", to.String(peConn.Name))
			rerr := az.DeletePEConn(service, to.String(pls.Name), to.String(peConn.Name))
			if rerr != nil {
				return rerr
			}
		}
	}

	rerr := az.DeletePLS(service, to.String(pls.Name), to.String((*pls.LoadBalancerFrontendIPConfigurations)[0].ID))
	if rerr != nil {
		return rerr
	}
	klog.V(2).Infof("safeDeletePLS(%s) finished", to.String(pls.Name))
	return nil
}

// getPrivateLinkServiceName() returns the name of private link service, or any error
func (az *Cloud) getPrivateLinkServiceName(
	existingPLS *network.PrivateLinkService,
	service *v1.Service,
	fipConfig *network.FrontendIPConfiguration,
) (string, error) {
	existingName := existingPLS.Name
	serviceName := getServiceName(service)

	if nameFromService, found := service.Annotations[consts.ServiceAnnotationPLSName]; found {
		nameFromService = strings.TrimSpace(nameFromService)
		if existingName != nil && !strings.EqualFold(to.String(existingName), nameFromService) {
			return "", fmt.Errorf(
				"getPrivateLinkServiceName(%s) failed: cannot change existing private link service name (%s) to (%s)",
				serviceName,
				to.String(existingName),
				nameFromService,
			)
		}
		return nameFromService, nil
	}

	if existingName != nil {
		return to.String(existingName), nil
	}

	// default PLS name: pls-<frontendIPConfigName>
	return fmt.Sprintf("%s-%s", "pls", *fipConfig.Name), nil
}

// getExpectedPrivateLinkService builds expected PLS object from service spec
func (az *Cloud) getExpectedPrivateLinkService(
	existingPLS *network.PrivateLinkService,
	plsName *string,
	clusterName *string,
	service *v1.Service,
	fipConfig *network.FrontendIPConfiguration,
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
	if existingPLS.ProvisioningState == network.ProvisioningStateFailed {
		dirtyPLS = true
	}

	// Shadow copy properties to avoid changing pls cache
	plsProperties := *existingPLS.PrivateLinkServiceProperties
	existingPLS.PrivateLinkServiceProperties = &plsProperties

	// Set LBFrontendIpConfiguration
	if existingPLS.LoadBalancerFrontendIPConfigurations == nil {
		existingPLS.LoadBalancerFrontendIPConfigurations = &[]network.FrontendIPConfiguration{{ID: fipConfig.ID}}
		dirtyPLS = true
	}

	changed, err := az.reconcilePLSIpConfigs(existingPLS, service)
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
	existingPLS *network.PrivateLinkService,
	service *v1.Service,
) (bool, error) {
	changed := false
	serviceName := getServiceName(service)

	subnetName := getPLSSubnetName(service)
	if subnetName == nil {
		subnetName = &az.SubnetName
	}
	subnet, existsSubnet, err := az.getSubnet(az.VnetName, *subnetName)
	if err != nil {
		return false, err
	}
	if !existsSubnet {
		return false, fmt.Errorf("checkAndUpdatePLSIPConfigs: failed to get private link service subnet(%s) for service(%s)", *subnetName, serviceName)
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

	if existingPLS.IPConfigurations == nil {
		existingPLS.IPConfigurations = &[]network.PrivateLinkServiceIPConfiguration{}
		changed = true
	}

	if int32(len(*existingPLS.IPConfigurations)) != ipConfigCount {
		changed = true
	}

	existingStaticIps := make([]string, 0)
	for _, ipConfig := range *existingPLS.IPConfigurations {
		if !strings.EqualFold(to.String(subnet.ID), to.String(ipConfig.Subnet.ID)) {
			changed = true
		}
		if strings.EqualFold(string(ipConfig.PrivateIPAllocationMethod), string(network.IPAllocationMethodStatic)) {
			klog.V(10).Infof("Found static IP: %s", to.String(ipConfig.PrivateIPAddress))
			if _, found := staticIps[to.String(ipConfig.PrivateIPAddress)]; !found {
				changed = true
			}
			existingStaticIps = append(existingStaticIps, to.String(ipConfig.PrivateIPAddress))
		}
		if *ipConfig.Primary {
			if strings.EqualFold(string(ipConfig.PrivateIPAllocationMethod), string(network.IPAllocationMethodStatic)) {
				if !strings.EqualFold(primaryIP, to.String(ipConfig.PrivateIPAddress)) {
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
		ipConfigs := []network.PrivateLinkServiceIPConfiguration{}
		for k := range staticIps {
			ip := k
			isPrimary := strings.EqualFold(ip, primaryIP)
			configName := fmt.Sprintf("%s-%s-static-%s", to.String(subnet.Name), to.String(existingPLS.Name), ip)
			ipConfigs = append(ipConfigs, network.PrivateLinkServiceIPConfiguration{
				Name: &configName,
				PrivateLinkServiceIPConfigurationProperties: &network.PrivateLinkServiceIPConfigurationProperties{
					PrivateIPAddress:          &ip,
					PrivateIPAllocationMethod: network.IPAllocationMethodStatic,
					Subnet: &network.Subnet{
						ID: subnet.ID,
					},
					Primary:                 &isPrimary,
					PrivateIPAddressVersion: network.IPVersionIPv4,
				},
			})
		}
		for i := 0; i < int(ipConfigCount)-len(staticIps); i++ {
			isPrimary := primaryIP == "" && i == 0
			configName := fmt.Sprintf("%s-%s-dynamic-%d", to.String(subnet.Name), to.String(existingPLS.Name), i)
			ipConfigs = append(ipConfigs, network.PrivateLinkServiceIPConfiguration{
				Name: &configName,
				PrivateLinkServiceIPConfigurationProperties: &network.PrivateLinkServiceIPConfigurationProperties{
					PrivateIPAllocationMethod: network.IPAllocationMethodDynamic,
					Subnet: &network.Subnet{
						ID: subnet.ID,
					},
					Primary:                 &isPrimary,
					PrivateIPAddressVersion: network.IPVersionIPv4,
				},
			})
		}
		existingPLS.IPConfigurations = &ipConfigs
	}
	return changed, nil
}

func serviceRequiresPLS(service *v1.Service) bool {
	return getBoolValueFromServiceAnnotations(service, consts.ServiceAnnotationPLSCreation)
}

func reconcilePLSEnableProxyProtocol(
	existingPLS *network.PrivateLinkService,
	service *v1.Service,
) bool {
	changed := false
	enableProxyProtocol := getBoolValueFromServiceAnnotations(service, consts.ServiceAnnotationPLSProxyProtocol)
	if enableProxyProtocol && (existingPLS.EnableProxyProtocol == nil || !*existingPLS.EnableProxyProtocol) {
		changed = true
	} else if !enableProxyProtocol && (existingPLS.EnableProxyProtocol != nil && *existingPLS.EnableProxyProtocol) {
		changed = true
	}
	if changed {
		existingPLS.EnableProxyProtocol = &enableProxyProtocol
	}
	return changed
}

func reconcilePLSFqdn(
	existingPLS *network.PrivateLinkService,
	service *v1.Service,
) bool {
	changed := false
	fqdns := getPLSFqdns(service)
	if existingPLS.Fqdns == nil {
		if len(fqdns) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(fqdns, *existingPLS.Fqdns) {
		changed = true
	}

	if changed {
		existingPLS.Fqdns = &fqdns
	}
	return changed
}

func reconcilePLSVisibility(
	existingPLS *network.PrivateLinkService,
	service *v1.Service,
) (bool, error) {
	changed := false
	visibilitySubs, _ := getPLSVisibility(service)
	autoApprovalSubs := getPLSAutoApproval(service)

	if existingPLS.Visibility == nil || existingPLS.Visibility.Subscriptions == nil {
		if len(visibilitySubs) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(visibilitySubs, *existingPLS.Visibility.Subscriptions) {
		changed = true
	}

	if existingPLS.AutoApproval == nil || existingPLS.AutoApproval.Subscriptions == nil {
		if len(autoApprovalSubs) != 0 {
			changed = true
		}
	} else if !sameContentInSlices(autoApprovalSubs, *existingPLS.AutoApproval.Subscriptions) {
		changed = true
	}

	if changed {
		existingPLS.Visibility = &network.PrivateLinkServicePropertiesVisibility{
			Subscriptions: &visibilitySubs,
		}
		existingPLS.AutoApproval = &network.PrivateLinkServicePropertiesAutoApproval{
			Subscriptions: &autoApprovalSubs,
		}
	}
	return changed, nil
}

func (az *Cloud) reconcilePLSTags(
	existingPLS *network.PrivateLinkService,
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

func isManagedPrivateLinkSerivce(existingPLS *network.PrivateLinkService, clusterName string) bool {
	tags := existingPLS.Tags
	v, ok := tags[consts.ClusterNameTagKey]
	return ok && v != nil && strings.EqualFold(strings.TrimSpace(*v), clusterName)
}

// find owner service for an existing private link service from its tags
func getPrivateLinkServiceOwner(existingPLS *network.PrivateLinkService) string {
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
