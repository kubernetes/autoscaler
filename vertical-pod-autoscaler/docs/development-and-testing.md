# Development and testing

## Contents

<!-- toc -->
- [Introduction](#introduction)
- [Running e2e tests](#running-e2e-tests)
  - [Feature gates](#feature-gates)
  - [Parallelism](#parallelism)
  - [External Metrics Tests](#external-metrics-tests)
- [Running integration tests](#running-integration-tests)
<!-- /toc -->

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

### External Metrics Tests

The external metrics tests (`recommender-externalmetrics`, available in `run-e2e-locally.sh` and `deploy-for-e2e-locally.sh`)
use a stack of 4 additional programs to support testing:

1. `hack/emit-metrics.py` to generate random CPU and RAM metrics for every pod in the local cluster.
2. Prometheus Pushgateway to accept metrics from `hack/emit-metrics`.
3. Prometheus to store the metrics accepted by the Pushgateway.
4. Prometheus Adapter to provide an External Metrics interface to Prometheus.

The External Metrics tests run by configuring a `recommender` to use the External Metrics interface
from the Prometheus Adapter.  With that configuration, it runs the standard `recommender` test suite.

## Running integration tests

There are two different test suites in this repository referred to as "integration" tests, and they are not the same kind of test:

* `test/e2e/integration` contains Ginkgo tests that, despite the directory name, are e2e style: they start a component with specific flags against a real Kubernetes cluster and verify its behavior directly, without going through the full deployment manifests in `deploy/`. This is useful for covering flags that the e2e suite's default deployments don't exercise. These tests are heavy to run and, unlike the e2e suite, cannot run in parallel. The plan is to remove this suite over time in favor of the lighter `test/integration/recommender` style below.
* `test/integration/recommender` contains standard Go tests, closer to how Kubernetes itself writes integration tests: the component runs inside the test process, against an in-process API server and etcd, rather than as a separate deployment against a real cluster. `test/` is its own Go module, so run them from that directory: `cd test && go test ./integration/...`.

The `test/e2e/integration` suite can be run using the `./hack/run-integration-locally.sh recommender` helper script. `recommender` is currently the only supported suite. Like `run-e2e-locally.sh`, this deletes any existing local [kind](https://kind.sigs.k8s.io) cluster before creating a fresh one.
