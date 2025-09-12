# Vertical Pod Autoscaler FAQ

## Contents

- [VPA restarts my pods but does not modify CPU or memory settings. Why?](#vpa-restarts-my-pods-but-does-not-modify-cpu-or-memory-settings)
- [How can I apply VPA to my Custom Resource?](#how-can-i-apply-vpa-to-my-custom-resource)
- [How can I use Prometheus as a history provider for the VPA recommender?](#how-can-i-use-prometheus-as-a-history-provider-for-the-vpa-recommender)
- [I get recommendations for my single pod replicaSet, but they are not applied. Why?](#i-get-recommendations-for-my-single-pod-replicaset-but-they-are-not-applied)
- [Can I run the VPA in an HA configuration?](#can-i-run-the-vpa-in-an-ha-configuration)
- [What are the parameters to VPA recommender?](#what-are-the-parameters-to-vpa-recommender)
- [What are the parameters to VPA updater?](#what-are-the-parameters-to-vpa-updater)
- [What are the parameters to VPA admission-controller?](#what-are-the-parameters-to-vpa-admission-controller)
- [How can I configure VPA to manage only specific resources?](#how-can-i-configure-vpa-to-manage-only-specific-resources)
- [How can I have Pods in the kube-system namespace under VPA control in AKS?](#how-can-i-have-pods-in-the-kube-system-namespace-under-vpa-control-in-aks)
- [How can I configure VPA when running in EKS with Cilium?](#how-can-i-configure-vpa-when-running-in-eks-with-cilium)

### VPA restarts my pods but does not modify CPU or memory settings

First check that the VPA admission controller is running correctly:

```console
$ kubectl get pod -n kube-system | grep vpa-admission-controller
vpa-admission-controller-69645795dc-sm88s            1/1       Running   0          1m
```

Check the logs of the admission controller:

```$ kubectl logs -n kube-system vpa-admission-controller-69645795dc-sm88s```

If the admission controller is up and running, but there is no indication of it
actually processing created pods or VPA objects in the logs, the webhook is not registered correctly.

Check the output of:

```$ kubectl describe mutatingWebhookConfiguration vpa-webhook-config```

This should be correctly configured to point to the VPA admission webhook service.
Example:

```yaml
Name:         vpa-webhook-config
Namespace:
Labels:       <none>
Annotations:  <none>
API Version:  admissionregistration.k8s.io/v1beta1
Kind:         MutatingWebhookConfiguration
Metadata:
  Creation Timestamp:  2019-01-18T15:44:42Z
  Generation:          1
  Resource Version:    1250
  Self Link:           /apis/admissionregistration.k8s.io/v1beta1/mutatingwebhookconfigurations/vpa-webhook-config
  UID:                 f8ccd13d-1b37-11e9-8906-42010a84002f
Webhooks:
  Client Config:
    Ca Bundle: <redacted>
    Service:
      Name:        vpa-webhook
      Namespace:   kube-system
  Failure Policy:  Ignore
  Name:            vpa.k8s.io
  Namespace Selector:
  Rules:
    API Groups:

    API Versions:
      v1
    Operations:
      CREATE
    Resources:
      pods
    API Groups:
      autoscaling.k8s.io
    API Versions:
      v1beta1
    Operations:
      CREATE
      UPDATE
    Resources:
      verticalpodautoscalers
```

If the webhook config doesn't exist, something got wrong with webhook
registration for admission controller. Check the logs for more info.

From the above config following part defines the webhook service:

```yaml
Service:
  Name:        vpa-webhook
  Namespace:   kube-system
```

Check that the service actually exists:

```$ kubectl describe -n kube-system service vpa-webhook```

```yaml
Name:              vpa-webhook
Namespace:         kube-system
Labels:            <none>
Annotations:       <none>
Selector:          app=vpa-admission-controller
Type:              ClusterIP
IP:                <some_ip>
Port:              <unset>  443/TCP
TargetPort:        8000/TCP
Endpoints:         <some_endpoint>
Session Affinity:  None
Events:            <none>
```

You can also curl the service's endpoint from within the cluster to make sure it
is serving.

Note: the commands will differ if you deploy VPA in a different namespace.

### How can I apply VPA to my Custom Resource?

The VPA can scale not only the built-in resources like Deployment or StatefulSet, but also Custom Resources which manage
Pods. Just like the Horizontal Pod Autoscaler, the VPA requires that the Custom Resource implements the
[`/scale` subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource)
with the optional field `labelSelector`,
which corresponds to `.scale.status.selector`. VPA doesn't use the `/scale` subresource for the actual scaling, but uses
this label selector to identify the Pods managed by a Custom Resource. As VPA relies on Pod eviction to apply new
resource recommendations, this ensures that all Pods with a matching VPA object are managed by a controller that will
recreate them after eviction. Furthermore, it avoids misconfigurations that happened in the past when label selectors
were specified manually.

### How can I use Prometheus as a history provider for the VPA recommender

Configure your Prometheus to get metrics from cadvisor. Make sure that the metrics from the cadvisor have the label `job=kubernetes-cadvisor`

Set the flags `--storage=prometheus` and `--prometheus-address=<your-prometheus-address>` in the deployment for the `VPA recommender`. The `args` for the container should look something like this:

```yaml
spec:
  containers:
  - args:
    - --v=4
    - --storage=prometheus
    - --prometheus-address=http://prometheus.default.svc.cluster.local:9090
```

In this example, Prometheus is running in the default namespace.

Now deploy the `VPA recommender` and check the logs.

```$ kubectl logs -n kube-system vpa-recommender-bb655b4b9-wk5x2```

Here you should see the flags that you set for the VPA recommender and you should see:
```Initializing VPA from history provider```

This means that the VPA recommender is now using Prometheus as the history provider.


For authentication to Prometheus, you can provide credentials in following ways:

1) Set the flags `--username=<user>` and `--password=<password>` in the `VPA recommender deployment`. The `args` for the container should look something like this:

```yaml
spec:
  containers:
  - args:
    - --v=4
    - --storage=prometheus
    - --prometheus-address=http://prometheus.default.svc.cluster.local:9090
    - --username=example-user
    - --password=example-password
```

2) Set the environment variables `PROMETHEUS_USERNAME` and `PROMETHEUS_PASSWORD` in the `VPA recommender deployment`.

```yaml
spec:
  containers:
  - args:
    - --storage=prometheus
    - --prometheus-address=http://prometheus.default.svc.cluster.local:9090
  env:
  - name: PROMETHEUS_USERNAME
    valueFrom:
      secretKeyRef:
        name: prometheus-auth
        key: example-user
  - name: PROMETHEUS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: prometheus-auth
        key: example-password
```

3) Set the flag `prometheus-bearer-token=<token>`, to use bearer token auth.

```yaml
spec:
  containers:
  - args:
    - --v=4
    - --storage=prometheus
    - --prometheus-address=http://prometheus.default.svc.cluster.local:9090
    - --prometheus-bearer-token=<example-token>
```

### I get recommendations for my single pod replicaset but they are not applied

By default, the [`--min-replicas`](https://github.com/kubernetes/autoscaler/tree/master/pkg/updater/main.go#L44) flag on the updater is set to 2. To change this, you can supply the arg in the [deploys/updater-deployment.yaml](https://github.com/kubernetes/autoscaler/tree/master/deploy/updater-deployment.yaml) file:

```yaml
spec:
  containers:
  - name: updater
    args:
    - "--min-replicas=1"
    - "--v=4"
    - "--stderrthreshold=info"
```

and then deploy it manually if your vpa is already configured.

### Can I run the VPA in an HA configuration?

The VPA admission-controller can be run with multiple active Pods at any given time.

Both the updater and recommender can only run a single active Pod at a time. Should you
want to run a Deployment with more than one pod, it's recommended to enable a lease
election with the `--leader-elect=true` parameter.

**NOTE**: If using GKE, you must set `--leader-elect-resource-name` to something OTHER than "vpa-recommender", for example "vpa-recommender-lease".

### What are the parameters to VPA recommender?

See the [full list of parameters in the VPA recommender](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/flags.md#what-are-the-parameters-to-vpa-recommender).

### What are the parameters to VPA updater?

See the [full list of parameters in the VPA updater](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/flags.md#what-are-the-parameters-to-vpa-updater).

### What are the parameters to VPA admission controller?

See the [full list of parameters in the VPA admission controller](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/flags.md#what-are-the-parameters-to-vpa-admission-controller).

### How can I configure VPA to manage only specific resources?

You can configure VPA to manage only specific resources (CPU or memory) using the controlledResources field in the resourcePolicy section of your VPA configuration. This is particularly useful when you want to:

* Combine VPA with HPA without resource conflicts
* Focus VPA's management on specific resource types
* Implement separate scaling strategies for different resources

Example configuration:
```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: resource-specific-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: "Recreate"  # Use explicit mode instead of deprecated "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      controlledResources: ["memory"]  # Only manage memory resources
```

The controlledResources field accepts the following values:
* ["cpu"] - VPA will only manage CPU resources
* ["memory"] - VPA will only manage memory resources
* ["cpu", "memory"] - VPA will manage both resources (default behavior)

Common use cases:
1. Memory-only VPA with CPU-based HPA:
* Configure VPA to manage only memory using controlledResources: ["memory"]
* Set up HPA to handle CPU-based scaling
* This prevents conflicts between VPA and HPA

2. CPU-only VPA:
* Use controlledResources: ["cpu"] when you want to automate CPU resource allocation
* Useful when memory requirements are stable but CPU usage varies

### How can I have Pods in the kube-system namespace under VPA control in AKS?

When running a webhook in AKS, it blocks webhook requests for the kube-system namespace in order to protect the system.
See the [AKS FAQ page](https://learn.microsoft.com/en-us/azure/aks/faq#can-admission-controller-webhooks-impact-kube-system-and-internal-aks-namespaces-) for more info.

The `--webhook-labels` parameter for the VPA admission-controller can be used to bypass this behaviour, if required by the user.

### How can I configure VPA when running in EKS with Cilium?

When running in EKS with Cilium, the EKS API server cannot route traffic to the overlay network. The VPA admission-controller
Pods either need to use host networking or be exposed through a service or ingress.
See the [Cilium Helm installation page](https://docs.cilium.io/en/stable/installation/k8s-install-helm/) for more info.
