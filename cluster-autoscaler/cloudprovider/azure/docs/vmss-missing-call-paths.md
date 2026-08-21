# Azure Missing-VMSS Call Paths

This note describes how core Cluster Autoscaler callers reach the Azure VMSS helpers and where missing-VMSS errors are handled.

## Related change

**August 2026:** [kubernetes/autoscaler#10166](https://github.com/kubernetes/autoscaler/pull/10166) prevents an Azure node group without a backing VMSS from blocking cluster-wide autoscaling.

## Node-group inventory and refresh

```text
StaticAutoscaler loop
  -> CloudProvider.Refresh()
  -> AzureManager.Refresh()
  -> AzureManager.forceRefresh()
     -> fetchAutoNodeGroups()
        -> getFilteredNodeGroups(autoDiscoverySpecs)
        -> getFilteredScaleSets()
           -> current VMSS cache entries matching configured tag selectors
        -> register/update matching autodiscovered groups
        -> compare against getRegisteredNodeGroups()
        -> unregister non-explicit groups that no longer match
     -> azureCache.regenerate()
        -> fetchAzureResources()
           -> VMSS.List() replaces the VMSS cache snapshot
        -> Nodes() for every registered group
```

`getRegisteredNodeGroups()` returns the complete provider configuration inventory. Groups enter it through explicit node-group specs or tag-based autodiscovery. Registration does not prove that a backing Azure resource currently exists.

`getNodeGroups()` returns the operational subset exposed by `CloudProvider.NodeGroups()`. It removes registered `ScaleSet` groups that have no VMSS in the current cache snapshot. Internal autodiscovery reconciliation uses the unfiltered registered inventory so a hidden group can still be updated or unregistered.

## Why there are two size helpers

`getCurSize()` is the lower-level Azure-aware helper. It owns synchronization for `curSize` and `lastSizeRefresh`, reads VMSS capacity from the cache (or a direct GET for Spot), and returns two pieces of information:

- the current or last-known size, using `-1` when no valid size has ever been learned;
- a typed `GetVMSSFailedError` that distinguishes VMSS not-found from other Azure failures.

The size and typed error are intentionally returned together. `Nodes()` needs the not-found classification to represent a missing VMSS as an empty group. Read-only `TargetSize()` can use a non-negative last-known size during a narrow snapshot transition. Neither behavior should apply to mutations.

`getScaleSetSize()` is the strict wrapper for mutation paths. It converts the Azure-specific result into the conventional `(int64, error)` contract, propagates every provider error, and also turns an unexplained `-1` size into an explicit error. `IncreaseSize()`, `AtomicIncreaseSize()`, `DecreaseTargetSize()`, and `DeleteNodes()` use this wrapper so they never act on stale or invalid capacity.

`TargetSize()` and `Nodes()` call `getCurSize()` directly because they deliberately interpret VMSS not-found differently. The separate helpers keep those caller-specific read behaviors out of mutation paths.

## Target-size flow

```text
StaticAutoscaler.updateClusterState()
  -> ClusterStateRegistry.UpdateNodes()
  -> getTargetSizes()
  -> CloudProvider.NodeGroups()
  -> AzureManager.getNodeGroups()
     -> filter ScaleSets absent from the VMSS cache
  -> TargetSize() for every returned group
  -> ScaleSet.getCurSize()
     -> getVMSSFromCache()
```

`getTargetSizes()` returns an empty map and an error as soon as any `TargetSize()` call fails. `UpdateNodes()` propagates that error, aborting cluster-state update and the autoscaler loop.

`TargetSize()` is read-only. If a VMSS disappears between node-group enumeration and its size lookup, it may use a non-negative last-known size returned with the typed not-found error. This fallback is defensive; normal missing VMSS groups are removed by `getNodeGroups()` first.

## Instance-list flow

```text
AzureManager.forceRefresh()
  -> azureCache.regenerate()
  -> Nodes() for every registered group
  -> ScaleSet.getCurSize()
     -> getVMSSFromCache()
```

`azureCache.regenerate()` returns immediately if any `Nodes()` call returns an error. `ScaleSet.Nodes()` therefore translates VMSS not-found into an empty instance list and nil error. Other errors still propagate.

Core CAS also calls `Nodes()` through its node-instance cache and scale-down processing. Those callers normally see only the groups exposed by `CloudProvider.NodeGroups()`.

## Mutation flows

```text
IncreaseSize() / AtomicIncreaseSize()
  -> canIncreaseSize()
  -> getScaleSetSize()
  -> getCurSize()

DecreaseTargetSize()
  -> getScaleSetSize()
  -> getCurSize()

DeleteNodes()
  -> getScaleSetSize()
  -> getCurSize()
```

`getScaleSetSize()` is strict: it propagates a missing-VMSS error even when `getCurSize()` also returns a last-known size. Mutating operations must not make decisions from stale capacity.

## Call path targeted by [kubernetes/autoscaler#10166](https://github.com/kubernetes/autoscaler/pull/10166)

The reported issue followed the target-size flow:

```text
phantom pool remains registered
  -> CloudProvider.NodeGroups() returns it
  -> ClusterStateRegistry.getTargetSizes()
  -> phantom TargetSize()
  -> missing VMSS error
  -> UpdateNodes() fails
  -> no scaling decisions for healthy groups
```

The primary fix filters cache-absent VMSS groups from the public `NodeGroups()` view. The group remains in the registered inventory for configuration reconciliation. The secondary protection lets read-only `TargetSize()` use a known cached size during a narrow snapshot transition, while mutations remain strict.

## Call paths targeted by previous fixes

[PR #7708](https://github.com/kubernetes/autoscaler/pull/7708) (`fix: don't crash when vmss not present or has no nodes`) targeted the instance-list flow. It made `GetOptions()` accept defaults and made `Nodes()` return an empty list for VMSS not-found, preventing Azure manager construction or cache regeneration from failing. It did not filter the group or change the `TargetSize()` path.

[PR #9779](https://github.com/kubernetes/autoscaler/pull/9779) (`fix(azure): VMSS size cache handling after delete failures`) included several related cache fixes. Its target-size change handled an inconsistent result from `getCurSize()`: `(-1, nil)`. It made the size wrapper return an explicit error instead of dereferencing a nil `GetVMSSFailedError`, but did not handle a real VMSS not-found error.

The same [PR #9779](https://github.com/kubernetes/autoscaler/pull/9779) also prevents deletion from publishing a negative cached size and invalidates size state after delete failure. Those changes target delete-related cache correctness; they do not affect node-group filtering or missing-VMSS error routing.