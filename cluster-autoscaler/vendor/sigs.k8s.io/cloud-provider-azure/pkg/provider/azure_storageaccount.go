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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// SkipMatchingTag skip account matching tag
const SkipMatchingTag = "skip-matching"
const LocationGlobal = "global"
const privateDNSZoneNameFmt = "privatelink.%s.%s"

type StorageType string

const (
	StorageTypeBlob StorageType = "blob"
	StorageTypeFile StorageType = "file"
)

// AccountOptions contains the fields which are used to create storage account.
type AccountOptions struct {
	SubscriptionID                            string
	Name, Type, Kind, ResourceGroup, Location string
	EnableHTTPSTrafficOnly                    bool
	// indicate whether create new account when Name is empty or when account does not exists
	CreateAccount                           bool
	CreatePrivateEndpoint                   bool
	StorageType                             StorageType
	StorageEndpointSuffix                   string
	DisableFileServiceDeleteRetentionPolicy *bool
	EnableLargeFileShare                    *bool
	IsHnsEnabled                            *bool
	EnableNfsV3                             *bool
	AllowBlobPublicAccess                   *bool
	RequireInfrastructureEncryption         *bool
	AllowSharedKeyAccess                    *bool
	IsMultichannelEnabled                   *bool
	KeyName                                 *string
	KeyVersion                              *string
	KeyVaultURI                             *string
	Tags                                    map[string]string
	VirtualNetworkResourceIDs               []string
	VNetResourceGroup                       string
	VNetName                                string
	SubnetName                              string
	AccessTier                              string
	MatchTags                               bool
	EnableBlobVersioning                    *bool
	SoftDeleteBlobs                         int32
	SoftDeleteContainers                    int32
}

type accountWithLocation struct {
	Name, StorageType, Location string
}

// getStorageAccounts get matching storage accounts
func (az *Cloud) getStorageAccounts(ctx context.Context, accountOptions *AccountOptions) ([]accountWithLocation, error) {
	if az.StorageAccountClient == nil {
		return nil, fmt.Errorf("StorageAccountClient is nil")
	}
	result, rerr := az.StorageAccountClient.ListByResourceGroup(ctx, accountOptions.SubscriptionID, accountOptions.ResourceGroup)
	if rerr != nil {
		return nil, rerr.Error()
	}

	accounts := []accountWithLocation{}
	for _, acct := range result {
		if acct.Name != nil && acct.Location != nil && acct.Sku != nil {
			if !(isStorageTypeEqual(acct, accountOptions) &&
				isAccountKindEqual(acct, accountOptions) &&
				isLocationEqual(acct, accountOptions) &&
				AreVNetRulesEqual(acct, accountOptions) &&
				isLargeFileSharesPropertyEqual(acct, accountOptions) &&
				isTagsEqual(acct, accountOptions) &&
				isTaggedWithSkip(acct) &&
				isHnsPropertyEqual(acct, accountOptions) &&
				isEnableNfsV3PropertyEqual(acct, accountOptions) &&
				isAllowBlobPublicAccessEqual(acct, accountOptions) &&
				isRequireInfrastructureEncryptionEqual(acct, accountOptions) &&
				isAllowSharedKeyAccessEqual(acct, accountOptions) &&
				isAccessTierEqual(acct, accountOptions) &&
				az.isMultichannelEnabledEqual(ctx, acct, accountOptions) &&
				az.isDisableFileServiceDeleteRetentionPolicyEqual(ctx, acct, accountOptions) &&
				az.isEnableBlobDataProtectionEqual(ctx, acct, accountOptions) &&
				isPrivateEndpointAsExpected(acct, accountOptions)) {
				continue
			}

			accounts = append(accounts, accountWithLocation{Name: *acct.Name, StorageType: string((*acct.Sku).Name), Location: *acct.Location})
		}
	}
	return accounts, nil
}

// GetStorageAccesskey gets the storage account access key
func (az *Cloud) GetStorageAccesskey(ctx context.Context, subsID, account, resourceGroup string) (string, error) {
	if az.StorageAccountClient == nil {
		return "", fmt.Errorf("StorageAccountClient is nil")
	}

	result, rerr := az.StorageAccountClient.ListKeys(ctx, subsID, resourceGroup, account)
	if rerr != nil {
		return "", rerr.Error()
	}
	if result.Keys == nil {
		return "", fmt.Errorf("empty keys")
	}

	for _, k := range *result.Keys {
		if k.Value != nil && *k.Value != "" {
			v := *k.Value
			if ind := strings.LastIndex(v, " "); ind >= 0 {
				v = v[(ind + 1):]
			}
			return v, nil
		}
	}
	return "", fmt.Errorf("no valid keys")
}

