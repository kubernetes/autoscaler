---
name: run-azure-e2e-tests
description: 'Run Azure CAS end-to-end tests — per-suite execution with focus filtering, background execution, and local/CI workflows. Use when: running e2e tests, debugging test failures, adding new test suites.'
---

# E2E Tests for Azure CAS

## Test Structure

```
cluster-autoscaler/cloudprovider/azure/test/
├── suites/
│   └── scaleup/             # Scale-up/down test
│       └── suite_test.go
├── pkg/
│   └── environment/         # Shared Environment struct + helpers
│       └── environment.go
├── Makefile                  # Local + CI targets
└── go.mod
```

## Local Developer Workflow

From `cluster-autoscaler/cloudprovider/azure/test/`:

### First-time setup

```bash
az login
make setup-cluster      # Creates AKS + ACR + workload identity (~5 min)
make deploy-local       # Builds + deploys CAS via skaffold (~1 min)
```

### Running tests

```bash
export AZURE_SUBSCRIPTION_ID="$(az account show --query id -o tsv)"
export AZURE_RESOURCE_GROUP="MC_..."  # Node resource group (printed by setup-cluster)

make e2etests                         # Run all suites
make e2etests TEST_SUITE=scaleup      # Run single suite
make e2etests FOCUS="scales up"       # Focus filter
```

### After code changes

```bash
make deploy-local       # Rebuild + redeploy CAS
make e2etests TEST_SUITE=scaleup
```

### Utility commands

- `make list-suites` — list available test suites
- `make validate-env` — check required env vars
- `make deploy-local-dev` — skaffold watch mode (auto-redeploy on changes)

### Background execution (survives VPN drops)

```bash
nohup make e2etests TEST_SUITE=scaleup > e2e.log 2>&1 &
tail -f e2e.log
```

## CI (Prow)

`make test-e2e` builds the CAS image and deploys via Helm (inside BeforeSuite), using cluster info from CAPZ. The Helm deploy is triggered by `-cas-image-repository` and `-cas-image-tag` flags — when absent (local path), Helm is skipped.

## Monitoring

- **Logs**: `tail -f e2e.log`
- **Cluster**: `kubectl get nodes,pods -w`
- **Events**: `kubectl get events -A --field-selector source=cluster-autoscaler --watch`
- **VMSS**: `az vmss list -g $AZURE_RESOURCE_GROUP -o table`
- **CAS logs**: `kubectl logs -n kube-system deploy/cluster-autoscaler -f`

## Adding a New Test Suite

1. Create `test/suites/<name>/suite_test.go`
2. Import `pkg/environment` for shared helpers
3. Register `-resource-group` flag in `init()`
4. Create `Environment` in `BeforeSuite`, call `EnsureHelmRelease(...)` for CI compatibility
5. Run: `make e2etests TEST_SUITE=<name>`
