# Release History

## 8.3.0-beta.1 (2025-12-26)
### Features Added

- New value `AgentPoolModeMachines`, `AgentPoolModeManagedSystem` added to enum type `AgentPoolMode`
- New value `AgentPoolSSHAccessEntraID` added to enum type `AgentPoolSSHAccess`
- New value `OSSKUFlatcar`, `OSSKUMariner`, `OSSKUWindows2025`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AccelerationMode` with values `AccelerationModeBpfVeth`, `AccelerationModeNone`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriftAction` with values `DriftActionRecreate`, `DriftActionSynced`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IdentityBindingProvisioningState` with values `IdentityBindingProvisioningStateCanceled`, `IdentityBindingProvisioningStateCreating`, `IdentityBindingProvisioningStateDeleting`, `IdentityBindingProvisioningStateFailed`, `IdentityBindingProvisioningStateSucceeded`, `IdentityBindingProvisioningStateUpdating`
- New enum type `InfrastructureEncryption` with values `InfrastructureEncryptionDisabled`, `InfrastructureEncryptionEnabled`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `JWTAuthenticatorProvisioningState` with values `JWTAuthenticatorProvisioningStateCanceled`, `JWTAuthenticatorProvisioningStateCreating`, `JWTAuthenticatorProvisioningStateDeleting`, `JWTAuthenticatorProvisioningStateFailed`, `JWTAuthenticatorProvisioningStateSucceeded`, `JWTAuthenticatorProvisioningStateUpdating`
- New enum type `ManagedGatewayType` with values `ManagedGatewayTypeDisabled`, `ManagedGatewayTypeStandard`
- New enum type `MeshMembershipProvisioningState` with values `MeshMembershipProvisioningStateCanceled`, `MeshMembershipProvisioningStateCreating`, `MeshMembershipProvisioningStateDeleting`, `MeshMembershipProvisioningStateFailed`, `MeshMembershipProvisioningStateSucceeded`, `MeshMembershipProvisioningStateUpdating`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`, `ModeNFTABLES`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `ProxyRedirectionMechanism` with values `ProxyRedirectionMechanismCNIChaining`, `ProxyRedirectionMechanismInitContainers`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New enum type `UpgradeStrategy` with values `UpgradeStrategyBlueGreen`, `UpgradeStrategyRolling`
- New enum type `VMState` with values `VMStateDeleted`, `VMStateRunning`
- New function `*AgentPoolsClient.BeginCompleteUpgrade(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, options *AgentPoolsClientBeginCompleteUpgradeOptions) (*runtime.Poller[AgentPoolsClientCompleteUpgradeResponse], error)`
- New function `NewClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(location string, options *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewIdentityBindingsClient() *IdentityBindingsClient`
- New function `*ClientFactory.NewJWTAuthenticatorsClient() *JWTAuthenticatorsClient`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewMeshMembershipsClient() *MeshMembershipsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewIdentityBindingsClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*IdentityBindingsClient, error)`
- New function `*IdentityBindingsClient.BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, identityBindingName string, parameters IdentityBinding, options *IdentityBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[IdentityBindingsClientCreateOrUpdateResponse], error)`
- New function `*IdentityBindingsClient.BeginDelete(ctx context.Context, resourceGroupName string, resourceName string, identityBindingName string, options *IdentityBindingsClientBeginDeleteOptions) (*runtime.Poller[IdentityBindingsClientDeleteResponse], error)`
- New function `*IdentityBindingsClient.Get(ctx context.Context, resourceGroupName string, resourceName string, identityBindingName string, options *IdentityBindingsClientGetOptions) (IdentityBindingsClientGetResponse, error)`
- New function `*IdentityBindingsClient.NewListByManagedClusterPager(resourceGroupName string, resourceName string, options *IdentityBindingsClientListByManagedClusterOptions) *runtime.Pager[IdentityBindingsClientListByManagedClusterResponse]`
- New function `NewJWTAuthenticatorsClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*JWTAuthenticatorsClient, error)`
- New function `*JWTAuthenticatorsClient.BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, jwtAuthenticatorName string, parameters JWTAuthenticator, options *JWTAuthenticatorsClientBeginCreateOrUpdateOptions) (*runtime.Poller[JWTAuthenticatorsClientCreateOrUpdateResponse], error)`
- New function `*JWTAuthenticatorsClient.BeginDelete(ctx context.Context, resourceGroupName string, resourceName string, jwtAuthenticatorName string, options *JWTAuthenticatorsClientBeginDeleteOptions) (*runtime.Poller[JWTAuthenticatorsClientDeleteResponse], error)`
- New function `*JWTAuthenticatorsClient.Get(ctx context.Context, resourceGroupName string, resourceName string, jwtAuthenticatorName string, options *JWTAuthenticatorsClientGetOptions) (JWTAuthenticatorsClientGetResponse, error)`
- New function `*JWTAuthenticatorsClient.NewListByManagedClusterPager(resourceGroupName string, resourceName string, options *JWTAuthenticatorsClientListByManagedClusterOptions) *runtime.Pager[JWTAuthenticatorsClientListByManagedClusterResponse]`
- New function `NewLoadBalancersClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, loadBalancerName string, parameters LoadBalancer, options *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(ctx context.Context, resourceGroupName string, resourceName string, loadBalancerName string, options *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(ctx context.Context, resourceGroupName string, resourceName string, loadBalancerName string, options *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(resourceGroupName string, resourceName string, options *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*MachinesClient.BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, machineName string, parameters Machine, options *MachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[MachinesClientCreateOrUpdateResponse], error)`
- New function `NewManagedClusterSnapshotsClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, parameters ManagedClusterSnapshot, options *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(ctx context.Context, resourceGroupName string, resourceName string, options *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(ctx context.Context, resourceGroupName string, resourceName string, options *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(resourceGroupName string, options *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(options *ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(ctx context.Context, resourceGroupName string, resourceName string, parameters TagsObject, options *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(ctx context.Context, location string, version string, options *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(ctx context.Context, location string, version string, options *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(location string, options *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(location string, options *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(ctx context.Context, resourceGroupName string, resourceName string, parameters RebalanceLoadBalancersRequestBody, options *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewMeshMembershipsClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*MeshMembershipsClient, error)`
- New function `*MeshMembershipsClient.BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, meshMembershipName string, parameters MeshMembership, options *MeshMembershipsClientBeginCreateOrUpdateOptions) (*runtime.Poller[MeshMembershipsClientCreateOrUpdateResponse], error)`
- New function `*MeshMembershipsClient.BeginDelete(ctx context.Context, resourceGroupName string, resourceName string, meshMembershipName string, options *MeshMembershipsClientBeginDeleteOptions) (*runtime.Poller[MeshMembershipsClientDeleteResponse], error)`
- New function `*MeshMembershipsClient.Get(ctx context.Context, resourceGroupName string, resourceName string, meshMembershipName string, options *MeshMembershipsClientGetOptions) (MeshMembershipsClientGetResponse, error)`
- New function `*MeshMembershipsClient.NewListByManagedClusterPager(resourceGroupName string, resourceName string, options *MeshMembershipsClientListByManagedClusterOptions) *runtime.Pager[MeshMembershipsClientListByManagedClusterResponse]`
- New function `NewOperationStatusResultClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(ctx context.Context, resourceGroupName string, resourceName string, operationID string, options *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, operationID string, options *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(resourceGroupName string, resourceName string, options *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworkingPerformance`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolBlueGreenUpgradeSettings`
- New struct `AgentPoolRecentlyUsedVersion`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `IdentityBinding`
- New struct `IdentityBindingListResult`
- New struct `IdentityBindingManagedIdentityProfile`
- New struct `IdentityBindingOidcIssuerProfile`
- New struct `IdentityBindingProperties`
- New struct `JWTAuthenticator`
- New struct `JWTAuthenticatorClaimMappingExpression`
- New struct `JWTAuthenticatorClaimMappings`
- New struct `JWTAuthenticatorExtraClaimMappingExpression`
- New struct `JWTAuthenticatorIssuer`
- New struct `JWTAuthenticatorListResult`
- New struct `JWTAuthenticatorProperties`
- New struct `JWTAuthenticatorValidationRule`
- New struct `KubernetesResourceObjectEncryptionProfile`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `MachineHardwareProfile`
- New struct `MachineKubernetesProfile`
- New struct `MachineOSProfile`
- New struct `MachineOSProfileLinuxProfile`
- New struct `MachineSecurityProfile`
- New struct `MachineStatus`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterHostedSystemProfile`
- New struct `ManagedClusterIngressDefaultDomainProfile`
- New struct `ManagedClusterIngressProfileApplicationLoadBalancer`
- New struct `ManagedClusterIngressProfileGatewayConfiguration`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `MeshMembership`
- New struct `MeshMembershipProperties`
- New struct `MeshMembershipsListResult`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeCustomizationProfile`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `Performance` in struct `AdvancedNetworking`
- New field `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `ComponentsByReleases`, `RecentlyUsedVersions` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes` in struct `AgentPoolUpgradeSettings`
- New field `DriverType` in struct `GPUProfile`
- New field `ProxyRedirectionMechanism` in struct `IstioComponents`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `EnableNodePublicIP`, `NodePublicIPPrefixID`, `NodePublicIPTags`, `PodSubnetID`, `VnetSubnetID` in struct `MachineNetworkProperties`
- New field `ETag`, `Hardware`, `Kubernetes`, `Mode`, `NodeImageVersion`, `OperatingSystem`, `Priority`, `ProvisioningState`, `Security`, `Status`, `Tags` in struct `MachineProperties`
- New field `ArtifactStreamingProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy`, `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `ApplicationLoadBalancer`, `GatewayAPI` in struct `ManagedClusterIngressProfile`
- New field `DefaultDomain` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `CreationData`, `EnableNamespaceResources`, `HostedSystemProfile`, `SchedulerProfile` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `KubernetesResourceObjectEncryptionProfile`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 8.2.0 (2025-12-25)
### Features Added

- New value `OSSKUUbuntu2404` added to enum type `OSSKU`


## 8.2.0-beta.1 (2025-11-21)
### Features Added

- New value `AgentPoolModeMachines`, `AgentPoolModeManagedSystem` added to enum type `AgentPoolMode`
- New value `AgentPoolSSHAccessEntraID` added to enum type `AgentPoolSSHAccess`
- New value `OSSKUFlatcar`, `OSSKUMariner`, `OSSKUUbuntu2404`, `OSSKUWindows2025`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AccelerationMode` with values `AccelerationModeBpfVeth`, `AccelerationModeNone`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriftAction` with values `DriftActionRecreate`, `DriftActionSynced`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IdentityBindingProvisioningState` with values `IdentityBindingProvisioningStateCanceled`, `IdentityBindingProvisioningStateCreating`, `IdentityBindingProvisioningStateDeleting`, `IdentityBindingProvisioningStateFailed`, `IdentityBindingProvisioningStateSucceeded`, `IdentityBindingProvisioningStateUpdating`
- New enum type `InfrastructureEncryption` with values `InfrastructureEncryptionDisabled`, `InfrastructureEncryptionEnabled`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `JWTAuthenticatorProvisioningState` with values `JWTAuthenticatorProvisioningStateCanceled`, `JWTAuthenticatorProvisioningStateCreating`, `JWTAuthenticatorProvisioningStateDeleting`, `JWTAuthenticatorProvisioningStateFailed`, `JWTAuthenticatorProvisioningStateSucceeded`, `JWTAuthenticatorProvisioningStateUpdating`
- New enum type `ManagedGatewayType` with values `ManagedGatewayTypeDisabled`, `ManagedGatewayTypeStandard`
- New enum type `MeshMembershipProvisioningState` with values `MeshMembershipProvisioningStateCanceled`, `MeshMembershipProvisioningStateCreating`, `MeshMembershipProvisioningStateDeleting`, `MeshMembershipProvisioningStateFailed`, `MeshMembershipProvisioningStateSucceeded`, `MeshMembershipProvisioningStateUpdating`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`, `ModeNFTABLES`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `ProxyRedirectionMechanism` with values `ProxyRedirectionMechanismCNIChaining`, `ProxyRedirectionMechanismInitContainers`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New enum type `UpgradeStrategy` with values `UpgradeStrategyBlueGreen`, `UpgradeStrategyRolling`
- New enum type `VMState` with values `VMStateDeleted`, `VMStateRunning`
- New function `*AgentPoolsClient.BeginCompleteUpgrade(context.Context, string, string, string, *AgentPoolsClientBeginCompleteUpgradeOptions) (*runtime.Poller[AgentPoolsClientCompleteUpgradeResponse], error)`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewIdentityBindingsClient() *IdentityBindingsClient`
- New function `*ClientFactory.NewJWTAuthenticatorsClient() *JWTAuthenticatorsClient`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewMeshMembershipsClient() *MeshMembershipsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewIdentityBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*IdentityBindingsClient, error)`
- New function `*IdentityBindingsClient.BeginCreateOrUpdate(context.Context, string, string, string, IdentityBinding, *IdentityBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[IdentityBindingsClientCreateOrUpdateResponse], error)`
- New function `*IdentityBindingsClient.BeginDelete(context.Context, string, string, string, *IdentityBindingsClientBeginDeleteOptions) (*runtime.Poller[IdentityBindingsClientDeleteResponse], error)`
- New function `*IdentityBindingsClient.Get(context.Context, string, string, string, *IdentityBindingsClientGetOptions) (IdentityBindingsClientGetResponse, error)`
- New function `*IdentityBindingsClient.NewListByManagedClusterPager(string, string, *IdentityBindingsClientListByManagedClusterOptions) *runtime.Pager[IdentityBindingsClientListByManagedClusterResponse]`
- New function `NewJWTAuthenticatorsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*JWTAuthenticatorsClient, error)`
- New function `*JWTAuthenticatorsClient.BeginCreateOrUpdate(context.Context, string, string, string, JWTAuthenticator, *JWTAuthenticatorsClientBeginCreateOrUpdateOptions) (*runtime.Poller[JWTAuthenticatorsClientCreateOrUpdateResponse], error)`
- New function `*JWTAuthenticatorsClient.BeginDelete(context.Context, string, string, string, *JWTAuthenticatorsClientBeginDeleteOptions) (*runtime.Poller[JWTAuthenticatorsClientDeleteResponse], error)`
- New function `*JWTAuthenticatorsClient.Get(context.Context, string, string, string, *JWTAuthenticatorsClientGetOptions) (JWTAuthenticatorsClientGetResponse, error)`
- New function `*JWTAuthenticatorsClient.NewListByManagedClusterPager(string, string, *JWTAuthenticatorsClientListByManagedClusterOptions) *runtime.Pager[JWTAuthenticatorsClientListByManagedClusterResponse]`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*MachinesClient.BeginCreateOrUpdate(context.Context, string, string, string, string, Machine, *MachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[MachinesClientCreateOrUpdateResponse], error)`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewMeshMembershipsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MeshMembershipsClient, error)`
- New function `*MeshMembershipsClient.BeginCreateOrUpdate(context.Context, string, string, string, MeshMembership, *MeshMembershipsClientBeginCreateOrUpdateOptions) (*runtime.Poller[MeshMembershipsClientCreateOrUpdateResponse], error)`
- New function `*MeshMembershipsClient.BeginDelete(context.Context, string, string, string, *MeshMembershipsClientBeginDeleteOptions) (*runtime.Poller[MeshMembershipsClientDeleteResponse], error)`
- New function `*MeshMembershipsClient.Get(context.Context, string, string, string, *MeshMembershipsClientGetOptions) (MeshMembershipsClientGetResponse, error)`
- New function `*MeshMembershipsClient.NewListByManagedClusterPager(string, string, *MeshMembershipsClientListByManagedClusterOptions) *runtime.Pager[MeshMembershipsClientListByManagedClusterResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworkingPerformance`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolBlueGreenUpgradeSettings`
- New struct `AgentPoolRecentlyUsedVersion`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `IdentityBinding`
- New struct `IdentityBindingListResult`
- New struct `IdentityBindingManagedIdentityProfile`
- New struct `IdentityBindingOidcIssuerProfile`
- New struct `IdentityBindingProperties`
- New struct `JWTAuthenticator`
- New struct `JWTAuthenticatorClaimMappingExpression`
- New struct `JWTAuthenticatorClaimMappings`
- New struct `JWTAuthenticatorExtraClaimMappingExpression`
- New struct `JWTAuthenticatorIssuer`
- New struct `JWTAuthenticatorListResult`
- New struct `JWTAuthenticatorProperties`
- New struct `JWTAuthenticatorValidationRule`
- New struct `KubernetesResourceObjectEncryptionProfile`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `MachineHardwareProfile`
- New struct `MachineKubernetesProfile`
- New struct `MachineOSProfile`
- New struct `MachineOSProfileLinuxProfile`
- New struct `MachineStatus`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterHostedSystemProfile`
- New struct `ManagedClusterIngressDefaultDomainProfile`
- New struct `ManagedClusterIngressProfileApplicationLoadBalancer`
- New struct `ManagedClusterIngressProfileGatewayConfiguration`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `MeshMembership`
- New struct `MeshMembershipProperties`
- New struct `MeshMembershipsListResult`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeCustomizationProfile`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `Performance` in struct `AdvancedNetworking`
- New field `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `ComponentsByReleases`, `RecentlyUsedVersions` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes`, `MinSurge` in struct `AgentPoolUpgradeSettings`
- New field `DriverType` in struct `GPUProfile`
- New field `ProxyRedirectionMechanism` in struct `IstioComponents`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `EnableNodePublicIP`, `NodePublicIPPrefixID`, `NodePublicIPTags`, `PodSubnetID`, `VnetSubnetID` in struct `MachineNetworkProperties`
- New field `ETag`, `Hardware`, `Kubernetes`, `Mode`, `NodeImageVersion`, `OperatingSystem`, `Priority`, `ProvisioningState`, `Security`, `Status`, `Tags` in struct `MachineProperties`
- New field `ArtifactStreamingProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy`, `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `ApplicationLoadBalancer`, `GatewayAPI` in struct `ManagedClusterIngressProfile`
- New field `DefaultDomain` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `CreationData`, `EnableNamespaceResources`, `HostedSystemProfile`, `SchedulerProfile` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `KubernetesResourceObjectEncryptionProfile`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 8.1.0 (2025-11-20)
### Features Added

- New value `WorkloadRuntimeKataVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AdoptionPolicy` with values `AdoptionPolicyAlways`, `AdoptionPolicyIfIdentical`, `AdoptionPolicyNever`
- New enum type `DeletePolicy` with values `DeletePolicyDelete`, `DeletePolicyKeep`
- New enum type `LocalDNSForwardDestination` with values `LocalDNSForwardDestinationClusterCoreDNS`, `LocalDNSForwardDestinationVnetDNS`
- New enum type `LocalDNSForwardPolicy` with values `LocalDNSForwardPolicyRandom`, `LocalDNSForwardPolicyRoundRobin`, `LocalDNSForwardPolicySequential`
- New enum type `LocalDNSMode` with values `LocalDNSModeDisabled`, `LocalDNSModePreferred`, `LocalDNSModeRequired`
- New enum type `LocalDNSProtocol` with values `LocalDNSProtocolForceTCP`, `LocalDNSProtocolPreferUDP`
- New enum type `LocalDNSQueryLogging` with values `LocalDNSQueryLoggingError`, `LocalDNSQueryLoggingLog`
- New enum type `LocalDNSServeStale` with values `LocalDNSServeStaleDisable`, `LocalDNSServeStaleImmediate`, `LocalDNSServeStaleVerify`
- New enum type `LocalDNSState` with values `LocalDNSStateDisabled`, `LocalDNSStateEnabled`
- New enum type `NamespaceProvisioningState` with values `NamespaceProvisioningStateCanceled`, `NamespaceProvisioningStateCreating`, `NamespaceProvisioningStateDeleting`, `NamespaceProvisioningStateFailed`, `NamespaceProvisioningStateSucceeded`, `NamespaceProvisioningStateUpdating`
- New enum type `PolicyRule` with values `PolicyRuleAllowAll`, `PolicyRuleAllowSameNamespace`, `PolicyRuleDenyAll`
- New function `*ClientFactory.NewManagedNamespacesClient() *ManagedNamespacesClient`
- New function `NewManagedNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedNamespacesClient, error)`
- New function `*ManagedNamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, ManagedNamespace, *ManagedNamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[ManagedNamespacesClientCreateOrUpdateResponse], error)`
- New function `*ManagedNamespacesClient.BeginDelete(context.Context, string, string, string, *ManagedNamespacesClientBeginDeleteOptions) (*runtime.Poller[ManagedNamespacesClientDeleteResponse], error)`
- New function `*ManagedNamespacesClient.Get(context.Context, string, string, string, *ManagedNamespacesClientGetOptions) (ManagedNamespacesClientGetResponse, error)`
- New function `*ManagedNamespacesClient.NewListByManagedClusterPager(string, string, *ManagedNamespacesClientListByManagedClusterOptions) *runtime.Pager[ManagedNamespacesClientListByManagedClusterResponse]`
- New function `*ManagedNamespacesClient.ListCredential(context.Context, string, string, string, *ManagedNamespacesClientListCredentialOptions) (ManagedNamespacesClientListCredentialResponse, error)`
- New function `*ManagedNamespacesClient.Update(context.Context, string, string, string, TagsObject, *ManagedNamespacesClientUpdateOptions) (ManagedNamespacesClientUpdateResponse, error)`
- New struct `LocalDNSOverride`
- New struct `LocalDNSProfile`
- New struct `ManagedNamespace`
- New struct `ManagedNamespaceListResult`
- New struct `NamespaceProperties`
- New struct `NetworkPolicies`
- New struct `ResourceQuota`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `LocalDNSProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `LocalDNSProfile` in struct `ManagedClusterAgentPoolProfileProperties`


## 8.1.0-beta.1 (2025-10-23)
### Features Added

- New value `AgentPoolModeMachines`, `AgentPoolModeManagedSystem` added to enum type `AgentPoolMode`
- New value `AgentPoolSSHAccessEntraID` added to enum type `AgentPoolSSHAccess`
- New value `OSSKUFlatcar`, `OSSKUMariner`, `OSSKUUbuntu2404`, `OSSKUWindows2025`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AccelerationMode` with values `AccelerationModeBpfVeth`, `AccelerationModeNone`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AdoptionPolicy` with values `AdoptionPolicyAlways`, `AdoptionPolicyIfIdentical`, `AdoptionPolicyNever`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DeletePolicy` with values `DeletePolicyDelete`, `DeletePolicyKeep`
- New enum type `DriftAction` with values `DriftActionRecreate`, `DriftActionSynced`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IdentityBindingProvisioningState` with values `IdentityBindingProvisioningStateCanceled`, `IdentityBindingProvisioningStateCreating`, `IdentityBindingProvisioningStateDeleting`, `IdentityBindingProvisioningStateFailed`, `IdentityBindingProvisioningStateSucceeded`, `IdentityBindingProvisioningStateUpdating`
- New enum type `InfrastructureEncryption` with values `InfrastructureEncryptionDisabled`, `InfrastructureEncryptionEnabled`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `JWTAuthenticatorProvisioningState` with values `JWTAuthenticatorProvisioningStateCanceled`, `JWTAuthenticatorProvisioningStateCreating`, `JWTAuthenticatorProvisioningStateDeleting`, `JWTAuthenticatorProvisioningStateFailed`, `JWTAuthenticatorProvisioningStateSucceeded`, `JWTAuthenticatorProvisioningStateUpdating`
- New enum type `LocalDNSForwardDestination` with values `LocalDNSForwardDestinationClusterCoreDNS`, `LocalDNSForwardDestinationVnetDNS`
- New enum type `LocalDNSForwardPolicy` with values `LocalDNSForwardPolicyRandom`, `LocalDNSForwardPolicyRoundRobin`, `LocalDNSForwardPolicySequential`
- New enum type `LocalDNSMode` with values `LocalDNSModeDisabled`, `LocalDNSModePreferred`, `LocalDNSModeRequired`
- New enum type `LocalDNSProtocol` with values `LocalDNSProtocolForceTCP`, `LocalDNSProtocolPreferUDP`
- New enum type `LocalDNSQueryLogging` with values `LocalDNSQueryLoggingError`, `LocalDNSQueryLoggingLog`
- New enum type `LocalDNSServeStale` with values `LocalDNSServeStaleDisable`, `LocalDNSServeStaleImmediate`, `LocalDNSServeStaleVerify`
- New enum type `LocalDNSState` with values `LocalDNSStateDisabled`, `LocalDNSStateEnabled`
- New enum type `ManagedGatewayType` with values `ManagedGatewayTypeDisabled`, `ManagedGatewayTypeStandard`
- New enum type `MeshMembershipProvisioningState` with values `MeshMembershipProvisioningStateCanceled`, `MeshMembershipProvisioningStateCreating`, `MeshMembershipProvisioningStateDeleting`, `MeshMembershipProvisioningStateFailed`, `MeshMembershipProvisioningStateSucceeded`, `MeshMembershipProvisioningStateUpdating`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NamespaceProvisioningState` with values `NamespaceProvisioningStateCanceled`, `NamespaceProvisioningStateCreating`, `NamespaceProvisioningStateDeleting`, `NamespaceProvisioningStateFailed`, `NamespaceProvisioningStateSucceeded`, `NamespaceProvisioningStateUpdating`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `PolicyRule` with values `PolicyRuleAllowAll`, `PolicyRuleAllowSameNamespace`, `PolicyRuleDenyAll`
- New enum type `ProxyRedirectionMechanism` with values `ProxyRedirectionMechanismCNIChaining`, `ProxyRedirectionMechanismInitContainers`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New enum type `UpgradeStrategy` with values `UpgradeStrategyBlueGreen`, `UpgradeStrategyRolling`
- New enum type `VMState` with values `VMStateDeleted`, `VMStateRunning`
- New function `*AgentPoolsClient.BeginCompleteUpgrade(context.Context, string, string, string, *AgentPoolsClientBeginCompleteUpgradeOptions) (*runtime.Poller[AgentPoolsClientCompleteUpgradeResponse], error)`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewIdentityBindingsClient() *IdentityBindingsClient`
- New function `*ClientFactory.NewJWTAuthenticatorsClient() *JWTAuthenticatorsClient`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewManagedNamespacesClient() *ManagedNamespacesClient`
- New function `*ClientFactory.NewMeshMembershipsClient() *MeshMembershipsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewIdentityBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*IdentityBindingsClient, error)`
- New function `*IdentityBindingsClient.BeginCreateOrUpdate(context.Context, string, string, string, IdentityBinding, *IdentityBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[IdentityBindingsClientCreateOrUpdateResponse], error)`
- New function `*IdentityBindingsClient.BeginDelete(context.Context, string, string, string, *IdentityBindingsClientBeginDeleteOptions) (*runtime.Poller[IdentityBindingsClientDeleteResponse], error)`
- New function `*IdentityBindingsClient.Get(context.Context, string, string, string, *IdentityBindingsClientGetOptions) (IdentityBindingsClientGetResponse, error)`
- New function `*IdentityBindingsClient.NewListByManagedClusterPager(string, string, *IdentityBindingsClientListByManagedClusterOptions) *runtime.Pager[IdentityBindingsClientListByManagedClusterResponse]`
- New function `NewJWTAuthenticatorsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*JWTAuthenticatorsClient, error)`
- New function `*JWTAuthenticatorsClient.BeginCreateOrUpdate(context.Context, string, string, string, JWTAuthenticator, *JWTAuthenticatorsClientBeginCreateOrUpdateOptions) (*runtime.Poller[JWTAuthenticatorsClientCreateOrUpdateResponse], error)`
- New function `*JWTAuthenticatorsClient.BeginDelete(context.Context, string, string, string, *JWTAuthenticatorsClientBeginDeleteOptions) (*runtime.Poller[JWTAuthenticatorsClientDeleteResponse], error)`
- New function `*JWTAuthenticatorsClient.Get(context.Context, string, string, string, *JWTAuthenticatorsClientGetOptions) (JWTAuthenticatorsClientGetResponse, error)`
- New function `*JWTAuthenticatorsClient.NewListByManagedClusterPager(string, string, *JWTAuthenticatorsClientListByManagedClusterOptions) *runtime.Pager[JWTAuthenticatorsClientListByManagedClusterResponse]`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*MachinesClient.BeginCreateOrUpdate(context.Context, string, string, string, string, Machine, *MachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[MachinesClientCreateOrUpdateResponse], error)`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewManagedNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedNamespacesClient, error)`
- New function `*ManagedNamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, ManagedNamespace, *ManagedNamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[ManagedNamespacesClientCreateOrUpdateResponse], error)`
- New function `*ManagedNamespacesClient.BeginDelete(context.Context, string, string, string, *ManagedNamespacesClientBeginDeleteOptions) (*runtime.Poller[ManagedNamespacesClientDeleteResponse], error)`
- New function `*ManagedNamespacesClient.Get(context.Context, string, string, string, *ManagedNamespacesClientGetOptions) (ManagedNamespacesClientGetResponse, error)`
- New function `*ManagedNamespacesClient.NewListByManagedClusterPager(string, string, *ManagedNamespacesClientListByManagedClusterOptions) *runtime.Pager[ManagedNamespacesClientListByManagedClusterResponse]`
- New function `*ManagedNamespacesClient.ListCredential(context.Context, string, string, string, *ManagedNamespacesClientListCredentialOptions) (ManagedNamespacesClientListCredentialResponse, error)`
- New function `*ManagedNamespacesClient.Update(context.Context, string, string, string, TagsObject, *ManagedNamespacesClientUpdateOptions) (ManagedNamespacesClientUpdateResponse, error)`
- New function `NewMeshMembershipsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MeshMembershipsClient, error)`
- New function `*MeshMembershipsClient.BeginCreateOrUpdate(context.Context, string, string, string, MeshMembership, *MeshMembershipsClientBeginCreateOrUpdateOptions) (*runtime.Poller[MeshMembershipsClientCreateOrUpdateResponse], error)`
- New function `*MeshMembershipsClient.BeginDelete(context.Context, string, string, string, *MeshMembershipsClientBeginDeleteOptions) (*runtime.Poller[MeshMembershipsClientDeleteResponse], error)`
- New function `*MeshMembershipsClient.Get(context.Context, string, string, string, *MeshMembershipsClientGetOptions) (MeshMembershipsClientGetResponse, error)`
- New function `*MeshMembershipsClient.NewListByManagedClusterPager(string, string, *MeshMembershipsClientListByManagedClusterOptions) *runtime.Pager[MeshMembershipsClientListByManagedClusterResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworkingPerformance`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolBlueGreenUpgradeSettings`
- New struct `AgentPoolRecentlyUsedVersion`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `IdentityBinding`
- New struct `IdentityBindingListResult`
- New struct `IdentityBindingManagedIdentityProfile`
- New struct `IdentityBindingOidcIssuerProfile`
- New struct `IdentityBindingProperties`
- New struct `JWTAuthenticator`
- New struct `JWTAuthenticatorClaimMappingExpression`
- New struct `JWTAuthenticatorClaimMappings`
- New struct `JWTAuthenticatorExtraClaimMappingExpression`
- New struct `JWTAuthenticatorIssuer`
- New struct `JWTAuthenticatorListResult`
- New struct `JWTAuthenticatorProperties`
- New struct `JWTAuthenticatorValidationRule`
- New struct `KubernetesResourceObjectEncryptionProfile`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `LocalDNSOverride`
- New struct `LocalDNSProfile`
- New struct `MachineHardwareProfile`
- New struct `MachineKubernetesProfile`
- New struct `MachineOSProfile`
- New struct `MachineOSProfileLinuxProfile`
- New struct `MachineStatus`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterHostedSystemProfile`
- New struct `ManagedClusterIngressProfileGatewayConfiguration`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedNamespace`
- New struct `ManagedNamespaceListResult`
- New struct `MeshMembership`
- New struct `MeshMembershipProperties`
- New struct `MeshMembershipsListResult`
- New struct `NamespaceProperties`
- New struct `NetworkPolicies`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeCustomizationProfile`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `ResourceQuota`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `Performance` in struct `AdvancedNetworking`
- New field `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `ComponentsByReleases`, `RecentlyUsedVersions` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes`, `MinSurge` in struct `AgentPoolUpgradeSettings`
- New field `DriverType` in struct `GPUProfile`
- New field `ProxyRedirectionMechanism` in struct `IstioComponents`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `EnableNodePublicIP`, `NodePublicIPPrefixID`, `NodePublicIPTags`, `PodSubnetID`, `VnetSubnetID` in struct `MachineNetworkProperties`
- New field `ETag`, `Hardware`, `Kubernetes`, `Mode`, `NodeImageVersion`, `OperatingSystem`, `Priority`, `ProvisioningState`, `Security`, `Status`, `Tags` in struct `MachineProperties`
- New field `ArtifactStreamingProfile`, `LocalDNSProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `LocalDNSProfile`, `NodeCustomizationProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy`, `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `GatewayAPI` in struct `ManagedClusterIngressProfile`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `CreationData`, `EnableNamespaceResources`, `HostedSystemProfile`, `SchedulerProfile` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `KubernetesResourceObjectEncryptionProfile`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 8.0.0 (2025-10-23)
### Breaking Changes

- Type of `AgentPoolStatus.ProvisioningError` has been changed from `*CloudErrorBody` to `*ErrorDetail`
- Type of `ManagedClusterStatus.ProvisioningError` has been changed from `*CloudErrorBody` to `*ErrorDetail`
- Struct `CloudErrorBody` has been removed

### Features Added

- New enum type `AdvancedNetworkPolicies` with values `AdvancedNetworkPoliciesFQDN`, `AdvancedNetworkPoliciesL7`, `AdvancedNetworkPoliciesNone`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New field `AdvancedNetworkPolicies` in struct `AdvancedNetworkingSecurity`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`


## 8.0.0-beta.1 (2025-09-25)
### Breaking Changes

- Type of `AgentPoolStatus.ProvisioningError` has been changed from `*CloudErrorBody` to `*ErrorDetail`
- Type of `ManagedClusterStatus.ProvisioningError` has been changed from `*CloudErrorBody` to `*ErrorDetail`
- Struct `CloudErrorBody` has been removed

### Features Added

- New value `AgentPoolModeMachines`, `AgentPoolModeManagedSystem` added to enum type `AgentPoolMode`
- New value `OSSKUMariner`, `OSSKUUbuntu2404`, `OSSKUWindows2025`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AccelerationMode` with values `AccelerationModeBpfVeth`, `AccelerationModeNone`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AdoptionPolicy` with values `AdoptionPolicyAlways`, `AdoptionPolicyIfIdentical`, `AdoptionPolicyNever`
- New enum type `AdvancedNetworkPolicies` with values `AdvancedNetworkPoliciesFQDN`, `AdvancedNetworkPoliciesL7`, `AdvancedNetworkPoliciesNone`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DeletePolicy` with values `DeletePolicyDelete`, `DeletePolicyKeep`
- New enum type `DriftAction` with values `DriftActionRecreate`, `DriftActionSynced`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IdentityBindingProvisioningState` with values `IdentityBindingProvisioningStateCanceled`, `IdentityBindingProvisioningStateCreating`, `IdentityBindingProvisioningStateDeleting`, `IdentityBindingProvisioningStateFailed`, `IdentityBindingProvisioningStateSucceeded`, `IdentityBindingProvisioningStateUpdating`
- New enum type `InfrastructureEncryption` with values `InfrastructureEncryptionDisabled`, `InfrastructureEncryptionEnabled`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `JWTAuthenticatorProvisioningState` with values `JWTAuthenticatorProvisioningStateCanceled`, `JWTAuthenticatorProvisioningStateCreating`, `JWTAuthenticatorProvisioningStateDeleting`, `JWTAuthenticatorProvisioningStateFailed`, `JWTAuthenticatorProvisioningStateSucceeded`, `JWTAuthenticatorProvisioningStateUpdating`
- New enum type `LocalDNSForwardDestination` with values `LocalDNSForwardDestinationClusterCoreDNS`, `LocalDNSForwardDestinationVnetDNS`
- New enum type `LocalDNSForwardPolicy` with values `LocalDNSForwardPolicyRandom`, `LocalDNSForwardPolicyRoundRobin`, `LocalDNSForwardPolicySequential`
- New enum type `LocalDNSMode` with values `LocalDNSModeDisabled`, `LocalDNSModePreferred`, `LocalDNSModeRequired`
- New enum type `LocalDNSProtocol` with values `LocalDNSProtocolForceTCP`, `LocalDNSProtocolPreferUDP`
- New enum type `LocalDNSQueryLogging` with values `LocalDNSQueryLoggingError`, `LocalDNSQueryLoggingLog`
- New enum type `LocalDNSServeStale` with values `LocalDNSServeStaleDisable`, `LocalDNSServeStaleImmediate`, `LocalDNSServeStaleVerify`
- New enum type `LocalDNSState` with values `LocalDNSStateDisabled`, `LocalDNSStateEnabled`
- New enum type `ManagedGatewayType` with values `ManagedGatewayTypeDisabled`, `ManagedGatewayTypeStandard`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NamespaceProvisioningState` with values `NamespaceProvisioningStateCanceled`, `NamespaceProvisioningStateCreating`, `NamespaceProvisioningStateDeleting`, `NamespaceProvisioningStateFailed`, `NamespaceProvisioningStateSucceeded`, `NamespaceProvisioningStateUpdating`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `PolicyRule` with values `PolicyRuleAllowAll`, `PolicyRuleAllowSameNamespace`, `PolicyRuleDenyAll`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New enum type `UpgradeStrategy` with values `UpgradeStrategyBlueGreen`, `UpgradeStrategyRolling`
- New enum type `VMState` with values `VMStateDeleted`, `VMStateRunning`
- New function `*AgentPoolsClient.BeginCompleteUpgrade(context.Context, string, string, string, *AgentPoolsClientBeginCompleteUpgradeOptions) (*runtime.Poller[AgentPoolsClientCompleteUpgradeResponse], error)`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewIdentityBindingsClient() *IdentityBindingsClient`
- New function `*ClientFactory.NewJWTAuthenticatorsClient() *JWTAuthenticatorsClient`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewManagedNamespacesClient() *ManagedNamespacesClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewIdentityBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*IdentityBindingsClient, error)`
- New function `*IdentityBindingsClient.BeginCreateOrUpdate(context.Context, string, string, string, IdentityBinding, *IdentityBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[IdentityBindingsClientCreateOrUpdateResponse], error)`
- New function `*IdentityBindingsClient.BeginDelete(context.Context, string, string, string, *IdentityBindingsClientBeginDeleteOptions) (*runtime.Poller[IdentityBindingsClientDeleteResponse], error)`
- New function `*IdentityBindingsClient.Get(context.Context, string, string, string, *IdentityBindingsClientGetOptions) (IdentityBindingsClientGetResponse, error)`
- New function `*IdentityBindingsClient.NewListByManagedClusterPager(string, string, *IdentityBindingsClientListByManagedClusterOptions) *runtime.Pager[IdentityBindingsClientListByManagedClusterResponse]`
- New function `NewJWTAuthenticatorsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*JWTAuthenticatorsClient, error)`
- New function `*JWTAuthenticatorsClient.BeginCreateOrUpdate(context.Context, string, string, string, JWTAuthenticator, *JWTAuthenticatorsClientBeginCreateOrUpdateOptions) (*runtime.Poller[JWTAuthenticatorsClientCreateOrUpdateResponse], error)`
- New function `*JWTAuthenticatorsClient.BeginDelete(context.Context, string, string, string, *JWTAuthenticatorsClientBeginDeleteOptions) (*runtime.Poller[JWTAuthenticatorsClientDeleteResponse], error)`
- New function `*JWTAuthenticatorsClient.Get(context.Context, string, string, string, *JWTAuthenticatorsClientGetOptions) (JWTAuthenticatorsClientGetResponse, error)`
- New function `*JWTAuthenticatorsClient.NewListByManagedClusterPager(string, string, *JWTAuthenticatorsClientListByManagedClusterOptions) *runtime.Pager[JWTAuthenticatorsClientListByManagedClusterResponse]`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*MachinesClient.BeginCreateOrUpdate(context.Context, string, string, string, string, Machine, *MachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[MachinesClientCreateOrUpdateResponse], error)`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewManagedNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedNamespacesClient, error)`
- New function `*ManagedNamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, ManagedNamespace, *ManagedNamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[ManagedNamespacesClientCreateOrUpdateResponse], error)`
- New function `*ManagedNamespacesClient.BeginDelete(context.Context, string, string, string, *ManagedNamespacesClientBeginDeleteOptions) (*runtime.Poller[ManagedNamespacesClientDeleteResponse], error)`
- New function `*ManagedNamespacesClient.Get(context.Context, string, string, string, *ManagedNamespacesClientGetOptions) (ManagedNamespacesClientGetResponse, error)`
- New function `*ManagedNamespacesClient.NewListByManagedClusterPager(string, string, *ManagedNamespacesClientListByManagedClusterOptions) *runtime.Pager[ManagedNamespacesClientListByManagedClusterResponse]`
- New function `*ManagedNamespacesClient.ListCredential(context.Context, string, string, string, *ManagedNamespacesClientListCredentialOptions) (ManagedNamespacesClientListCredentialResponse, error)`
- New function `*ManagedNamespacesClient.Update(context.Context, string, string, string, TagsObject, *ManagedNamespacesClientUpdateOptions) (ManagedNamespacesClientUpdateResponse, error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworkingPerformance`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolBlueGreenUpgradeSettings`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `IdentityBinding`
- New struct `IdentityBindingListResult`
- New struct `IdentityBindingManagedIdentityProfile`
- New struct `IdentityBindingOidcIssuerProfile`
- New struct `IdentityBindingProperties`
- New struct `JWTAuthenticator`
- New struct `JWTAuthenticatorClaimMappingExpression`
- New struct `JWTAuthenticatorClaimMappings`
- New struct `JWTAuthenticatorExtraClaimMappingExpression`
- New struct `JWTAuthenticatorIssuer`
- New struct `JWTAuthenticatorListResult`
- New struct `JWTAuthenticatorProperties`
- New struct `JWTAuthenticatorValidationRule`
- New struct `KubernetesResourceObjectEncryptionProfile`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `LocalDNSOverride`
- New struct `LocalDNSProfile`
- New struct `MachineHardwareProfile`
- New struct `MachineKubernetesProfile`
- New struct `MachineOSProfile`
- New struct `MachineOSProfileLinuxProfile`
- New struct `MachineStatus`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterIngressProfileGatewayConfiguration`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedNamespace`
- New struct `ManagedNamespaceListResult`
- New struct `NamespaceProperties`
- New struct `NetworkPolicies`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `ResourceQuota`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `Performance` in struct `AdvancedNetworking`
- New field `AdvancedNetworkPolicies`, `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes`, `MinSurge` in struct `AgentPoolUpgradeSettings`
- New field `DriverType` in struct `GPUProfile`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `EnableNodePublicIP`, `NodePublicIPPrefixID`, `NodePublicIPTags`, `PodSubnetID`, `VnetSubnetID` in struct `MachineNetworkProperties`
- New field `ETag`, `Hardware`, `Kubernetes`, `Mode`, `NodeImageVersion`, `OperatingSystem`, `Priority`, `ProvisioningState`, `Security`, `Status`, `Tags` in struct `MachineProperties`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `LocalDNSProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `LocalDNSProfile`, `NodeInitializationTaints`, `UpgradeSettingsBlueGreen`, `UpgradeStrategy` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy`, `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `GatewayAPI` in struct `ManagedClusterIngressProfile`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `CreationData`, `EnableNamespaceResources`, `SchedulerProfile` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `KubernetesResourceObjectEncryptionProfile`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 7.3.0 (2025-09-25)
### Features Added

- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUAzureLinux3` added to enum type `OSSKU`
- New field `Kind` in struct `ManagedCluster`


## 7.3.0-beta.1 (2025-07-17)
### Features Added

- New value `AgentPoolModeMachines`, `AgentPoolModeManagedSystem` added to enum type `AgentPoolMode`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUUbuntu2404`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AdoptionPolicy` with values `AdoptionPolicyAlways`, `AdoptionPolicyIfIdentical`, `AdoptionPolicyNever`
- New enum type `AdvancedNetworkPolicies` with values `AdvancedNetworkPoliciesFQDN`, `AdvancedNetworkPoliciesL7`, `AdvancedNetworkPoliciesNone`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DeletePolicy` with values `DeletePolicyDelete`, `DeletePolicyKeep`
- New enum type `DriftAction` with values `DriftActionRecreate`, `DriftActionSynced`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `LocalDNSForwardDestination` with values `LocalDNSForwardDestinationClusterCoreDNS`, `LocalDNSForwardDestinationVnetDNS`
- New enum type `LocalDNSForwardPolicy` with values `LocalDNSForwardPolicyRandom`, `LocalDNSForwardPolicyRoundRobin`, `LocalDNSForwardPolicySequential`
- New enum type `LocalDNSMode` with values `LocalDNSModeDisabled`, `LocalDNSModePreferred`, `LocalDNSModeRequired`
- New enum type `LocalDNSProtocol` with values `LocalDNSProtocolForceTCP`, `LocalDNSProtocolPreferUDP`
- New enum type `LocalDNSQueryLogging` with values `LocalDNSQueryLoggingError`, `LocalDNSQueryLoggingLog`
- New enum type `LocalDNSServeStale` with values `LocalDNSServeStaleDisable`, `LocalDNSServeStaleImmediate`, `LocalDNSServeStaleVerify`
- New enum type `LocalDNSState` with values `LocalDNSStateDisabled`, `LocalDNSStateEnabled`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NamespaceProvisioningState` with values `NamespaceProvisioningStateCanceled`, `NamespaceProvisioningStateCreating`, `NamespaceProvisioningStateDeleting`, `NamespaceProvisioningStateFailed`, `NamespaceProvisioningStateSucceeded`, `NamespaceProvisioningStateUpdating`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `PolicyRule` with values `PolicyRuleAllowAll`, `PolicyRuleAllowSameNamespace`, `PolicyRuleDenyAll`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New enum type `VMState` with values `VMStateDeleted`, `VMStateRunning`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewManagedNamespacesClient() *ManagedNamespacesClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*MachinesClient.BeginCreateOrUpdate(context.Context, string, string, string, string, Machine, *MachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[MachinesClientCreateOrUpdateResponse], error)`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewManagedNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedNamespacesClient, error)`
- New function `*ManagedNamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, ManagedNamespace, *ManagedNamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[ManagedNamespacesClientCreateOrUpdateResponse], error)`
- New function `*ManagedNamespacesClient.BeginDelete(context.Context, string, string, string, *ManagedNamespacesClientBeginDeleteOptions) (*runtime.Poller[ManagedNamespacesClientDeleteResponse], error)`
- New function `*ManagedNamespacesClient.Get(context.Context, string, string, string, *ManagedNamespacesClientGetOptions) (ManagedNamespacesClientGetResponse, error)`
- New function `*ManagedNamespacesClient.NewListByManagedClusterPager(string, string, *ManagedNamespacesClientListByManagedClusterOptions) *runtime.Pager[ManagedNamespacesClientListByManagedClusterResponse]`
- New function `*ManagedNamespacesClient.ListCredential(context.Context, string, string, string, *ManagedNamespacesClientListCredentialOptions) (ManagedNamespacesClientListCredentialResponse, error)`
- New function `*ManagedNamespacesClient.Update(context.Context, string, string, string, TagsObject, *ManagedNamespacesClientUpdateOptions) (ManagedNamespacesClientUpdateResponse, error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `LocalDNSOverride`
- New struct `LocalDNSProfile`
- New struct `MachineHardwareProfile`
- New struct `MachineKubernetesProfile`
- New struct `MachineOSProfile`
- New struct `MachineOSProfileLinuxProfile`
- New struct `MachineStatus`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedNamespace`
- New struct `ManagedNamespaceListResult`
- New struct `NamespaceProperties`
- New struct `NetworkPolicies`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `ResourceQuota`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `AdvancedNetworkPolicies`, `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes` in struct `AgentPoolUpgradeSettings`
- New field `DriverType` in struct `GPUProfile`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `EnableNodePublicIP`, `NodePublicIPPrefixID`, `NodePublicIPTags`, `PodSubnetID`, `VnetSubnetID` in struct `MachineNetworkProperties`
- New field `ETag`, `Hardware`, `Kubernetes`, `Mode`, `NodeImageVersion`, `OperatingSystem`, `Priority`, `ProvisioningState`, `Security`, `Status`, `Tags` in struct `MachineProperties`
- New field `Kind` in struct `ManagedCluster`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `LocalDNSProfile`, `NodeInitializationTaints` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `LocalDNSProfile`, `NodeInitializationTaints` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy`, `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `CreationData`, `EnableNamespaceResources`, `SchedulerProfile` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 7.2.0 (2025-07-17)
### Features Added

- New enum type `NodeProvisioningDefaultNodePools` with values `NodeProvisioningDefaultNodePoolsAuto`, `NodeProvisioningDefaultNodePoolsNone`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New field `AiToolchainOperatorProfile`, `NodeProvisioningProfile` in struct `ManagedClusterProperties`


## 7.1.0 (2025-06-18)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New struct `AgentPoolGatewayProfile`
- New struct `AgentPoolStatus`
- New struct `CloudErrorBody`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManagedClusterStatus`
- New struct `ManualScaleProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `Zones` in struct `Machine`
- New field `GatewayProfile`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `GatewayProfile`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Status` in struct `ManagedClusterProperties`
- New field `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 7.0.0 (2025-06-16)
### Breaking Changes

- Field `EnablePodSecurityPolicy` of struct `ManagedClusterProperties` has been removed

### Features Added

- New value `OSSKUUbuntu2204` added to enum type `OSSKU`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New field `MaxUnavailable`, `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`


## 7.0.0-beta.3 (2025-05-29)
### Breaking Changes

- Type of `AdvancedNetworkingSecurity.TransitEncryption` has been changed from `*TransitEncryption` to `*AdvancedNetworkingSecurityTransitEncryption`
- Type of `ManagedClusterAgentPoolProfile.GpuProfile` has been changed from `*AgentPoolGPUProfile` to `*GPUProfile`
- Type of `ManagedClusterAgentPoolProfileProperties.GpuProfile` has been changed from `*AgentPoolGPUProfile` to `*GPUProfile`
- Enum `TransitEncryption` has been removed
- Function `*ClientFactory.NewNamespacesClient` has been removed
- Function `NewNamespacesClient` has been removed
- Function `*NamespacesClient.BeginCreateOrUpdate` has been removed
- Function `*NamespacesClient.BeginDelete` has been removed
- Function `*NamespacesClient.Get` has been removed
- Function `*NamespacesClient.NewListByManagedClusterPager` has been removed
- Function `*NamespacesClient.ListCredential` has been removed
- Function `*NamespacesClient.Update` has been removed
- Struct `AgentPoolGPUProfile` has been removed
- Struct `Namespace` has been removed
- Struct `NamespaceListResult` has been removed
- Field `EnablePodSecurityPolicy` of struct `ManagedClusterProperties` has been removed

### Features Added

- New value `OSSKUUbuntu2204`, `OSSKUUbuntu2404` added to enum type `OSSKU`
- New enum type `GPUDriver` with values `GPUDriverInstall`, `GPUDriverNone`
- New enum type `LocalDNSForwardDestination` with values `LocalDNSForwardDestinationClusterCoreDNS`, `LocalDNSForwardDestinationVnetDNS`
- New enum type `LocalDNSForwardPolicy` with values `LocalDNSForwardPolicyRandom`, `LocalDNSForwardPolicyRoundRobin`, `LocalDNSForwardPolicySequential`
- New enum type `LocalDNSMode` with values `LocalDNSModeDisabled`, `LocalDNSModePreferred`, `LocalDNSModeRequired`
- New enum type `LocalDNSProtocol` with values `LocalDNSProtocolForceTCP`, `LocalDNSProtocolPreferUDP`
- New enum type `LocalDNSQueryLogging` with values `LocalDNSQueryLoggingError`, `LocalDNSQueryLoggingLog`
- New enum type `LocalDNSServeStale` with values `LocalDNSServeStaleDisable`, `LocalDNSServeStaleImmediate`, `LocalDNSServeStaleVerify`
- New enum type `LocalDNSState` with values `LocalDNSStateDisabled`, `LocalDNSStateEnabled`
- New enum type `NodeProvisioningDefaultNodePools` with values `NodeProvisioningDefaultNodePoolsAuto`, `NodeProvisioningDefaultNodePoolsNone`
- New enum type `SchedulerConfigMode` with values `SchedulerConfigModeDefault`, `SchedulerConfigModeManagedByCRD`
- New enum type `TransitEncryptionType` with values `TransitEncryptionTypeNone`, `TransitEncryptionTypeWireGuard`
- New function `*ClientFactory.NewManagedNamespacesClient() *ManagedNamespacesClient`
- New function `NewManagedNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedNamespacesClient, error)`
- New function `*ManagedNamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, ManagedNamespace, *ManagedNamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[ManagedNamespacesClientCreateOrUpdateResponse], error)`
- New function `*ManagedNamespacesClient.BeginDelete(context.Context, string, string, string, *ManagedNamespacesClientBeginDeleteOptions) (*runtime.Poller[ManagedNamespacesClientDeleteResponse], error)`
- New function `*ManagedNamespacesClient.Get(context.Context, string, string, string, *ManagedNamespacesClientGetOptions) (ManagedNamespacesClientGetResponse, error)`
- New function `*ManagedNamespacesClient.NewListByManagedClusterPager(string, string, *ManagedNamespacesClientListByManagedClusterOptions) *runtime.Pager[ManagedNamespacesClientListByManagedClusterResponse]`
- New function `*ManagedNamespacesClient.ListCredential(context.Context, string, string, string, *ManagedNamespacesClientListCredentialOptions) (ManagedNamespacesClientListCredentialResponse, error)`
- New function `*ManagedNamespacesClient.Update(context.Context, string, string, string, TagsObject, *ManagedNamespacesClientUpdateOptions) (ManagedNamespacesClientUpdateResponse, error)`
- New struct `AdvancedNetworkingSecurityTransitEncryption`
- New struct `GPUProfile`
- New struct `LocalDNSOverrides`
- New struct `LocalDNSProfile`
- New struct `ManagedNamespace`
- New struct `ManagedNamespaceListResult`
- New struct `SchedulerInstanceProfile`
- New struct `SchedulerProfile`
- New struct `SchedulerProfileSchedulerInstanceProfiles`
- New field `LocalDNSProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `LocalDNSProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Enabled` in struct `ManagedClusterHTTPProxyConfig`
- New field `DefaultNodePools` in struct `ManagedClusterNodeProvisioningProfile`
- New field `SchedulerProfile` in struct `ManagedClusterProperties`


## 7.0.0-beta.2 (2025-04-24)
### Breaking Changes

- Struct `AutoScaleProfile` has been removed
- Field `OSDiskSizeGB`, `OSDiskType`, `Sizes` of struct `ManualScaleProfile` has been removed

### Features Added

- Type of `ScaleProfile.Autoscale` has been changed from `[]*AutoScaleProfile` to `any`
- New enum type `AdoptionPolicy` with values `AdoptionPolicyAlways`, `AdoptionPolicyIfIdentical`, `AdoptionPolicyNever`
- New enum type `AdvancedNetworkPolicies` with values `AdvancedNetworkPoliciesFQDN`, `AdvancedNetworkPoliciesL7`, `AdvancedNetworkPoliciesNone`
- New enum type `DeletePolicy` with values `DeletePolicyDelete`, `DeletePolicyKeep`
- New enum type `NamespaceProvisioningState` with values `NamespaceProvisioningStateCanceled`, `NamespaceProvisioningStateCreating`, `NamespaceProvisioningStateDeleting`, `NamespaceProvisioningStateFailed`, `NamespaceProvisioningStateSucceeded`, `NamespaceProvisioningStateUpdating`
- New enum type `PolicyRule` with values `PolicyRuleAllowAll`, `PolicyRuleAllowSameNamespace`, `PolicyRuleDenyAll`
- New enum type `TransitEncryption` with values `TransitEncryptionNone`, `TransitEncryptionWireGuard`
- New function `*ClientFactory.NewNamespacesClient() *NamespacesClient`
- New function `NewNamespacesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*NamespacesClient, error)`
- New function `*NamespacesClient.BeginCreateOrUpdate(context.Context, string, string, string, Namespace, *NamespacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[NamespacesClientCreateOrUpdateResponse], error)`
- New function `*NamespacesClient.BeginDelete(context.Context, string, string, string, *NamespacesClientBeginDeleteOptions) (*runtime.Poller[NamespacesClientDeleteResponse], error)`
- New function `*NamespacesClient.Get(context.Context, string, string, string, *NamespacesClientGetOptions) (NamespacesClientGetResponse, error)`
- New function `*NamespacesClient.NewListByManagedClusterPager(string, string, *NamespacesClientListByManagedClusterOptions) *runtime.Pager[NamespacesClientListByManagedClusterResponse]`
- New function `*NamespacesClient.ListCredential(context.Context, string, string, string, *NamespacesClientListCredentialOptions) (NamespacesClientListCredentialResponse, error)`
- New function `*NamespacesClient.Update(context.Context, string, string, string, TagsObject, *NamespacesClientUpdateOptions) (NamespacesClientUpdateResponse, error)`
- New struct `Namespace`
- New struct `NamespaceListResult`
- New struct `NamespaceProperties`
- New struct `NetworkPolicies`
- New struct `ResourceQuota`
- New field `AdvancedNetworkPolicies`, `TransitEncryption` in struct `AdvancedNetworkingSecurity`
- New field `Size` in struct `ManualScaleProfile`


## 6.6.0 (2025-04-22)
### Features Added

- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New struct `ManagedClusterIngressProfileNginx`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`


