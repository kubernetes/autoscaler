# charts

The Helm charts for the Autoscaler project reside within this folder. If making changes to the Helm charts, make sure you follow the instructions below for the pre-commit checks.

## Pre-commit hooks

This Helm repository has pre-commit hooks for Helm specific needs:

* Makes sure all charts pass a `helm lint` check.
* Updates the `README.md` file of all charts based on comments in that chart's `values.yaml` file.

### Install `pre-commit` binary

The binary for `pre-commit` can be installed via Homebrew:

```shell
$ brew install pre-commit
```

### Install git hooks

After the `pre-commit` binary is installed, go to this repository's directory, and run the following command to install the git hook:

```shell
$ pre-commit install
```

### Install hook dependencies

The pre-commit hooks themselves call binaries under the hood; they can be installed via the [instructions found here](https://github.com/norwoodj/helm-docs#installation).

Note: You should ensure that whichever installation method you are using you either install the same version of helm-docs as used in the PR workflow to ensure your PR passes CI checks or update the version used by the workflow to match.
