# VPA Recommender

- [Intro](#intro)
- [Running](#running)
- [Implementation](#implmentation)
## Intro

Recommender is the core binary of Vertical Pod Autoscaler system.
It computes the recommended resource requests for pods based on
historical and current usage of the resources.
The current recommendations are put in status of the VPA resource, where they
can be inspected.

## Running

* In order to have historical data pulled in by the recommender, install
  Prometheus in your cluster and pass its address through a flag.
* Create RBAC configuration from `../deploy/vpa-rbac.yaml`.
* Create a deployment with the recommender pod from
  `../deploy/recommender-deployment.yaml`.
* The recommender will start running and pushing its recommendations to VPA
  object statuses.

## Implementation

The recommender is based on a model of the cluster that it builds in its memory.
The model contains Kubernetes resources: *Pods*, *VerticalPodAutoscalers*, with
their configuration (e.g. labels) as well as other information, e.g. usage data for
each container.

After starting the binary, recommender reads the history of running pods and
their usage from Prometheus into the model.
It then runs in a loop and at each step performs the following actions:

* update model with recent information on resources (using listers based on
  watch),
* update model with fresh usage samples from Metrics API,
* compute new recommendation for each VPA,
* put any changed recommendations into the VPA resources.
