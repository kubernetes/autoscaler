# Cluster Autoscaler on AWS
The cluster autoscaler on AWS scales worker nodes within any specified autoscaling group. It will run as a `Deployment` in your cluster. This README covers the steps required to configure and run the cluster autoscaler.

## Kubernetes Version
Cluster autoscaler must run on v1.3.0 or greater.

## Permissions
The pod running the cluster autoscaler will need access to certain resources and actions. If using AWS EKS it is recommend to attach the IAM policy to the cluster austoscaler pod using [IAM roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html). For non-EKS kubernetes clusters attaching the IAM policy to the NodeGroup is recommended instead of using AWS credentials directly unless you have special requirements.

### Attach IAM policy to Service Account
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
                "autoscaling:DescribeLaunchConfigurations",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
```
If you'd like to scale node groups from 0, an
`autoscaling:DescribeLaunchConfigurations` or
`ec2:DescribeLaunchTemplateVersions` permission is required depending on if you
made your ASG with Launch Configuration or Launch Template.

If you'd like the cluster autoscaler to [automatically
discover](#auto-discovery-setup) EC2 AutoScalingGroups, the
`autoscaling:DescribeTags` permission is also required.

**NOTE**: You can restrict the target resources for the autoscaling actions by
specifying autoscaling group ARNS. More information can be found
[here](https://docs.aws.amazon.com/autoscaling/latest/userguide/control-access-using-iam.html#policy-auto-scaling-resources).

### Using AWS Credentials
For on premise users wishing to scale out to AWS, the above approach of attaching policy to a nodegroup role won't work. Instead, you can create an aws secret manually and add following environment variables to cluster-autoscaler deployment manifest. Cluster autoscaler will use credential to authenticate and authorize itself. Please make sure your role has above permissions.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-secret
type: Opaque
data:
  aws_access_key_id: BASE64_OF_YOUR_AWS_ACCESS_KEY_ID
  aws_secret_access_key: BASE64_OF_YOUR_AWS_SECRET_ACCESS_KEY
```
Please check [guidance](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret-manually) for creating a secret manually.

```yaml
env:
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: aws-secret
      key: aws_access_key_id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: aws-secret
      key: aws_secret_access_key
- name: AWS_REGION
  value: YOUR_AWS_REGION
```

## Deployment Specification
Auto-Discovery Setup is always preferred option to avoid multiple, potentially different configuration for min/max values. If you want to adjust minimum and maximum size of the group, please adjust size on ASG directly, CA will fetch latest change when talking to ASG.

If you use one or multiple ASG setup, the min/max configuration change in CA will not make the corresponding change to ASG. Please make sure CA min/max values are within the boundary of ASG minSize and maxSize.

### One ASG Setup (min: 1, max: 10, ASG Name: k8s-worker-asg-1)
```
kubectl apply -f examples/cluster-autoscaler-one-asg.yaml
```

### Multiple ASG Setup
```
kubectl apply -f examples/cluster-autoscaler-multi-asg.yaml
```

### Master Node Setup

To run a CA pod in master node - CA deployment should tolerate the master `taint` and `nodeSelector` should be used to schedule the pods in master node.
Please replace `{{ node_asg_min }}`, `{{ node_asg_max }}` and `{{ name }}` with your ASG setting in the yaml file.
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
If your pods request resources other than `cpu` and `memory`, you need to tag ASG with key `k8s.io/cluster-autoscaler/node-template/resources/`.

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
If you request other resources on the node, like `vpc.amazonaws.com/PrivateIPv4Address` for Windows nodes, `ephemeral-storage`, etc, you would tag ASG with

```json
{
    "ResourceType": "auto-scaling-group",
    "ResourceId": "foo.example.com",
    "PropagateAtLaunch": true,
    "Value": "2",
    "Key": "k8s.io/cluster-autoscaler/node-template/resources/vpc.amazonaws.com/PrivateIPv4Address"
}
```
> Note: This is only supported in CA 1.14.x and above

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

