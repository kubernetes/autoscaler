## Overview:

This document, and directory are focused on the ability to deploy and test a working version of autoscaler from a development branch onto an AKS cluster for testing out a set of changes.

## Steps:

1. Create a codespace using one of the devcontainer setups from the `devcontainers` branch of https://github.com/azure/autoscaler

2. In the codespace switch to whatever branch you want to test
    - Note: for testing an upstream branch use: `git checkout upstream/<branch-name>`
        - This might require a `git fetch upstream`

5. run `cd cluster-autoscaler/cloudprovider/azure/examples/dev`

6. run `az login`

7. run `./aks-dev-deploy.sh`

8. run `cd ../../../../`

9. run `skaffold run --filename cloudprovider/azure/examples/dev/skaffold.yaml`

10. inspect the cluster with `kubectl`, and scale the `inflate` deployment for testing as desired.
