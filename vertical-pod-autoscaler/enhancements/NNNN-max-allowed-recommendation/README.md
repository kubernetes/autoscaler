# KEP-NNNN: Max allowed recommendation configurable

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Test Plan](#test-plan)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary
VPA currently allows configuring the maximum allowed recommendation per resource using the field `.spec.resourcePolicy.containerPolicies[].maxAllowed`. This helps ensure that recommendations don't exceed the allocatable capacity of the largest node. Without this limit, a high recommendation could make the Pod unschedulable, causing downtime. We aim to improve the VPA recommender by making it aware of the maximum node size, preventing it from generating recommendations that would render Pods unschedulable.

## Motivation
In dynamic environments, manual configuration of .maxAllowed can be error-prone, especially as cluster sizes and node configurations change over time. Automating this process by making the VPA recommender aware of the largest node size ensures more reliable scaling decisions. This prevents situations where Pods become unschedulable due to excessive resource recommendations.

### Goals
* Ensure Accurate and Schedulable Recommendations: Guarantee that the VPA recommender provides resource recommendations that do not exceed the allocatable capacity of the largest node, ensuring Pods remain schedulable and avoiding unnecessary downtime.
* Support for Changing Environments: Adapt recommendations as the cluster size or node configurations evolve.

## Proposal
VPA could query the Cluster Autoscaler (CA) for feedback, potentially using the `provisioningrequest` API or a similar mechanism. This approach aligns well with CA’s design, as it is built to handle questions related to resource availability and scheduling. However, this solution comes with some challenges:  

- Not all clusters use CA, and Karpenter currently does not support the `provisioningrequest` API and is unlikely to support it in the future.  
- While we’ve previously discussed building a unified API supported by both CA and Karpenter, it is not a high priority and is unlikely to be implemented soon.  
- Some clusters may not use any form of node autoscaling at all.  
- Additionally, the `provisioningrequest` API, in its current form, is not optimized for this specific use case, meaning adjustments would be needed in CA to fully support it.  

## Design Details
### Test Plan
* Ensure that recommendations do not exceed the allocatable capacity of the largest node.
* Test behavior when no node information is available (e.g., fallback scenarios).
* Simulate scenarios where new nodes are added, removed, or resized, ensuring the recommender adapts its recommendations dynamically.
* Test interactions between VPA and the Cluster Autoscaler (if enabled) to confirm smooth cooperation via the provisioningrequest API or other communication mechanisms.


## Alternatives
Regarding the simpler solution, two suggestions were proposed:
### Check Current Cluster Nodes
Analyze the available nodes and, if needed, perform calculations based on the largest node. I’ve started implementing this approach here: [PR #7345](https://github.com/kubernetes/autoscaler/pull/7345)

### Set Fixed Values: 
Use predefined limits, similar to the implementation [here](../../pkg/recommender/logic/recommender.go#L32).
