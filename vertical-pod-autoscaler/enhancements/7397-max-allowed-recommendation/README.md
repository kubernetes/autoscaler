# KEP-7397: Max allowed recommendation configurable

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Proposed Solution](#proposed-solution)
   - [Alternatives](#alternatives)
   - [Challenges](#challenges)
   - [Test Plan](#test-plan)
<!-- /toc -->

## Summary
VPA currently allows configuring the maximum allowed recommendation per resource using the field `.spec.resourcePolicy.containerPolicies[].maxAllowed`. This helps ensure that recommendations don't exceed the allocatable capacity of the largest node. Without this limit, a high recommendation could make the Pod unschedulable, causing downtime. We aim to improve the VPA recommender by making it aware of the maximum node size, preventing it from generating recommendations that would render Pods unschedulable.

## Motivation
In dynamic environments, manual configuration of .maxAllowed can be error-prone, especially as cluster sizes and node configurations change over time. Automating this process by making the VPA recommender aware of the largest node size ensures more reliable scaling decisions. This helps prevents situations where Pods become unschedulable due to excessive resource recommendations.

### Goals
* Ensure Pods will still schedulable when a recommendation would be applied: Have the VPA recommender only suggest resources that fit within the largest node's capacity, preventing situation which require manual interaction.
* Adapt to Changing Environments: Adjust recommendations as the cluster size or node configurations change to better reflect the current state of the cluster.

## Proposal
We propose to improve the VPA recommender by making it aware of the maximum node size, preventing it from generating recommendations that would render Pods unschedulable. This will involve querying for cluster information and adjusting recommendations accordingly.

## Design Details

### Proposed Solution
We propose to use the Cluster Autoscaler's provisioning-request API to obtain information about the maximum node size. This approach aligns well with CA's design and capabilities. The implementation will include:

1. Feature Flag:
   The behavior of querying Cluster Autoscaler will be controlled by a clearly defined flag. This ensures that users can explicitly enable or disable this feature, providing clarity on whether they can rely on CA for recommendation limits.

2. Caching Mechanism:
   To optimize performance and reduce the load on the Cluster Autoscaler, a caching mechanism will be implemented. This would involve storing the retrieved node size information for a configurable period before refreshing.

3. Cluster Autoscaler Availability Warning:
   The system should implement a warning mechanism to alert users when Cluster Autoscaler is not present in the cluster. This warning will help users understand why the feature might not be functioning as expected and guide them towards proper configuration.

4. Query Cluster Autoscaler: 
   VPA could query the Cluster Autoscaler (CA) for feedback, potentially using the [provisioning-request](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/provisioning-request.md) API or a similar mechanism. This approach aligns well with CA's design, as it is built to handle questions related to resource availability and scheduling.
   For more information on supported provisioning classes, refer to the [Cluster Autoscaler FAQ](https://github.com/kubernetes/autoscaler/blob/3a29dc2690102a6758cd085e9d6a3bcf4d7c29d8/cluster-autoscaler/FAQ.md#supported-provisioningclasses).

### Alternatives

1. Check Current Cluster Nodes:
   Analyze the available nodes and, if needed, perform calculations based on the largest node. Implementation has been started in [PR #7345](https://github.com/kubernetes/autoscaler/pull/7345).

2. Set Fixed Values:
   Use predefined limits, similar to the implementation [here](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/recommender/logic/recommender.go).

### Challenges
- Not all clusters use CA, and Karpenter currently does not support the `provisioningrequest` API and is unlikely to support it in the future.
- While we've previously discussed building a unified API supported by both CA and Karpenter, it is not a high priority and is unlikely to be implemented soon.
- Some clusters may not use any form of node autoscaling at all.
- The `provisioningrequest` API, in its current form, is not optimized for this specific use case, meaning adjustments would be needed in CA to fully support it.

### Test Plan
* Ensure that recommendations do not exceed the allocatable capacity of the largest node.
* Test behavior when no node information is available (e.g., fallback scenarios).
* Simulate scenarios where new nodes are added, removed, or resized, ensuring the recommender adapts its recommendations dynamically.
* Test interactions between VPA and the Cluster Autoscaler (if enabled) to confirm smooth cooperation via the provisioningrequest API or other communication mechanisms.
* Test timeout scenarios for API calls to CA/Karpenter.