// EnsureStorageAccount search storage account, create one storage account(with genAccountNamePrefix) if not found, return accountName, accountKey
func (az *Cloud) EnsureStorageAccount(ctx context.Context, accountOptions *AccountOptions, genAccountNamePrefix string) (string, string, error) {
	if accountOptions == nil {
		return "", "", fmt.Errorf("account options is nil")
	}

	accountName := accountOptions.Name
	accountType := accountOptions.Type
	accountKind := accountOptions.Kind
	resourceGroup := accountOptions.ResourceGroup
	location := accountOptions.Location
	enableHTTPSTrafficOnly := accountOptions.EnableHTTPSTrafficOnly
	vnetResourceGroup := accountOptions.VNetResourceGroup
	vnetName := accountOptions.VNetName
	if vnetName == "" {
		vnetName = az.VnetName
	}

	subnetName := accountOptions.SubnetName
	if subnetName == "" {
		subnetName = az.SubnetName
	}

	if accountOptions.SubscriptionID != "" && !strings.EqualFold(accountOptions.SubscriptionID, az.Config.SubscriptionID) && accountOptions.ResourceGroup == "" {
		return "", "", fmt.Errorf("resourceGroup must be specified when subscriptionID(%s) is not empty", accountOptions.SubscriptionID)
	}

	subsID := az.Config.SubscriptionID
	if accountOptions.SubscriptionID != "" {
		subsID = accountOptions.SubscriptionID
	}

	if location == "" {
		location = az.Location
	}

	var privateDNSZoneName string
	if accountOptions.CreatePrivateEndpoint {
		if accountOptions.StorageType == "" {
			klog.V(2).Info("set StorageType as file when not specified")
			accountOptions.StorageType = StorageTypeFile
		}

		if len(accountOptions.StorageEndpointSuffix) == 0 {
			accountOptions.StorageEndpointSuffix = az.cloud.Environment.StorageEndpointSuffix
		}
		privateDNSZoneName = fmt.Sprintf(privateDNSZoneNameFmt, accountOptions.StorageType, accountOptions.StorageEndpointSuffix)
	}

	if len(accountOptions.Tags) == 0 {
		accountOptions.Tags = make(map[string]string)
	}
	// set built-in tags
	accountOptions.Tags[consts.CreatedByTag] = "azure"

	var createNewAccount bool
	if len(accountName) == 0 {
		createNewAccount = true
		if !accountOptions.CreateAccount {
			// find a storage account that matches accountType
			accounts, err := az.getStorageAccounts(ctx, accountOptions)
			if err != nil {
				return "", "", fmt.Errorf("could not list storage accounts for account type %s: %w", accountType, err)
			}

			if len(accounts) > 0 {
				accountName = accounts[0].Name
				createNewAccount = false
				klog.V(4).Infof("found a matching account %s type %s location %s", accounts[0].Name, accounts[0].StorageType, accounts[0].Location)
			}
		}

		if len(accountName) == 0 {
			accountName = generateStorageAccountName(genAccountNamePrefix)
		}
	} else {
		createNewAccount = false
		if accountOptions.CreateAccount {
			// check whether account exists
			if _, err := az.GetStorageAccesskey(ctx, subsID, accountName, resourceGroup); err != nil {
				klog.V(2).Infof("get storage key for storage account %s returned with %v", accountName, err)
				createNewAccount = true
			}
		}
	}

	if vnetResourceGroup == "" {
		vnetResourceGroup = az.ResourceGroup
		if len(az.VnetResourceGroup) > 0 {
			vnetResourceGroup = az.VnetResourceGroup
		}
	}

	if accountOptions.CreatePrivateEndpoint {
		if _, err := az.privatednsclient.Get(ctx, vnetResourceGroup, privateDNSZoneName); err != nil {
			klog.V(2).Infof("get private dns zone %s returned with %v", privateDNSZoneName, err.Error())
			// Create DNS zone first, this could make sure driver has write permission on vnetResourceGroup
			if err := az.createPrivateDNSZone(ctx, vnetResourceGroup, privateDNSZoneName); err != nil {
				return "", "", fmt.Errorf("create private DNS zone(%s) in resourceGroup(%s): %w", privateDNSZoneName, vnetResourceGroup, err)
			}
		}

		// Create virtual link to the private DNS zone
		vNetLinkName := vnetName + "-vnetlink"
		if _, err := az.virtualNetworkLinksClient.Get(ctx, vnetResourceGroup, privateDNSZoneName, vNetLinkName); err != nil {
			klog.V(2).Infof("get virtual link for vnet(%s) and DNS Zone(%s) returned with %v", vnetName, privateDNSZoneName, err.Error())
			if err := az.createVNetLink(ctx, vNetLinkName, vnetResourceGroup, vnetName, privateDNSZoneName); err != nil {
				return "", "", fmt.Errorf("create virtual link for vnet(%s) and DNS Zone(%s) in resourceGroup(%s): %w", vnetName, privateDNSZoneName, vnetResourceGroup, err)
			}
		}
	}

	if createNewAccount {
		// set network rules for storage account
		var networkRuleSet *storage.NetworkRuleSet
		virtualNetworkRules := []storage.VirtualNetworkRule{}
		for i, subnetID := range accountOptions.VirtualNetworkResourceIDs {
			vnetRule := storage.VirtualNetworkRule{
				VirtualNetworkResourceID: &accountOptions.VirtualNetworkResourceIDs[i],
				Action:                   storage.ActionAllow,
			}
			virtualNetworkRules = append(virtualNetworkRules, vnetRule)
			klog.V(4).Infof("subnetID(%s) has been set", subnetID)
		}
		if len(virtualNetworkRules) > 0 {
			networkRuleSet = &storage.NetworkRuleSet{
				VirtualNetworkRules: &virtualNetworkRules,
				DefaultAction:       storage.DefaultActionDeny,
			}
		}

		if accountOptions.CreatePrivateEndpoint {
			networkRuleSet = &storage.NetworkRuleSet{
				DefaultAction: storage.DefaultActionDeny,
			}
		}

		if accountType == "" {
			accountType = consts.DefaultStorageAccountType
		}

		// use StorageV2 by default per https://docs.microsoft.com/en-us/azure/storage/common/storage-account-options
		kind := consts.DefaultStorageAccountKind
		if accountKind != "" {
			kind = storage.Kind(accountKind)
		}
		tags := convertMapToMapPointer(accountOptions.Tags)

		klog.V(2).Infof("azure - no matching account found, begin to create a new account %s in resource group %s, location: %s, accountType: %s, accountKind: %s, tags: %+v",
			accountName, resourceGroup, location, accountType, kind, accountOptions.Tags)

		cp := storage.AccountCreateParameters{
			Sku:  &storage.Sku{Name: storage.SkuName(accountType)},
			Kind: kind,
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{
				EnableHTTPSTrafficOnly: &enableHTTPSTrafficOnly,
				NetworkRuleSet:         networkRuleSet,
				IsHnsEnabled:           accountOptions.IsHnsEnabled,
				EnableNfsV3:            accountOptions.EnableNfsV3,
				MinimumTLSVersion:      storage.MinimumTLSVersionTLS12,
			},
			Tags:     tags,
			Location: &location}

		if accountOptions.EnableLargeFileShare != nil {
			state := storage.LargeFileSharesStateDisabled
			if *accountOptions.EnableLargeFileShare {
				state = storage.LargeFileSharesStateEnabled
			}
			klog.V(2).Infof("enable LargeFileShare(%s) for storage account(%s)", state, accountName)
			cp.AccountPropertiesCreateParameters.LargeFileSharesState = state
		}
		if accountOptions.AllowBlobPublicAccess != nil {
			klog.V(2).Infof("set AllowBlobPublicAccess(%v) for storage account(%s)", *accountOptions.AllowBlobPublicAccess, accountName)
			cp.AccountPropertiesCreateParameters.AllowBlobPublicAccess = accountOptions.AllowBlobPublicAccess
		}
		if accountOptions.RequireInfrastructureEncryption != nil {
			klog.V(2).Infof("set RequireInfrastructureEncryption(%v) for storage account(%s)", *accountOptions.RequireInfrastructureEncryption, accountName)
			cp.AccountPropertiesCreateParameters.Encryption = &storage.Encryption{
				RequireInfrastructureEncryption: accountOptions.RequireInfrastructureEncryption,
				KeySource:                       storage.KeySourceMicrosoftStorage,
				Services: &storage.EncryptionServices{
					File: &storage.EncryptionService{Enabled: pointer.Bool(true)},
					Blob: &storage.EncryptionService{Enabled: pointer.Bool(true)},
				},
			}
		}
		if accountOptions.AllowSharedKeyAccess != nil {
			klog.V(2).Infof("set Allow SharedKeyAccess (%v) for storage account (%s)", *accountOptions.AllowSharedKeyAccess, accountName)
			cp.AccountPropertiesCreateParameters.AllowSharedKeyAccess = accountOptions.AllowSharedKeyAccess
		}
		if accountOptions.KeyVaultURI != nil {
			klog.V(2).Infof("set KeyVault(%v) for storage account(%s)", accountOptions.KeyVaultURI, accountName)
			cp.AccountPropertiesCreateParameters.Encryption = &storage.Encryption{
				KeyVaultProperties: &storage.KeyVaultProperties{
					KeyName:     accountOptions.KeyName,
					KeyVersion:  accountOptions.KeyVersion,
					KeyVaultURI: accountOptions.KeyVaultURI,
				},
				KeySource: storage.KeySourceMicrosoftKeyvault,
				Services: &storage.EncryptionServices{
					File: &storage.EncryptionService{Enabled: pointer.Bool(true)},
					Blob: &storage.EncryptionService{Enabled: pointer.Bool(true)},
				},
			}
		}
		if az.StorageAccountClient == nil {
			return "", "", fmt.Errorf("StorageAccountClient is nil")
		}

		if rerr := az.StorageAccountClient.Create(ctx, subsID, resourceGroup, accountName, cp); rerr != nil {
			return "", "", fmt.Errorf("failed to create storage account %s, error: %v", accountName, rerr)
		}

		if pointer.BoolDeref(accountOptions.EnableBlobVersioning, false) ||
			accountOptions.SoftDeleteBlobs > 0 ||
			accountOptions.SoftDeleteContainers > 0 {
			var blobPolicy, containerPolicy *storage.DeleteRetentionPolicy
			var enableBlobVersioning *bool

			if accountOptions.SoftDeleteContainers > 0 {
				containerPolicy = &storage.DeleteRetentionPolicy{
					Enabled: pointer.Bool(accountOptions.SoftDeleteContainers > 0),
					Days:    pointer.Int32(accountOptions.SoftDeleteContainers),
				}
			}
			if accountOptions.SoftDeleteBlobs > 0 {
				blobPolicy = &storage.DeleteRetentionPolicy{
					Enabled: pointer.Bool(accountOptions.SoftDeleteBlobs > 0),
					Days:    pointer.Int32(accountOptions.SoftDeleteBlobs),
				}
			}

			if accountOptions.EnableBlobVersioning != nil {
				enableBlobVersioning = pointer.Bool(*accountOptions.EnableBlobVersioning)
			}

			property := storage.BlobServiceProperties{
				BlobServicePropertiesProperties: &storage.BlobServicePropertiesProperties{
					IsVersioningEnabled:            enableBlobVersioning,
					ContainerDeleteRetentionPolicy: containerPolicy,
					DeleteRetentionPolicy:          blobPolicy,
				},
			}

			if _, err := az.BlobClient.SetServiceProperties(ctx, subsID, resourceGroup, accountName, property); err != nil {
				return "", "", fmt.Errorf("failed to set blob service properties for storage account %s, error: %w", accountName, err)
			}
		}

		if accountOptions.DisableFileServiceDeleteRetentionPolicy != nil || accountOptions.IsMultichannelEnabled != nil {
			prop, err := az.FileClient.WithSubscriptionID(subsID).GetServiceProperties(ctx, resourceGroup, accountName)
			if err != nil {
				return "", "", err
			}
			if prop.FileServicePropertiesProperties == nil {
				return "", "", fmt.Errorf("FileServicePropertiesProperties of account(%s), subscription(%s), resource group(%s) is nil", accountName, subsID, resourceGroup)
			}
			prop.FileServicePropertiesProperties.ProtocolSettings = nil
			prop.FileServicePropertiesProperties.Cors = nil
			if accountOptions.DisableFileServiceDeleteRetentionPolicy != nil {
				enable := !*accountOptions.DisableFileServiceDeleteRetentionPolicy
				klog.V(2).Infof("set ShareDeleteRetentionPolicy(%v) on account(%s), subscription(%s), resource group(%s)",
					enable, accountName, subsID, resourceGroup)
				prop.FileServicePropertiesProperties.ShareDeleteRetentionPolicy = &storage.DeleteRetentionPolicy{Enabled: &enable}
			}
			if accountOptions.IsMultichannelEnabled != nil {
				klog.V(2).Infof("enable SMB Multichannel setting on account(%s), subscription(%s), resource group(%s)", accountName, subsID, resourceGroup)
				enabled := *accountOptions.IsMultichannelEnabled
				prop.FileServicePropertiesProperties.ProtocolSettings = &storage.ProtocolSettings{Smb: &storage.SmbSetting{Multichannel: &storage.Multichannel{Enabled: &enabled}}}
			}
			if _, err := az.FileClient.WithSubscriptionID(subsID).SetServiceProperties(ctx, resourceGroup, accountName, prop); err != nil {
				return "", "", err
			}
		}

		if accountOptions.AccessTier != "" {
			klog.V(2).Infof("set AccessTier(%s) on account(%s), subscription(%s), resource group(%s)", accountOptions.AccessTier, accountName, subsID, resourceGroup)
			cp.AccountPropertiesCreateParameters.AccessTier = storage.AccessTier(accountOptions.AccessTier)
		}
	}

	if accountOptions.CreatePrivateEndpoint {
		// Get properties of the storageAccount
		storageAccount, err := az.StorageAccountClient.GetProperties(ctx, subsID, resourceGroup, accountName)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the properties of storage account(%s), resourceGroup(%s), error: %v", accountName, resourceGroup, err)
		}

		// Create private endpoint
		privateEndpointName := accountName + "-pvtendpoint"
		if accountOptions.StorageType == StorageTypeBlob {
			privateEndpointName = privateEndpointName + "-blob"
		}
		if err := az.createPrivateEndpoint(ctx, accountName, storageAccount.ID, privateEndpointName, vnetResourceGroup, vnetName, subnetName, location, accountOptions.StorageType); err != nil {
			return "", "", fmt.Errorf("create private endpoint for storage account(%s), resourceGroup(%s): %w", accountName, vnetResourceGroup, err)
		}

		// Create dns zone group
		dnsZoneGroupName := accountName + "-dnszonegroup"
		if accountOptions.StorageType == StorageTypeBlob {
			dnsZoneGroupName = dnsZoneGroupName + "-blob"
		}
		if err := az.createPrivateDNSZoneGroup(ctx, dnsZoneGroupName, privateEndpointName, vnetResourceGroup, vnetName, privateDNSZoneName); err != nil {
			return "", "", fmt.Errorf("create private DNS zone group - privateEndpoint(%s), vNetName(%s), resourceGroup(%s): %w", privateEndpointName, vnetName, vnetResourceGroup, err)
		}
	}

	// find the access key with this account
	accountKey, err := az.GetStorageAccesskey(ctx, subsID, accountName, resourceGroup)
	if err != nil {
		return "", "", fmt.Errorf("could not get storage key for storage account %s: %w", accountName, err)
	}

	return accountName, accountKey, nil
}

