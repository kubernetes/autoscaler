# Development and testing

## Contents

- [Development and testing](#development-and-testing)
  - [Introduction](#introduction)
  - [Running e2e tests](#running-e2e-tests)
    - [Feature gates](#feature-gates)
    - [Parallelism](#parallelism)

## Introduction

This project contains various scripts and tools to aid in the development of the three VPA components.

## Running e2e tests

The VPA contains some e2e tests that test how each component interacts with Pods and VPA resources inside a real Kubernetes cluster.

They can be run using the `./hack/run-e2e-locally.sh` helper script. Please note that this script will delete any existing [kind](https://kind.sigs.k8s.io) cluster running on the local machine before creating a fresh cluster for executing the tests.

### Feature gates

By default, the e2e test suite only runs feature-gated tests for features that are enabled by default (typically beta and GA). Alpha features, which are disabled by default, are not tested.

Setting the environment variable `ENABLE_ALL_FEATURE_GATES=true` will enable all feature gates and run all feature-gated tests.

### Parallelism

By default, the e2e tests create 4 worker processes, each one running its own test. This can be changed by setting the `NUMPROC=<workers>` variable.
