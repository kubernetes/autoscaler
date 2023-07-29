# MPA Updater

- [Introduction](#introduction)
- [Current implementation](current-implementation)
- [Missing parts](#missing-parts)

## Introduction
Updater component for Multidimensional Pod Autoscaler described in https://github.com/kubernetes/community/pull/338 (To be updated)

Updater runs in Kubernetes cluster and decides which pods should be restarted
based on resources allocation recommendation calculated by Recommender.
If a pod should be updated, Updater will try to evict the pod.
It respects the pod disruption budget, by using Eviction API to evict pods.
Updater does not perform the actual resources update, but relies on Multidimensional Pod Autoscaler admission plugin
to update pod resources when the pod is recreated after eviction.

## Running the Updater

* Create RBAC configuration from `../../deploy/mpa-rbac.yaml` if not yet done so.
* Create a deployment with the updater pod from `../../deploy/updater-deployment.yaml`.
* The updater will start running and watch MPA object statuses for pod eviction and replica updates.

## Current implementation
Runs in a loop. On one iteration performs:
* Fetching Multidimensional Pod Autoscaler configuration using a lister implementation.
* Fetching live pods information with their current resource allocation.
* For each replicated pods group calculating if pod update is required and how many replicas can be evicted.
Updater will always allow eviction of at least one pod in replica set. Maximum ratio of evicted replicas is specified by flag.
* Evicting pods if recommended resources significantly vary from the actual resources allocation.
Threshold for evicting pods is specified by recommended min/max values from VPA resource.
Priority of evictions within a set of replicated pods is proportional to sum of percentages of changes in resources
(i.e. pod with 15% memory increase 15% cpu decrease recommended will be evicted
before pod with 20% memory increase and no change in cpu).
* Updating the Deployment with the desired number of replicas.

## Missing parts
* Recommendation API for fetching data from Multidimensional Pod Autoscaler Recommender.

## Building the Docker Image

```
make build-binary-with-vendor-amd64
make docker-build-amd64
```
