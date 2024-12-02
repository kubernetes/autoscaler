# Multi-dimensional Pod Autoscaler (MPA)

## Intro

Multi-dimensional Pod Autoscaler (MPA) combines Kubernetes HPA and VPA so that scaling actions can be considered and actuated together in a holistic manner.
MPA separates the scaling recommendation and actuation completely so that any multi-dimensional autoscaling algorithms can be used as the "recommender".
The default recommender is a simple combination of HPA and VPA algorithm which (1) set the requests automatically based on usage and (2) set the number of replicas based on a target metric.

Same as VPA, MPA is configured with a [Custom Resource Definition object](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
called [MultidimPodAutoscaler](https://github.com/IIDA-Institute/autoscaler/blob/mpa-dev/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1/types.go).
It allows to specify which pods should be vertically and horizontally autoscaled as well as if/how the resource recommendations are applied.

To enable multi-dimensional pod autoscaling on your cluster please follow the installation procedure described below.

## Prerequisites

* `kubectl` should be connected to the cluster you want to install MPA.
* The metrics server must be deployed in your cluster. Read more about [Metrics Server](https://github.com/kubernetes-incubator/metrics-server).
* If you are using a GKE Kubernetes cluster, you will need to grant your current Google
  identity `cluster-admin` role. Otherwise, you won't be authorized to grant extra
  privileges to the MPA system components.
  ```console
  $ gcloud info | grep Account    # get current google identity
  Account: [myname@example.org]

  $ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org
  Clusterrolebinding "myname-cluster-admin-binding" created
  ```
* You should make sure your API server supports Mutating Webhooks.
  Its `--admission-control` flag should have `MutatingAdmissionWebhook` as one of
  the values on the list and its `--runtime-config` flag should include
  `admissionregistration.k8s.io/v1beta1=true`.
  To change those flags, ssh to your API Server instance, edit
  `/etc/kubernetes/manifests/kube-apiserver.manifest` and restart kubelet to pick
  up the changes: ```sudo systemctl restart kubelet.service```
* Please upgrade `openssl` to version 1.1.1 or higher (needs to support `-addext` option).

## Installation

To install MPA, please download the source code of MPA
and run the following command inside the `multidimensional-pod-autoscaler` directory:

```
./deploy/mpa-up.sh
```

Note: the script currently reads environment variables: `$REGISTRY` and `$TAG`.
Make sure you leave them unset unless you want to use a non-default version of MPA.

The script issues multiple `kubectl` commands to the
cluster that insert the configuration and start all needed pods
in the `kube-system` namespace. It also generates
and uploads a secret (a CA cert) used by MPA Admission Controller when communicating
with the API server.

## Tearing Down

To remove MPA installation:

```
./deploy/mpa-down.sh
```

## Quick Start

After [installation](#installation) the system is ready to recommend and set
resource requests for your pods.
In order to use it, you need to insert a *Multidimensional Pod Autoscaler* resource for
each controller that you want to have automatically computed resource requirements.
This will be most commonly a **Deployment**.
There are four modes in which *MPAs* operate, same as [VPA modes](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#quick-start).

#### Example Deployment

A simple way to check if Multidimensional Pod Autoscaler is fully operational in your
cluster is to create a sample deployment:

```
kubectl create -f examples/hamster.yaml
```

The above command creates a deployment with two pods, each running a single container
that requests 100 millicores and tries to utilize slightly above 200 millicores.

#### Example MPA Config

```
---
apiVersion: "autoscaling.k8s.io/v1alpha1"
kind: MultidimPodAutoscaler
metadata:
  name: hamster-mpa
  namespace: default
spec:
  # recommenders field can be unset when using the default recommender.
  # recommenders: 
  #   - name: 'hamster-recommender'
  scaleTargetRef:
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
  constraints:
    minReplicas: 1
    maxReplicas: 6
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 30
```

Create an MPA:

```
kubectl create -f examples/hamster-mpa.yaml
```

To see MPA config and current recommended resource requests run:

```
kubectl describe mpa
```
