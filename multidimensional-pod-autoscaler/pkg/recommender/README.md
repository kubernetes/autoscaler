# MPA Recommender

- [Intro](#intro)
- [Running](#running)
- [Implementation](#implementation)

## Intro

Recommender is the core binary of Multi-dimensiional Pod Autoscaler (MPA) system.
It consists of both vertical and horizontal scaling of resources:
- Vertical: It computes the recommended resource requests for pods based on historical and current usage of the resources. Like VPA, the current recommendations are put in status of the MPA object, where they can be inspected.
- Horizontal: It updates the number of replicas based on specified target metrics threshold according to the following formula:

```
desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]
```

- Combined: To be released.
  - The current way of combining vertical and horizontal scaling is simple: Each dimension is alternatively being considered. In the future, we will design and implement prioritization (e.g., to prioritize horizontal scaling for CPU-instensive workloads) and conflict-resolving (e.g., scaling in and up simultaneuously) mechanisms.

## Running

* In order to have historical data pulled in by the recommender, install Prometheus in your cluster and pass its address through a flag.
* Create RBAC configuration from `../../deploy/mpa-rbac.yaml` if not yet.
* Create a deployment with the recommender pod from `../../deploy/recommender-deployment.yaml`.
* The recommender will start running and pushing its recommendations to MPA object statuses.

## Implementation

The recommender is based on a model of the cluster that it builds in its memory.
The model contains Kubernetes resources: *Pods*, *MultidimPodAutoscalers*, with their configuration (e.g. labels) as well as other information, e.g., usage data for each container.

After starting the binary, the recommender reads the history of running pods and their usage from Prometheus into the model.
It then runs in a loop and at each step performs the following actions:

* update model with recent information on resources (using listers based on watch),
* update model with fresh usage samples from Metrics API,
* compute new recommendation for each MPA,
* put any changed recommendations into the MPA objects.

## Building the Docker Image

```
make build-binary-with-vendor-amd64
make docker-build-amd64
```
