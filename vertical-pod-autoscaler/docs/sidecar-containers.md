# VPA Sidecar Container Management

<!-- toc -->
- [Understanding VPA and Container Policies](#understanding-vpa-and-container-policies)
  - [Default Container Policies](#default-container-policies)
- [Default Behavior: Ignoring Sidecar Containers](#default-behavior-ignoring-sidecar-containers)
  - [The vpaObservedContainers Annotation](#the-vpaobservedcontainers-annotation)
  - [Webhook Ordering Importance](#webhook-ordering-importance)
  - [The Eviction Loop Problem](#the-eviction-loop-problem)
- [Native Kubernetes Sidecars](#native-kubernetes-sidecars)
- [Customizing VPA Behavior for Sidecar Containers](#customizing-vpa-behavior-for-sidecar-containers)
  - [Option 1: Webhook Ordering](#option-1-webhook-ordering)
  - [Option 2: Webhook Reinvocation](#option-2-webhook-reinvocation)
- [Implementation Examples](#implementation-examples)
  - [Example 1: Istio Sidecar Integration](#example-1-istio-sidecar-integration)
  - [Example 2: Custom VPA Configuration with Sidecars](#example-2-custom-vpa-configuration-with-sidecars)
<!-- /toc -->

This document distinguishes additional regular `containers` (often injected by a webhook) from [native Kubernetes sidecars](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/), which are `initContainers` with `restartPolicy: Always`. The webhook-ordering guidance below applies to injected containers; native sidecar support is described separately.

The Vertical Pod Autoscaler (VPA) handles webhook-injected regular containers and native sidecars differently.

## Understanding VPA and Container Policies

### Default Container Policies

To understand why sidecar handling is important, let's first look at how VPA manages containers by default. This default behavior is at the root of why special consideration is needed for sidecars.

When you create a VPA resource for a pod, it automatically attempts to manage all eligible containers in that pod - not just the ones you explicitly configure. This happens because:

1. VPA applies a default `containerPolicy` with `mode: Auto` to any container not explicitly configured
2. This automatic inclusion means VPA will try to manage resources for every eligible container it sees
3. Even sidecars injected into the pod will fall under VPA's management unless special steps are taken

This default "manage everything" approach can cause problems with sidecars because:
- Sidecar containers often have their own resource requirements set by their injection webhooks
- VPA's automatic management may conflict with these requirements
- Without proper handling, this can lead to problems like endless eviction loops

Example of a VPA resource that explicitly configures one container but affects all:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: main-container
      minAllowed:
        cpu: 100m
        memory: 50Mi
# Note: Other containers will still get the default Auto mode
```

## Default Behavior: Ignoring Sidecar Containers

This section concerns webhook-injected regular containers that VPA did not observe at pod admission time, not all sidecars.

### The vpaObservedContainers Annotation

VPA uses a special annotation to track which containers were present when its admission webhook ran:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    vpaObservedContainers: "main-container,logging-sidecar"
```

This annotation is crucial because:
1. It's added by the VPA admission controller webhook
2. Only containers listed in this annotation will be managed by VPA
3. The annotation must be added before sidecar injection for the default behavior to work

### Webhook Ordering Importance

The order of webhook execution is determined alphabetically by webhook names. For example:

```yaml
webhooks:
- name: a.sidecar-injector.kubernetes.io  # Executes first
- name: b.vpa.kubernetes.io               # Executes second
- name: c.another-sidecar.kubernetes.io   # Executes third
```

### The Eviction Loop Problem

Without proper handling of sidecar containers, the following problematic sequence could occur:

1. VPA admission controller sets resources for all containers
2. Sidecar webhook reconciles the injected sidecar container to the original resource requirements
3. The pod starts with mismatched resources
4. VPA detects the mismatch and evicts the pod
5. The process repeats, creating an endless loop

## Native Kubernetes Sidecars

Starting with VPA v1.9, native sidecar support is an alpha feature, disabled by default. Enable `--feature-gates=NativeSidecar=true` on the admission controller, recommender, and updater to manage `initContainers` whose `restartPolicy` is `Always`. Plain init containers are **not** managed, even when the gate is enabled. With the gate disabled, VPA continues to ignore all init containers.

For example, in a pod targeted by a VPA, the gate allows VPA to recommend and apply resources for `log-agent` as well as `app`, but not for `setup`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: native-sidecar-example
spec:
  initContainers:
  - name: setup
    image: registry.k8s.io/pause:3.9
  - name: log-agent
    image: registry.k8s.io/pause:3.9
    restartPolicy: Always
  containers:
  - name: app
    image: registry.k8s.io/pause:3.9
```

The admission controller can apply recommendations to a native sidecar on pod creation. The updater also considers native sidecars for updates and eviction decisions. As with regular containers, an injected native sidecar must be present when the VPA webhook runs to appear in `vpaObservedContainers`; the ordering considerations above still matter for injected sidecars.

## Customizing VPA Behavior for Sidecar Containers

### Option 1: Webhook Ordering

If you know that you only use sidecar injecting webhooks which _don't_ reconcile Container resources, you can choose to have VPA manage sidecar resources. Ensure your webhook names follow this pattern, resulting in the VPA admission-controller webhook to be executed last:

```yaml
webhooks:
- name: sidecar-injector.kubernetes.io    # Executes first
- name: zz.vpa.kubernetes.io              # Executes last
```

This ensures:
1. Sidecars are injected first
2. VPA sees the complete pod with all sidecars
3. The `vpaObservedContainers` annotation includes all containers

### Option 2: Webhook Reinvocation

Configure the VPA webhook to reinvoke after sidecar injection:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: vpa-webhook-config
webhooks:
- name: vpa.kubernetes.io
  reinvocationPolicy: IfNeeded
  rules:
    - apiGroups: [""]
      apiVersions: ["v1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["pods"]
```

This configuration:
1. Allows the VPA webhook to be called multiple times
2. Ensures resource recommendations are applied after all sidecars are injected
3. Prevents the eviction loop problem

## Implementation Examples

### Example 1: Istio Sidecar Integration

When using Istio, which injects proxy sidecars:

```yaml
# Istio sidecar injector webhook
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: istio-sidecar-injector
webhooks:
- name: a.sidecar-injector.istio.io  # Alphabetically first

# VPA webhook configuration
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: vpa-webhook-config
webhooks:
- name: z.vpa.kubernetes.io          # Alphabetically last
  reinvocationPolicy: IfNeeded
```

### Example 2: Custom VPA Configuration with Sidecars

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: web-app-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: web-app
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: main-container
      minAllowed:
        cpu: 100m
        memory: 128Mi
    - containerName: logging-sidecar  # Explicitly configure sidecar
      minAllowed:
        cpu: 50m
        memory: 64Mi
```
