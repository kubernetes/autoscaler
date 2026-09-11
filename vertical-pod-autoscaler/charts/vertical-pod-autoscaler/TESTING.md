# Helm Chart Unit Tests

<!-- toc -->
- [1. Installing helm-unittest](#1-installing-helm-unittest)
- [2. Running the tests](#2-running-the-tests)
- [3. Writing new test cases](#3-writing-new-test-cases)
- [4. Troubleshooting](#4-troubleshooting)
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
# Helm 4.x (plugin verification enabled by default)
helm plugin install https://github.com/helm-unittest/helm-unittest.git --version 1.1.2 --verify=false

# Helm 3.x (no --verify flag exists, omit it)
helm plugin install https://github.com/helm-unittest/helm-unittest.git --version 1.1.2
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

## 4. Troubleshooting

- `unknown flag: --verify` means you're on Helm 3; omit the flag.
- `plugin already exists` means check `helm plugin list`; the plugin registers as
  `unittest`, not `helm-unittest`. Run `helm plugin uninstall unittest`, then
  reinstall.
- Unexpected `helm version --short` output; run `which -a helm` to check for
  multiple binaries competing on your `PATH`.
