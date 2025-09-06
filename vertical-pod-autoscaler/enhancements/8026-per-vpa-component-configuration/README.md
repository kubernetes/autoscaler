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
  - [API Changes](#api-changes)
    - [OOM Parameter Consolidation](#oom-parameter-consolidation)
    - [Parameter Coexistence and Global Configuration](#parameter-coexistence-and-global-configuration)
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

Currently, VPA components (recommender, updater, admission controller) are configured through global flags. This makes it challenging to support different workloads with varying resource optimization needs within the same cluster. This proposal introduces the ability to specify configuration parameters at the individual VPA object level, allowing for workload-specific optimization strategies.

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
  - oomMinBumpUp
  - memoryAggregationInterval
  - memoryAggregationIntervalCount
  - evictAfterOOMThreshold

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
    evictAfterOOMThreshold: "5m"
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      oomBumpUp: "1.5"
      memoryAggregationInterval: "12h"
      memoryAggregationIntervalCount: 5
```

### Parameter Descriptions

#### Container Policy Parameters
* `oomBumpUp` (Quantity):
  - Controls memory increase after OOM events
  - Can be specified as either:
    - A ratio (e.g., "1.5" for 50% increase)
    - An absolute value (e.g., "100Mi" for fixed increase)
  - When specified, overrides global OOM configuration
  - Allows fine-tuning of OOM behavior per container
  
* `memoryAggregationInterval` (duration):
  - Time window for aggregating memory usage data
  - Affects how quickly VPA responds to memory usage changes

* `memoryAggregationIntervalCount` (integer):
  - Number of consecutive memory aggregation intervals
  - Used to calculate the total memory aggregation window length
  - Total window length = memoryAggregationInterval * memoryAggregationIntervalCount

#### Update Policy Parameters
* `evictAfterOOMThreshold` (duration):
  - Time to wait after OOM before considering pod eviction
  - Helps prevent rapid eviction cycles while maintaining stability

Each parameter can be configured independently, falling back to global defaults if not specified. Values should be chosen based on workload characteristics and stability requirements.

## Design Details

### Configuration Level Considerations

When designing the configuration parameters, we analyzed each parameter to determine the most appropriate configuration level:

#### Parameter Level Analysis

1. **OOM-Related Parameters (`oomBumpUp`)**
- **Recommended Level**: Container-level
- **Rationale**:
  - Different containers in the same pod may have different memory requirements and OOM patterns
  - Memory recommendations in VPA are already handled at container level
  - Consistent with how VPA handles other resource-related configurations
  - While `oomBumpUp` is container-level, VPA-wide configuration can still be achieved using the wildcard container name "*"

2. **Memory Aggregation Parameters (`memoryAggregationInterval`, `memoryAggregationIntervalCount`)**
- **Recommended Level**: Container-level
- **Rationale**:
  - Different containers may have different memory usage patterns
  - Some containers might need faster reaction times to memory changes
  - Consistent with existing VPA container-level resource policies
  - While `memoryAggregationInterval` and `memoryAggregationIntervalCount` are container-level, VPA-wide configuration can still be achieved using the wildcard container name "*"

3. **Eviction Parameters (`evictAfterOOMThreshold`)**
- **Recommended Level**: VPA-level
- **Rationale**:
  - Eviction decisions affect the entire pod
  - Pod-level operation that shouldn't vary by container
  - Simpler operational model for pod lifecycle management
  - Consistent with how Kubernetes handles pod evictions


### API Changes

### OOM Parameter Consolidation

In this proposal, we are consolidating `oomBumpUpRatio` and `oomMinBumpUp` into a single `oomBumpUp` parameter for several reasons:
1. **Simplified Configuration**:
   - Instead of managing two separate parameters that affect the same behavior, users can specify their intent with a single parameter
   - Makes configuration more intuitive and less error-prone

2. **Flexible Expression**:
   - `oomBumpUp` can express both ratio-based and absolute increases:
     - Ratio format: "1.5" (50% increase)
     - Absolute format: "100Mi" (100 MiB increase)
   - This flexibility eliminates the need for separate parameters while maintaining all functionality

3. **Consistent with Kubernetes Patterns**:
   - Uses the standard Kubernetes Quantity type
   - Similar to how other Kubernetes components handle resource specifications

4. **Backward Compatibility**:
   - The new parameter can express all use cases covered by the previous parameters
   - Migration path is straightforward:
     - For ratio-based increases: Use the same value in `oomBumpUp`
     - For minimum increases: Convert to absolute quantity in `oomBumpUp`

### Parameter Coexistence and Global Configuration

The VPA recommender maintains global configuration values for OOM-related parameters. The new `oomBumpUp` parameter interacts with these global configurations as follows:

1. **Priority Order**:
   - If `oomBumpUp` is specified in the VPA object, it takes precedence over any global configuration
   - If `oomBumpUp` is not specified, the global configuration values from the recommender will be used
   - This allows for a smooth transition and maintains backward compatibility

2. **Validation**:
   - The admission controller can only validate the syntax and range of `oomBumpUp` when specified
   - Cannot validate against global configuration as these values live in the recommender
   - Values must be either:
     - A ratio >= 1 (e.g., "1.5")
     - An absolute memory value > 0 (e.g., "100Mi")

3. **Example Scenarios**:
   ```yaml
   # Scenario 1: oomBumpUp specified - overrides global config
   containerPolicies:
   - containerName: "*"
     oomBumpUp: "1.5"

   # Scenario 2: No oomBumpUp - uses global config from recommender
   containerPolicies:
   - containerName: "*"


#### Phase 1 (Current Proposal)

Extend `ContainerResourcePolicy` with:
* `oomBumpUp`
* `memoryAggregationInterval`
* `memoryAggregationIntervalCount`

Extend `PodUpdatePolicy` with:
* `evictAfterOOMThreshold`

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
* `oomBumpUP` >= 1
* `memoryAggregationInterval` > 0
* `memoryAggregationIntervalCount` > 0
* `evictAfterOOMThreshold` > 0

Validation via Admission Controller:
Some components cann't be validated using Common Expression Language (CEL). This validation is performed within the admission controller.

* `oomBumpUP` – Using Kubernetes Quantity type for validation. The value must be greater than or equal to 1.

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
- 2025-09-06: Consolidated `oomBumpUpRatio` and `oomMinBumpUp` into single `oomBumpUp` parameter
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