### Gotchas

* Without these tags, when the cluster autoscaler needs to increase the number of nodes, if a node group creates nodes with taints that the pending pod does not tolerate then the cluster autoscaler will only learn about this after the node has been created and it sees that it is tainted. From this point on this information will be cached and subsequent scaling operations will take this into account, but it means that the behaviour of the cluster autoscaler differs between the first and subsequent scale up requests and can lead to confusion.

* The device plugin on nodes which provide GPU resources take a little while to advertise the GPU resource to the APIServer so the AutoScaler may unnecessarily scale up again. See the guidance below for how to avoid this

## GPU Node Groups

If you launch a pod that requires a GPU in it's resource requirements then you must add the following node label to the node (via the kubelet arguments for example)

### Cluster AutoScaler Version < 1.15.x

```bash
--node-labels=cloud.google.com/gke-accelerator=<GPU TYPE YOU ARE USING>
```

E.g. on an AWS P2.X instance

```bash
--kubelet-extra-args '--node-labels=cloud.google.com/gke-accelerator=nvidia-tesla-k80'
```

### Cluster AutoScaler Version >= 1.15.x

```bash
--node-labels=k8s.amazonaws.com/accelerator=<GPU TYPE YOU ARE USING>
```

E.g. on an AWS P2.X instance

```bash
--kubelet-extra-args '--node-labels=k8s.amazonaws.com/accelerator=nvidia-tesla-k80'
```

This is because the GPU resource does not become available immediately after the instance is ready and so without this label, the cluster autoscaler will think that no suitable GPU resource is available and add an additional node.

## Using AutoScalingGroup MixedInstancesPolicy

> Note: The minimum version of cluster autoscaler to support MixedInstancePolicy is v1.14.x.

If your workloads can tolerate interruption, consider taking advantage of Spot Instances for a lower price point. To enable diversity among On Demand and Spot Instances, as well as specify multiple EC2 instance types in order to tap into multiple Spot capacity pools, use a [mixed instances policy](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-autoscalinggroup-mixedinstancespolicy.html) on your ASG. Note that the instance types should have the same amount of RAM and number of CPU cores, since this is fundamental to CA's scaling calculations. Using mismatched instances types can produce unintended results. See an example below.

