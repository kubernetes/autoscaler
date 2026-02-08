# Development and testing

## Contents

<!-- toc -->
- [Introduction](#introduction)
- [Running integration tests](#running-integration-tests)
- [Running e2e tests](#running-e2e-tests)
  - [Feature gates](#feature-gates)
  - [Parallelism](#parallelism)
  - [External Metrics Tests](#external-metrics-tests)
<!-- /toc -->

## Introduction

This project contains various scripts and tools to aid in the development of the three VPA components.

## Running integration tests

The VPA contains integration tests that test individual components in isolation using [controller-runtime's envtest](https://github.com/kubernetes-sigs/controller-runtime/tree/main/tools/setup-envtest) environment.

They can be run using the `./hack/run-integration-tests.sh` helper script. This script automatically downloads and manages the required Kubernetes API server and etcd binaries for testing.

Integration tests are faster than e2e tests and do not require a full Kubernetes cluster. They are suitable for testing:

- Component-specific behavior and logic
- Configuration handling
- API object validation
- Resource filtering and selection logic

**Example usage:**

```bash
# Run all integration tests
./hack/run-integration-tests.sh

# Run a specific test
./hack/run-integration-tests.sh -run TestRecommenderWithNamespaceFiltering

# Run tests with custom parallelism
./hack/run-integration-tests.sh -parallel 8
```

By default, integration tests run with 4 parallel workers.

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
