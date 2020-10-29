# Cluster autoscaler tests

This folder is home to a test suite extracted from the [Kubernetes e2e
suite](https://github.com/kubernetes/kubernetes/blob/2729b8e375143434fc4977fe49eaea572567dac3/test/e2e/autoscaling/cluster_size_autoscaling.go).
These tests were very coupled to running only on GCE/GKE, and the intention here
was to create a staging ground for making them more platform agnostic.

In order to allow platform providers to implement the [Provider
interface](./types.go) without needing all of them to live in one
shared repository, the tests are setup as a function that can be invoked from a
separate test binary. An example of this can be found in the [CAPI autoscaler
repo](https://github.com/benmoss/cluster-api-autoscaler-provider/blob/aea52e04287ec24faec0e5fb208cd623f1884a34/test/suite_test.go#L9-L11).
This repository is meant to solely be a test library, and the actual executable
tests that import this library are meant to live outside of this repo.