Additionally, there are other factors which affect scaling, such as node labels. If you are currently using `nodeSelector` with the [beta.kubernetes.io/instance-type](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#interlude-built-in-node-labels) label, you will need to apply a common propagating label to the ASG and use that instead, since the instance-type label can no longer be relied upon. One may also use auto-generated tags such as `aws:cloudformation:stack-name` for this purpose. [Node affinity and anti-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity) are not affected in the same way, since these selectors natively accept multiple values; one must add all the configured instances types to the list of values, for example:

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: beta.kubernetes.io/instance-type
            operator: In
            values:
            - r5.2xlarge
            - r5d.2xlarge
            - r5a.2xlarge
            - r5ad.2xlarge
            - r5n.2xlarge
            - r5dn.2xlarge
            - r4.2xlarge
            - i3.2xlarge
```

### Example usage:

* Create a [Launch Template](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-autoscalinggroup-launchtemplate.html) (LT) with an instance type, for example, r5.2xlarge. Consider this the 'base' instance type. Do not define any spot purchase options here.
* Create an ASG with a MixedInstancesPolicy that refers to the newly-created LT.
* Set LaunchTemplateOverrides to include the 'base' instance type r5.2xlarge and suitable alternatives, e.g. r5d.2xlarge, i3.2xlarge, r5a.2xlarge and r5ad.2xlarge. Differing processor types and speeds should be evaluated depending on your use-case(s).
* Set [InstancesDistribution](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_InstancesDistribution.html) according to your needs.
* See [Allocation Strategies](https://docs.aws.amazon.com/autoscaling/ec2/userguide/asg-purchase-options.html#asg-allocation-strategies) for information about how the ASG fulfils capacity from the specified instance types. It is recommended to use the capacity-optimized allocation strategy, which will automatically launch Spot Instances into the most available pools by looking at real-time capacity data and.
* For the same workload or for the generic capacity in your cluster, you can also create more node groups with a vCPU/Mem ratio that is a good fit for your workloads, but from different instance sizes. For example:
Node group 1: m5.xlarge, m5a.xlarge, m5d.xlarge, m5ad.xlarge, m4.xlarge. 
Node group 2: m5.2xlarge, m5a.2xlarge, m5d.2xlarge, m5ad.2xlarge, m4.2xlarge. 
This approach increases the chance of achieving your desired scale at the lowest cost by tapping into many Spot capacity pools.

See CloudFormation example [here](MixedInstancePolicy.md).

## Use Static Instance List
The set of the latest supported EC2 instance types will be fetched by the CA at run time. You can find all the available instance types in the CA logs.
If your network access is restricted such that fetching this set is infeasible, you can specify the command-line flag `--aws-use-static-instance-list=true` to switch the CA back to its original use of a statically defined set.

To refresh static list, please run `go run ec2_instance_types/gen.go` under `cluster-autoscaler/cloudprovider/aws/` and update `staticListLastUpdateTime` in `aws_util.go`

## Common Notes and Gotchas:
- The `/etc/ssl/certs/ca-bundle.crt` should exist by default on ec2 instance in your EKS cluster. If you use other cluster privision tools like [kops](https://github.com/kubernetes/kops) with different operating systems other than Amazon Linux 2, please use `/etc/ssl/certs/ca-certificates.crt` or correct path on your host instead for the volume hostPath in your cluster autoscaler manifest.
- If you’re using Persistent Volumes, your deployment needs to run in the same AZ as where the EBS volume is, otherwise the pod scheduling could fail if it is scheduled in a different AZ and cannot find the EBS volume. To overcome this, either use a single AZ ASG for this use case, or an ASG-per-AZ while enabling [--balance-similar-node-groups](../../FAQ.md#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler). Alternately, and depending on your use-case, you might be able to switch from using EBS to using shared storage that is available across AZs (for each pod in its respective AZ). Consider AWS services like Amazon EFS or Amazon FSx for Lustre.
- On creation time, the ASG will have the [AZRebalance process](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-benefits.html#AutoScalingBehavior.InstanceUsage) enabled, which means it will actively work to balance the number of instances between AZs, and possibly terminate instances. If your applications could be impacted from sudden termination, you can either suspend the AZRebalance feature, or use a tool for automatic draining upon ASG scale-in such as the [k8s-node-drainer]https://github.com/aws-samples/amazon-k8s-node-drainer. The [AWS Node Termination Handler](https://github.com/aws/aws-node-termination-handler/issues/95) will also support this use-case in the future.
- By default, cluster autoscaler will not terminate nodes running pods in the kube-system namespace. You can override this default behaviour by passing in the `--skip-nodes-with-system-pods=false` flag.
- By default, cluster autoscaler will wait 10 minutes between scale down operations, you can adjust this using the `--scale-down-delay-after-add`, `--scale-down-delay-after-delete`, and `--scale-down-delay-after-failure` flag. E.g. `--scale-down-delay-after-add=5m` to decrease the scale down delay to 5 minutes after a node has been added.
- If you're running multiple ASGs, the `--expander` flag supports three options: `random`, `most-pods` and `least-waste`. `random` will expand a random ASG on scale up. `most-pods` will scale up the ASG that will schedule the most amount of pods. `least-waste` will expand the ASG that will waste the least amount of CPU/MEM resources. In the event of a tie, cluster autoscaler will fall back to `random`.
- If you're managing your own kubelets, they need to be started with the `--provider-id` flag. The provider id has the format `aws:///<availability-zone>/<instance-id>`, e.g. `aws:///us-east-1a/i-01234abcdef`.
- If you want to use regional STS endpoints (e.g. when using VPC endpoint for STS) the env `AWS_STS_REGIONAL_ENDPOINTS=regional` should be set.
