# Helm Chart Unit Tests

<!-- toc -->
- [1. Installing helm-unittest](#1-installing-helm-unittest)
- [2. Running the tests](#2-running-the-tests)
- [3. Writing new test cases](#3-writing-new-test-cases)
<!-- /toc -->

The `vertical-pod-autoscaler` chart uses [helm-unittest](https://github.com/helm-unittest/helm-unittest)
to test template rendering logic, for example, the recommender's automatic
leader-election defaults based on replica count.

## 1. Installing helm-unittest

Check your Helm major version first:

```bash
helm version --short
```

Then install accordingly:

```bash
# Helm 4.x (plugin verification is enabled by default). --verify=false is
# required because helm-unittest is installed from a git source, which Helm
# cannot verify. See the helm-unittest README:
# https://github.com/helm-unittest/helm-unittest/blob/33c48cac798e465deda9a66c8e6c07c0973cf53d/README.md#L69
helm plugin install https://github.com/helm-unittest/helm-unittest.git --version 1.2.0 --verify=false

# Helm 3.x (no --verify flag exists, omit it)
helm plugin install https://github.com/helm-unittest/helm-unittest.git --version 1.2.0
```

Confirm with `helm plugin list`; it should show a plugin named `unittest`.

## 2. Running the tests

```bash
helm unittest vertical-pod-autoscaler/charts/vertical-pod-autoscaler
```

## 3. Writing new test cases

Test suites live under `tests/` and follow the `<template-name>_test.yaml`
naming convention. See `tests/recommender-deployment_test.yaml` for examples
using `equal`, `contains`, `notContains`, and `matchRegex` assertions.
