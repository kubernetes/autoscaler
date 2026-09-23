# AEP-8026: Allow per-VPA component configuration parameters

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [Parameter Descriptions](#parameter-descriptions)
    - [Container Policy Parameters](#container-policy-parameters)
    - [Update Policy Parameters](#update-policy-parameters)
- [Design Details](#design-details)
  - [Configuration Level Considerations](#configuration-level-considerations)
    - [Parameter Level Analysis](#parameter-level-analysis)
    - [Parameter Precedence: Per-VPA vs Global Configuration](#parameter-precedence-per-vpa-vs-global-configuration)
  - [API Changes](#api-changes)
    - [Phase 1 (Current Proposal)](#phase-1-current-proposal)
    - [Future Extensions](#future-extensions)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
    - [How can this feature be enabled / disabled in a live cluster?](#how-can-this-feature-be-enabled--disabled-in-a-live-cluster)
  - [Kubernetes version compatibility](#kubernetes-version-compatibility)
  - [Validation via CEL and Testing](#validation-via-cel-and-testing)
  - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
- [Future Work](#future-work)
- [Alternatives](#alternatives)
  - [Multiple VPA Deployments](#multiple-vpa-deployments)
  - [Environment-Specific Configuration](#environment-specific-configuration)
<!-- /toc -->

## Summary

Currently, VPA components (recommender, updater, admission controller) are configured through global flags. This makes it challenging to support different workloads with varying resource optimization needs within the same cluster. This proposal introduces the ability to specify configuration parameters at the individual VPA object level, allowing for workload-specific optimization strategies. The goal is not to introduce new configuration options but rather to make existing internal configurations accessible and customizable per VPA object.

## Motivation

Different types of workloads in a Kubernetes cluster often have different resource optimization requirements. For example:
- Batch processing jobs might benefit from aggressive OOM handling and frequent adjustments
- User-facing services might need more conservative growth patterns for stability
- Development environments might need different settings than production

Currently, supporting these different needs requires running multiple VPA component instances with different configurations, which increases operational complexity and resource usage.

### Goals

- Allow specification of component-specific parameters in individual VPA objects
- Support different optimization strategies for different workloads in the same cluster
- Maintain backward compatibility with existing global configuration
- Initially support the following parameters:
  - oomBumpUpRatio
  - oomMinBumpUp
  - memoryAggregationInterval
  - memoryAggregationIntervalCount
  - evictAfterOOMSeconds

### Non-Goals

- Converting all existing VPA flags to per-object configuration
- Changing the core VPA algorithm or its decision-making process
- Adding new optimization strategies

## Proposal

The configuration will be split into two sections: container-specific recommendations under `containerPolicies` and updater configuration under `updatePolicy`. This structure is designed to be extensible, allowing for additional parameters to be added in future iterations of this enhancement.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: simple-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: Auto
    evictAfterOOMSeconds: 300
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      oomBumpUpRatio: "1.5"
      oomMinBumpUp: 104857600
      memoryAggregationInterval: "12h"
      memoryAggregationIntervalCount: 5
```

### Parameter Descriptions

#### Container Policy Parameters

* `oomBumpUpRatio` and `oomMinBumpUp`:
  These parameters work together to determine memory increases after OOM events.
  The VPA selects the larger of:
    - The absolute increase specified by `oomMinBumpUp`
    - A relative increase calculated using `oomBumpUpRatio`
    - If both are unset or set to their neutral values (`oomBumpUpRatio = 1`, `oomMinBumpUp = 0`), OOM-based increases are disabled.

  This ensures:
    - Small containers receive a guaranteed minimum bump (via `oomMinBumpUp`, e.g., +100MB).
    - Larger containers receive proportional scaling (via `oomBumpUpRatio`, e.g., ×1.5).

  Implementation logic:

  ```golang
    memoryNeeded := ResourceAmountMax(
      memoryUsed + MemoryAmountFromBytes(GetAggregationsConfig().OOMMinBumpUp),
      ScaleResource(memoryUsed, GetAggregationsConfig().OOMBumpUpRatio)
    )
  ```
  Example: with oomBumpUpRatio: "1.5" and oomMinBumpUp: 104857600 (100MB):
    - For a container using 50MB: max(50MB + 100MB, 50MB * 1.5) = 150MB
    - For a container using 1GB: max(1GB + 100MB, 1GB * 1.5) = 1.5GB

  Note: Using a single field approach (e.g., a unified `oomBumpUp` field) would not provide sufficient flexibility for users who need both a minimum absolute increase and a proportional ratio.
  For example, if a user wants to ensure a minimum increase of 100MB while also applying a 1.5x ratio for larger containers, a single field cannot express this combined behavior. The current dual-field design allows users to specify both constraints independently, ensuring small containers get a guaranteed minimum bump while larger containers receive appropriate proportional scaling. This approach provides more precise control over memory recommendation adjustments after OOM events than a simplified single-field model could offer.


  - `oomBumpUpRatio` (Quantity):
    - Multiplier applied to memory recommendations after OOM events
    - Represented as a Quantity (e.g., "1.5")
    - Must be greater than or equal to 1
    - Setting to 1 effectively disables the OOM ratio-based increase
    - Controls how aggressively memory is increased after container crashes

  - `oomMinBumpUp` (bytes):
    - Minimum absolute memory increase after OOM events
    - Setting to 0 effectively disables the OOM minimum increase

* `memoryAggregationInterval` (duration):
  - Time window for aggregating memory usage data
  - Affects how quickly VPA responds to memory usage changes

* `memoryAggregationIntervalCount` (integer):
  - Number of consecutive memory aggregation intervals
  - Used to calculate the total memory aggregation window length
  - Total window length = memoryAggregationInterval * memoryAggregationIntervalCount

#### Update Policy Parameters
* `evictAfterOOMSeconds` (int32):
  - Time in seconds to wait after OOM before considering pod eviction
  - Helps prevent rapid eviction cycles while maintaining stability

Each parameter can be configured independently, falling back to global defaults if not specified. Values should be chosen based on workload characteristics and stability requirements.

## Design Details

### Configuration Level Considerations

When designing the configuration parameters, we analyzed each parameter to determine the most appropriate configuration level:

#### Parameter Level Analysis

1. **OOM-Related Parameters (`oomBumpUpRatio`, `oomMinBumpUp`)**
- **Recommended Level**: Container-level
- **Rationale**:
  - Different containers in the same pod may have different memory requirements and OOM patterns
  - Memory recommendations in VPA are already handled at container level
  - Consistent with how VPA handles other resource-related configurations
  - While these parameters are container-level, VPA-wide configuration can still be achieved using the wildcard container name "*"

2. **Memory Aggregation Parameters (`memoryAggregationInterval`, `memoryAggregationIntervalCount`)**
- **Recommended Level**: Container-level
- **Rationale**:
  - Different containers may have different memory usage patterns
  - Some containers might need faster reaction times to memory changes
  - Consistent with existing VPA container-level resource policies
  - While `memoryAggregationInterval` and `memoryAggregationIntervalCount` are container-level, VPA-wide configuration can still be achieved using the wildcard container name "*"

3. **Eviction Parameters (`evictAfterOOMSeconds`)**
- **Recommended Level**: VPA-level
- **Rationale**:
  - Eviction decisions affect the entire pod
  - Pod-level operation that shouldn't vary by container
  - Simpler operational model for pod lifecycle management
  - Consistent with how Kubernetes handles pod evictions

#### Parameter Precedence: Per-VPA vs Global Configuration

The per-VPA configuration parameters introduced in this proposal are designed to **override** the corresponding global flags when specified in a VPA object. If a parameter is not defined at the VPA level, the VPA components will fall back to using the value set via global flags.

This approach ensures backward compatibility and allows users to adopt per-VPA configuration incrementally, without requiring changes to existing setups. Users can continue relying on global defaults while gradually introducing workload-specific tuning where needed.

For example:
- If `oomBumpUpRatio` is set in a VPA's `containerPolicy`, that value will be used for recommendations for that container.
- If it is omitted, the global flag value (e.g., from the recommender component) will apply.

This override behavior applies to all parameters introduced in this AEP:
- `oomBumpUpRatio`
- `oomMinBumpUp`
- `memoryAggregationInterval`
- `memoryAggregationIntervalCount`
- `evictAfterOOMSeconds`

Validation and error handling will ensure that invalid or conflicting values are caught early, either through CEL rules or admission controller logic.

### API Changes

#### Phase 1 (Current Proposal)

Extend `ContainerResourcePolicy` with:
* `oomBumpUpRatio`
* `oomMinBumpUp`
* `memoryAggregationInterval`
* `memoryAggregationIntervalCount`

Extend `PodUpdatePolicy` with:
* `evictAfterOOMSeconds`

#### Future Extensions

This AEP will be updated as additional parameters are identified for per-object configuration. Potential candidates include:
* `confidenceIntervalCPU`
* `confidenceIntervalMemory`
* `recommendationMarginFraction`
* Other parameters that benefit from workload-specific tuning

### Feature Enablement and Rollback

#### How can this feature be enabled / disabled in a live cluster?

- Feature gate name: `PerVPAConfig`
- Default: Off (Alpha)
- Components depending on the feature gate:
  - admission-controller
  - recommender
  - updater

The feature gate will remain in alpha (default off) until:
- All planned configuration parameters have been implemented and tested
- Performance impact has been thoroughly evaluated
- Documentation is complete for all parameters

Disabling of feature gate `PerVPAConfig` will cause the following to happen:

- Any per-VPA configuration parameters specified in VPA objects will be ignored
- Components will fall back to using their global configuration values

Enabling of feature gate `PerVPAConfig` will cause the following to happen:

- VPA components will honor the per-VPA configuration parameters specified in VPA objects
- Validation will be performed on the configuration parameters
- Configuration parameters will override global defaults for the specific VPA object

### Kubernetes version compatibility

The `PerVPAConfig` feature requires VPA version 1.5.0 or higher. The feature is being introduced as alpha and will follow the standard Kubernetes feature gate graduation process:
- Alpha: v1.5.0 (default off)
- Beta: TBD (default on)
- GA: TBD (default on)

### Validation via CEL and Testing

Initial validation rules (CEL):
* `oomMinBumpUp` >= 0
* `memoryAggregationInterval` > 0
* `memoryAggregationIntervalCount` > 0
* `evictAfterOOMSeconds` > 0

Validation via Admission Controller:
Some components cann't be validated using Common Expression Language (CEL). This validation is performed within the admission controller.

* `oomBumpUpRatio` – Using Kubernetes Quantity type for validation. The value must be greater than or equal to 1.

Additional validation rules will be added as new parameters are introduced.
E2E tests will be included to verify:
* Different configurations are properly applied and respected by VPA components
* VPA behavior matches expected outcomes for different parameter combinations
* Proper fallback to global configuration when parameters are not specified

### Test Plan

- Unit tests for new API fields and validation
- Integration tests verifying different configurations are properly applied
- E2E tests comparing behavior with different configurations
- Upgrade tests ensuring backward compatibility

## Implementation History

- 2025-04-12: Initial proposal
- 2025-09-06: Specify `oomBumpUpRatio` and `oomMinBumpUp` as container-level parameters
- Future: Additional parameters will be added based on user feedback and requirements

## Future Work

This enhancement is designed to be extensible. As the VPA evolves and users provide feedback, additional parameters may be added to the per-object configuration. Each new parameter will:
1. Be documented in this AEP
2. Include appropriate validation rules
3. Maintain backward compatibility
4. Follow the same pattern of falling back to global configuration when not specified

The decision to add new parameters will be based on:
- User feedback and requirements
- Performance impact analysis
- Implementation complexity
- Maintenance considerations

## Alternatives

### Multiple VPA Deployments

Continue with current approach of running multiple VPA deployments with different configurations:
- Pros: No API changes needed
- Cons: Higher resource usage, operational complexity

### Environment-Specific Configuration

Use different VPA deployments per environment (dev/staging/prod):
- Pros: Simpler than per-workload configuration
- Cons: Less flexible, doesn't address varying needs within same environment