## 7.0.0-beta.1 (2025-04-15)
### Breaking Changes

- Type of `ManagedClusterAgentPoolProfile.GpuProfile` has been changed from `*GPUProfile` to `*AgentPoolGPUProfile`
- Type of `ManagedClusterAgentPoolProfileProperties.GpuProfile` has been changed from `*GPUProfile` to `*AgentPoolGPUProfile`
- Enum `GPUDriver` has been removed
- Struct `GPUProfile` has been removed

### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AgentPoolStatus`
- New struct `AutoScaleProfile`
- New struct `CloudErrorBody`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterIngressProfileNginx`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGating`
- New struct `ManagedClusterSecurityProfileDefenderSecurityGatingIdentitiesItem`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManagedClusterStatus`
- New struct `ManualScaleProfile`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxBlockedNodes`, `MaxUnavailable`, `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `NodeInitializationTaints`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `NodeInitializationTaints`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `AiToolchainOperatorProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `SafeguardsProfile`, `Status` in struct `ManagedClusterProperties`
- New field `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `SecurityGating` in struct `ManagedClusterSecurityProfileDefender`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 6.5.0 (2025-03-26)
### Features Added

- New value `OutboundTypeNone` added to enum type `OutboundType`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `GPUDriver` with values `GPUDriverInstall`, `GPUDriverNone`
- New struct `GPUProfile`
- New struct `ManagedClusterBootstrapProfile`
- New field `GpuProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `GpuProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `BootstrapProfile` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates` in struct `ManagedClusterSecurityProfile`


