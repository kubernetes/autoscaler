# Running Integration Tests locally

Included in parallel with `run-e2e.sh` and `deploy-for-e2e.sh` are two alternate versions
with `-locally` as part of their names.  They use Kubernetes in Docker (`kind`) to run a local
cluster in Docker.  Using them will require `docker` and `kind` in your `PATH`.

## External Metrics Tests

The external metrics tests (`recommender-externalmetrics`, available on the `-locally` variants)
use a stack of 4 additional programs to support testing:

1. `hack/emit-metrics.py` to generate random CPU and RAM metrics for every pod in the local cluster.
2. Prometheus Pushgateway to accept metrics from `hack/emit-metrics`.
3. Prometheus to store the metrics accepted by the Pushgateway.
4. Prometheus Adapter to provide an External Metrics interface to Prometheus.

The External Metrics tests run by configuring a `recommender` to use the External Metrics interface
from the Prometheus Adapter.  With that configuration, it runs the standard `recommender` test suite. 

## Non-recommender tests

The `recommender` and `recommender-externalmetrics` test work locally, but none of the others do;
they require more Makefile work.

# Configuration Notes

To support the regular `recommender` tests locally, we've added the stock Kubernetes Metrics Server.
Unfortunately, it doesn't work with TLS turned on.  The metrics server is being run in insecure mode
to work around this.  This only runs in the local `kind` case, not in a real cluster.

# RBAC Changes

The local test cases support running the `recommender` with external metrics.  This requires
additional permissions we don't want to automatically enable for all customers via the 
configuration given in `deploy/vpa-rbac.yaml`.  The scripts use a context diff `hack/e2e/vpa-rbac.diff`
to enable those permission when running locally.

# Quick Integration Tests

`run-integration-locally.sh` is a quicker way to integration test compared to `run-e2e-locally.sh`. Only used for simple tests.
