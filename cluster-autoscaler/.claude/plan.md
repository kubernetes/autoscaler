# Azure SDK v2 Migration Plan for Cluster Autoscaler

## Executive Summary

This plan outlines the migration of the Azure cluster-autoscaler backend from Azure SDK v1 (using `sigs.k8s.io/cloud-provider-azure/pkg/azureclients`) to Azure SDK v2 (using `sigs.k8s.io/cloud-provider-azure/pkg/azclient`).

## Current State Analysis

### Current Azure SDK v1 Dependencies

The codebase currently uses:
- `sigs.k8s.io/cloud-provider-azure/pkg/azureclients/*` - SDK v1 clients
- `github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute` - SDK v1 compute types
- `github.com/Azure/go-autorest/autorest` - SDK v1 authentication

### Files Requiring Changes

**Core client files:**
1. `cloudprovider/azure/azure_client.go` - Client initialization and authentication
2. `cloudprovider/azure/azure_manager.go` - AzureManager setup
3. `cloudprovider/azure/azure_cache.go` - Resource caching logic

**Files using Azure clients:**
4. `cloudprovider/azure/azure_scale_set.go` - VMSS operations
5. `cloudprovider/azure/azure_agent_pool.go` - Agent pool operations
6. `cloudprovider/azure/azure_vms_pool.go` - VMs pool operations
7. `cloudprovider/azure/azure_util.go` - Utility functions (VM deletion, storage)
8. `cloudprovider/azure/azure_force_delete_scale_set.go` - Force deletion logic

**Test files:**
9. `cloudprovider/azure/azure_scale_set_test.go`
10. `cloudprovider/azure/azure_cloud_provider_test.go`
11. `cloudprovider/azure/azure_vms_pool_test.go`
12. `cloudprovider/azure/azure_agent_pool_test.go`
13. `cloudprovider/azure/azure_manager_test.go`
14. `cloudprovider/azure/azure_config_test.go`
15. `cloudprovider/azure/azure_client_test.go`

**Mock/fake files:**
16. `cloudprovider/azure/azure_fakes.go`
17. `cloudprovider/azure/azure_mock_agentpool_client.go`

## Target State - Azure SDK v2

### New Dependencies

The migration will use:
- `sigs.k8s.io/cloud-provider-azure/pkg/azclient` - SDK v2 client factory
- `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5` - SDK v2 compute types
- `github.com/Azure/azure-sdk-for-go/sdk/azcore` - SDK v2 core
- `github.com/Azure/azure-sdk-for-go/sdk/azidentity` - SDK v2 authentication

### Key SDK v2 Interfaces

**ClientFactory** - Centralized client creation:
```go
type ClientFactory interface {
    GetVirtualMachineScaleSetClient() virtualmachinescalesetclient.Interface
    GetVirtualMachineScaleSetVMClient() virtualmachinescalesetvmclient.Interface
    GetVirtualMachineClient() virtualmachineclient.Interface
    GetDeploymentClient() deploymentclient.Interface
    GetInterfaceClient() interfaceclient.Interface
    GetDiskClient() diskclient.Interface
    GetAccountClient() accountclient.Interface
    // ... and more
}
```

**AuthProvider** - Handles multiple auth methods:
- Federated workload identity
- Managed identity (system & user-assigned)
- Client secret
- Client certificate

## Migration Strategy

### Phase 1: Update Client Initialization

**File: `cloudprovider/azure/azure_client.go`**

Changes:
1. Replace `azClient` struct fields:
   - OLD: Individual client imports from `azureclients/*`
   - NEW: Single `azclient.ClientFactory` instance

2. Update `newAzClient()` function:
   - Remove `newAuthorizer()` (SDK v1 specific)
   - Create `ARMClientConfig` and `AzureAuthConfig` for SDK v2
   - Use `azclient.NewAuthProvider()` for authentication
   - Use `azclient.NewClientFactory()` to create all clients

3. Keep `newAgentpoolClient()` as-is (already using SDK v2)