## 6.5.0-beta.1 (2025-03-11)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `OutboundTypeNone` added to enum type `OutboundType`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `NewClient(string, azcore.TokenCredential, *arm.ClientOptions) (*Client, error)`
- New function `*Client.NewListNodeImageVersionsPager(string, *ClientListNodeImageVersionsOptions) *runtime.Pager[ClientListNodeImageVersionsResponse]`
- New function `*ClientFactory.NewClient() *Client`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AgentPoolStatus`
- New struct `AutoScaleProfile`
- New struct `CloudErrorBody`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterIngressProfileNginx`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManagedClusterStatus`
- New struct `ManualScaleProfile`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `NodeImageVersion`
- New struct `NodeImageVersionsListResult`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `IsOutOfSupport` in struct `AgentPoolUpgradeProfilePropertiesUpgradesItem`
- New field `MaxUnavailable`, `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `GatewayConfigurationName`, `Name`, `Namespace` in struct `IstioEgressGateway`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `NodeInitializationTaints`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `NodeInitializationTaints`, `PodIPAllocationMode`, `Status`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IsOutOfSupport` in struct `ManagedClusterPoolUpgradeProfileUpgradesItem`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `SafeguardsProfile`, `Status` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 6.4.0 (2025-02-27)
### Features Added

- New field `MessageOfTheDay` in struct `ManagedClusterAgentPoolProfile`
- New field `MessageOfTheDay` in struct `ManagedClusterAgentPoolProfileProperties`


## 6.4.0-beta.1 (2024-12-09)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `OutboundTypeNone` added to enum type `OutboundType`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `SeccompDefault` with values `SeccompDefaultRuntimeDefault`, `SeccompDefaultUnconfined`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterIngressProfileNginx`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManualScaleProfile`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `MaxUnavailable`, `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `SeccompDefault` in struct `KubeletConfig`
- New field `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `PodLinkLocalAccess`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 6.3.0 (2024-11-18)
### Features Added

- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New struct `AdvancedNetworking`
- New struct `AdvancedNetworkingObservability`
- New struct `AdvancedNetworkingSecurity`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New field `IfMatch`, `IfNoneMatch` in struct `AgentPoolsClientBeginCreateOrUpdateOptions`
- New field `IfMatch` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `ETag` in struct `ManagedCluster`
- New field `ETag` in struct `ManagedClusterAgentPoolProfile`
- New field `ETag` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `NodeResourceGroupProfile` in struct `ManagedClusterProperties`
- New field `IfMatch`, `IfNoneMatch` in struct `ManagedClustersClientBeginCreateOrUpdateOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginUpdateTagsOptions`
- New field `AdvancedNetworking` in struct `NetworkProfile`


## 6.2.0 (2024-10-24)
### Features Added

- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`


