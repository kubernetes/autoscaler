# cluster-autoscaler-chart

Scales Kubernetes worker nodes within autoscaling groups.

## TL;DR:

```console
$ helm repo add autoscaler https://kubernetes.github.io/autoscaler

# Method 1 - Using Autodiscovery
$ helm install my-release autoscaler/cluster-autoscaler-chart \
--set 'autoDiscovery.clusterName'=<CLUSTER NAME>

# Method 2 - Specifying groups manually
$ helm install my-release autoscaler/cluster-autoscaler-chart \
--set "autoscalingGroups[0].name=your-asg-name" \
--set "autoscalingGroups[0].maxSize=10" \
--set "autoscalingGroups[0].minSize=1"
```

## Introduction

This chart bootstraps a cluster-autoscaler deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Helm 3+
- Kubernetes 1.8+
  - [Older versions](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler#releases) may work by overriding the `image`. Cluster autoscaler internally simulates the scheduler and bugs between mismatched versions may be subtle.
- Azure AKS specific Prerequisites:
  - Kubernetes 1.10+ with RBAC-enabled.

## Previous Helm Chart

The previous `cluster-autoscaler` Helm chart hosted at [helm/charts](https://github.com/helm/charts) has been moved to this repository in accordance with the [Deprecation timeline](https://github.com/helm/charts#deprecation-timeline). Note that a few things have changed between this version and the old version:

- This repository **only** supports Helm chart installations using Helm 3+ since the `apiVersion` on the charts has been marked as `v2`.
- Previous versions of the Helm chart have not been migrated, and the version was reset to `1.0.0` at the onset. If you are looking for old versions of the chart, it's best to run `helm pull stable/cluster-autoscaler --version <your-version>` until you are ready to move to this repository's version.

## Installing the Chart

**By default, no deployment is created and nothing will autoscale**.

You must provide some minimal configuration, either to specify instance groups or enable auto-discovery. It is not recommended to do both.

Either:

- Set `autoDiscovery.clusterName` and provide additional autodiscovery options if necessary **or**
- Set static node group configurations for one or more node groups (using `autoscalingGroups` or `autoscalingGroupsnamePrefix`).

To create a valid configuration, follow instructions for your cloud provider:

* [AWS](#aws---using-auto-discovery-of-tagged-instance-groups)
* [GCE](#gce)
* [Azure AKS](#azure-aks)
* [OpenStack Magnum](#openstack-magnum)

### AWS - Using auto-discovery of tagged instance groups

Auto-discovery finds ASGs tags as below and automatically manages them based on the min and max size specified in the ASG. `cloudProvider=aws` only.

- Tag the ASGs with keys to match `.Values.autoDiscovery.tags`, by default: `k8s.io/cluster-autoscaler/enabled` and `k8s.io/cluster-autoscaler/<YOUR CLUSTER NAME>`
- Verify the [IAM Permissions](#aws---iam)
- Set `autoDiscovery.clusterName=<YOUR CLUSTER NAME>`
- Set `awsRegion=<YOUR AWS REGION>`
- Set `awsAccessKeyID=<YOUR AWS KEY ID>` and `awsSecretAccessKey=<YOUR AWS SECRET KEY>` if you want to [use AWS credentials directly instead of an instance role](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md#using-aws-credentials)

```console
$ helm install my-release autoscaler/cluster-autoscaler-chart --set autoDiscovery.clusterName=<CLUSTER NAME>
```

#### Specifying groups manually

Without autodiscovery, specify an array of elements each containing ASG name, min size, max size. The sizes specified here will be applied to the ASG, assuming IAM permissions are correctly configured.

- Verify the [IAM Permissions](#aws---iam)
- Either provide a yaml file setting `autoscalingGroups` (see values.yaml) or use `--set` e.g.:

```console
$ helm install my-release autoscaler/cluster-autoscaler-chart \
--set "autoscalingGroups[0].name=your-asg-name" \
--set "autoscalingGroups[0].maxSize=10" \
--set "autoscalingGroups[0].minSize=1"
```

#### Auto-discovery

For auto-discovery of instances to work, they must be tagged with the keys in `.Values.autoDiscovery.tags`, which by default are
`k8s.io/cluster-autoscaler/enabled` and `k8s.io/cluster-autoscaler/<ClusterName>`

The value of the tag does not matter, only the key.

An example kops spec excerpt:

```yaml
apiVersion: kops/v1alpha2
kind: Cluster
metadata:
  name: my.cluster.internal
spec:
  additionalPolicies:
    node: |
      [
        {"Effect":"Allow","Action":["autoscaling:DescribeAutoScalingGroups","autoscaling:DescribeAutoScalingInstances","autoscaling:DescribeLaunchConfigurations","autoscaling:DescribeTags","autoscaling:SetDesiredCapacity","autoscaling:TerminateInstanceInAutoScalingGroup"],"Resource":"*"}
      ]
      ...
---
apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: my.cluster.internal
  name: my-instances
spec:
  cloudLabels:
    k8s.io/cluster-autoscaler/enabled: ""
    k8s.io/cluster-autoscaler/my.cluster.internal: ""
  image: kops.io/k8s-1.8-debian-jessie-amd64-hvm-ebs-2018-01-14
  machineType: r4.large
  maxSize: 4
  minSize: 0
```

In this example you would need to `--set autoDiscovery.clusterName=my.cluster.internal` when installing.

It is not recommended to try to mix this with setting `autoscalingGroups`

See [autoscaler AWS documentation](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md#auto-discovery-setup) for a more discussion of the setup.

### GCE

The following parameters are required:

- `autoDiscovery.clusterName=any-name`
- `cloud-provider=gce`
- `autoscalingGroupsnamePrefix[0].name=your-ig-prefix,autoscalingGroupsnamePrefix[0].maxSize=10,autoscalingGroupsnamePrefix[0].minSize=1`

To use Managed Instance Group (MIG) auto-discovery, provide a YAML file setting `autoscalingGroupsnamePrefix` (see values.yaml) or use `--set` when installing the Chart - e.g.

```console
$ helm install my-release autoscaler/cluster-autoscaler-chart \
--set "autoscalingGroupsnamePrefix[0].name=your-ig-prefix,autoscalingGroupsnamePrefix[0].maxSize=10,autoscalingGroupsnamePrefi[0].minSize=1" \
--set autoDiscovery.clusterName=<CLUSTER NAME> \
--set cloudProvider=gce
```

Note that `your-ig-prefix` should be a _prefix_ matching one or more MIGs, and _not_ the full name of the MIG. For example, to match multiple instance groups - `k8s-node-group-a-standard`, `k8s-node-group-b-gpu`, you would use a prefix of `k8s-node-group-`.

In the event you want to explicitly specify MIGs instead of using auto-discovery, set members of the `autoscalingGroups` array directly - e.g.

```
# where 'n' is the index, starting at 0
-- set autoscalingGroups[n].name=https://content.googleapis.com/compute/v1/projects/$PROJECTID/zones/$ZONENAME/instanceGroupManagers/$FULL-MIG-NAME,autoscalingGroups[n].maxSize=$MAXSIZE,autoscalingGroups[n].minSize=$MINSIZE
```

### Azure AKS

The following parameters are required:

- `cloudProvider=azure`
- `autoscalingGroups[0].name=your-agent-pool,autoscalingGroups[0].maxSize=10,autoscalingGroups[0].minSize=1`
- `azureClientID: "your-service-principal-app-id"`
- `azureClientSecret: "your-service-principal-client-secret"`
- `azureSubscriptionID: "your-azure-subscription-id"`
- `azureTenantID: "your-azure-tenant-id"`
- `azureClusterName: "your-aks-cluster-name"`
- `azureResourceGroup: "your-aks-cluster-resource-group-name"`
- `azureVMType: "AKS"`
- `azureNodeResourceGroup: "your-aks-cluster-node-resource-group"`

### OpenStack Magnum

`cloudProvider: magnum` must be set, and then one of

- `magnumClusterName=<cluster name or ID>` and `autoscalingGroups` with the names of node groups and min/max node counts
- or `autoDiscovery.clusterName=<cluster name or ID>` with one or more `autoDiscovery.roles`.

Additionally, `cloudConfigPath: "/etc/kubernetes/cloud-config"` must be set as this should be the location
of the cloud-config file on the host.

Example values files can be found [here](../../cluster-autoscaler/cloudprovider/magnum/examples).

Install the chart with

```
$ helm install my-release autoscaler/cluster-autoscaler-chart -f myvalues.yaml
```

## Uninstalling the Chart

To uninstall `my-release`:

```console
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

> **Tip**: List all releases using `helm list` or start clean with `helm uninstall my-release`

## Additional Configuration

### AWS - IAM

The worker running the cluster autoscaler will need access to certain resources and actions:

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
        "autoscaling:DescribeTags",
        "autoscaling:SetDesiredCapacity",
        "autoscaling:TerminateInstanceInAutoScalingGroup"
      ],
      "Resource": "*"
    }
  ]
}
```

- `DescribeTags` is required for autodiscovery.
- `DescribeLaunchConfigurations` is required to scale up an ASG from 0.

If you would like to limit the scope of the Cluster Autoscaler to ***only*** modify ASGs for a particular cluster, use the following policy instead:

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
        "autoscaling:DescribeTags",
        "ec2:DescribeLaunchTemplateVersions"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:SetDesiredCapacity",
        "autoscaling:TerminateInstanceInAutoScalingGroup",
        "autoscaling:UpdateAutoScalingGroup"
      ],
      "Resource": [
        "arn:aws:autoscaling:<aws-region>:<account-id>:autoScalingGroup:<some-random-id>:autoScalingGroupName/node-group-1",
        "arn:aws:autoscaling:<aws-region>:<account-id>:autoScalingGroup:<some-random-id>:autoScalingGroupName/node-group-2",
        "arn:aws:autoscaling:<aws-region>:<account-id>:autoScalingGroup:<some-random-id>:autoScalingGroupName/node-group-3"
      ],
      "Condition": {
        "StringEquals": {
          "autoscaling:ResourceTag/k8s.io/cluster-autoscaler/enabled": "true",
          "autoscaling:ResourceTag/kubernetes.io/cluster/<cluster-name>": "owned"
        }
      }
    }
  ]
}
```

Make sure to replace the variables `<aws-region>`, `<cluster-name>`, `<account-id>`, and the ARNs of the ASGs where applicable.

### AWS - IAM Roles for Service Accounts (IRSA)

For Kubernetes clusters that use Amazon EKS, the service account can be configured with an IAM role using [IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) to avoid needing to grant access to the worker nodes for AWS resources.

In order to accomplish this, you will first need to create a new IAM role with the above mentions policies.  Take care in [configuring the trust relationship](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts-technical-overview.html#iam-role-configuration) to restrict access just to the service account used by cluster autoscaler.

Once you have the IAM role configured, you would then need to `--set rbac.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::123456789012:role/MyRoleName` when installing.

## Troubleshooting

The chart will succeed even if the container arguments are incorrect. A few minutes after starting
`kubectl logs -l "app=aws-cluster-autoscaler" --tail=50` should loop through something like

```
polling_autoscaler.go:111] Poll finished
static_autoscaler.go:97] Starting main loop
utils.go:435] No pod using affinity / antiaffinity found in cluster, disabling affinity predicate for this loop
static_autoscaler.go:230] Filtering out schedulables
```

If not, find a pod that the deployment created and `describe` it, paying close attention to the arguments under `Command`. e.g.:

```
Containers:
  cluster-autoscaler:
    Command:
      ./cluster-autoscaler
      --cloud-provider=aws
# if specifying ASGs manually
      --nodes=1:10:your-scaling-group-name
# if using autodiscovery
      --node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/<ClusterName>
      --v=4
```

### PodSecurityPolicy

Though enough for the majority of installations, the default PodSecurityPolicy _could_ be too restrictive depending on the specifics of your release. Please make sure to check that the template fits with any customizations made or disable it by setting `rbac.pspEnabled` to `false`.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod assignment |
| autoDiscovery.clusterName | string | `nil` | Enable autodiscovery for `cloudProvider=aws`, for groups matching `autoDiscovery.tags`. Enable autodiscovery for `cloudProvider=gce`, but no MIG tagging required. Enable autodiscovery for `cloudProvider=magnum`, for groups matching `autoDiscovery.roles`. |
| autoDiscovery.roles | list | `["worker"]` | Magnum node group roles to match. |
| autoDiscovery.tags | list | `["k8s.io/cluster-autoscaler/enabled","k8s.io/cluster-autoscaler/{{ .Values.autoDiscovery.clusterName }}"]` | ASG tags to match, run through `tpl`. |
| autoscalingGroups | list | `[]` | For AWS, Azure AKS or Magnum. At least one element is required if not using `autoDiscovery`. For example: <pre> - name: asg1<br />   maxSize: 2<br />   minSize: 1 </pre> |
| autoscalingGroupsnamePrefix | list | `[]` | For GCE. At least one element is required if not using `autoDiscovery`. For example: <pre> - name: ig01<br />   maxSize: 10<br />   minSize: 0 </pre> |
| awsAccessKeyID | string | `""` | AWS access key ID ([if AWS user keys used](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md#using-aws-credentials)) |
| awsRegion | string | `"us-east-1"` | AWS region (required if `cloudProvider=aws`) |
| awsSecretAccessKey | string | `""` | AWS access secret key ([if AWS user keys used](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md#using-aws-credentials)) |
| azureClientID | string | `""` | Service Principal ClientID with contributor permission to Cluster and Node ResourceGroup. Required if `cloudProvider=azure` |
| azureClientSecret | string | `""` | Service Principal ClientSecret with contributor permission to Cluster and Node ResourceGroup. Required if `cloudProvider=azure` |
| azureClusterName | string | `""` | Azure AKS cluster name. Required if `cloudProvider=azure` |
| azureNodeResourceGroup | string | `""` | Azure resource group where the cluster's nodes are located, typically set as `MC_<cluster-resource-group-name>_<cluster-name>_<location>`. Required if `cloudProvider=azure` |
| azureResourceGroup | string | `""` | Azure resource group that the cluster is located. Required if `cloudProvider=azure` |
| azureSubscriptionID | string | `""` | Azure subscription where the resources are located. Required if `cloudProvider=azure` |
| azureTenantID | string | `""` | Azure tenant where the resources are located. Required if `cloudProvider=azure` |
| azureUseManagedIdentityExtension | bool | `false` | Whether to use Azure's managed identity extension for credentials. If using MSI, ensure subscription ID and resource group are set. |
| azureVMType | string | `"AKS"` | Azure VM type. |
| cloudConfigPath | string | `"/etc/gce.conf"` | Configuration file for cloud provider. |
| cloudProvider | string | `"aws"` | The cloud provider where the autoscaler runs. Currently only `gce`, `aws`, `azure` and `magnum` are supported. `aws` supported for AWS. `gce` for GCE. `azure` for Azure AKS. `magnum` for OpenStack Magnum. |
| containerSecurityContext | object | `{}` | [Security context for container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) |
| dnsPolicy | string | `"ClusterFirst"` | Defaults to `ClusterFirst`. Valid values are: `ClusterFirstWithHostNet`, `ClusterFirst`, `Default` or `None`. If autoscaler does not depend on cluster DNS, recommended to set this to `Default`. |
| expanderPriorities | object | `{}` | The expanderPriorities is used if `extraArgs.expander` is set to `priority` and expanderPriorities is also set with the priorities. If `extraArgs.expander` is set to `priority`, then expanderPriorities is used to define cluster-autoscaler-priority-expander priorities. See: https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/expander/priority/readme.md |
| extraArgs | object | `{"logtostderr":true,"stderrthreshold":"info","v":4}` | Additional container arguments. |
| extraEnv | object | `{}` | Additional container environment variables. |
| fullnameOverride | string | `""` | String to fully override `cluster-autoscaler.fullname` template. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.pullSecrets | list | `[]` | Image pull secrets |
| image.repository | string | `"us.gcr.io/k8s-artifacts-prod/autoscaling/cluster-autoscaler"` | Image repository |
| image.tag | string | `"v1.18.1"` | Image tag |
| kubeTargetVersionOverride | string | `""` | Allow overridding the `.Capabilities.KubeVersion.GitVersion` check. Useful for `helm template` commands. |
| magnumCABundlePath | string | `"/etc/kubernetes/ca-bundle.crt"` | Path to the host's CA bundle, from `ca-file` in the cloud-config file. |
| magnumClusterName | string | `""` | Cluster name or ID in Magnum. Required if `cloudProvider=magnum` and not setting `autoDiscovery.clusterName`. |
| nameOverride | string | `""` | String to partially override `cluster-autoscaler.fullname` template (will maintain the release name) |
| nodeSelector | object | `{}` | Node labels for pod assignment. Ref: https://kubernetes.io/docs/user-guide/node-selection/. |
| podAnnotations | object | `{}` | Annotations to add to each pod. |
| podDisruptionBudget | object | `{"maxUnavailable":1}` | Pod disruption budget. |
| podLabels | object | `{}` | Labels to add to each pod. |
| priorityClassName | string | `""` | priorityClassName |
| rbac.create | bool | `true` | If `true`, create and use RBAC resources. |
| rbac.pspEnabled | bool | `false` | If `true`, creates and uses RBAC resources required in the cluster with [Pod Security Policies](https://kubernetes.io/docs/concepts/policy/pod-security-policy/) enabled. Must be used with `rbac.create` set to `true`. |
| rbac.serviceAccount.annotations | object | `{}` | Additional Service Account annotations. |
| rbac.serviceAccount.create | bool | `true` | If `true` and `rbac.create` is also true, a Service Account will be created. |
| rbac.serviceAccount.name | string | `""` | The name of the ServiceAccount to use. If not set and create is `true`, a name is generated using the fullname template. |
| replicaCount | int | `1` | Desired number of pods |
| resources | object | `{}` | Pod resource requests and limits. |
| securityContext | object | `{}` | [Security context for pod](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) |
| service.annotations | object | `{}` | Annotations to add to service |
| service.externalIPs | list | `[]` | List of IP addresses at which the service is available. Ref: https://kubernetes.io/docs/user-guide/services/#external-ips. |
| service.labels | object | `{}` | Labels to add to service |
| service.loadBalancerIP | string | `""` | IP address to assign to load balancer (if supported). |
| service.loadBalancerSourceRanges | list | `[]` | List of IP CIDRs allowed access to load balancer (if supported). |
| service.portName | string | `"http"` | Name for service port. |
| service.servicePort | int | `8085` | Service port to expose. |
| service.type | string | `"ClusterIP"` | Type of service to create. |
| serviceMonitor.enabled | bool | `false` | If true, creates a Prometheus Operator ServiceMonitor. |
| serviceMonitor.interval | string | `"10s"` | Interval that Prometheus scrapes Cluster Autoscaler metrics. |
| serviceMonitor.namespace | string | `"monitoring"` | Namespace which Prometheus is running in. |
| serviceMonitor.path | string | `"/metrics"` | The path to scrape for metrics; autoscaler exposes `/metrics` (this is standard) |
| serviceMonitor.selector | object | `{"release":"prometheus-operator"}` | Default to kube-prometheus install (CoreOS recommended), but should be set according to Prometheus install. |
| tolerations | list | `[]` | List of node taints to tolerate (requires Kubernetes >= 1.6). |
| updateStrategy | object | `{}` | [Deployment update strategy](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy) |
