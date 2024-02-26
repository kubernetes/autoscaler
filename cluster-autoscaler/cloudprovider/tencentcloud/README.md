# Cluster Autoscaler on TencentCloud

On TencentCloud, Cluster Autoscaler utilizes CVM Auto Scaling Groups to manage node
groups. Cluster Autoscaler typically runs as a `Deployment` in your cluster.

## Requirements

Cluster Autoscaler requires [TKE](https://intl.cloud.tencent.com/document/product/457) v1.10.x or greater.

## Permissions

### CAM Policy

The following policy provides the minimum privileges necessary for Cluster Autoscaler to run:

```json
{
    "version": "2.0",
    "statement": [
        {
            "effect": "allow",
            "action": [
                "as:ModifyAutoScalingGroup",
                "as:RemoveInstances",
                "as:DescribeAutoScalingGroups",
                "as:DescribeAutoScalingInstances",
                "as:DescribeLaunchConfigurations",
                "as:DescribeAutoScalingActivities",
                "cvm:DescribeZones",
                "cvm:DescribeInstanceTypeConfigs",
                "vpc:DescribeSubnets"
            ],
            "resource": [
                "*"
            ]
        }
    ]
}
```

### Using TencentCloud Credentials

> NOTICE: Make sure the [access key](https://intl.cloud.tencent.com/document/product/598/32675) you will be using has all the above permissions


```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tencentcloud-secret
type: Opaque
data:
  tencentcloud_secret_id: BASE64_OF_YOUR_TENCENTCLOUD_SECRET_ID
  tencentcloud_secret_key: BASE64_OF_YOUR_TENCENTCLOUD_SECRET_KEY
```

Please refer to the [relevant Kubernetes
documentation](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret-manually)
for creating a secret manually.

```yaml
env:
  - name: SECRET_ID
    valueFrom:
      secretKeyRef:
        name: tencentcloud-secret
        key: tencentcloud_secret_id
  - name: SECRET_KEY
    valueFrom:
      secretKeyRef:
        name: tencentcloud-secret
        key: tencentcloud_secret_key
  - name: REGION
    value: YOUR_TENCENCLOUD_REGION
  - name: CLUSTER_ID
    value: YOUR_TKE_CLUSTER_ID
```

## Setup

### cluster-autoscaler deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  selector:
    matchLabels:
      qcloud-app: cluster-autoscaler
  template:
    metadata:
      labels:
        qcloud-app: cluster-autoscaler
    spec:
      containers:
      - args:
        - --cloud-provider=tencentcloud
        - --v=4
        - --ok-total-unready-count=3
        - --cloud-config=/etc/kubernetes/qcloud.conf
        - --scale-down-utilization-threshold=0.8
        - --scale-down-enabled=true
        - --max-total-unready-percentage=33
        - --nodes=[min]:[max]:[ASG_ID]
        - --logtostderr
        - --kubeconfig=/kubeconfig/config
        command:
        - /cluster-autoscaler
        env:
        - name: SECRET_ID
          valueFrom:
            secretKeyRef:
              name: tencentcloud-secret
              key: tencentcloud_secret_id
        - name: SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: tencentcloud-secret
              key: tencentcloud_secret_key
        - name: REGION
          value: YOUR_TENCENCLOUD_REGION
        image: ccr.ccs.tencentyun.com/tkeimages/cluster-autoscaler:v1.18.4-49692187a
        imagePullPolicy: Always
        name: cluster-autoscaler
        resources:
          limits:
            cpu: "1"
            memory: 1Gi
          requests:
            cpu: 250m
            memory: 256Mi
        volumeMounts:
        - mountPath: /etc/localtime
          name: tz-config
      hostAliases:
      - hostnames:
        - as.tencentcloudapi.com
        - cvm.tencentcloudapi.com
        - vpc.tencentcloudapi.com
        ip: 169.254.0.95
      restartPolicy: Always
      serviceAccount: kube-admin
      serviceAccountName: kube-admin
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
      volumes:
      - hostPath:
          path: /etc/localtime
          type: ""
        name: tz-config
```

### Scaling up from 0 nodes

When scaling up from 0 nodes, the Cluster Autoscaler reads ASG tags to derive information about the specifications of the nodes
i.e labels and taints in that ASG. Note that it does not actually apply these labels or taints - this is done by an AWS generated
user data script. It gives the Cluster Autoscaler information about whether pending pods will be able to be scheduled should a new node
be spun up for a particular ASG with the asumption the ASG tags accurately reflect the labels/taint actually applied.

The following is only required if scaling up from 0 nodes. The Cluster Autoscaler will require the label tag
on the ASG should a deployment have a NodeSelector, else no scaling will occur as the Cluster Autoscaler does not realise
the ASG has that particular label. The tag is of the format
`k8s.io/cluster-autoscaler/node-template/label/<label-name>`: `<label-value>` or `tencentcloud:<label-name>`: `<label-value>` is
the name of the label and the value of each tag specifies the label value.

Example tags:

- `k8s.io/cluster-autoscaler/node-template/label/foo`: `bar`
- `tencentcloud:foo`:`bar`

The following is only required if scaling up from 0 nodes. The Cluster Autoscaler will require the taint tag
on the ASG, else tainted nodes may get spun up that cannot actually have the pending pods run on it. The tag is of the format
`k8s.io/cluster-autoscaler/node-template/taint/<taint-name>`:`<taint-value:taint-effect>` is
the name of the taint and the value of each tag specifies the taint value and effect with the format `<taint-value>:<taint-effect>`.

Example tags:

- `k8s.io/cluster-autoscaler/node-template/taint/dedicated`: `true:NoSchedule`

From version 1.14, Cluster Autoscaler can also determine the resources provided
by each Auto Scaling Group via tags. The tag is of the format
`k8s.io/cluster-autoscaler/node-template/resources/<resource-name>`.
`<resource-name>` is the name of the resource, such as `ephemeral-storage`. The
value of each tag specifies the amount of resource provided. The units are
identical to the units used in the `resources` field of a Pod specification.

Example tags:

- `k8s.io/cluster-autoscaler/node-template/resources/ephemeral-storage`: `100G`