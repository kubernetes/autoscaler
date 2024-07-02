# Release History

## 1.4.0 (2023-11-24)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.


## 1.3.0 (2023-10-27)
### Features Added

- New value `ManagedHsmSKUNameCustomB6` added to enum type `ManagedHsmSKUName`
- New enum type `ManagedServiceIdentityType` with values `ManagedServiceIdentityTypeNone`, `ManagedServiceIdentityTypeSystemAssigned`, `ManagedServiceIdentityTypeSystemAssignedUserAssigned`, `ManagedServiceIdentityTypeUserAssigned`
- New struct `ManagedServiceIdentity`
- New struct `UserAssignedIdentity`
- New field `Identity` in struct `MHSMPrivateEndpointConnection`
- New field `Identity` in struct `MHSMPrivateLinkResource`
- New field `Identity` in struct `ManagedHsm`


## 1.2.0 (2023-04-28)
### Features Added

- New value `JSONWebKeyOperationRelease` added to enum type `JSONWebKeyOperation`
- New value `KeyPermissionsGetrotationpolicy`, `KeyPermissionsRelease`, `KeyPermissionsRotate`, `KeyPermissionsSetrotationpolicy` added to enum type `KeyPermissions`
- New enum type `ActivationStatus` with values `ActivationStatusActive`, `ActivationStatusFailed`, `ActivationStatusNotActivated`, `ActivationStatusUnknown`
- New enum type `GeoReplicationRegionProvisioningState` with values `GeoReplicationRegionProvisioningStateCleanup`, `GeoReplicationRegionProvisioningStateDeleting`, `GeoReplicationRegionProvisioningStateFailed`, `GeoReplicationRegionProvisioningStatePreprovisioning`, `GeoReplicationRegionProvisioningStateProvisioning`, `GeoReplicationRegionProvisioningStateSucceeded`
- New enum type `KeyRotationPolicyActionType` with values `KeyRotationPolicyActionTypeNotify`, `KeyRotationPolicyActionTypeRotate`
- New function `*ClientFactory.NewMHSMRegionsClient() *MHSMRegionsClient`
- New function `*ClientFactory.NewManagedHsmKeysClient() *ManagedHsmKeysClient`
- New function `NewMHSMRegionsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MHSMRegionsClient, error)`
- New function `*MHSMRegionsClient.NewListByResourcePager(string, string, *MHSMRegionsClientListByResourceOptions) *runtime.Pager[MHSMRegionsClientListByResourceResponse]`
- New function `NewManagedHsmKeysClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedHsmKeysClient, error)`
- New function `*ManagedHsmKeysClient.CreateIfNotExist(context.Context, string, string, string, ManagedHsmKeyCreateParameters, *ManagedHsmKeysClientCreateIfNotExistOptions) (ManagedHsmKeysClientCreateIfNotExistResponse, error)`
- New function `*ManagedHsmKeysClient.Get(context.Context, string, string, string, *ManagedHsmKeysClientGetOptions) (ManagedHsmKeysClientGetResponse, error)`
- New function `*ManagedHsmKeysClient.GetVersion(context.Context, string, string, string, string, *ManagedHsmKeysClientGetVersionOptions) (ManagedHsmKeysClientGetVersionResponse, error)`
- New function `*ManagedHsmKeysClient.NewListPager(string, string, *ManagedHsmKeysClientListOptions) *runtime.Pager[ManagedHsmKeysClientListResponse]`
- New function `*ManagedHsmKeysClient.NewListVersionsPager(string, string, string, *ManagedHsmKeysClientListVersionsOptions) *runtime.Pager[ManagedHsmKeysClientListVersionsResponse]`
- New function `*ManagedHsmsClient.CheckMhsmNameAvailability(context.Context, CheckMhsmNameAvailabilityParameters, *ManagedHsmsClientCheckMhsmNameAvailabilityOptions) (ManagedHsmsClientCheckMhsmNameAvailabilityResponse, error)`
- New struct `Action`
- New struct `CheckMhsmNameAvailabilityParameters`
- New struct `CheckMhsmNameAvailabilityResult`
- New struct `KeyReleasePolicy`
- New struct `KeyRotationPolicyAttributes`
- New struct `LifetimeAction`
- New struct `MHSMGeoReplicatedRegion`
- New struct `MHSMRegionsListResult`
- New struct `ManagedHSMSecurityDomainProperties`
- New struct `ManagedHsmAction`
- New struct `ManagedHsmKey`
- New struct `ManagedHsmKeyAttributes`
- New struct `ManagedHsmKeyCreateParameters`
- New struct `ManagedHsmKeyListResult`
- New struct `ManagedHsmKeyProperties`
- New struct `ManagedHsmKeyReleasePolicy`
- New struct `ManagedHsmKeyRotationPolicyAttributes`
- New struct `ManagedHsmLifetimeAction`
- New struct `ManagedHsmRotationPolicy`
- New struct `ManagedHsmTrigger`
- New struct `ProxyResourceWithoutSystemData`
- New struct `RotationPolicy`
- New struct `Trigger`
- New field `ReleasePolicy` in struct `KeyProperties`
- New field `RotationPolicy` in struct `KeyProperties`
- New field `Etag` in struct `MHSMPrivateEndpointConnectionItem`
- New field `ID` in struct `MHSMPrivateEndpointConnectionItem`
- New field `Regions` in struct `ManagedHsmProperties`
- New field `SecurityDomainProperties` in struct `ManagedHsmProperties`


## 1.1.0 (2023-04-06)
### Features Added

- New struct `ClientFactory` which is a client factory used to create any client in this module


## 1.0.0 (2022-05-16)

The package of `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault` is using our [next generation design principles](https://azure.github.io/azure-sdk/general_introduction.html) since version 1.0.0, which contains breaking changes.

To migrate the existing applications to the latest version, please refer to [Migration Guide](https://aka.ms/azsdk/go/mgmt/migration).

To learn more, please refer to our documentation [Quick Start](https://aka.ms/azsdk/go/mgmt).