# Cluster Autoscaler on AWS
The cluster autoscaler on AWS scales worker nodes within an autoscaling group. It will run as a `Deployment` in your cluster. This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Kubernetes Version
Cluster autoscaler must run on v1.3.0 or greater.

## Permissions
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
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
```
Unfortunately AWS does not support ARNs for autoscaling groups yet so you must use "*" as the resource. More information [here](http://docs.aws.amazon.com/autoscaling/latest/userguide/IAM.html#UsingWithAutoScaling_Actions).

## Deployment Specification
Your deployment configuration should look something like this:
```yaml
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: cluster-autoscaler
  labels:
    app: cluster-autoscaler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cluster-autoscaler
  template:
    metadata:
      labels:
        app: cluster-autoscaler
    spec:
      containers:
        - image: {{ YOUR IMAGE HERE }}
          name: cluster-autoscaler
          resources:
            limits:
              cpu: 100m
              memory: 300Mi
            requests:
              cpu: 100m
              memory: 300Mi
          command:
            - ./cluster-autoscaler
            - -v=4
            - --cloud-provider=aws
            - --skip-nodes-with-local-storage=false
            - --nodes={{ ASG MIN e.g. 1 }}:{{ASG MAX e.g. 5}}:{{ASG NAME e.g. k8s-worker-asg}}
          env:
            - name: AWS_REGION
              value: us-east-1
          volumeMounts:
            - name: ssl-certs
              mountPath: /etc/ssl/certs/ca-certificates.crt
              readOnly: true
          imagePullPolicy: "Always"
      volumes:
        - name: ssl-certs
          hostPath:
            path: "/etc/ssl/certs/ca-certificates.crt"
```
Note:
- The `/etc/ssl/certs/ca-certificates.crt` should exist by default on your ec2 instance.
- The autoscaling group should span 1 availability zone for the cluster autoscaler to work. If you want to distribute workloads evenly across zones, set up multiple ASGs, with a cluster autoscaler for each ASG. At the time of writing this, cluster autoscaler is unaware of availability zones and although autoscaling groups can contain instances in multiple availability zones when configured so, the cluster autoscaler can't reliably add nodes to desired zones. That's because AWS AutoScaling determines which zone to add nodes which is out of the control of the cluster autoscaler. For more information, see https://github.com/kubernetes/contrib/pull/1552#discussion_r75533090.
