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
                "tke:DeleteClusterInstances",
                "tke:DescribeClusterAsGroups",
                "as:ModifyAutoScalingGroup",
                "as:RemoveInstances",
                "as:StopAutoScalingInstances",
                "as:DescribeAutoScalingGroups",
                "as:DescribeAutoScalingInstances",
                "as:DescribeLaunchConfigurations",
                "as:DescribeAutoScalingActivities"
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
  - name: REGION_NAME
    value: YOUR_TENCENCLOUD_REGION_NAME
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
        - name: REGION_NAME
          value: YOUR_TENCENCLOUD_REGION_NAME
        - name: CLUSTER_ID
          value: YOUR_TKE_CLUSTER_ID
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
        - cbs.api.qcloud.com
        - cvm.api.qcloud.com
        - lb.api.qcloud.com
        - tag.api.qcloud.com
        - snapshot.api.qcloud.com
        - monitor.api.qcloud.com
        - scaling.api.qcloud.com
        - ccs.api.qcloud.com
        ip: 169.254.0.28
      - hostnames:
        - tke.internal.tencentcloudapi.com
        - clb.internal.tencentcloudapi.com
        - cvm.internal.tencentcloudapi.com
        - tag.internal.tencentcloudapi.com
        - as.tencentcloudapi.com
        - cbs.tencentcloudapi.com
        - cvm.tencentcloudapi.com
        - vpc.tencentcloudapi.com
        - tke.tencentcloudapi.com
        ip: 169.254.0.95
      restartPolicy: Always
      serviceAccount: kube-admin
      serviceAccountName: kube-admin
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      - hostPath:
          path: /etc/localtime
          type: ""
        name: tz-config
```

### Auto-Discovery Setup

Auto Discovery is not supported in TencentCloud currently.