**Authentication mapping:**
- `UseManagedIdentityExtension` → `ManagedIdentityCredential`
- `AADClientID + AADClientSecret` → `ClientSecretCredential`
- `AADClientCertPath` → `ClientCertificateCredential`
- `UseFederatedWorkloadIdentityExtension` → `FederatedIdentityCredential`
- CLI auth → Remove (not commonly used, SDK v2 doesn't support easily)

### Phase 2: Update Type Definitions

**Files: All files using Azure types**

Type mappings:
- `compute.VirtualMachineScaleSet` → `armcompute.VirtualMachineScaleSet`
- `compute.VirtualMachineScaleSetVM` → `armcompute.VirtualMachineScaleSetVM`
- `compute.VirtualMachine` → `armcompute.VirtualMachine`
- `compute.ResourceSkusClient` → Use `armcompute` SKU client from factory
- `retry.Error` → Standard Go `error` (SDK v2 uses standard errors)

### Phase 3: Update Client Method Calls

**Key interface changes:**

**VMSS Client:**
- OLD: `List(ctx, resourceGroup)` returns `([]compute.VirtualMachineScaleSet, *retry.Error)`
- NEW: `List(ctx, resourceGroup)` returns `([]*armcompute.VirtualMachineScaleSet, error)`

- OLD: `Get(ctx, rg, name)` returns `(compute.VirtualMachineScaleSet, *retry.Error)`
- NEW: `Get(ctx, rg, name, expand)` returns `(*armcompute.VirtualMachineScaleSet, error)`

- OLD: `CreateOrUpdateAsync(ctx, rg, name, params)` returns `(*azure.Future, *retry.Error)`
- NEW: `CreateOrUpdate(ctx, rg, name, params)` returns `(*armcompute.VirtualMachineScaleSet, error)`

- OLD: `DeleteInstancesAsync(ctx, rg, name, ids, forceDelete)` returns `(*azure.Future, *retry.Error)`
- NEW: `DeleteInstances(ctx, rg, name, ids, forceDelete)` returns `(pollers, error)`

**VM Client:**
- OLD: `Get(ctx, rg, name, expand)` returns `(compute.VirtualMachine, *retry.Error)`
- NEW: `Get(ctx, rg, name, expand)` returns `(*armcompute.VirtualMachine, error)`

- OLD: `Delete(ctx, rg, name)` returns `*retry.Error`
- NEW: `Delete(ctx, rg, name)` returns `(*pollers, error)`

- OLD: `List(ctx, rg)` returns `([]compute.VirtualMachine, *retry.Error)`
- NEW: `List(ctx, rg)` returns `([]*armcompute.VirtualMachine, error)`

**VMSS VM Client:**
- OLD: `List(ctx, rg, vmssName, expand)` returns `([]compute.VirtualMachineScaleSetVM, *retry.Error)`
- NEW: `List(ctx, rg, vmssName, expand)` returns `([]*armcompute.VirtualMachineScaleSetVM, error)`

### Phase 4: Error Handling Updates

**Changes needed:**
1. Replace `*retry.Error` with standard `error`
2. Remove `checkResourceExistsFromRetryError()` calls
3. Use standard error checking: `if err != nil`
4. Update error unwrapping for 404/NotFound checks

**Error checking pattern:**
```go
// OLD
result, rerr := client.Get(ctx, rg, name)
if rerr != nil {
    if exists, _ := checkResourceExistsFromRetryError(rerr); !exists {
        // Not found
    }
}

// NEW
result, err := client.Get(ctx, rg, name, nil)
if err != nil {
    var respErr *azcore.ResponseError
    if errors.As(err, &respErr) && respErr.StatusCode == 404 {
        // Not found
    }
}
```

### Phase 5: Update Async Operations

**OLD pattern:**
```go
future, rerr := client.CreateOrUpdateAsync(ctx, rg, name, params)
if rerr != nil {
    return rerr
}
err := future.WaitForCompletionRef(ctx, client.Client)
```

**NEW pattern:**
```go
result, err := client.CreateOrUpdate(ctx, rg, name, params)
if err != nil {
    return err
}
// Result is available immediately (SDK v2 waits internally)
```

For truly async operations:
```go
poller, err := client.BeginDelete(ctx, rg, name, nil)
if err != nil {
    return err
}
_, err = poller.PollUntilDone(ctx, nil)
```

### Phase 6: Update Cache Logic

**File: `cloudprovider/azure/azure_cache.go`**

Changes:
1. Update `fetchAzureResources()` to handle pointer types
2. Update `scaleSets` map: `map[string]compute.VirtualMachineScaleSet` → `map[string]*armcompute.VirtualMachineScaleSet`
3. Update `virtualMachines` map: `map[string][]compute.VirtualMachine` → `map[string][]*armcompute.VirtualMachine`
4. Handle nil pointer dereferences when accessing fields

### Phase 7: Update Mock/Fake Clients

**Files: `azure_fakes.go`, test files**

Create new mock implementations:
1. Mock `azclient.ClientFactory`
2. Mock individual client interfaces from `azclient/*client` packages
3. Update test setup to use factory pattern
4. Regenerate mocks using mockgen (for `azure_mock_agentpool_client.go`)

### Phase 8: Update go.mod Dependencies

Remove SDK v1 dependencies:
- Remove: `github.com/Azure/azure-sdk-for-go v68.0.0+incompatible`
- Remove: `github.com/Azure/go-autorest/autorest`
- Remove: `sigs.k8s.io/cloud-provider-azure` (old version with azureclients)

Add/update SDK v2 dependencies:
- Keep: `github.com/Azure/azure-sdk-for-go/sdk/azcore`
- Keep: `github.com/Azure/azure-sdk-for-go/sdk/azidentity`
- Keep: `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5`
- Keep: `sigs.k8s.io/cloud-provider-azure/pkg/azclient`

## Implementation Order

1. **azure_client.go** - Core client initialization
2. **azure_manager.go** - Manager setup
3. **azure_cache.go** - Cache with type changes
4. **azure_util.go** - Utility functions
5. **azure_scale_set.go** - VMSS operations
6. **azure_force_delete_scale_set.go** - Force deletion
7. **azure_agent_pool.go** - Agent pool
8. **azure_vms_pool.go** - VMs pool
9. **azure_fakes.go** - Mock implementations
10. **Test files** - All test files
11. **go.mod** - Dependency cleanup
12. **Verification** - Build and test

## Risk Mitigation

### Backwards Compatibility Concerns

**Authentication:**
- Risk: Different auth methods might behave differently
- Mitigation: Thoroughly test all auth methods (MSI, service principal, workload identity)

**API Behavior:**
- Risk: SDK v2 might have different retry/timeout behavior
- Mitigation: Preserve existing timeout and retry configurations

**Type Changes:**
- Risk: Pointer vs value types might cause nil pointer panics
- Mitigation: Add nil checks where needed, especially in cache operations

### Testing Strategy

1. **Unit tests** - Ensure all existing tests pass with new mocks
2. **Build verification** - Confirm build command succeeds
3. **Test command** - Confirm test command succeeds
4. **Manual testing** - Test in real environment if possible

## Success Criteria

✅ Build command succeeds:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cluster-autoscaler-arm64 --ldflags="-s"
```

✅ Test command succeeds:
```bash
go test ./cloudprovider/azure/... -vet="atomic,bool,buildtags,directive,errorsas,ifaceassert,nilfunc,slog,stringintconv,tests"
```

✅ All existing functionality preserved:
- Scale up/down operations work
- Node discovery works
- Authentication methods work
- Cache invalidation works

✅ Code is cleaner:
- Fewer dependencies
- Modern SDK patterns
- Consistent error handling

## Open Questions

1. **CLI authentication**: Should we keep CLI auth support or remove it?
   - Recommendation: Remove - not well supported in SDK v2, rarely used in production

2. **ResourceSkusClient**: How to handle the SKU client?
   - OLD: `compute.NewResourceSkusClientWithBaseURI()`
   - NEW: Use factory or create separately?
   - Recommendation: Create separately, it's not frequently used

3. **Async operations**: Should we preserve async patterns or make synchronous?
   - Recommendation: Make synchronous where possible (SDK v2 handles waiting internally)

4. **Error types**: Keep custom error types or use standard errors?
   - Recommendation: Use standard errors, check for `azcore.ResponseError` for HTTP status codes

## Timeline Estimate

- Phase 1-2 (Client setup + types): 2-3 hours
- Phase 3-4 (Method calls + errors): 3-4 hours
- Phase 5-6 (Async + cache): 2-3 hours
- Phase 7-8 (Mocks + deps): 2-3 hours
- Testing and fixes: 2-3 hours

**Total: 11-16 hours**

## Notes

- This migration is primarily a refactoring effort - the underlying Azure APIs remain the same
- The goal is minimal behavioral changes while modernizing the SDK
- All existing configuration options and authentication methods should continue to work
- The new SDK provides better performance and is actively maintained by Microsoft