func (az *Cloud) createPrivateEndpoint(ctx context.Context, accountName string, accountID *string, privateEndpointName, vnetResourceGroup, vnetName, subnetName, location string, storageType StorageType) error {
	klog.V(2).Infof("Creating private endpoint(%s) for account (%s)", privateEndpointName, accountName)

	subnet, _, err := az.getSubnet(vnetName, subnetName)
	if err != nil {
		return err
	}
	if subnet.SubnetPropertiesFormat == nil {
		klog.Errorf("SubnetPropertiesFormat of (%s, %s) is nil", vnetName, subnetName)
	} else {
		// Disable the private endpoint network policies before creating private endpoint
		subnet.SubnetPropertiesFormat.PrivateEndpointNetworkPolicies = network.VirtualNetworkPrivateEndpointNetworkPoliciesDisabled
	}
	if rerr := az.SubnetsClient.CreateOrUpdate(ctx, vnetResourceGroup, vnetName, subnetName, subnet); rerr != nil {
		return rerr.Error()
	}

	//Create private endpoint
	privateLinkServiceConnectionName := accountName + "-pvtsvcconn"
	if storageType == StorageTypeBlob {
		privateLinkServiceConnectionName = privateLinkServiceConnectionName + "-blob"
	}
	privateLinkServiceConnection := network.PrivateLinkServiceConnection{
		Name: &privateLinkServiceConnectionName,
		PrivateLinkServiceConnectionProperties: &network.PrivateLinkServiceConnectionProperties{
			GroupIds:             &[]string{string(storageType)},
			PrivateLinkServiceID: accountID,
		},
	}
	privateLinkServiceConnections := []network.PrivateLinkServiceConnection{privateLinkServiceConnection}
	privateEndpoint := network.PrivateEndpoint{
		Location:                  &location,
		PrivateEndpointProperties: &network.PrivateEndpointProperties{Subnet: &subnet, PrivateLinkServiceConnections: &privateLinkServiceConnections},
	}

	return az.privateendpointclient.CreateOrUpdate(ctx, vnetResourceGroup, privateEndpointName, privateEndpoint, "", true).Error()
}

