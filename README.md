# Kubernetes Autoscaler

This repository contains autoscaling-related components for Kubernetes.

## What's inside

A component that automatically adjusts the size of a Kubernetes

## Getting the Code

The code must be checked out as a subdirectory of `k8s.io`, and not `github.com`.

```shell
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github orgnization of github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
cd autoscaler
```

The current dev branch is `cluster`.

Build:
```
cd cluster-autoscaler
git checkout cluster
make build-binary
```

Run:
```
./cluster-autoscaler --v=5 --stderrthreshold=error --logtostderr=true --cloud-provider=aztools --skip-nodes-with-local-storage=false --nodes=1:10:dlws-worker-asg --leader-elect=false --scale-down-enabled=false --kubeconfig=./deploy/kubeconfig.yaml
```

Note:

`--nodes=1:10:dlws-worker-asg`: dlws-worker-asg is the cluster name of you deploy, with 1 node as min, 10 nodes as max for scaling.