## 6.2.0-beta.1 (2024-09-29)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `OutboundTypeNone` added to enum type `OutboundType`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `DriverType` with values `DriverTypeCUDA`, `DriverTypeGRID`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `TLSManagement` with values `TLSManagementManaged`, `TLSManagementNone`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworking`
- New struct `AdvancedNetworkingFQDNPolicy`
- New struct `AdvancedNetworkingObservability`
- New struct `AdvancedNetworkingSecurity`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterIngressProfileNginx`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManualScaleProfile`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `SSHAccess` in struct `AgentPoolSecurityProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `IfMatch`, `IfNoneMatch` in struct `AgentPoolsClientBeginCreateOrUpdateOptions`
- New field `IfMatch`, `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `ETag`, `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `ETag`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `ETag`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `NodeResourceGroupProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IfMatch`, `IfNoneMatch` in struct `ManagedClustersClientBeginCreateOrUpdateOptions`
- New field `IfMatch`, `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginUpdateTagsOptions`
- New field `AdvancedNetworking`, `KubeProxyConfig`, `PodLinkLocalAccess`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 6.1.0 (2024-09-27)
### Features Added

- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolSecurityProfile`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New field `SecurityProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `SecurityProfile` in struct `ManagedClusterAgentPoolProfileProperties`


## 6.1.0-beta.2 (2024-09-26)
### Features Added

- New enum type `NginxIngressControllerType` with values `NginxIngressControllerTypeAnnotationControlled`, `NginxIngressControllerTypeExternal`, `NginxIngressControllerTypeInternal`, `NginxIngressControllerTypeNone`
- New enum type `TLSManagement` with values `TLSManagementManaged`, `TLSManagementNone`
- New struct `AdvancedNetworkingFQDNPolicy`
- New struct `AdvancedNetworkingSecurity`
- New struct `ManagedClusterIngressProfileNginx`
- New field `Security` in struct `AdvancedNetworking`
- New field `TLSManagement` in struct `AdvancedNetworkingObservability`
- New field `Nginx` in struct `ManagedClusterIngressProfileWebAppRouting`


## 6.1.0-beta.1 (2024-08-22)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `OutboundTypeNone` added to enum type `OutboundType`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AdvancedNetworking`
- New struct `AdvancedNetworkingObservability`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManualScaleProfile`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `RebalanceLoadBalancersRequestBody`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `IfMatch`, `IfNoneMatch` in struct `AgentPoolsClientBeginCreateOrUpdateOptions`
- New field `IfMatch`, `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `ETag`, `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `ETag`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `ETag`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `NodeResourceGroupProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IfMatch`, `IfNoneMatch` in struct `ManagedClustersClientBeginCreateOrUpdateOptions`
- New field `IfMatch`, `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginUpdateTagsOptions`
- New field `AdvancedNetworking`, `KubeProxyConfig`, `PodLinkLocalAccess`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 6.0.0 (2024-07-26)
### Breaking Changes

- Function `*ManagedClustersClient.GetOSOptions` has been removed
- Struct `OSOptionProfile` has been removed
- Struct `OSOptionProperty` has been removed
- Struct `OSOptionPropertyList` has been removed

### Features Added

- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`


