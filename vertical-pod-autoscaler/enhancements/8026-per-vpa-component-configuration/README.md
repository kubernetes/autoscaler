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
  - oomBumpUpRatio
  - oomMinBumpUp
  - memoryAggregationInterval
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
  name: oom-test-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: oom-test
  updatePolicy:
    updateMode: Auto
    evictAfterOOMThreshold: "5m"
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      oomBumpUpRatio: 1.5
      oomMinBumpUp: 104857600
      memoryAggregationInterval: "12h"
```

### Parameter Descriptions

#### Container Policy Parameters
* `oomBumpUpRatio` (float): 
  - Multiplier applied to memory recommendations after OOM events
  - Controls how aggressively memory is increased after container crashes

* `oomMinBumpUp` (bytes): 
  - Minimum absolute memory increase after OOM events
  - Ensures meaningful increases even for small containers

* `memoryAggregationInterval` (duration):
  - Time window for aggregating memory usage data
  - Affects how quickly VPA responds to memory usage changes

#### Update Policy Parameters
* `evictAfterOOMThreshold` (duration):
  - Time to wait after OOM before considering pod eviction
  - Helps prevent rapid eviction cycles while maintaining stability

Each parameter can be configured independently, falling back to global defaults if not specified. Values should be chosen based on workload characteristics and stability requirements.

## Design Details

### API Changes

#### Phase 1 (Current Proposal)

Extend `ContainerResourcePolicy` with:
* `oomBumpUpRatio`
* `oomMinBumpUp`
* `memoryAggregationInterval`

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
- Components depending on the feature gate:
  - admission-controller
  - recommender
  - updater

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

Initial validation rules:
* oomBumpUpRatio > 1.0
* oomMinBumpUp > 0
* memoryAggregationInterval > 0
* evictAfterOOMThreshold > 0

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