# VPA Sidecar Container Management

In this document, "sidecar container" refers to any additional Container that isn't the main application Container in a Pod. This is distinct from the [native Kubernetes sidecar pattern](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/), which makes use of `initContainers`. Our usage here applies to all additional regular `containers` only, as VPA does not support `initContainers` yet.

The Vertical Pod Autoscaler (VPA) has specific behavior when dealing with these additional containers that are injected into pods via admission webhooks.

## Understanding VPA and Container Policies

### Default Container Policies

To understand why sidecar handling is important, let's first look at how VPA manages containers by default. This default behavior is at the root of why special consideration is needed for sidecars.

When you create a VPA resource for a pod, it automatically attempts to manage ALL containers in that pod - not just the ones you explicitly configure. This happens because:

1. VPA applies a default `containerPolicy` with `mode: Auto` to any container not explicitly configured
2. This automatic inclusion means VPA will try to manage resources for every container it sees
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

### The vpaObservedContainers Annotation

VPA uses a special annotation to track which containers were present in the pod before any webhook injections:

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