func (az *Cloud) createPrivateDNSZone(ctx context.Context, vnetResourceGroup, privateDNSZoneName string) error {
	klog.V(2).Infof("Creating private dns zone(%s) in resourceGroup (%s)", privateDNSZoneName, vnetResourceGroup)
	location := LocationGlobal
	privateDNSZone := privatedns.PrivateZone{Location: &location}
	if err := az.privatednsclient.CreateOrUpdate(ctx, vnetResourceGroup, privateDNSZoneName, privateDNSZone, "", true); err != nil {
		if strings.Contains(err.Error().Error(), "exists already") {
			klog.V(2).Infof("private dns zone(%s) in resourceGroup (%s) already exists", privateDNSZoneName, vnetResourceGroup)
			return nil
		}
		return err.Error()
	}
	return nil
}

func (az *Cloud) createVNetLink(ctx context.Context, vNetLinkName, vnetResourceGroup, vnetName, privateDNSZoneName string) error {
	klog.V(2).Infof("Creating virtual link for vnet(%s) and DNS Zone(%s) in resourceGroup(%s)", vNetLinkName, privateDNSZoneName, vnetResourceGroup)
	location := LocationGlobal
	vnetID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", az.SubscriptionID, vnetResourceGroup, vnetName)
	parameters := privatedns.VirtualNetworkLink{
		Location: &location,
		VirtualNetworkLinkProperties: &privatedns.VirtualNetworkLinkProperties{
			VirtualNetwork:      &privatedns.SubResource{ID: &vnetID},
			RegistrationEnabled: pointer.Bool(false)},
	}
	return az.virtualNetworkLinksClient.CreateOrUpdate(ctx, vnetResourceGroup, privateDNSZoneName, vNetLinkName, parameters, "", false).Error()
}

