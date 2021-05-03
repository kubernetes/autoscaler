# Cluster Autoscaler for Bizflycloud

The cluster autoscaler for Bizflycloud scales worker nodes within any
specified Bizflycloud Kubernetes Engine cluster's worker pool.

# Configuration

Bizflycloud Kubernetes Engine (BKE) will authenticate with Bizflycloud providers using your application credentials automatics created by BKE.

The scaling option about enable autoscaler, min-nodes, max-nodes will be configure though our dashboard

**Note**: Do not install cluster-autoscaler deployment in manifest since it already install by BKE.
