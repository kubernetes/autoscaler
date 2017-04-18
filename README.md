# Kubernetes Autoscaler

[![Build Status](https://travis-ci.org/kubernetes/autoscaler.svg?branch=master)](https://travis-ci.org/kubernetes/autoscaler)

This repository contains autoscaling-related components for Kubernetes.

## Getting the Code

The code must be checked out as a subdirectory of `k8s.io`, and not `github.com`.

```shell
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
cd autoscaler
```