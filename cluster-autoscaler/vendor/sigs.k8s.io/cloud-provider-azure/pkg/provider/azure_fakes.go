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
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/availabilitysetclient/mock_availabilitysetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/backendaddresspoolclient/mock_backendaddresspoolclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/diskclient/mock_diskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/interfaceclient/mock_interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/loadbalancerclient/mock_loadbalancerclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/mock_azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privateendpointclient/mock_privateendpointclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privatelinkserviceclient/mock_privatelinkserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privatezoneclient/mock_privatezoneclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/publicipaddressclient/mock_publicipaddressclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/routetableclient/mock_routetableclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/securitygroupclient/mock_securitygroupclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/subnetclient/mock_subnetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient/mock_virtualmachineclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualnetworklinkclient/mock_virtualnetworklinkclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/privatelinkservice"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/routetable"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/securitygroup"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/subnet"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/zone"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

// NewTestScaleSet creates a fake ScaleSet for unit test
func NewTestScaleSet(ctrl *gomock.Controller) (*ScaleSet, error) {
	return newTestScaleSetWithState(ctrl)
}

func newTestScaleSetWithState(ctrl *gomock.Controller) (*ScaleSet, error) {
	cloud := GetTestCloud(ctrl)
	ss, err := newScaleSet(cloud)
	if err != nil {
		return nil, err
	}

	return ss.(*ScaleSet), nil
}

func NewTestFlexScaleSet(ctrl *gomock.Controller) (*FlexScaleSet, error) {
	cloud := GetTestCloud(ctrl)
	fs, err := newFlexScaleSet(cloud)
	if err != nil {
		return nil, err
	}

	return fs.(*FlexScaleSet), nil
}

// GetTestCloud returns a fake azure cloud for unit tests in Azure related CSI drivers
func GetTestCloud(ctrl *gomock.Controller) (az *Cloud) {
	az = &Cloud{
		Config: config.Config{
			AzureClientConfig: config.AzureClientConfig{
				ARMClientConfig: azclient.ARMClientConfig{
					TenantID: "TenantID",
				},
				AzureAuthConfig: azclient.AzureAuthConfig{},
				SubscriptionID:  "subscription",
			},
			ResourceGroup:                            "rg",
			VnetResourceGroup:                        "rg",
			RouteTableResourceGroup:                  "rg",
			SecurityGroupResourceGroup:               "rg",
			PrivateLinkServiceResourceGroup:          "rg",
			Location:                                 "westus",
			VnetName:                                 "vnet",
			SubnetName:                               "subnet",
			SecurityGroupName:                        "nsg",
			RouteTableName:                           "rt",
			PrimaryAvailabilitySetName:               "as",
			PrimaryScaleSetName:                      "vmss",
			MaximumLoadBalancerRuleCount:             250,
			VMType:                                   consts.VMTypeStandard,
			LoadBalancerBackendPoolConfigurationType: consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration,
		},
		nodeZones:                map[string]*utilsets.IgnoreCaseSet{},
		nodeInformerSynced:       func() bool { return true },
		nodeResourceGroups:       map[string]string{},
		unmanagedNodes:           utilsets.NewString(),
		excludeLoadBalancerNodes: utilsets.NewString(),
		nodePrivateIPs:           map[string]*utilsets.IgnoreCaseSet{},
		routeCIDRs:               map[string]string{},
		eventRecorder:            &record.FakeRecorder{},
		Environment:              &azclient.Environment{},
	}
	clientFactory := mock_azclient.NewMockClientFactory(ctrl)
	az.ComputeClientFactory = clientFactory
	az.NetworkClientFactory = clientFactory
	disksClient := mock_diskclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetDiskClient().Return(disksClient).AnyTimes()
	interfacesClient := mock_interfaceclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetInterfaceClient().Return(interfacesClient).AnyTimes()
	loadBalancerClient := mock_loadbalancerclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetLoadBalancerClient().Return(loadBalancerClient).AnyTimes()
	publicIPAddressesClient := mock_publicipaddressclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetPublicIPAddressClient().Return(publicIPAddressesClient).AnyTimes()
	subnetsClient := mock_subnetclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetSubnetClient().Return(subnetsClient).AnyTimes()
	bpClient := mock_backendaddresspoolclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetBackendAddressPoolClient().Return(bpClient).AnyTimes()
	vmasClient := mock_availabilitysetclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetAvailabilitySetClient().Return(vmasClient).AnyTimes()
	virtualMachineScaleSetsClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetVirtualMachineScaleSetClient().Return(virtualMachineScaleSetsClient).AnyTimes()
	virtualMachineScaleSetVMsClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetVirtualMachineScaleSetVMClient().Return(virtualMachineScaleSetVMsClient).AnyTimes()

	virtualMachinesClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetVirtualMachineClient().Return(virtualMachinesClient).AnyTimes()

	securtyGrouptrack2Client := mock_securitygroupclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetSecurityGroupClient().Return(securtyGrouptrack2Client).AnyTimes()
	mockPrivateDNSClient := mock_privatezoneclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetPrivateZoneClient().Return(mockPrivateDNSClient).AnyTimes()
	virtualNetworkLinkClient := mock_virtualnetworklinkclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetVirtualNetworkLinkClient().Return(virtualNetworkLinkClient).AnyTimes()
	subnetTrack2Client := mock_subnetclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetSubnetClient().Return(subnetTrack2Client).AnyTimes()
	privatelinkserviceClient := mock_privatelinkserviceclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetPrivateLinkServiceClient().Return(privatelinkserviceClient).AnyTimes()
	routetableClient := mock_routetableclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetRouteTableClient().Return(routetableClient).AnyTimes()
	privateendpointTrack2Client := mock_privateendpointclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetPrivateEndpointClient().Return(privateendpointTrack2Client).AnyTimes()
	az.AuthProvider = &azclient.AuthProvider{
		ComputeCredential: mock_azclient.NewMockTokenCredential(ctrl),
	}
	az.VMSet, _ = newAvailabilitySet(az)
	az.vmCache, _ = az.newVMCache()
	az.lbCache, _ = az.newLBCache()
	az.nsgRepo, _ = securitygroup.NewSecurityGroupRepo(az.SecurityGroupResourceGroup, az.SecurityGroupName, az.NsgCacheTTLInSeconds, az.DisableAPICallCache, securtyGrouptrack2Client)
	az.subnetRepo = subnet.NewMockRepository(ctrl)
	az.pipCache, _ = az.newPIPCache()
	az.LoadBalancerBackendPool = NewMockBackendPool(ctrl)

	az.plsRepo = privatelinkservice.NewMockRepository(ctrl)
	az.routeTableRepo = routetable.NewMockRepository(ctrl)
	az.zoneRepo = zone.NewMockRepository(ctrl)
	az.regionZonesMap = map[string][]string{az.Location: {"1", "2", "3"}}

	{
		kubeClient := fake.NewSimpleClientset() // FIXME: inject kubeClient
		informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
		az.serviceLister = informerFactory.Core().V1().Services().Lister()
		az.nodeLister = informerFactory.Core().V1().Nodes().Lister()
		informerFactory.Start(wait.NeverStop)
		informerFactory.WaitForCacheSync(wait.NeverStop)
	}
	return az
}

// GetTestCloudWithExtendedLocation returns a fake azure cloud for unit tests in Azure related CSI drivers with extended location.
func GetTestCloudWithExtendedLocation(ctrl *gomock.Controller) (az *Cloud) {
	az = GetTestCloud(ctrl)
	az.ExtendedLocationName = "microsoftlosangeles1"
	az.ExtendedLocationType = "EdgeZone"
	return az
}
