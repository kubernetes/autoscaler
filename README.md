# Kubernetes Autoscaler

[![Build Status](https://travis-ci.org/kubernetes/autoscaler.svg?branch=master)](https://travis-ci.org/kubernetes/autoscaler)

This repository contains autoscaling-related components for Kubernetes.

## Contact Info

Interested in Autoscaling? Want to talk? Have questions, concenrns or great ideas?

Please join us on #sig-autoscaling at https://kubernetes.slack.com/.
Moreover, every Thursday we host a 30min sig-autoscaling meeting on 
https://plus.google.com/hangouts/_/google.com/k8s-autoscaling at
17:30 CEST/CET,  8:30 am PST/PDT. 

## Getting the Code

The code must be checked out as a subdirectory of `k8s.io`, and not `github.com`.

```shell
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/autoscaler.git
cd autoscaler
```