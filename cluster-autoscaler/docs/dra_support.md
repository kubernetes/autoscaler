Cluster Autoscaler DRA support
==============================

This page documents Cluster Autoscaler support for Dynamic Resource Allocation (DRA).

The documentation is organized in the following high-level sections:
* [Background on DRA](#background-on-dra): summary of key information about DRA in the context of Cluster Autoscaler
* [User guide](#user-guide): user-facing documentation
* [Developer guide](#developer-guide): developer-facing documentation 

## Background on DRA

### Basic overview

[Dynamic Resource Allocation (DRA)](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
is Kubernetes feature that lets you request and share devices (e.g. GPUs) among Pods. DRA reached stable status, and so was
enabled by default, in K8s 1.35.

DRA solves similar problems as the older [Device Plugin feature](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/),
which allows adding and requesting extended resources (e.g. GPUs) to Node objects. DRA was created
because Device Plugin was not flexible enough to handle some new device usage patterns (e.g. selecting devices by attributes,
sharing devices between Pods).

At a high level, DRA works as follows:

* DRA Drivers (usually DaemonSets) publish ResourceSlices which describe the available Devices, logically grouped into ResourcePools. Each Device contains attributes
  (e.g. driver version), capacity (e.g. the amount of VRAM), and config (i.e. the config data used when configuring the Device).
* Pods request the devices by referencing ResourceClaims, which select from the available Devices using CEL expressions. CEL expressions reference Device attributes
  and/or capacity.
* DRA Drivers or Cluster Admins publish DeviceClasses, which encapsulate the CEL expressions needed to select particular devices.
  DeviceClasses can be referenced from ResourceClaims instead of/in addition to explicit CEL expressions.
* When scheduling a Pod, kube-scheduler checks if all the ResourceClaims it references can be satisfied on a given Node (based on the published ResourceSlices).
* When binding a Pod to a Node, kube-scheduler modifies the Status of its ResourceClaims to reflect that some of the Devices expressed
  in ResourceSlices are _allocated_ to them. This makes the Devices unavailable to satisfy further claims, until the Pod is deleted
  and the claims are _deallocated_.
* When admitting a Pod to a Node, kubelet configures all the Devices from referenced ResourceClaims (using the config part of the Device), and makes them available
  to the Pod via [Container Device Interface (CDI)](https://docs.docker.com/build/building/cdi/).

### Important details

Diving a bit deeper, the following DRA details are also important to understand for Cluster Autoscaler:

* The Devices published in ResourceSlices specify on which Nodes they are available. A device can be available on every Node, on a subset
  of Nodes, or on a single particular Node - i.e. Node-local.
* The Devices published in ResourceSlices are different depending on if they use the Partitionable Devices feature (added in a later KEP):
  * Without Partitionable Devices, the Devices cannot "overlap" each other. Each Device within a ResourcePool is typically identical and represents a distinct, full hardware device.
  * With Partitionable Devices, the Devices within a ResourcePool can overlap each other, exposing different parts of the same hardware device. Such ResourcePools contain SharedCounters in addition
    to Devices. Each SharedCounter is a collection of Counters, and Devices _consume_ portions of these counters. When a Device is allocated, the consumed portion of the counters
    is no longer available to other Devices - so the overlapping Devices can no longer be allocated. This also means that the individual Devices within a ResourcePool are not
    identical - one could be representing a full hardware device, and another just a tiny portion of it.
* ResourceClaims can be referenced by Pods in two different ways:
  * Pods can directly reference a ResourceClaim by its name. In this case, the lifetime of the ResourceClaim is not tied to the lifetime of the Pod. The ResourceClaim has
    to be created by a user before the Pods are created, and then cleaned up after the claim is no longer used by any Pod. This mode is mostly used for sharing the same
    ResourceClaim across multiple Pods.
  * Pods can reference a ResourceClaimTemplate instead of referencing a ResourceClaim directly. In this case, the lifetime of the ResourceClaim is bound to the lifetime of
    the Pod. When such a Pod is created, the ResourceClaim controller creates the ResourceClaim object based on the referenced template. When the Pod is deleted, the controller
    deletes the ResourceClaim object as well. This is the default mode when claims don't have to be shared between multiple Pods.

## User guide

DRA is an extremely flexible API with many use-cases, but Cluster Autoscaler support mostly focuses on Node-local Devices.
Such Node-local devices are treated similarly to the Node-local devices exposed via Device Plugin extended resources (e.g. nvidia.com/gpu).

### Configuration

In order to use DRA with Cluster Autoscaler, you have to configure a NodeGroup in which the Nodes have Node-local Devices exposed
in ResourceSlice objects by some DRA driver. Every Node in the NodeGroup should have an identical set of Devices exposed - which is in line with the
general requirement that all Nodes within a NodeGroup should be homogeneous.

Similarly to many other Cluster Autoscaler features, scaling such a DRA NodeGroup up from 0 Nodes only works if your chosen cloud provider
has explicitly implemented support for it. This support will typically be implemented per-DRA-driver, check the documentation of the
cloud provider integration to see which DRA drivers are supported for scale-from-0. If the DRA driver you want to run is not supported for
scale-from-0, you have to configure the NodeGroup to have at least 1 Node.

Consult the [Limitations/Caveats section](#limitationscaveats) for other limitations and caveats to look out for.

Cluster Autoscaler then provisions new Nodes from the configured NodeGroup in response to pending Pods referencing ResourceClaims, and consolidates existing Nodes
when they're no longer needed.

### Node provisioning

When performing Node provisioning/scale-up simulations, Cluster Autoscaler assumes that adding a new Node to the NodeGroup will also an identical set of
Node-local Devices.

If a pending Pod referencing ResourceClaims can't be scheduled on any existing Node in a cluster, but could be scheduled
on a new Node from the configured NodeGroup, Cluster Autoscaler will provision a new Node from the NodeGroup for the Pod.

Cluster Autoscaler validates that the pending Pod's ResourceClaims can be satisfied by the Node-local DRA Devices of the potential Node, using the exact same
logic as kube-scheduler. If the claims can't be satisfied by Nodes from any of the configured NodeGroups, Cluster Autoscaler won't provision
any Nodes unnecessarily.

### Node consolidation

When analyzing whether to consolidate/scale-down a given Node that has Node-local DRA Devices exposed, Cluster Autoscaler calculates the
Node utilization very similarly to how it calculates GPU utilization:
* If a Node has Node-local DRA Devices exposed, only the utilization of these DRA Devices matters for the final utilization value - CPU and
  memory utilization is ignored.
* The utilization is calculated separately for each ResourcePool among the Node-local Devices, then the final utilization is a maximum of the
  ResourcePool utilization. In a typical case, there's only a single ResourcePool.
* The utilization of a given ResourcePool of Node-local Devices is calculated differently depending on if the ResourcePool utilizes Partitionable Devices:
  * If a ResourcePool doesn't utilize Partitionable Devices, the utilization is simply calculated as the number of allocated Devices in the pool divided by the total
    number of Devices in the pool.
  * If a ResourcePool utilizes Partitionable Devices, the utilization is calculated as the sum of utilizations of every SharedCounter in the pool. The utilization of
    a given SharedCounter is calculated as the maximum of utilizations among its Counters.

### Observability

The following metrics were extended or introduced to provide observability for DRA autoscaling:

* A `dra_drivers` label was added to the metrics tracking scaling decisions - `scaled_up_nodes_total`, `scaled_down_nodes_total`, `failed_scale_ups_total`. 
  Using the new labels, the metrics can be drilled down to scaling decisions involving DRA Nodes.
* A new `dra_node_template_resources_mismatch` gauge metric was introduced, with a single driver label. This metric reports differences in DRA devices between
  scale-from-0 predictions and real Nodes.

### Cluster Autoscaler supported versions

DRA support in Cluster Autoscaler has been production-ready since the [1.35.0 release](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.35.0).
The support can be enabled in earlier versions, but it has various gaps and is not recommended to be used in production.

#### Additional KEP support

There's a number of KEPs which add new features to DRA, on top of the main one which introduced DRA itself. Cluster Autoscaler has been integrated with some of them in
later releases.

* [KEP-5018: DRA: AdminAccess for ResourceClaims and ResourceClaimTemplates](https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/5018-dra-adminaccess/README.md):
  supported since [Cluster Autoscaler 1.35.0](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.35.0)
* [KEP-4815: DRA: Add support for partitionable devices](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/4815-dra-partitionable-devices/README.md):
  supported since [Cluster Autoscaler 1.36.0](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.36.0)

TODO(#10009): Document support for the rest of DRA KEPs.

### Limitations/caveats

Limitations:

* Not all DRA drivers are supported by Cluster Autoscaler. DRA drivers publishing Devices for autoscaled NodeGroups have the following requirements:
  * They should publish an identical set of Devices for every Node in the same NodeGroup.
  * They should only publish Node-local Devices. Such Node-local Devices shouldn't form larger multi-Node topologies.
  * If Partitionable Devices are not utilized, all Devices within a published ResourcePool should be identical.
  * If Partitionable Devices are utilized, all SharedCounters within a published ResourcePool should be identical.
* The logic for calculating the "DRA utilization" of a Node should work well for "GPU-like devices" - i.e. a set of identical hardware devices attached to a Node,
  the cost of which is much higher than the cost of CPU/memory. The logic will likely not work well for other kinds of devices (e.g. virtual devices that don't actually
  cost anything, or different kinds of devices on the same Node), and Cluster Autoscaler might not consolidate Nodes when you expect it to.
* Pods referencing ResourceClaims shouldn't have a PriorityClass with `preemptionPolicy: PreemptLowerPriority` (see [#7863](https://github.com/kubernetes/autoscaler/issues/7683) for details).
* ResourceClaims referenced by DaemonSet Pods targeting an autoscaled NodeGroup have the following limitations:
  * None of the limitations apply as long as all requests in the ResourceClaim have AdminAccess configured. We recommend that DS Pods only reference ResourceClaims with AdminAccess configured.
  * ResourceClaims that aren't owned by the DS Pod ("shared") can only be used by at most 128 Pods (this is a limitation in the DRA API for the length of ReservedFor field). It's the user's
    responsibility to configure autoscaling so that the number of DS Pods referencing each "shared" claim cannot exceed 128. For example, a user could create a DaemonSet per NodeGroup, and
    configure `max_nodes<=128`. We recommend not referencing "shared" claims from DS Pods.
  * ResourceClaims that are owned by the DS Pod should have selectors configured so that only Node-local Devices can get allocated for the claim. DS Pods referencing claims with non-Node-local
    allocations are not supported.

Caveats:

* Despite the limitations on supported DRA drivers outlined above, pending Pods can reference ResourceClaims that can only be satisfied by some Devices that are not Node-local. Cluster Autoscaler
  will still provision new Nodes for such Pods if they can be scheduled on Nodes from some NodeGroup in the simulations. This can be true if the claims are already allocated (unlikely but possible),
  or if there already are some unallocated non-Node-local Devices in the cluster that could satisfy the claims. Cluster Autoscaler will recognize that such claims could be satisfied, and will allocate the
  non-Node-local Devices to them in the simulation. But adding a new Node in the simulation will never add new non-Node-local claims, so Cluster Autoscaler will only be able to provision new Nodes until
  the existing non-Node-local Devices are all allocated.
  * For example, a Pod could reference two ResourceClaims - one for a GPU, and another one for a "meta Device" used for some bookkeeping, available globally and expected to be created by some controller for every Pod
    in the cluster. If the global Device is indeed created for the Pod, Cluster Autoscaler will be able to provision a new Node for the Pod from a configured DRA GPU NodeGroup. If the global Device is for some reason
    not created, and all existing ones are already allocated, Cluster Autoscaler won't be able to provision a new Node for the Pod (and if it did provision a new Node, kube-scheduler also wouldn't be able to schedule
    the Pod there).

## Developer guide

### Implementation details

The following parts of Cluster Autoscaler logic interact with DRA:
* New NodeInfo and PodInfo objects were introduced, wrapping the corresponding scheduler framework objects. These new objects track information about Nodes and Pods
  specific to Cluster Autoscaler logic - like the Node-local ResourceSlices and the ResourceClaims referenced by Pod.
* ClusterSnapshot tracks DRA-related objects in addition to Pods and Nodes:
  * Correlates Node-local Devices from ResourceSlices to the relevant NodeInfos.
  * Correlates ResourceClaims referenced by Pods to the relevant PodInfos.
* Cluster Autoscaler had to start calling an additional phase of scheduler framework - Reserve:
  * During the Filter phase, the DRA scheduler plugin has to determine which Devices can satisfy the ResourceClaims referenced by the Pod. This information is cached in CycleState,
    and then used during the Reserve phase to persist the allocation in the ResourceClaim status.
  * Cluster Autoscaler hooks into the DRA scheduler plugin so that the ResourceClaim status modifications modify the claims
    in ClusterSnapshot instead of the real API.
  * Any time a Pod referencing ResourceClaims is scheduled on a Node in Cluster Autoscaler ClusterSnapshot simulations, CA has to
    call the Reserve phase in addition to PreFilter and Filter, so that the claim allocations are persisted in ClusterSnapshot.
* Scale-from-0 logic used for predicting how a new Node should look like in a NodeGroup with 0 Nodes now also needs to predict ResourceSlices
  containing Node-local Devices for that Node. This is fully deferred to cloud-provider-specific logic via `NodeGroup.TemplateNodeInfo()`.
* Readiness-hacking logic in CustomResourcesProcessor hacks DRA Nodes to be not-Ready until all the expected ResourceSlices are published. This is similar to the
  existing "GPU hack" mechanism for Device Plugin.
  * It can take some time for the ResourceSlices to be published, even after the Node is already in a Ready state. Cluster Autoscaler has to modify the Node in-memory
    to be not-Ready until all the expected ResourceSlices are published.
  * This logic also utilizes template NodeInfos to know what exactly ResourceSlices it should wait for:
    * If a given NodeGroup has existing Nodes, its existing ResourceSlices are used.
    * If a given NodeGroup is at 0 Nodes, `NodeGroup.TemplateNodeInfo()` is used to determine the slices.
* Template NodeInfo sanitization logic now also has to sanitize ResourceSlices and potentially ResourceClaims (if DaemonSet/static Pods reference them).
  * Sanitizing ResourceSlices with Node-local Devices is straightforward - the Devices are kept as-is, and only the ResourcePool name is adapted to match the new Node.
  * Sanitizing ResourceClaims referenced by DaemonSet Pods is tricky. See the [Limitations/Caveats section](#limitationscaveats) for details.
* Utilization logic now calculates DRA utilization for Nodes with Node-local DRA Devices, similarly to the existing Device Plugin GPU logic.
  * A Node can have multiple ResourcePools with Node-local DRA Devices. A _DRA utilization_ of a Node is defined as the highest utilization among the Node-local ResourcePools.
  * For each ResourcePool with non-partitioned Devices, the utilization is calculated by dividing the number of allocated Devices by the total number of Devices.
  * For each ResourcePool with partitioned Devices, the utilization is calculated [as explained in this section](#node-consolidation).
  * This logic mirrors the existing logic for GPU utilization, and should work well for Devices similar to GPUs. The logic won't work well for other kinds of devices,
    see the [Limitations/Caveats](#limitationscaveats) section for details.

### CloudProvider integration

CloudProvider integrations can implement support for scaling up a DRA NodeGroup up from 0 Nodes. This is done similarly to the existing scale-from-0 support,
through the `NodeGroup.TemplateNodeInfo()` method:

* A new CA-specific `NodeInfo` and `PodInfo` objects were introduced, wrapping the corresponding scheduler framework objects. `NodeGroup.TemplateNodeInfo()` was changed to
  return these new objects instead of the scheduler ones.
* In order to implement scale-from-0 support for a given DRA driver, create the ResourceSlices expected to be created by the driver for the Node in `NodeGroup.TemplateNodeInfo()`,
  and include them in the return value - in `NodeInfo.LocalResourceSlices`.
* The ResourceSlices included in `NodeInfo.LocalResourceSlices` returned from `NodeGroup.TemplateNodeInfo()` are added to the cluster simulation state together with the
  Node. Similarly to scale-from-0 logic in general - if the Device attributes inside the ResourceSlices are predicted incorrectly, Cluster Autoscaler might make wrong
  scaling decisions.
* The Node readiness-hacking logic for DRA (corresponding to the old GPU hack logic) is a bit more resilient to wrong predictions. Instead of waiting for the exact Devices
  returned from `NodeGroup.TemplateNodeInfo()` to be published, it waits for the expected number of complete ResourcePools from each driver to be published. So the logic
  should work as long as `NodeGroup.TemplateNodeInfo()` predicts the same number of ResourcePools for each driver as there published in reality. If e.g. a single parameter
  of some Device is predicted incorrectly, the readiness-hacking logic should still work (but scheduler filters will still not work correctly if a pending Pod references
  a ResourceClaim which references that incorrectly predicted parameter).
* If a DaemonSet Pod targeting an autoscaled NodeGroup references ResourceClaims, the ResourceClaims should be included in the result of `NodeGroup.TemplateNodeInfo()` -
  in `PodInfo.NeededResourceClaims`. The claims should be returned already allocated. There are limitations on the kinds of ResourceClaims that a DS Pod can reference,
  see the [Limitations/Caveats](#limitationscaveats) section for details.
* There is a new `dra_node_template_resources_mismatch` metric that will report differences in the DRA objects returned for scale-from-0 from `NodeGroup.TemplateNodeInfo()`,
  and the real DRA objects created in the cluster.
