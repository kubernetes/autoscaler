# charts

The Helm chart for the Cluster Autoscaler project resides within this folder. If making changes to the Helm charts, make sure you follow the instructions below for the pre-commit checks.

## Pre-commit hooks

This Helm repository has pre-commit hooks for Helm specific needs:

* Makes sure all charts pass a `helm lint` check.
* Updates the `README.md` file of all charts based on comments in that chart's `values.yaml` file.

### Install `pre-commit` binary

The binary for `pre-commit` can be installed via Homebrew:

```shell
$ brew install pre-commit
```

For those without Homebrew, Pre-commit has [other installation methods available](https://pre-commit.com/#install).

### Install git hooks

After the `pre-commit` binary is installed, go to this repository's directory, and run the following command to install the git hook:

```shell
$ pre-commit install
```

### Install hook dependencies

The pre-commit hooks themselves call binaries under the hood; they can be installed via the [instructions found here](https://github.com/norwoodj/helm-docs#installation).

Note: You should ensure that whichever installation method you are using you either install the same version of helm-docs as used in the PR workflow to ensure your PR passes CI checks or update the version used by the workflow to match.

## OCI Helm Chart Publishing & Promotion

When `cluster-autoscaler/charts/cluster-autoscaler/Chart.yaml` is updated on `master`, the Prow postsubmit job [`post-autoscaler-push-cluster-autoscaler-chart`](https://testgrid.k8s.io/sig-autoscaling-cluster-autoscaler-images#post-autoscaler-push-cluster-autoscaler-chart) automatically packages and pushes the chart to the Kubernetes staging Artifact Registry (`oci://us-central1-docker.pkg.dev/k8s-staging-images/autoscaling/charts/cluster-autoscaler`).

To promote a staged chart version to `oci://registry.k8s.io/autoscaling/charts/cluster-autoscaler`:

1. Obtain the chart's `sha256` digest from the `post-autoscaler-push-cluster-autoscaler-chart` build log (or run `gcloud artifacts docker images describe us-central1-docker.pkg.dev/k8s-staging-images/autoscaling/charts/cluster-autoscaler:<chart-version> --format='value(image_summary.digest)'`).
2. Open a PR in [`kubernetes/k8s.io`](https://github.com/kubernetes/k8s.io) adding the digest and chart version under `charts/cluster-autoscaler` in [`registry.k8s.io/images/k8s-staging-autoscaling/images.yaml`](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-autoscaling/images.yaml).
3. Once merged and promoted by `post-k8sio-image-promo`, the chart is available at `oci://registry.k8s.io/autoscaling/charts/cluster-autoscaler`.
