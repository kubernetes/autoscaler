# Cluster Autoscaler on AWS
On AWS, Cluster Autoscaler utilizes Amazon EC2 Auto Scaling Groups to manage node
groups. Cluster Autoscaler typically runs as a `Deployment` in your cluster.

## Requirements
Cluster Autoscaler requires Kubernetes v1.3.0 or greater.

## Permissions
Cluster Autoscaler requires the ability to examine and modify EC2 Auto Scaling
Groups. We recommend using [IAM roles for Service
Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
to associate the Service Account that the Cluster Autoscaler Deployment runs as
with an IAM role that is able to perform these functions. If you are unable to
use IAM Roles for Service Accounts, you may associate an IAM service role with
the EC2 instance on which the Cluster Autoscaler pod runs.

### IAM Policy
The following policy provides the minimum privileges necessary for Cluster Autoscaler to run:

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
            "Resource": ["*"]
        }
    ]
}
```

If you'd like Cluster Autoscaler to [automatically
discover](#auto-discovery-setup) EC2 Auto Scaling Groups **(recommended)**, add
`autoscaling:DescribeTags` to the `Action` list. Also add
`autoscaling:DescribeLaunchConfigurations` (if you created your ASG using a
Launch Configuration) and/or `ec2:DescribeLaunchTemplateVersions` (if you
created your ASG using a Launch Template) to the `Action` list.

If you prefer, you can restrict the target resources for the autoscaling actions
by specifying Auto Scaling Group ARNs in the `Resource` list of the policy. More
information can be found
[here](https://docs.aws.amazon.com/autoscaling/latest/userguide/control-access-using-iam.html#policy-auto-scaling-resources).

### Using OIDC Federated Authentication
OIDC federated authentication allows your service to assume an IAM role and interact with AWS services without having to store credentials as environment variables. For an example of how to use AWS IAM OIDC with the Cluster Autoscaler please see [here](CA_with_AWS_IAM_OIDC.md).

### Using AWS Credentials

**NOTE** The following is not recommended for Kubernetes clusters running on
AWS. If you are using Amazon EKS, consider using [IAM roles for Service
Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
instead.

For on-premise clusters, you may create an IAM user subject to the above policy
and provide the IAM credentials as environment variables in the Cluster
Autoscaler deployment manifest. Cluster Autoscaler will use these credentials to
authenticate and authorize itself.

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
Please refer to the [relevant Kubernetes
documentation](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret-manually)
for creating a secret manually.

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

## Auto-Discovery Setup

Auto-Discovery Setup is the preferred method to configure Cluster Autoscaler.

To enable this, provide the `--node-group-auto-discovery` flag as an argument
whose value is a list of tag keys that should be looked for. For example,
`--node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/<cluster-name>`
will find the ASGs where those tag keys _exist_. It does not matter what value
the tags have.

Example deployment:

```
kubectl apply -f examples/cluster-autoscaler-autodiscover.yaml
```

Cluster Autoscaler will respect the minimum and maximum values of each Auto
Scaling Group. It will only adjust the desired value.

Each Auto Scaling Group should be composed of instance types that provide
approximately equal capacity. For example, ASG "xlarge" could be composed of
m5a.xlarge, m4.xlarge, m5.xlarge, and m5d.xlarge instance types, because each of
those provide 4 vCPUs and 16GiB RAM. Separately, ASG "2xlarge" could be
composed of m5a.2xlarge, m4.2xlarge, m5.2xlarge, and m5d.2xlarge instance
types, because each of those provide 8 vCPUs and 32GiB RAM.

Cluster Autoscaler will attempt to determine the CPU, memory, and GPU resources
provided by an Auto Scaling Group based on the instance type specified in its
Launch Configuration or Launch Template. It will also examine any overrides
provided in an ASG's Mixed Instances Policy. If any such overrides are found,
only the first instance type found will be used.  See [Using Mixed Instances
Policies and Spot Instances](#Using-Mixed-Instances-Policies-and-Spot-Instances)
for details.

From version 1.14, Cluster Autoscaler can also determine the resources provided
by each Auto Scaling Group via tags. The tag is of the format
`k8s.io/cluster-autoscaler/node-template/resources/<resource-name>`.
`<resource-name>` is the name of the resource, such as `ephemeral-storage`. The
value of each tag specifies the amount of resource provided. The units are
identical to the units used in the `resources` field of a Pod specification.

Example tags:

* `k8s.io/cluster-autoscaler/node-template/resources/ephemeral-storage`: `100G`

You may also provide additional hints to Cluster Autoscaler that the nodes will
be labeled or tainted when they join the cluster, such as:

* `k8s.io/cluster-autoscaler/node-template/label/foo`: `bar`
* `k8s.io/cluster-autoscaler/node-template/taint/dedicated`: `NoSchedule`

**NOTE:** It is your responsibility to ensure such labels and/or taints are
applied via the node's kubelet configuration at startup.

Recommendations:

* It is recommended to use a second tag like
  `k8s.io/cluster-autoscaler/<cluster-name>` when
  `k8s.io/cluster-autoscaler/enabled` is used across many clusters to prevent
  ASGs from different clusters recognized as the node groups.
* To prevent conflicts, do not provide a `--nodes` argument if
  `--node-group-auto-discovery` is specified.
* Be sure to add `autoscaling:DescribeLaunchConfigurations` or
  `ec2:DescribeLaunchTemplateVersions` to the `Action` list of the IAM Policy
  used by Cluster Autoscaler, depending on whether your ASG utilizes Launch
  Configurations or Launch Templates.
* If Cluster Autoscaler adds a node to the cluster, and the node has taints applied
  when it joins the cluster that Cluster Autoscaler was unaware of (because the tag
  wasn't supplied), this can lead to significant confusion and misbehavior.

### Special note on GPU instances

The device plugin on nodes that provides GPU resources can take some time to
advertise the GPU resource to the cluster. This may cause Cluster Autoscaler to
unnecessarily scale out multiple times.

To avoid this, you can configure `kubelet` on your GPU nodes to label the node
before it joins the cluster by passing it the `--node-labels` flag. The label
format is as follows:

* Cluster Autoscaler < 1.15: `cloud.google.com/gke-accelerator=<gpu-type>`
* Cluster Autoscaler >= 1.15: `k8s.amazonaws.com/accelerator=<gpu-type>`

`<gpu-type>` varies by instance type. On P2 instances, for example, the
value is `nvidia-tesla-k80`.

## Manual configuration

Cluster Autoscaler can also be configured manually if you wish by passing the
`--nodes` argument at startup. The format of the argument is
`--nodes=<min>:<max>:<asg-name>`, where `<min>` is the minimum number of nodes,
`<max>` is the maximum number of nodes, and `<asg-name>` is the Auto Scaling
Group name.

You can pass multiple `--nodes` arguments if you have multiple Auto Scaling Groups
you want Cluster Autoscaler to use.

**NOTES**:
* Both `<min>` and `<max>` must be within the range of the minimum and maximum
  instance counts specified by the Auto Scaling group.
* When manual configuration is used, all Auto Scaling groups must use EC2
  instance types that provide equal CPU and memory capacity.

Examples:

### One ASG Setup (min: 1, max: 10, ASG Name: k8s-worker-asg-1)
```
kubectl apply -f examples/cluster-autoscaler-one-asg.yaml
```

### Multiple ASG Setup
```
kubectl apply -f examples/cluster-autoscaler-multi-asg.yaml
```

<!--TODO: Remove "previously referred to as master" references from this doc once this terminology is fully removed from k8s-->
## Control Plane (previously referred to as master) Node Setup

**NOTE**: This setup is not compatible with Amazon EKS.

To run a CA pod on a control plane node the CA deployment should tolerate the `master`
taint and `nodeSelector` should be used to schedule the pods on a control plane node.
Please replace `{{ node_asg_min }}`, `{{ node_asg_max }}` and `{{ name }}` with
your ASG setting in the yaml file.
```
kubectl apply -f examples/cluster-autoscaler-run-on-control-plane.yaml
```

## Using Mixed Instances Policies and Spot Instances

**NOTE:** The minimum version of cluster autoscaler to support MixedInstancePolicy is v1.14.x.

If your workloads can tolerate interruption, consider taking advantage of Spot
Instances for a lower price point. To enable diversity among On Demand and Spot
Instances, as well as specify multiple EC2 instance types in order to tap into
multiple Spot capacity pools, use a [mixed instances
policy](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-autoscalinggroup-mixedinstancespolicy.html)
on your ASG. Note that the instance types should have the same amount of RAM and
number of CPU cores, since this is fundamental to CA's scaling calculations.
Using mismatched instances types can produce unintended results. See an example
below.

Additionally, there are other factors which affect scaling, such as node labels.
If you are currently using `nodeSelector` with the
[beta.kubernetes.io/instance-type](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#interlude-built-in-node-labels)
label, you will need to apply a common propagating label to the ASG and use that
instead, since the instance-type label can no longer be relied upon. One may
also use auto-generated tags such as `aws:cloudformation:stack-name` for this
purpose. [Node affinity and
anti-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)
are not affected in the same way, since these selectors natively accept multiple
values; one must add all the configured instances types to the list of values,
for example:

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

* Create a [Launch
  Template](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-autoscalinggroup-launchtemplate.html)
  (LT) with an instance type, for example, r5.2xlarge. Consider this the 'base'
  instance type. Do not define any spot purchase options here.
* Create an ASG with a MixedInstancesPolicy that refers to the newly-created LT.
* Set LaunchTemplateOverrides to include the 'base' instance type r5.2xlarge and
  suitable alternatives, e.g. r5d.2xlarge, i3.2xlarge, r5a.2xlarge and
  r5ad.2xlarge. Differing processor types and speeds should be evaluated
  depending on your use-case(s).
* Set
  [InstancesDistribution](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_InstancesDistribution.html)
  according to your needs.
* See [Allocation
  Strategies](https://docs.aws.amazon.com/autoscaling/ec2/userguide/asg-purchase-options.html#asg-allocation-strategies)
  for information about how the ASG fulfills capacity from the specified instance
  types. It is recommended to use the capacity-optimized allocation strategy,
  which will automatically launch Spot Instances into the most available pools
  by looking at real-time capacity data and.
* For the same workload or for the generic capacity in your cluster, you can
  also create more node groups with a vCPU/Mem ratio that is a good fit for your
  workloads, but from different instance sizes. For example: Node group 1:
  m5.xlarge, m5a.xlarge, m5d.xlarge, m5ad.xlarge, m4.xlarge. Node group 2:
  m5.2xlarge, m5a.2xlarge, m5d.2xlarge, m5ad.2xlarge, m4.2xlarge. This approach
  increases the chance of achieving your desired scale at the lowest cost by
  tapping into many Spot capacity pools.

See CloudFormation example [here](MixedInstancePolicy.md).

## Use Static Instance List
The set of the latest supported EC2 instance types will be fetched by the CA at
run time. You can find all the available instance types in the CA logs. If your
network access is restricted such that fetching this set is infeasible, you can
specify the command-line flag `--aws-use-static-instance-list=true` to switch
the CA back to its original use of a statically defined set.

To refresh static list, please run `go run ec2_instance_types/gen.go` under
`cluster-autoscaler/cloudprovider/aws/` and update `staticListLastUpdateTime` in
`aws_util.go`

## Common Notes and Gotchas:
* The `/etc/ssl/certs/ca-bundle.crt` should exist by default on ec2 instance in
  your EKS cluster. If you use other cluster privision tools like
  [kops](https://github.com/kubernetes/kops) with different operating systems
  other than Amazon Linux 2, please use `/etc/ssl/certs/ca-certificates.crt` or
  correct path on your host instead for the volume hostPath in your cluster
  autoscaler manifest.
* If you’re using Persistent Volumes, your deployment needs to run in the same
  AZ as where the EBS volume is, otherwise the pod scheduling could fail if it
  is scheduled in a different AZ and cannot find the EBS volume. To overcome
  this, either use a single AZ ASG for this use case, or an ASG-per-AZ while
  enabling
  [--balance-similar-node-groups](../../FAQ.md#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler).
  Alternately, and depending on your use-case, you might be able to switch from
  using EBS to using shared storage that is available across AZs (for each pod
  in its respective AZ). Consider AWS services like Amazon EFS or Amazon FSx for
  Lustre.
* On creation time, the ASG will have the [AZRebalance
  process](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-benefits.html#AutoScalingBehavior.InstanceUsage)
  enabled, which means it will actively work to balance the number of instances
  between AZs, and possibly terminate instances. If your applications could be
  impacted from sudden termination, you can either suspend the AZRebalance
  feature, or use a tool for automatic draining upon ASG scale-in such as the
  [k8s-node-drainer](https://github.com/aws-samples/amazon-k8s-node-drainer). The
  [AWS Node Termination
  Handler](https://github.com/aws/aws-node-termination-handler/issues/95) will
  also support this use-case in the future.
* By default, cluster autoscaler will not terminate nodes running pods in the
  kube-system namespace. You can override this default behaviour by passing in
  the `--skip-nodes-with-system-pods=false` flag.
* By default, cluster autoscaler will wait 10 minutes between scale down
  operations, you can adjust this using the `--scale-down-delay-after-add`,
  `--scale-down-delay-after-delete`, and `--scale-down-delay-after-failure`
  flag. E.g. `--scale-down-delay-after-add=5m` to decrease the scale down delay
  to 5 minutes after a node has been added.
* If you're running multiple ASGs, the `--expander` flag supports three options:
  `random`, `most-pods` and `least-waste`. `random` will expand a random ASG on
  scale up. `most-pods` will scale up the ASG that will schedule the most amount
  of pods. `least-waste` will expand the ASG that will waste the least amount of
  CPU/MEM resources. In the event of a tie, cluster autoscaler will fall back to
  `random`.
* If you're managing your own kubelets, they need to be started with the
  `--provider-id` flag. The provider id has the format
  `aws:///<availability-zone>/<instance-id>`, e.g.
  `aws:///us-east-1a/i-01234abcdef`.
* If you want to use regional STS endpoints (e.g. when using VPC endpoint for
  STS) the env `AWS_STS_REGIONAL_ENDPOINTS=regional` should be set.
