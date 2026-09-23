# Quick start

## Contents

<!-- toc -->
- [Test your installation](#test-your-installation)
- [Example VPA configuration](#example-vpa-configuration)
- [Troubleshooting](#troubleshooting)
<!-- /toc -->

After [installation](./installation.md) the system is ready to recommend and set
resource requests for your pods.
In order to use it, you need to insert a *Vertical Pod Autoscaler* resource for
each controller that you want to have automatically computed resource requirements.
This will be most commonly a **Deployment**.
There are several modes in which *VPAs* operate:

- `"Auto"` [__deprecated__]: VPA assigns resource requests on pod creation as well as updates
  them on existing pods using the preferred update mechanism. Currently, this is
  equivalent to `"Recreate"` (see below). **This mode is deprecated and will be removed in a future API version.**
  **Use explicit modes like "Recreate", "Initial", or "InPlaceOrRecreate" instead.**
- `"Recreate"` [__default__]: VPA assigns resource requests on pod creation as well as updates
  them on existing pods by evicting them when the requested resources differ significantly
  from the new recommendation (respecting the Pod Disruption Budget, if defined).
  This mode should be used rarely, only if you need to ensure that the pods are restarted
  whenever the resource request changes.
- `"InPlaceOrRecreate"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods by leveraging [Kubernetes `in-place` update](https://kubernetes.io/blog/2025/05/16/kubernetes-v1-33-in-place-pod-resize-beta/) capability.
  If `in-place` update fails, it falls back to evicting the pods, performing a _recreation_.
  For more details, see the [In-Place Updates documentation](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/features.md#in-place-updates-inplaceorrecreate).
- `"InPlace"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods using only Kubernetes `in-place` pod resize capability.
  Unlike `"InPlaceOrRecreate"`, this mode never evicts pods. If an `in-place`
  resize cannot be performed, VPA retries the update later when cluster conditions
  change. This mode is recommended for workloads where any disruption is unacceptable.
- `"Initial"`: VPA only assigns resource requests on pod creation and never changes them
  later.
- `"Off"`: VPA does not automatically change the resource requirements of the pods.
  The recommendations are calculated and can be inspected in the VPA object.

## Test your installation

A simple way to check if Vertical Pod Autoscaler is fully operational in your
cluster is to create a sample deployment and a corresponding VPA config:

```console
kubectl apply -f - <<EOF
apiVersion: "autoscaling.k8s.io/v1"
kind: VerticalPodAutoscaler
metadata:
  name: hamster-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: hamster
  resourcePolicy:
    containerPolicies:
      - containerName: '*'
        minAllowed:
          cpu: 100m
          memory: 50Mi
        maxAllowed:
          cpu: 1
          memory: 500Mi
        controlledResources: ["cpu", "memory"]
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hamster
spec:
  selector:
    matchLabels:
      app: hamster
  replicas: 2
  template:
    metadata:
      labels:
        app: hamster
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      containers:
        - name: hamster
          image: registry.k8s.io/ubuntu-slim:0.14
          resources:
            requests:
              cpu: 100m
              memory: 50Mi
          command: ["/bin/sh"]
          args:
            - "-c"
            - "while true; do timeout 0.5s yes >/dev/null; sleep 0.5s; done"
EOF
```

The above command creates a deployment with two pods, each running a single container
that requests 100 millicores and tries to utilize slightly above 500 millicores.
The command also creates a VPA config pointing at the deployment.
VPA will observe the behaviour of the pods, and after about 5 minutes, they should get
updated with a higher CPU request
(note that VPA does not modify the template in the deployment, but the actual requests
of the pods are updated). To see VPA config and current recommended resource requests run:

```console
kubectl describe vpa
```

*Note: if your cluster has little free capacity these pods may be unable to schedule.
You may need to add more nodes or adjust the above deployment to use less CPU.*

## Example VPA configuration

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind:       Deployment
    name:       my-app
  updatePolicy:
    updateMode: "Recreate"  # Use explicit mode instead of deprecated "Auto"
```

## Troubleshooting

To diagnose problems with a VPA installation, perform the following steps:

- Check if all system components are running:

```console
kubectl --namespace=kube-system get pods|grep vpa
```

The above command should list 3 pods (recommender, updater and admission-controller)
all in state Running.

- Check if the system components log any errors.
  For each of the pods returned by the previous command do:

```console
kubectl --namespace=kube-system logs [pod name] | grep -e '^E[0-9]\{4\}'
```

- Check that the VPA Custom Resource Definition was created:

```console
kubectl get customresourcedefinition | grep verticalpodautoscalers
```
