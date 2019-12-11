#### The following is an example to make use of the AWS IAM OIDC with the Cluster Autoscaler in an EKS cluster. 


#### Prerequisites 

  - An Active EKS cluster (1.14 preferred since it is the latest) against which the user is able to run kubectl commands. 
  - Cluster must consist of at least one worker node ASG. 

A) Create an IAM OIDC identity provider for your cluster with the AWS Management Console using the [documentation] . 

B) Create a test [IAM policy] for your service accounts.

```sh
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::my-pod-secrets-bucket/*"
      ]
    }
  ]
}
```

C) Create an IAM role for your service accounts in the console.
- Retrieve the OIDC issuer URL from the Amazon EKS console description of your cluster . It will look something identical to: 
'https://oidc.eks.us-east-1.amazonaws.com/id/xxxxxxxxxx'
- While creating a new IAM role, In the "Select type of trusted entity" section, choose "Web identity".
- In the "Choose a web identity provider" section:
For Identity provider, choose the URL for your cluster.
For Audience, type sts.amazonaws.com.

- In the "Attach Policy" section, select the policy to use for your service account, that you created in Section B above. 
- After the role is created, choose the role in the console to open it for editing.
- Choose the "Trust relationships" tab, and then choose "Edit trust relationship".
Edit the OIDC provider suffix and change it from :aud to :sub.
Replace sts.amazonaws.com to your service account ID.
- Update trust policy to finish. 

D) Set up [Cluster Autoscaler Auto-Discovery] using the [tutorial] . 
- Open the Amazon EC2 console, and then choose EKS worker node Auto Scaling Groups from the navigation pane.
- In the "Add/Edit Auto Scaling Group Tags" window, please make sure you enter the following tags by replacing 'awsExampleClusterName' with the name of your EKS cluster. Then, choose "Save".

| Plugin | README |
| ------ | ------ |
| Key: | k8s.io/cluster-autoscaler/enabled |
| Key: | k8s.io/cluster-autoscaler/'awsExampleClusterName' |

Note: The keys for the tags that you entered don't have values. Cluster Autoscaler ignores any value set for the keys.

- Create an IAM Policy for cluster autoscaler and to enable AutoDiscovery. 

```sh
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

NOTE: ``` autoscaling:DescribeTags ``` is very important if you are making use of the AutoDiscovery feature of the Cluster AutoScaler. 

- Attach the above created policy to the *instance role* that's attached to your Amazon EKS worker nodes.
- Download a deployment example file provided by the Cluster Autoscaler project on GitHub, run the following command:

```sh
$ wget https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml
```

- Open the downloaded YAML file in an editor. 

##### Change 1: 

Set the EKS cluster name (awsExampleClusterName) and environment variable (us-east-1) based on the following example. 

```sh
    spec:
      serviceAccountName: cluster-autoscaler
      containers:
        - image: gcr.io/google-containers/cluster-autoscaler:v1.14.6     #cluster-autoscaler image
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
            - --v=4
            - --stderrthreshold=info
            - --cloud-provider=aws
            - --skip-nodes-with-local-storage=false
            - --expander=least-waste
            - --node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/<<awsExampleClusterName>>
          env:
            - name: AWS_REGION
              value: <<us-east-1>>
```

##### Change 2: 

To use IAM with OIDC, you will have to make the below changes to the file as well. 

```sh
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::xxxxx:role/Amazon_CA_role   # Add the IAM role created in the above C section.
  name: cluster-autoscaler
  namespace: kube-system
```

- Following this setup, you can test if the cluster-autoscaler kicked in and if the role was attached using the below commands:

```sh
$ kubectl get pods -n kube-system
$ kubectl exec -n kube-system cluster-autoscaler-xxxxxx-xxxxx  env | grep AWS
```

Output of the exec command should ideally display the values for AWS_REGION, AWS_ROLE_ARN and AWS_WEB_IDENTITY_TOKEN_FILE where the role arn must be the same as the role provided in the service account annotations. 

The cluster autoscaler scaling the worker nodes can also be tested: 

```sh
$ kubectl scale deployment autoscaler-demo --replicas=50
deployment.extensions/autoscaler-demo scaled
 
$ kubectl get deployment
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
autoscaler-demo   55/55   55           55          143m
```

Snippet of the cluster-autoscaler pod logs while scaling:

```sh
I1025 13:48:42.975037       1 scale_up.go:529] Final scale-up plan: [{eksctl-xxx-xxx-xxx-nodegroup-ng-xxxxx-NodeGroup-xxxxxxxxxx 2->3 (max: 8)}]
```


[//]: # 

   [Cluster Autoscaler Auto-Discovery]: <https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml>
   [IAM OIDC]: <https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html> 
   [IAM policy]: <https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html>
   [documentation]: <https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html> 
   [tutorial]: <https://aws.amazon.com/premiumsupport/knowledge-center/eks-cluster-autoscaler-setup/>

   
   
  
