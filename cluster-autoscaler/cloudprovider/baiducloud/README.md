# Cluster Autoscaler on BaiduCloud
The cluster autoscaler on BaiduCloud scales worker nodes within any specified autoscaling group. It will run as a `Deployment` in your cluster. This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Kubernetes Version
Cluster autoscaler must run on v1.8.6 or greater.

## Deployment Specification

### 1 ASG Setup (min: 1, max: 10, ASG Name: k8s-worker-asg-1)
```
kubectl apply -f examples/cluster-autoscaler-one-asg.yaml
```

### Multiple ASG Setup
Multiple ASG Setup is not supported in BaiduCloud currently.

## Common Notes and Gotchas:
- By default, cluster autoscaler will not terminate nodes running pods in the kube-system namespace. You can override this default behaviour by passing in the `--skip-nodes-with-system-pods=false` flag.
- By default, cluster autoscaler will wait 10 minutes between scale down operations, you can adjust this using the `--scale-down-delay` flag. E.g. `--scale-down-delay=5m` to decrease the scale down delay to 5 minutes.

## Maintainer
* Hongbin Mao [@hello2mao](https://github.com/hello2mao)  
* Ti Zhou [@tizhou86](https://github.com/tizhou86)