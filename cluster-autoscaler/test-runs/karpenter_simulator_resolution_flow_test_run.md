# Test Run: Karpenter Simulator Group-Expand-Balance Resolution Flow

This test run validates the implementation of the new **Group-Expand-Balance** resolution flow in Karpenter simulator ([karpenter_simulator.go](file:///usr/local/google/home/danielmk/src/k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator/karpenter_simulator.go)).

The changes implemented:
1.  Grouped virtual claims by requirement constraints.
2.  Built expander options using similar node groups (resolving the locking label issue).
3.  Evaluated options with the expander to find the best candidate.
4.  Applied resource quotas and total node limits to cap the scale-up delta.
5.  Balanced the allowed nodes across similar groups.
6.  Committed capped claims to the snapshot and consumed quotas in-place.
7.  Recycled pods from discarded claims back to the pending pool.
8.  Enforced cost-optimization with failover to more expensive types if cheaper types are full.
9.  **Resolved Custom Label Blocking Gap:** NodePool requirements are now configured with `NotIn [ca-ignore-reserved-value]` instead of a strict `In` union of all labels, ensuring node groups lacking custom labels are not blocked from scaling up when scheduling pods with no selectors. We also integrated Karpenter's native `Compatible` checker for high-fidelity custom label matching.

## Test Runner Script

The tests were executed using the runner script: [run_karpenter_simulator_tests.sh](file:///usr/local/google/home/danielmk/src/k8s.io/autoscaler/cluster-autoscaler/test-runs/run_karpenter_simulator_tests.sh).

```bash
#!/bin/bash
# Script to run Karpenter Simulator tests manually.
set -e

echo "Running Karpenter Simulator unit tests..."
go test -v ./core/scaleup/orchestrator/ -run "TestKarpenterSimulator.*"

echo "Running Builder validation tests..."
go test -v ./builder/... -run "TestAutoscalerBuilder.*"

echo "All tests passed successfully!"
```

## Execution Log

The execution of the runner script was recorded on 2026-07-10 (post-custom-label-fix):

```
Running Karpenter Simulator unit tests...
=== RUN   TestKarpenterSimulatorCorrectness
--- PASS: TestKarpenterSimulatorCorrectness (0.00s)
=== RUN   TestKarpenterSimulatorLabelRequirements
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_custom_label_present_on_one_node_group
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_custom_label_NOT_present_on_any_node_group
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_well-known_label_(zone)_present_on_one_node_group
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_restricted_label_(hostname)_-_should_NOT_match_(new_nodes_don't_have_hostname_yet)
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_label_in_restricted_domain_(karpenter.sh/)
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_requesting_multiple_labels
=== RUN   TestKarpenterSimulatorLabelRequirements/pod_not_requesting_custom_label_-_should_allow_scaling_up_group_lacking_it
--- PASS: TestKarpenterSimulatorLabelRequirements (0.01s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_custom_label_present_on_one_node_group (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_custom_label_NOT_present_on_any_node_group (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_well-known_label_(zone)_present_on_one_node_group (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_restricted_label_(hostname)_-_should_NOT_match_(new_nodes_don't_have_hostname_yet) (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_label_in_restricted_domain_(karpenter.sh/) (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_requesting_multiple_labels (0.00s)
    --- PASS: TestKarpenterSimulatorLabelRequirements/pod_not_requesting_custom_label_-_should_allow_scaling_up_group_lacking_it (0.00s)
=== RUN   TestKarpenterSimulatorSmallestInstanceTypeResolution
    karpenter_simulator_test.go:344: PLAN ITEM: Group=ng-standard-2, Delta=1
--- PASS: TestKarpenterSimulatorSmallestInstanceTypeResolution (0.00s)
=== RUN   TestKarpenterSimulatorMultiZoneReliabilityBalancing
--- PASS: TestKarpenterSimulatorMultiZoneReliabilityBalancing (0.00s)
=== RUN   TestKarpenterSimulatorDaemonSets
--- PASS: TestKarpenterSimulatorDaemonSets (0.00s)
=== RUN   TestKarpenterSimulatorClusteringWithDifferentLabels
--- PASS: TestKarpenterSimulatorClusteringWithDifferentLabels (0.00s)
=== RUN   TestKarpenterSimulatorQuotaCappingAndBalancing
--- PASS: TestKarpenterSimulatorQuotaCappingAndBalancing (0.01s)
PASS
ok  	k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator	0.123s
Running Builder validation tests...
=== RUN   TestAutoscalerBuilderNoError
I0101 01:00:00.000000  924319 autoscaler.go:325] Waiting for caches to sync...
--- PASS: TestAutoscalerBuilderNoError (0.08s)
=== RUN   TestAutoscalerBuilderConflictError
--- PASS: TestAutoscalerBuilderConflictError (0.00s)
PASS
ok  	k8s.io/autoscaler/cluster-autoscaler/builder	0.200s
All tests passed successfully!
```

## Detailed Test Cases

*   **`TestKarpenterSimulatorCorrectness`**: Validates basic scheduling and simulation flow.
*   **`TestKarpenterSimulatorLabelRequirements`**: Validates node requirement label matching logic.
*   **`TestKarpenterSimulatorLabelRequirements/pod_not_requesting_custom_label_-_should_allow_scaling_up_group_lacking_it`**: **[NEW]** Validates that when a pod does not request custom labels, the simulator correctly allows scale-up of node groups lacking custom labels, ensuring they are not blocked by other labeled node groups in the same pool.
*   **`TestKarpenterSimulatorSmallestInstanceTypeResolution`**: Validates cost-optimization preference (cheapest instance type selected first) and correct failover logic.
*   **`TestKarpenterSimulatorMultiZoneReliabilityBalancing`**: Validates that scheduling identical pods on multi-zone node groups results in balanced allocation across similar node groups.
*   **`TestKarpenterSimulatorDaemonSets`**: Validates daemonset pod scheduling and handling.
*   **`TestKarpenterSimulatorClusteringWithDifferentLabels`**: Validates cluster snapshot tracking.
*   **`TestKarpenterSimulatorQuotaCappingAndBalancing`**: Validates resource quota capping, in-place quota consumption, pod recycling of capped claims, and balancing of allowed claims.
*   **`TestAutoscalerBuilderConflictError`**: Validates the validation logic that prevents enabling Karpenter simulator and ProvisioningRequests simultaneously.