## 5.1.0-beta.2 (2024-07-25)
### Breaking Changes

- Struct `NetworkMonitoring` has been removed
- Field `Monitoring` of struct `NetworkProfile` has been removed

### Features Added

- New value `ManagedClusterSKUNameAutomatic` added to enum type `ManagedClusterSKUName`
- New value `OutboundTypeNone` added to enum type `OutboundType`
- New enum type `Operator` with values `OperatorDoesNotExist`, `OperatorExists`, `OperatorIn`, `OperatorNotIn`
- New enum type `PodLinkLocalAccess` with values `PodLinkLocalAccessIMDS`, `PodLinkLocalAccessNone`
- New enum type `UndrainableNodeBehavior` with values `UndrainableNodeBehaviorCordon`, `UndrainableNodeBehaviorSchedule`
- New function `*ClientFactory.NewLoadBalancersClient() *LoadBalancersClient`
- New function `NewLoadBalancersClient(string, azcore.TokenCredential, *arm.ClientOptions) (*LoadBalancersClient, error)`
- New function `*LoadBalancersClient.CreateOrUpdate(context.Context, string, string, string, LoadBalancer, *LoadBalancersClientCreateOrUpdateOptions) (LoadBalancersClientCreateOrUpdateResponse, error)`
- New function `*LoadBalancersClient.BeginDelete(context.Context, string, string, string, *LoadBalancersClientBeginDeleteOptions) (*runtime.Poller[LoadBalancersClientDeleteResponse], error)`
- New function `*LoadBalancersClient.Get(context.Context, string, string, string, *LoadBalancersClientGetOptions) (LoadBalancersClientGetResponse, error)`
- New function `*LoadBalancersClient.NewListByManagedClusterPager(string, string, *LoadBalancersClientListByManagedClusterOptions) *runtime.Pager[LoadBalancersClientListByManagedClusterResponse]`
- New function `*ManagedClustersClient.BeginRebalanceLoadBalancers(context.Context, string, string, RebalanceLoadBalancersRequestBody, *ManagedClustersClientBeginRebalanceLoadBalancersOptions) (*runtime.Poller[ManagedClustersClientRebalanceLoadBalancersResponse], error)`
- New struct `AdvancedNetworking`
- New struct `AdvancedNetworkingObservability`
- New struct `AutoScaleProfile`
- New struct `Component`
- New struct `ComponentsByRelease`
- New struct `LabelSelector`
- New struct `LabelSelectorRequirement`
- New struct `LoadBalancer`
- New struct `LoadBalancerListResult`
- New struct `LoadBalancerProperties`
- New struct `RebalanceLoadBalancersRequestBody`
- New field `ComponentsByReleases` in struct `AgentPoolUpgradeProfileProperties`
- New field `UndrainableNodeBehavior` in struct `AgentPoolUpgradeSettings`
- New field `IfMatch`, `IfNoneMatch` in struct `AgentPoolsClientBeginCreateOrUpdateOptions`
- New field `IfMatch` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `ETag` in struct `ManagedCluster`
- New field `ETag` in struct `ManagedClusterAgentPoolProfile`
- New field `ETag` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `ComponentsByReleases` in struct `ManagedClusterPoolUpgradeProfile`
- New field `IfMatch`, `IfNoneMatch` in struct `ManagedClustersClientBeginCreateOrUpdateOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `IfMatch` in struct `ManagedClustersClientBeginUpdateTagsOptions`
- New field `AdvancedNetworking`, `PodLinkLocalAccess` in struct `NetworkProfile`
- New field `Autoscale` in struct `ScaleProfile`


## 5.1.0-beta.1 (2024-04-26)
### Features Added

- New value `AgentPoolModeGateway` added to enum type `AgentPoolMode`
- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `ClusterServiceLoadBalancerHealthProbeMode` with values `ClusterServiceLoadBalancerHealthProbeModeServiceNodePort`, `ClusterServiceLoadBalancerHealthProbeModeShared`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolGatewayProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringAutoInstrumentation`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryLogs`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManagedClusterStaticEgressGatewayProfile`
- New struct `ManualScaleProfile`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `Kind` in struct `ManagedCluster`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GatewayProfile`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `AppMonitoring`, `ContainerInsights` in struct `ManagedClusterAzureMonitorProfile`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `ClusterServiceLoadBalancerHealthProbeMode` in struct `ManagedClusterLoadBalancerProfile`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `NodeProvisioningProfile`, `NodeResourceGroupProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring`, `StaticEgressGatewayProfile` in struct `NetworkProfile`


## 5.0.0 (2024-04-26)
### Breaking Changes

- Field `NodeSelector` of struct `IstioEgressGateway` has been removed

### Features Added

- New struct `AgentPoolWindowsProfile`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterMetricsProfile`
- New field `IsDefault` in struct `KubernetesVersion`
- New field `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `MetricsProfile` in struct `ManagedClusterProperties`


## 4.9.0-beta.1 (2024-03-22)
### Features Added

- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `ArtifactSource` with values `ArtifactSourceCache`, `ArtifactSourceDirect`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `PodIPAllocationMode` with values `PodIPAllocationModeDynamicIndividual`, `PodIPAllocationModeStaticBlock`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterBootstrapProfile`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManualScaleProfile`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `PodIPAllocationMode`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `AiToolchainOperatorProfile`, `BootstrapProfile`, `CreationData`, `EnableNamespaceResources`, `MetricsProfile`, `NodeProvisioningProfile`, `NodeResourceGroupProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.8.0 (2024-03-22)
### Features Added

- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New field `IngressProfile` in struct `ManagedClusterProperties`


## 4.8.0-beta.1 (2024-02-23)
### Features Added

- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `SafeguardsSupport` with values `SafeguardsSupportPreview`, `SafeguardsSupportStable`
- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetSafeguardsVersions(context.Context, string, string, *ManagedClustersClientGetSafeguardsVersionsOptions) (ManagedClustersClientGetSafeguardsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListSafeguardsVersionsPager(string, *ManagedClustersClientListSafeguardsVersionsOptions) *runtime.Pager[ManagedClustersClientListSafeguardsVersionsResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManualScaleProfile`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `SafeguardsAvailableVersion`
- New struct `SafeguardsAvailableVersionsList`
- New struct `SafeguardsAvailableVersionsProperties`
- New struct `SafeguardsProfile`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `AiToolchainOperatorProfile`, `CreationData`, `EnableNamespaceResources`, `IngressProfile`, `MetricsProfile`, `NodeProvisioningProfile`, `NodeResourceGroupProfile`, `SafeguardsProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.7.0 (2024-01-26)
### Features Added

- New field `NodeSoakDurationInMinutes` in struct `AgentPoolUpgradeSettings`


## 4.7.0-beta.1 (2023-12-22)
### Features Added

- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New function `*AgentPoolsClient.BeginDeleteMachines(context.Context, string, string, string, AgentPoolDeleteMachinesParameter, *AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[AgentPoolsClientDeleteMachinesResponse], error)`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewOperationStatusResultClient() *OperationStatusResultClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `NewOperationStatusResultClient(string, azcore.TokenCredential, *arm.ClientOptions) (*OperationStatusResultClient, error)`
- New function `*OperationStatusResultClient.Get(context.Context, string, string, string, *OperationStatusResultClientGetOptions) (OperationStatusResultClientGetResponse, error)`
- New function `*OperationStatusResultClient.GetByAgentPool(context.Context, string, string, string, string, *OperationStatusResultClientGetByAgentPoolOptions) (OperationStatusResultClientGetByAgentPoolResponse, error)`
- New function `*OperationStatusResultClient.NewListPager(string, string, *OperationStatusResultClientListOptions) *runtime.Pager[OperationStatusResultClientListResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolDeleteMachinesParameter`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `GuardrailsProfile`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `ManualScaleProfile`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `OperationStatusResult`
- New struct `OperationStatusResultList`
- New struct `ScaleProfile`
- New struct `VirtualMachineNodes`
- New struct `VirtualMachinesProfile`
- New field `NodeSoakDurationInMinutes` in struct `AgentPoolUpgradeSettings`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NodeInitializationTaints`, `SecurityProfile`, `VirtualMachineNodesStatus`, `VirtualMachinesProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `AiToolchainOperatorProfile`, `CreationData`, `EnableNamespaceResources`, `GuardrailsProfile`, `IngressProfile`, `MetricsProfile`, `NodeProvisioningProfile`, `NodeResourceGroupProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.6.0 (2023-11-24)
### Features Added

- New enum type `BackendPoolType` with values `BackendPoolTypeNodeIP`, `BackendPoolTypeNodeIPConfiguration`
- New enum type `Protocol` with values `ProtocolTCP`, `ProtocolUDP`
- New struct `AgentPoolNetworkProfile`
- New struct `IPTag`
- New struct `PortRange`
- New field `CapacityReservationGroupID`, `NetworkProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `CapacityReservationGroupID`, `NetworkProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `BackendPoolType` in struct `ManagedClusterLoadBalancerProfile`


## 4.6.0-beta.1 (2023-11-24)
### Features Added

- New value `AgentPoolTypeVirtualMachines` added to enum type `AgentPoolType`
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner`, `OSSKUWindowsAnnual` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `BackendPoolType` with values `BackendPoolTypeNodeIP`, `BackendPoolTypeNodeIPConfiguration`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `NodeProvisioningMode` with values `NodeProvisioningModeAuto`, `NodeProvisioningModeManual`
- New enum type `Protocol` with values `ProtocolTCP`, `ProtocolUDP`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New struct `AgentPoolArtifactStreamingProfile`
- New struct `AgentPoolGPUProfile`
- New struct `AgentPoolNetworkProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `GuardrailsProfile`
- New struct `IPTag`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAIToolchainOperatorProfile`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeProvisioningProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `PortRange`
- New field `NodeSoakDurationInMinutes` in struct `AgentPoolUpgradeSettings`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `ArtifactStreamingProfile`, `CapacityReservationGroupID`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `ArtifactStreamingProfile`, `CapacityReservationGroupID`, `EnableCustomCATrust`, `GpuProfile`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `BackendPoolType` in struct `ManagedClusterLoadBalancerProfile`
- New field `AiToolchainOperatorProfile`, `CreationData`, `EnableNamespaceResources`, `GuardrailsProfile`, `IngressProfile`, `MetricsProfile`, `NodeProvisioningProfile`, `NodeResourceGroupProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `Expanders`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.5.0 (2023-11-24)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.
- New enum type `TrustedAccessRoleBindingProvisioningState` with values `TrustedAccessRoleBindingProvisioningStateCanceled`, `TrustedAccessRoleBindingProvisioningStateDeleting`, `TrustedAccessRoleBindingProvisioningStateFailed`, `TrustedAccessRoleBindingProvisioningStateSucceeded`, `TrustedAccessRoleBindingProvisioningStateUpdating`
- New function `*ClientFactory.NewTrustedAccessRoleBindingsClient() *TrustedAccessRoleBindingsClient`
- New function `*ClientFactory.NewTrustedAccessRolesClient() *TrustedAccessRolesClient`
- New function `NewTrustedAccessRoleBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRoleBindingsClient, error)`
- New function `*TrustedAccessRoleBindingsClient.BeginCreateOrUpdate(context.Context, string, string, string, TrustedAccessRoleBinding, *TrustedAccessRoleBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[TrustedAccessRoleBindingsClientCreateOrUpdateResponse], error)`
- New function `*TrustedAccessRoleBindingsClient.BeginDelete(context.Context, string, string, string, *TrustedAccessRoleBindingsClientBeginDeleteOptions) (*runtime.Poller[TrustedAccessRoleBindingsClientDeleteResponse], error)`
- New function `*TrustedAccessRoleBindingsClient.Get(context.Context, string, string, string, *TrustedAccessRoleBindingsClientGetOptions) (TrustedAccessRoleBindingsClientGetResponse, error)`
- New function `*TrustedAccessRoleBindingsClient.NewListPager(string, string, *TrustedAccessRoleBindingsClientListOptions) *runtime.Pager[TrustedAccessRoleBindingsClientListResponse]`
- New function `NewTrustedAccessRolesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRolesClient, error)`
- New function `*TrustedAccessRolesClient.NewListPager(string, *TrustedAccessRolesClientListOptions) *runtime.Pager[TrustedAccessRolesClientListResponse]`
- New struct `TrustedAccessRole`
- New struct `TrustedAccessRoleBinding`
- New struct `TrustedAccessRoleBindingListResult`
- New struct `TrustedAccessRoleBindingProperties`
- New struct `TrustedAccessRoleListResult`
- New struct `TrustedAccessRoleRule`


## 4.5.0-beta.1 (2023-10-27)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.
- New value `NetworkPolicyNone` added to enum type `NetworkPolicy`
- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AddonAutoscaling` with values `AddonAutoscalingDisabled`, `AddonAutoscalingEnabled`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `BackendPoolType` with values `BackendPoolTypeNodeIP`, `BackendPoolTypeNodeIPConfiguration`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `Protocol` with values `ProtocolTCP`, `ProtocolUDP`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `TrustedAccessRoleBindingProvisioningState` with values `TrustedAccessRoleBindingProvisioningStateCanceled`, `TrustedAccessRoleBindingProvisioningStateDeleting`, `TrustedAccessRoleBindingProvisioningStateFailed`, `TrustedAccessRoleBindingProvisioningStateSucceeded`, `TrustedAccessRoleBindingProvisioningStateUpdating`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewTrustedAccessRoleBindingsClient() *TrustedAccessRoleBindingsClient`
- New function `*ClientFactory.NewTrustedAccessRolesClient() *TrustedAccessRolesClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `NewTrustedAccessRoleBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRoleBindingsClient, error)`
- New function `*TrustedAccessRoleBindingsClient.BeginCreateOrUpdate(context.Context, string, string, string, TrustedAccessRoleBinding, *TrustedAccessRoleBindingsClientBeginCreateOrUpdateOptions) (*runtime.Poller[TrustedAccessRoleBindingsClientCreateOrUpdateResponse], error)`
- New function `*TrustedAccessRoleBindingsClient.BeginDelete(context.Context, string, string, string, *TrustedAccessRoleBindingsClientBeginDeleteOptions) (*runtime.Poller[TrustedAccessRoleBindingsClientDeleteResponse], error)`
- New function `*TrustedAccessRoleBindingsClient.Get(context.Context, string, string, string, *TrustedAccessRoleBindingsClientGetOptions) (TrustedAccessRoleBindingsClientGetResponse, error)`
- New function `*TrustedAccessRoleBindingsClient.NewListPager(string, string, *TrustedAccessRoleBindingsClientListOptions) *runtime.Pager[TrustedAccessRoleBindingsClientListResponse]`
- New function `NewTrustedAccessRolesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRolesClient, error)`
- New function `*TrustedAccessRolesClient.NewListPager(string, *TrustedAccessRolesClientListOptions) *runtime.Pager[TrustedAccessRolesClientListResponse]`
- New struct `AgentPoolNetworkProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `GuardrailsProfile`
- New struct `IPTag`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `PortRange`
- New struct `TrustedAccessRole`
- New struct `TrustedAccessRoleBinding`
- New struct `TrustedAccessRoleBindingListResult`
- New struct `TrustedAccessRoleBindingProperties`
- New struct `TrustedAccessRoleListResult`
- New struct `TrustedAccessRoleRule`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `CapacityReservationGroupID`, `EnableCustomCATrust`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `CapacityReservationGroupID`, `EnableCustomCATrust`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `BackendPoolType` in struct `ManagedClusterLoadBalancerProfile`
- New field `CreationData`, `EnableNamespaceResources`, `GuardrailsProfile`, `IngressProfile`, `MetricsProfile`, `NodeResourceGroupProfile` in struct `ManagedClusterProperties`
- New field `DaemonsetEvictionForEmptyNodes`, `DaemonsetEvictionForOccupiedNodes`, `Expanders`, `IgnoreDaemonsetsUtilization` in struct `ManagedClusterPropertiesAutoScalerProfile`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `AddonAutoscaling` in struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.4.0 (2023-10-27)
### Features Added

- New enum type `IstioIngressGatewayMode` with values `IstioIngressGatewayModeExternal`, `IstioIngressGatewayModeInternal`
- New enum type `ServiceMeshMode` with values `ServiceMeshModeDisabled`, `ServiceMeshModeIstio`
- New function `*ManagedClustersClient.GetMeshRevisionProfile(context.Context, string, string, *ManagedClustersClientGetMeshRevisionProfileOptions) (ManagedClustersClientGetMeshRevisionProfileResponse, error)`
- New function `*ManagedClustersClient.GetMeshUpgradeProfile(context.Context, string, string, string, *ManagedClustersClientGetMeshUpgradeProfileOptions) (ManagedClustersClientGetMeshUpgradeProfileResponse, error)`
- New function `*ManagedClustersClient.NewListMeshRevisionProfilesPager(string, *ManagedClustersClientListMeshRevisionProfilesOptions) *runtime.Pager[ManagedClustersClientListMeshRevisionProfilesResponse]`
- New function `*ManagedClustersClient.NewListMeshUpgradeProfilesPager(string, string, *ManagedClustersClientListMeshUpgradeProfilesOptions) *runtime.Pager[ManagedClustersClientListMeshUpgradeProfilesResponse]`
- New struct `CompatibleVersions`
- New struct `IstioCertificateAuthority`
- New struct `IstioComponents`
- New struct `IstioEgressGateway`
- New struct `IstioIngressGateway`
- New struct `IstioPluginCertificateAuthority`
- New struct `IstioServiceMesh`
- New struct `MeshRevision`
- New struct `MeshRevisionProfile`
- New struct `MeshRevisionProfileList`
- New struct `MeshRevisionProfileProperties`
- New struct `MeshUpgradeProfile`
- New struct `MeshUpgradeProfileList`
- New struct `MeshUpgradeProfileProperties`
- New struct `ServiceMeshProfile`
- New field `ResourceUID`, `ServiceMeshProfile` in struct `ManagedClusterProperties`


## 4.4.0-beta.2 (2023-10-09)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.

## 4.4.0-beta.1 (2023-09-22)
### Features Added

- New value `NodeOSUpgradeChannelSecurityPatch` added to enum type `NodeOSUpgradeChannel`
- New value `OSSKUMariner` added to enum type `OSSKU`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New value `SnapshotTypeManagedCluster` added to enum type `SnapshotType`
- New value `WorkloadRuntimeKataMshvVMIsolation` added to enum type `WorkloadRuntime`
- New enum type `AgentPoolSSHAccess` with values `AgentPoolSSHAccessDisabled`, `AgentPoolSSHAccessLocalUser`
- New enum type `BackendPoolType` with values `BackendPoolTypeNodeIP`, `BackendPoolTypeNodeIPConfiguration`
- New enum type `GuardrailsSupport` with values `GuardrailsSupportPreview`, `GuardrailsSupportStable`
- New enum type `IpvsScheduler` with values `IpvsSchedulerLeastConnection`, `IpvsSchedulerRoundRobin`
- New enum type `IstioIngressGatewayMode` with values `IstioIngressGatewayModeExternal`, `IstioIngressGatewayModeInternal`
- New enum type `Level` with values `LevelEnforcement`, `LevelOff`, `LevelWarning`
- New enum type `Mode` with values `ModeIPTABLES`, `ModeIPVS`
- New enum type `Protocol` with values `ProtocolTCP`, `ProtocolUDP`
- New enum type `RestrictionLevel` with values `RestrictionLevelReadOnly`, `RestrictionLevelUnrestricted`
- New enum type `ServiceMeshMode` with values `ServiceMeshModeDisabled`, `ServiceMeshModeIstio`
- New enum type `TrustedAccessRoleBindingProvisioningState` with values `TrustedAccessRoleBindingProvisioningStateCanceled`, `TrustedAccessRoleBindingProvisioningStateDeleting`, `TrustedAccessRoleBindingProvisioningStateFailed`, `TrustedAccessRoleBindingProvisioningStateSucceeded`, `TrustedAccessRoleBindingProvisioningStateUpdating`
- New function `*ClientFactory.NewMachinesClient() *MachinesClient`
- New function `*ClientFactory.NewManagedClusterSnapshotsClient() *ManagedClusterSnapshotsClient`
- New function `*ClientFactory.NewTrustedAccessRoleBindingsClient() *TrustedAccessRoleBindingsClient`
- New function `*ClientFactory.NewTrustedAccessRolesClient() *TrustedAccessRolesClient`
- New function `NewMachinesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*MachinesClient, error)`
- New function `*MachinesClient.Get(context.Context, string, string, string, string, *MachinesClientGetOptions) (MachinesClientGetResponse, error)`
- New function `*MachinesClient.NewListPager(string, string, string, *MachinesClientListOptions) *runtime.Pager[MachinesClientListResponse]`
- New function `NewManagedClusterSnapshotsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*ManagedClusterSnapshotsClient, error)`
- New function `*ManagedClusterSnapshotsClient.CreateOrUpdate(context.Context, string, string, ManagedClusterSnapshot, *ManagedClusterSnapshotsClientCreateOrUpdateOptions) (ManagedClusterSnapshotsClientCreateOrUpdateResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Delete(context.Context, string, string, *ManagedClusterSnapshotsClientDeleteOptions) (ManagedClusterSnapshotsClientDeleteResponse, error)`
- New function `*ManagedClusterSnapshotsClient.Get(context.Context, string, string, *ManagedClusterSnapshotsClientGetOptions) (ManagedClusterSnapshotsClientGetResponse, error)`
- New function `*ManagedClusterSnapshotsClient.NewListByResourceGroupPager(string, *ManagedClusterSnapshotsClientListByResourceGroupOptions) *runtime.Pager[ManagedClusterSnapshotsClientListByResourceGroupResponse]`
- New function `*ManagedClusterSnapshotsClient.NewListPager(*ManagedClusterSnapshotsClientListOptions) *runtime.Pager[ManagedClusterSnapshotsClientListResponse]`
- New function `*ManagedClusterSnapshotsClient.UpdateTags(context.Context, string, string, TagsObject, *ManagedClusterSnapshotsClientUpdateTagsOptions) (ManagedClusterSnapshotsClientUpdateTagsResponse, error)`
- New function `*ManagedClustersClient.GetGuardrailsVersions(context.Context, string, string, *ManagedClustersClientGetGuardrailsVersionsOptions) (ManagedClustersClientGetGuardrailsVersionsResponse, error)`
- New function `*ManagedClustersClient.GetMeshRevisionProfile(context.Context, string, string, *ManagedClustersClientGetMeshRevisionProfileOptions) (ManagedClustersClientGetMeshRevisionProfileResponse, error)`
- New function `*ManagedClustersClient.GetMeshUpgradeProfile(context.Context, string, string, string, *ManagedClustersClientGetMeshUpgradeProfileOptions) (ManagedClustersClientGetMeshUpgradeProfileResponse, error)`
- New function `*ManagedClustersClient.NewListGuardrailsVersionsPager(string, *ManagedClustersClientListGuardrailsVersionsOptions) *runtime.Pager[ManagedClustersClientListGuardrailsVersionsResponse]`
- New function `*ManagedClustersClient.NewListMeshRevisionProfilesPager(string, *ManagedClustersClientListMeshRevisionProfilesOptions) *runtime.Pager[ManagedClustersClientListMeshRevisionProfilesResponse]`
- New function `*ManagedClustersClient.NewListMeshUpgradeProfilesPager(string, string, *ManagedClustersClientListMeshUpgradeProfilesOptions) *runtime.Pager[ManagedClustersClientListMeshUpgradeProfilesResponse]`
- New function `NewTrustedAccessRoleBindingsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRoleBindingsClient, error)`
- New function `*TrustedAccessRoleBindingsClient.CreateOrUpdate(context.Context, string, string, string, TrustedAccessRoleBinding, *TrustedAccessRoleBindingsClientCreateOrUpdateOptions) (TrustedAccessRoleBindingsClientCreateOrUpdateResponse, error)`
- New function `*TrustedAccessRoleBindingsClient.Delete(context.Context, string, string, string, *TrustedAccessRoleBindingsClientDeleteOptions) (TrustedAccessRoleBindingsClientDeleteResponse, error)`
- New function `*TrustedAccessRoleBindingsClient.Get(context.Context, string, string, string, *TrustedAccessRoleBindingsClientGetOptions) (TrustedAccessRoleBindingsClientGetResponse, error)`
- New function `*TrustedAccessRoleBindingsClient.NewListPager(string, string, *TrustedAccessRoleBindingsClientListOptions) *runtime.Pager[TrustedAccessRoleBindingsClientListResponse]`
- New function `NewTrustedAccessRolesClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TrustedAccessRolesClient, error)`
- New function `*TrustedAccessRolesClient.NewListPager(string, *TrustedAccessRolesClientListOptions) *runtime.Pager[TrustedAccessRolesClientListResponse]`
- New struct `AgentPoolNetworkProfile`
- New struct `AgentPoolSecurityProfile`
- New struct `AgentPoolWindowsProfile`
- New struct `CompatibleVersions`
- New struct `GuardrailsAvailableVersion`
- New struct `GuardrailsAvailableVersionsList`
- New struct `GuardrailsAvailableVersionsProperties`
- New struct `GuardrailsProfile`
- New struct `IPTag`
- New struct `IstioCertificateAuthority`
- New struct `IstioComponents`
- New struct `IstioIngressGateway`
- New struct `IstioPluginCertificateAuthority`
- New struct `IstioServiceMesh`
- New struct `Machine`
- New struct `MachineIPAddress`
- New struct `MachineListResult`
- New struct `MachineNetworkProperties`
- New struct `MachineProperties`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoring`
- New struct `ManagedClusterAzureMonitorProfileAppMonitoringOpenTelemetryMetrics`
- New struct `ManagedClusterAzureMonitorProfileContainerInsights`
- New struct `ManagedClusterAzureMonitorProfileLogs`
- New struct `ManagedClusterAzureMonitorProfileWindowsHostLogs`
- New struct `ManagedClusterCostAnalysis`
- New struct `ManagedClusterIngressProfile`
- New struct `ManagedClusterIngressProfileWebAppRouting`
- New struct `ManagedClusterMetricsProfile`
- New struct `ManagedClusterNodeResourceGroupProfile`
- New struct `ManagedClusterPropertiesForSnapshot`
- New struct `ManagedClusterSecurityProfileImageIntegrity`
- New struct `ManagedClusterSecurityProfileNodeRestriction`
- New struct `ManagedClusterSnapshot`
- New struct `ManagedClusterSnapshotListResult`
- New struct `ManagedClusterSnapshotProperties`
- New struct `MeshRevision`
- New struct `MeshRevisionProfile`
- New struct `MeshRevisionProfileList`
- New struct `MeshRevisionProfileProperties`
- New struct `MeshUpgradeProfile`
- New struct `MeshUpgradeProfileList`
- New struct `MeshUpgradeProfileProperties`
- New struct `NetworkMonitoring`
- New struct `NetworkProfileForSnapshot`
- New struct `NetworkProfileKubeProxyConfig`
- New struct `NetworkProfileKubeProxyConfigIpvsConfig`
- New struct `PortRange`
- New struct `ServiceMeshProfile`
- New struct `TrustedAccessRole`
- New struct `TrustedAccessRoleBinding`
- New struct `TrustedAccessRoleBindingListResult`
- New struct `TrustedAccessRoleBindingProperties`
- New struct `TrustedAccessRoleListResult`
- New struct `TrustedAccessRoleRule`
- New field `IgnorePodDisruptionBudget` in struct `AgentPoolsClientBeginDeleteOptions`
- New field `EnableVnetIntegration`, `SubnetID` in struct `ManagedClusterAPIServerAccessProfile`
- New field `CapacityReservationGroupID`, `EnableCustomCATrust`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfile`
- New field `CapacityReservationGroupID`, `EnableCustomCATrust`, `MessageOfTheDay`, `NetworkProfile`, `SecurityProfile`, `WindowsProfile` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `Logs` in struct `ManagedClusterAzureMonitorProfile`
- New field `AppMonitoringOpenTelemetryMetrics` in struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `EffectiveNoProxy` in struct `ManagedClusterHTTPProxyConfig`
- New field `BackendPoolType` in struct `ManagedClusterLoadBalancerProfile`
- New field `CreationData`, `EnableNamespaceResources`, `GuardrailsProfile`, `IngressProfile`, `MetricsProfile`, `NodeResourceGroupProfile`, `ResourceUID`, `ServiceMeshProfile` in struct `ManagedClusterProperties`
- New field `CustomCATrustCertificates`, `ImageIntegrity`, `NodeRestriction` in struct `ManagedClusterSecurityProfile`
- New field `Version` in struct `ManagedClusterStorageProfileDiskCSIDriver`
- New field `IgnorePodDisruptionBudget` in struct `ManagedClustersClientBeginDeleteOptions`
- New field `KubeProxyConfig`, `Monitoring` in struct `NetworkProfile`


## 4.3.0 (2023-08-25)
### Features Added

- New struct `ClusterUpgradeSettings`
- New struct `UpgradeOverrideSettings`
- New field `UpgradeSettings` in struct `ManagedClusterProperties`


## 4.2.0 (2023-08-25)
### Features Added

- New enum type `NodeOSUpgradeChannel` with values `NodeOSUpgradeChannelNodeImage`, `NodeOSUpgradeChannelNone`, `NodeOSUpgradeChannelUnmanaged`
- New struct `DelegatedResource`
- New struct `ManagedClusterWorkloadAutoScalerProfileVerticalPodAutoscaler`
- New field `DrainTimeoutInMinutes` in struct `AgentPoolUpgradeSettings`
- New field `NodeOSUpgradeChannel` in struct `ManagedClusterAutoUpgradeProfile`
- New field `DelegatedResources` in struct `ManagedClusterIdentity`
- New field `VerticalPodAutoscaler` in struct `ManagedClusterWorkloadAutoScalerProfile`


## 4.1.0 (2023-07-28)
### Features Added

- New enum type `Type` with values `TypeFirst`, `TypeFourth`, `TypeLast`, `TypeSecond`, `TypeThird`
- New struct `AbsoluteMonthlySchedule`
- New struct `DailySchedule`
- New struct `DateSpan`
- New struct `MaintenanceWindow`
- New struct `RelativeMonthlySchedule`
- New struct `Schedule`
- New struct `WeeklySchedule`
- New field `MaintenanceWindow` in struct `MaintenanceConfigurationProperties`


## 4.0.0 (2023-05-26)
### Breaking Changes

- Field `DockerBridgeCidr` of struct `NetworkProfile` has been removed

### Features Added

- New value `OSSKUAzureLinux` added to enum type `OSSKU`


## 3.0.0 (2023-04-28)
### Breaking Changes

- Const `ManagedClusterSKUNameBasic` from type alias `ManagedClusterSKUName` has been removed
- Const `ManagedClusterSKUTierPaid` from type alias `ManagedClusterSKUTier` has been removed

### Features Added

- New value `ManagedClusterSKUTierPremium` added to enum type `ManagedClusterSKUTier`
- New value `NetworkPolicyCilium` added to enum type `NetworkPolicy`
- New enum type `KubernetesSupportPlan` with values `KubernetesSupportPlanAKSLongTermSupport`, `KubernetesSupportPlanKubernetesOfficial`
- New enum type `NetworkDataplane` with values `NetworkDataplaneAzure`, `NetworkDataplaneCilium`
- New enum type `NetworkPluginMode` with values `NetworkPluginModeOverlay`
- New function `*ManagedClustersClient.ListKubernetesVersions(context.Context, string, *ManagedClustersClientListKubernetesVersionsOptions) (ManagedClustersClientListKubernetesVersionsResponse, error)`
- New struct `KubernetesPatchVersion`
- New struct `KubernetesVersion`
- New struct `KubernetesVersionCapabilities`
- New struct `KubernetesVersionListResult`
- New struct `ManagedClusterSecurityProfileImageCleaner`
- New struct `ManagedClusterSecurityProfileWorkloadIdentity`
- New field `SupportPlan` in struct `ManagedClusterProperties`
- New field `ImageCleaner` in struct `ManagedClusterSecurityProfile`
- New field `WorkloadIdentity` in struct `ManagedClusterSecurityProfile`
- New field `NetworkDataplane` in struct `NetworkProfile`
- New field `NetworkPluginMode` in struct `NetworkProfile`


## 2.4.0 (2023-03-24)
### Features Added

- New struct `ClientFactory` which is a client factory used to create any client in this module
- New value `ManagedClusterSKUNameBase` added to enum type `ManagedClusterSKUName`
- New value `ManagedClusterSKUTierStandard` added to enum type `ManagedClusterSKUTier`
- New function `*AgentPoolsClient.BeginAbortLatestOperation(context.Context, string, string, string, *AgentPoolsClientBeginAbortLatestOperationOptions) (*runtime.Poller[AgentPoolsClientAbortLatestOperationResponse], error)`
- New function `*ManagedClustersClient.BeginAbortLatestOperation(context.Context, string, string, *ManagedClustersClientBeginAbortLatestOperationOptions) (*runtime.Poller[ManagedClustersClientAbortLatestOperationResponse], error)`
- New struct `ManagedClusterAzureMonitorProfile`
- New struct `ManagedClusterAzureMonitorProfileKubeStateMetrics`
- New struct `ManagedClusterAzureMonitorProfileMetrics`
- New field `AzureMonitorProfile` in struct `ManagedClusterProperties`


## 2.3.0 (2023-01-27)
### Features Added

- New value `ManagedClusterPodIdentityProvisioningStateCanceled`, `ManagedClusterPodIdentityProvisioningStateSucceeded` added to type alias `ManagedClusterPodIdentityProvisioningState`
- New value `PrivateEndpointConnectionProvisioningStateCanceled` added to type alias `PrivateEndpointConnectionProvisioningState`
- New struct `ManagedClusterWorkloadAutoScalerProfile`
- New struct `ManagedClusterWorkloadAutoScalerProfileKeda`
- New field `WorkloadAutoScalerProfile` in struct `ManagedClusterProperties`
- New field `Location` in struct `ManagedClustersClientGetCommandResultResponse`


## 2.2.0 (2022-10-26)
### Features Added

- New function `*ManagedClustersClient.BeginRotateServiceAccountSigningKeys(context.Context, string, string, *ManagedClustersClientBeginRotateServiceAccountSigningKeysOptions) (*runtime.Poller[ManagedClustersClientRotateServiceAccountSigningKeysResponse], error)`
- New struct `ManagedClusterOIDCIssuerProfile`
- New struct `ManagedClusterStorageProfileBlobCSIDriver`
- New struct `ManagedClustersClientBeginRotateServiceAccountSigningKeysOptions`
- New struct `ManagedClustersClientRotateServiceAccountSigningKeysResponse`
- New field `BlobCSIDriver` in struct `ManagedClusterStorageProfile`
- New field `OidcIssuerProfile` in struct `ManagedClusterProperties`


## 2.1.0 (2022-08-25)
### Features Added

- New const `OSSKUWindows2019`
- New const `OSSKUWindows2022`


## 2.0.0 (2022-07-22)
### Breaking Changes

- Struct `ManagedClusterSecurityProfileAzureDefender` has been removed
- Field `AzureDefender` of struct `ManagedClusterSecurityProfile` has been removed

### Features Added

- New const `KeyVaultNetworkAccessTypesPrivate`
- New const `NetworkPluginNone`
- New const `KeyVaultNetworkAccessTypesPublic`
- New function `PossibleKeyVaultNetworkAccessTypesValues() []KeyVaultNetworkAccessTypes`
- New struct `AzureKeyVaultKms`
- New struct `ManagedClusterSecurityProfileDefender`
- New struct `ManagedClusterSecurityProfileDefenderSecurityMonitoring`
- New field `HostGroupID` in struct `ManagedClusterAgentPoolProfileProperties`
- New field `HostGroupID` in struct `ManagedClusterAgentPoolProfile`
- New field `AzureKeyVaultKms` in struct `ManagedClusterSecurityProfile`
- New field `Defender` in struct `ManagedClusterSecurityProfile`


## 1.0.0 (2022-05-16)
### Other Changes

The package of `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice` is using our [next generation design principles](https://azure.github.io/azure-sdk/general_introduction.html) since version 1.0.0, which contains breaking changes.

To migrate the existing applications to the latest version, please refer to [Migration Guide](https://aka.ms/azsdk/go/mgmt/migration).

To learn more, please refer to our documentation [Quick Start](https://aka.ms/azsdk/go/mgmt).
