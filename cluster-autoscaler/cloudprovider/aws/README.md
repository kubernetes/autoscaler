# Cluster Autoscaler on AWS
The cluster autoscaler on AWS scales worker nodes within any specified autoscaling group. It will run as a `Deployment` in your cluster. This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Kubernetes Version
Cluster autoscaler must run on v1.3.0 or greater.

## Permissions
The worker running the cluster autoscaler will need access to certain resources and actions.

A minimum IAM policy would look like:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
```

If you'd like to auto-discover node groups by specifying the `--node-group-auto-discover` flag, a `DescribeTags` permission is also required:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:DescribeTags",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
```

AWS supports ARNs for autoscaling groups. More information [here](https://docs.aws.amazon.com/autoscaling/latest/userguide/control-access-using-iam.html#policy-auto-scaling-resources).

## Deployment Specification

### 1 ASG Setup (min: 1, max: 10, ASG Name: k8s-worker-asg-1)
```
kubectl apply -f examples/cluster-autoscaler-one-asg.yaml
```

### Multiple ASG Setup
```
kubectl apply -f examples/cluster-autoscaler-multi-asg.yaml
```

### Master Node Setup

To run a CA pod in master node - CA deployment should tolerate the master `taint` and `nodeSelector` should be used to schedule the pods in master node.
```
kubectl apply -f examples/cluster-autoscaler-run-on-master.yaml
```


### Auto-Discovery Setup

To run a cluster-autoscaler which auto-discovers ASGs with nodes use the `--node-group-auto-discovery` flag. For example, `--node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/<YOUR CLUSTER NAME>` will find the ASGs where those tag keys
_exist_. It does not matter what value the tags have.

Note that:

* It is recommended to use a second tag like `k8s.io/cluster-autoscaler/<YOUR CLUSTER NAME>` when `k8s.io/cluster-autoscaler/enabled` is used across many clusters to prevent ASGs from different clusters recognized as the node groups
* There are no `--nodes` flags passed to cluster-autoscaler because the node groups are automatically discovered by tags
* No min/max values are provided when using Auto-Discovery, cluster-autoscaler will respect the current min and max values of the ASG being targeted, and it will adjust only the "desired" value.

```
kubectl apply -f examples/cluster-autoscaler-autodiscover.yaml
```

## Scaling a node group to 0

From CA 0.6.1 - it is possible to scale a node group to 0 (and obviously from 0), assuming that all scale-down conditions are met.

If you are using `nodeSelector` you need to tag the ASG with a node-template key `"k8s.io/cluster-autoscaler/node-template/label/"` and `"k8s.io/cluster-autoscaler/node-template/taint/"` if you are using taints.

For example for a node label of `foo=bar` you would tag the ASG with:

```json
{
    "ResourceType": "auto-scaling-group",
    "ResourceId": "foo.example.com",
    "PropagateAtLaunch": true,
    "Value": "bar",
    "Key": "k8s.io/cluster-autoscaler/node-template/label/foo"
}
```

And for a taint of `"dedicated": "foo:NoSchedule"` you would tag the ASG with:

```json
{
    "ResourceType": "auto-scaling-group",
    "ResourceId": "foo.example.com",
    "PropagateAtLaunch": true,
    "Value": "foo:NoSchedule",
    "Key": "k8s.io/cluster-autoscaler/node-template/taint/dedicated"
}
```

If you'd like to scale node groups from 0, an `autoscaling:DescribeLaunchConfigurations` or `ec2:DescribeLaunchTemplateVersions` permission is required depending on if you made your ASG with Launch Configuration or Launch Template:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:DescribeTags",
                "autoscaling:DescribeLaunchConfigurations",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup",
                "ec2:DescribeLaunchTemplateVersions"
            ],
            "Resource": "*"
        }
    ]
}
```

## Common Notes and Gotchas:
- The `/etc/ssl/certs/ca-certificates.crt` should exist by default on your ec2 instance. If you use Amazon Linux 2, use `/etc/ssl/certs/ca-bundle.crt` instead.
- Cluster autoscaler is not zone aware (for now), so if you wish to span multiple availability zones in your autoscaling groups beware that cluster autoscaler will not evenly distribute them. For more information, see https://github.com/kubernetes/contrib/pull/1552#discussion_r75532949.
- By default, cluster autoscaler will not terminate nodes running pods in the kube-system namespace. You can override this default behaviour by passing in the `--skip-nodes-with-system-pods=false` flag.
- By default, cluster autoscaler will wait 10 minutes between scale down operations, you can adjust this using the `--scale-down-delay-after-add`, `--scale-down-delay-after-delete`, and `--scale-down-delay-after-failure` flag. E.g. `--scale-down-delay-after-add=5m` to decrease the scale down delay to 5 minutes after a node has been added.
- If you're running multiple ASGs, the `--expander` flag supports three options: `random`, `most-pods` and `least-waste`. `random` will expand a random ASG on scale up. `most-pods` will scale up the ASG that will scheduable the most amount of pods. `least-waste` will expand the ASG that will waste the least amount of CPU/MEM resources. In the event of a tie, cluster autoscaler will fall back to `random`.