func (az *Cloud) createPrivateDNSZoneGroup(ctx context.Context, dnsZoneGroupName, privateEndpointName, vnetResourceGroup, vnetName, privateDNSZoneName string) error {
	klog.V(2).Infof("Creating private DNS zone group(%s) with privateEndpoint(%s), vNetName(%s), resourceGroup(%s)", dnsZoneGroupName, privateEndpointName, vnetName, vnetResourceGroup)
	privateDNSZoneID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/privateDnsZones/%s", az.SubscriptionID, vnetResourceGroup, privateDNSZoneName)
	privateDNSZoneConfig := network.PrivateDNSZoneConfig{
		Name: &privateDNSZoneName,
		PrivateDNSZonePropertiesFormat: &network.PrivateDNSZonePropertiesFormat{
			PrivateDNSZoneID: &privateDNSZoneID},
	}
	privateDNSZoneConfigs := []network.PrivateDNSZoneConfig{privateDNSZoneConfig}
	privateDNSZoneGroup := network.PrivateDNSZoneGroup{
		PrivateDNSZoneGroupPropertiesFormat: &network.PrivateDNSZoneGroupPropertiesFormat{
			PrivateDNSZoneConfigs: &privateDNSZoneConfigs,
		},
	}
	return az.privatednszonegroupclient.CreateOrUpdate(ctx, vnetResourceGroup, privateEndpointName, dnsZoneGroupName, privateDNSZoneGroup, "", false).Error()
}

