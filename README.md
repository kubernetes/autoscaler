# Kubernetes Autoscaler

[![Build Status](https://travis-ci.org/kubernetes/autoscaler.svg?branch=master)](https://travis-ci.org/kubernetes/autoscaler) [![GoDoc Widget]][GoDoc]

This repository contains autoscaling-related components for Kubernetes.

## What's inside

[Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) - a component that automatically adjusts the size of a Kubernetes
Cluster so that all pods have a place to run and there are no unneeded nodes. Works with GCP, AWS and Azure. Version 1.0 (GA) was released with kubernetes 1.8.

[Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) - a set of components that automatically adjust the
amount of CPU and memory requested by pods running in the Kubernetes Cluster. Current state - beta.

[Addon Resizer](https://github.com/kubernetes/autoscaler/tree/master/addon-resizer) - a simplified version of vertical pod autoscaler that modifies
resource requests of a deployment based on the number of nodes in the Kubernetes Cluster. Current state - beta.

## Contact Info

Interested in autoscaling? Want to talk? Have questions, concerns or great ideas?

Please join us on #sig-autoscaling at https://kubernetes.slack.com/, or join one
of our weekly meetings.  See [the Kubernetes Community Repo](https://github.com/kubernetes/community/blob/master/sig-autoscaling/README.md) for more information.

## Getting the Code

Fork the repository in the cloud:
1. Visit https://github.com/kubernetes/autoscaler
2. Click Fork button (top right) to establish a cloud-based fork.

The code must be checked out as a subdirectory of `k8s.io`, and not `github.com`.

```shell
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
cd autoscaler
```

## Building the Cluster autoscaler from master

```shell
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
# Replace "GO_Versiom" below with the version of go you want to use(eg. go1.11)
wget https://dl.google.com/go/$GO_Version.linux-amd64.tar.gz
# Extract the tar file
tar -xvf $GO_Version.linux-amd64.tar.gz
# Move go to /usr/local
mv go /usr/local
# Set GOROOT,GOPATH,PATH and add to profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
source ~/.profile
# Create directory $GOPATH/src/k8s.io and copy autoscaler cloned earlier
cp -r ./autoscaler/  $GOPATH/src/k8s.io/
# get godep and move to the copied location 
go get github.com/tools/godep
cd $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler 
# start the build
GOOS=linux make build-binary
#Run the Dockerfile from $GOPATH/src/k8s.io/autoscaler directory
docker build -t autoscaler -f cluster-autoscaler/Dockerfile .
```


Please refer to Kubernetes [Github workflow guide] for more details.

[GoDoc]: https://godoc.org/k8s.io/autoscaler
[GoDoc Widget]: https://godoc.org/k8s.io/autoscaler?status.svg
[Github workflow guide]: https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md
