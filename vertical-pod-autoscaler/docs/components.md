# Components

## Contents

- [Components](#components)
  - [Introduction](#introduction)
  - [Recommender](#recommender)
    - [Running](#running-the-recommender)
    - [Implementation](#implementation-of-the-recommender)
  - [Updater](#updater)
    - [Current implementation](#current-implementation)
    - [Missing Parts](#missing-parts)
  - [Admission Controller](#admission-controller)
    - [Running](#running-the-admission-controller)
    - [Implementation](#implementation-of-the-admission-controller)

## Introduction

The VPA project consists of 3 components:

- [Recommender](#recommender) - monitors the current and past resource consumption and, based on it,
  provides recommended values for the containers' cpu and memory requests.

- [Updater](#updater) - checks which of the managed pods have correct resources set and, if not,
  kills them so that they can be recreated by their controllers with the updated requests.

- [Admission Controller](#admission-controller) - sets the correct resource requests on new pods (either just created
  or recreated by their controller due to Updater's activity).

For detailed information about configuration parameters for each component, see the [flags documentation](flags.md).

More on the architecture can be found [HERE](https://github.com/kubernetes/design-proposals-archive/blob/main/autoscaling/vertical-pod-autoscaler.md).

## Recommender

Recommender is the core binary of Vertical Pod Autoscaler system.
It computes the recommended resource requests for pods based on
historical and current usage of the resources.
The current recommendations are put in status of the VPA resource, where they
can be inspected.

## Running the recommender

- By default the VPA will use a VerticalPodAutoscalerCheckpoint to store history, but
  it is possible to fetch historical from Prometheus in your cluster.
- Create RBAC configuration from `../deploy/vpa-rbac.yaml`.
- Create a deployment with the recommender pod from
  `../deploy/recommender-deployment.yaml`.
- The recommender will start running and pushing its recommendations to VPA
  object statuses.

### Implementation of the recommender

The recommender is based on a model of the cluster that it builds in its memory.
The model contains Kubernetes resources: *Pods*, *VerticalPodAutoscalers*, with
their configuration (e.g. labels) as well as other information, e.g. usage data for
each container.

After starting the binary, recommender reads the history of running pods and
their usage from VerticalPodAutoscalerCheckpoint (or Prometheus) into the model.
It then runs in a loop and at each step performs the following actions:

- update model with recent information on resources (using listers based on
  watch),
- update model with fresh usage samples from Metrics API,
- compute new recommendation for each VPA,
- put any changed recommendations into the VPA resources.

## Updater

Updater component for Vertical Pod Autoscaler described in the [Vertical Pod Autoscaler - design proposal](https://github.com/kubernetes/community/pull/338)

Updater runs in Kubernetes cluster and decides which pods should be restarted
based on resources allocation recommendation calculated by Recommender.
If a pod should be updated, Updater will try to evict the pod.
It respects the pod disruption budget, by using Eviction API to evict pods.
Updater does not perform the actual resources update, but relies on Vertical Pod Autoscaler admission plugin
to update pod resources when the pod is recreated after eviction.

### Current implementation

Runs in a loop. On one iteration performs:

- Fetching Vertical Pod Autoscaler configuration using a lister implementation.
- Fetching live pods information with their current resource allocation.
- For each replicated pods group calculating if pod update is required and how many replicas can be evicted.
Updater will always allow eviction of at least one pod in replica set. Maximum ratio of evicted replicas is specified by flag.
- Evicting pods if recommended resources significantly vary from the actual resources allocation.
Threshold for evicting pods is specified by recommended min/max values from VPA resource.
Priority of evictions within a set of replicated pods is proportional to sum of percentages of changes in resources
(i.e. pod with 15% memory increase 15% cpu decrease recommended will be evicted
before pod with 20% memory increase and no change in cpu).

### Missing parts

- Recommendation API for fetching data from Vertical Pod Autoscaler Recommender.

## Admission-controller

This is a binary that registers itself as a Mutating Admission Webhook
and because of that is on the path of creating all pods.
For each pod creation, it will get a request from the apiserver and it will
either decide there's no matching VPA configuration or find the corresponding
one and use current recommendation to set resource requests in the pod.

### Running the admission-controller

1. You should make sure your API server supports Mutating Webhooks.
Its `--admission-control` flag should have `MutatingAdmissionWebhook` as one of
the values on the list and its `--runtime-config` flag should include
`admissionregistration.k8s.io/v1beta1=true`.
To change those flags, ssh to your API Server instance, edit
`/etc/kubernetes/manifests/kube-apiserver.manifest` and restart kubelet to pick
up the changes: ```sudo systemctl restart kubelet.service```
1. Generate certs by running `bash gencerts.sh`. This will use kubectl to create
   a secret in your cluster with the certs.
1. Create RBAC configuration for the admission controller pod by running
   `kubectl create -f ../deploy/admission-controller-rbac.yaml`
1. Create the pod:
   `kubectl create -f ../deploy/admission-controller-deployment.yaml`.
   The first thing this will do is it will register itself with the apiserver as
   Webhook Admission Controller and start changing resource requirements
   for pods on their creation & updates.
1. You can specify a path for it to register as a part of the installation process
   by setting `--register-by-url=true` and passing `--webhook-address` and `--webhook-port`.
1. You can specify a minimum TLS version with `--min-tls-version` with acceptable values being `tls1_2` (default), or `tls1_3`.
1. You can also specify a comma or colon separated list of ciphers for the server to use with `--tls-ciphers` if `--min-tls-version` is set to `tls1_2`.
1. You can specify a comma separated list to set webhook labels with `--webhook-labels`, example format: key1:value1,key2:value2.

### Implementation of the Admission Controller

All VPA configurations in the cluster are watched with a lister.
In the context of pod creation, there is an incoming https request from
apiserver.
The logic to serve that request involves finding the appropriate VPA, retrieving
current recommendation from it and encodes the recommendation as a json patch to
the Pod resource.