// AddStorageAccountTags add tags to storage account
func (az *Cloud) AddStorageAccountTags(ctx context.Context, subsID, resourceGroup, account string, tags map[string]*string) *retry.Error {
	if az.StorageAccountClient == nil {
		return retry.NewError(false, fmt.Errorf("StorageAccountClient is nil"))
	}
	result, rerr := az.StorageAccountClient.GetProperties(ctx, subsID, resourceGroup, account)
	if rerr != nil {
		return rerr
	}

	newTags := result.Tags
	if newTags == nil {
		newTags = make(map[string]*string)
	}

	// merge two tag map
	for k, v := range tags {
		newTags[k] = v
	}

	updateParams := storage.AccountUpdateParameters{Tags: newTags}
	return az.StorageAccountClient.Update(ctx, subsID, resourceGroup, account, updateParams)
}

// RemoveStorageAccountTag remove tag from storage account
func (az *Cloud) RemoveStorageAccountTag(ctx context.Context, subsID, resourceGroup, account, key string) *retry.Error {
	if az.StorageAccountClient == nil {
		return retry.NewError(false, fmt.Errorf("StorageAccountClient is nil"))
	}
	result, rerr := az.StorageAccountClient.GetProperties(ctx, subsID, resourceGroup, account)
	if rerr != nil {
		return rerr
	}

	if len(result.Tags) == 0 {
		return nil
	}

	originalLen := len(result.Tags)
	delete(result.Tags, key)
	if originalLen != len(result.Tags) {
		updateParams := storage.AccountUpdateParameters{Tags: result.Tags}
		return az.StorageAccountClient.Update(ctx, subsID, resourceGroup, account, updateParams)
	}
	return nil
}

func isStorageTypeEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.Type != "" && !strings.EqualFold(accountOptions.Type, string((*account.Sku).Name)) {
		return false
	}
	return true
}

func isAccountKindEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.Kind != "" && !strings.EqualFold(accountOptions.Kind, string(account.Kind)) {
		return false
	}
	return true
}

func isLocationEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.Location != "" && !strings.EqualFold(accountOptions.Location, *account.Location) {
		return false
	}
	return true
}

func AreVNetRulesEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if len(accountOptions.VirtualNetworkResourceIDs) > 0 {
		if account.AccountProperties == nil || account.AccountProperties.NetworkRuleSet == nil ||
			account.AccountProperties.NetworkRuleSet.VirtualNetworkRules == nil {
			return false
		}

		found := false
		for _, subnetID := range accountOptions.VirtualNetworkResourceIDs {
			for _, rule := range *account.AccountProperties.NetworkRuleSet.VirtualNetworkRules {
				if strings.EqualFold(pointer.StringDeref(rule.VirtualNetworkResourceID, ""), subnetID) && rule.Action == storage.ActionAllow {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func isLargeFileSharesPropertyEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.EnableLargeFileShare == nil {
		return true
	}
	if *accountOptions.EnableLargeFileShare {
		return account.LargeFileSharesState == storage.LargeFileSharesStateEnabled
	}
	return account.LargeFileSharesState == "" || account.LargeFileSharesState == storage.LargeFileSharesStateDisabled
}

func isTaggedWithSkip(account storage.Account) bool {
	if account.Tags != nil {
		// skip account with SkipMatchingTag tag
		if _, ok := account.Tags[SkipMatchingTag]; ok {
			klog.V(2).Infof("found %s tag for account %s, skip matching", SkipMatchingTag, *account.Name)
			return false
		}
	}
	return true
}

func isTagsEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if !accountOptions.MatchTags {
		// always return true when tags matching is false (by default)
		return true
	}

	// nil and empty map should be regarded as equal
	if len(account.Tags) == 0 && len(accountOptions.Tags) == 0 {
		return true
	}

	for k, v := range account.Tags {
		var value string
		// nil and empty value should be regarded as equal
		if v != nil {
			value = *v
		}
		if accountOptions.Tags[k] != value {
			return false
		}
	}

	return true
}

func isHnsPropertyEqual(account storage.Account, accountOptions *AccountOptions) bool {
	return pointer.BoolDeref(accountOptions.IsHnsEnabled, false) == pointer.BoolDeref(account.IsHnsEnabled, false)
}

func isEnableNfsV3PropertyEqual(account storage.Account, accountOptions *AccountOptions) bool {
	return pointer.BoolDeref(accountOptions.EnableNfsV3, false) == pointer.BoolDeref(account.EnableNfsV3, false)
}

func isPrivateEndpointAsExpected(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.CreatePrivateEndpoint && account.PrivateEndpointConnections != nil && len(*account.PrivateEndpointConnections) > 0 {
		return true
	}
	if !accountOptions.CreatePrivateEndpoint && (account.PrivateEndpointConnections == nil || len(*account.PrivateEndpointConnections) == 0) {
		return true
	}
	return false
}

func isAllowBlobPublicAccessEqual(account storage.Account, accountOptions *AccountOptions) bool {
	return pointer.BoolDeref(accountOptions.AllowBlobPublicAccess, false) == pointer.BoolDeref(account.AllowBlobPublicAccess, false)
}

func isRequireInfrastructureEncryptionEqual(account storage.Account, accountOptions *AccountOptions) bool {
	requireInfraEncryption := pointer.BoolDeref(accountOptions.RequireInfrastructureEncryption, false)
	if account.Encryption == nil {
		return !requireInfraEncryption
	}
	return requireInfraEncryption == pointer.BoolDeref(account.Encryption.RequireInfrastructureEncryption, false)
}

func isAllowSharedKeyAccessEqual(account storage.Account, accountOptions *AccountOptions) bool {
	return pointer.BoolDeref(accountOptions.AllowSharedKeyAccess, false) == pointer.BoolDeref(account.AllowSharedKeyAccess, false)
}

func isAccessTierEqual(account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.AccessTier == "" {
		return true
	}
	return accountOptions.AccessTier == string(account.AccessTier)
}

func (az *Cloud) isMultichannelEnabledEqual(ctx context.Context, account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.IsMultichannelEnabled == nil {
		return true
	}

	if account.Name == nil {
		klog.Warningf("account.Name under resource group(%s) is nil", accountOptions.ResourceGroup)
		return false
	}

	prop, err := az.FileClient.WithSubscriptionID(accountOptions.SubscriptionID).GetServiceProperties(ctx, accountOptions.ResourceGroup, *account.Name)
	if err != nil {
		klog.Warningf("GetServiceProperties(%s) under resource group(%s) failed with %v", *account.Name, accountOptions.ResourceGroup, err)
		return false
	}

	if prop.FileServicePropertiesProperties == nil ||
		prop.FileServicePropertiesProperties.ProtocolSettings == nil ||
		prop.FileServicePropertiesProperties.ProtocolSettings.Smb == nil ||
		prop.FileServicePropertiesProperties.ProtocolSettings.Smb.Multichannel == nil {
		return !*accountOptions.IsMultichannelEnabled
	}

	return *accountOptions.IsMultichannelEnabled == pointer.BoolDeref(prop.FileServicePropertiesProperties.ProtocolSettings.Smb.Multichannel.Enabled, false)
}

func (az *Cloud) isDisableFileServiceDeleteRetentionPolicyEqual(ctx context.Context, account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.DisableFileServiceDeleteRetentionPolicy == nil {
		return true
	}

	if account.Name == nil {
		klog.Warningf("account.Name under resource group(%s) is nil", accountOptions.ResourceGroup)
		return false
	}

	prop, err := az.FileClient.WithSubscriptionID(accountOptions.SubscriptionID).GetServiceProperties(ctx, accountOptions.ResourceGroup, *account.Name)
	if err != nil {
		klog.Warningf("GetServiceProperties(%s) under resource group(%s) failed with %v", *account.Name, accountOptions.ResourceGroup, err)
		return false
	}

	if prop.FileServicePropertiesProperties == nil ||
		prop.FileServicePropertiesProperties.ShareDeleteRetentionPolicy == nil ||
		prop.FileServicePropertiesProperties.ShareDeleteRetentionPolicy.Enabled == nil {
		// by default, ShareDeleteRetentionPolicy.Enabled is true if it's nil
		return !*accountOptions.DisableFileServiceDeleteRetentionPolicy
	}

	return *accountOptions.DisableFileServiceDeleteRetentionPolicy != *prop.FileServicePropertiesProperties.ShareDeleteRetentionPolicy.Enabled
}

func (az *Cloud) isEnableBlobDataProtectionEqual(ctx context.Context, account storage.Account, accountOptions *AccountOptions) bool {
	if accountOptions.SoftDeleteBlobs == 0 &&
		accountOptions.SoftDeleteContainers == 0 &&
		accountOptions.EnableBlobVersioning == nil {
		return true
	}

	property, err := az.BlobClient.GetServiceProperties(ctx, accountOptions.SubscriptionID, accountOptions.ResourceGroup, *account.Name)
	if err != nil {
		klog.Warningf("GetServiceProperties failed for account %s, err: %v", *account.Name, err)
		return false
	}

	return isSoftDeleteBlobsEqual(property, accountOptions) &&
		isSoftDeleteContainersEqual(property, accountOptions) &&
		isEnableBlobVersioningEqual(property, accountOptions)
}

func isSoftDeleteBlobsEqual(property storage.BlobServiceProperties, accountOptions *AccountOptions) bool {
	wantEnable := accountOptions.SoftDeleteBlobs > 0
	actualEnable := property.DeleteRetentionPolicy != nil &&
		pointer.BoolDeref(property.DeleteRetentionPolicy.Enabled, false)
	if wantEnable != actualEnable {
		return false
	}
	if !actualEnable {
		return true
	}

	return accountOptions.SoftDeleteBlobs == pointer.Int32Deref(property.DeleteRetentionPolicy.Days, 0)
}

func isSoftDeleteContainersEqual(property storage.BlobServiceProperties, accountOptions *AccountOptions) bool {
	wantEnable := accountOptions.SoftDeleteContainers > 0
	actualEnable := property.ContainerDeleteRetentionPolicy != nil &&
		pointer.BoolDeref(property.ContainerDeleteRetentionPolicy.Enabled, false)
	if wantEnable != actualEnable {
		return false
	}
	if !actualEnable {
		return true
	}

	return accountOptions.SoftDeleteContainers == pointer.Int32Deref(property.ContainerDeleteRetentionPolicy.Days, 0)
}

func isEnableBlobVersioningEqual(property storage.BlobServiceProperties, accountOptions *AccountOptions) bool {
	return pointer.BoolDeref(accountOptions.EnableBlobVersioning, false) == pointer.BoolDeref(property.IsVersioningEnabled, false)
}
