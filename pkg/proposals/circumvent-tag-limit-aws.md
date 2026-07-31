# Circumvent 50 tag ASG limit for EKS ManagedNodegroups
##### Author: MyannaHarris

## Table of Contents

<!-- toc -->
- [Summary](#summary)
- [IAM](#iam)
- [Design](#design)
- [Implementation Proposal](#implementation-proposal)
- [Testing](#testing)
- [Risks](#risks)
- [Other Solutions Considered](#other-solutions-considered)
<!-- /toc -->

## Summary

Currently an EC2 Autoscaling group can only have 50 tags. Many tags are already added to the ASG by standard components like the AWS cloudprovider for Kubernetes and by customers for billing and cost association purposes. Adding labels and taints to the ASG as tags will run into this 50 tag limit. The primary focus of this proposal is to get around the 50 tag limit for customers scaling to/from 0 nodes using Cluster Autoscaler on AWS EKS ManagedNodegroups in a way that will not limit the ManagedNodegroups service.

AWS provides the EKS ManagedNodegroups service which manages the lifecycle of EC2 worker nodes that can join an EKS Kubernetes cluster. Each EKS ManagedNodegroup has an underlying ASG. ASGs and Cluster Autoscaler support scaling to/from 0.

In the [AWS cloud provider][], if there are no nodes in the nodegroup, Cluster Autoscaler creates a nodeTemplate from ASG tags and some default allocatable resources. (See the [code][]) The tags for Cluster Autoscaler include resources, labels, and taints. The nodeTemplate is used to identify available node resources in the absence of an actual EC2 instance existing. We propose that, when a scaled-to-0 EKS ManagedNodegroup is being used, the Cluster Autoscaler cloud provider for AWS takes advantage of the EKS DescribeNodegroup API, which will return the latest lists of labels and taints. Unlike the ASG tags, the API will not limit the number of labels and taints that Cluster Autoscaler can discover to 50.

[AWS cloud provider]: https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/aws
[code]: https://github.com/kubernetes/autoscaler/blob/99936f3f3971c91b30def913c1c807c42c4b87a7/cluster-autoscaler/cloudprovider/aws/aws_manager.go#L365-L371

## IAM

The ServiceAccount running the Cluster Autoscaler Pod (if using IAM Roles for Service Accounts (IRSA)) or associated with the instance profile (if not using IRSA) will need to have one additional policy applied to it: `eks:DescribeNodegroup`. If Cluster Autoscaler doesn’t have the correct permissions to call the API, it will fall back to using just the ASG tags.

## Design

We propose that the AWS specific driver of Cluster Autoscaler checks for AWS EKS ManagedNodegroups tags on the ASG that marks the ASG as a ManagedNodegroup. If the tags are present, Cluster Autoscaler pulls information from a cache struct containing DescribeNodegroup API response data. The AWS EKS ManagedNodegroups tags look like  `eks:cluster-name : <CLUSTER_NAME>` and `eks:nodegroup-name : <NODEGROUP_NAME>`. We'll use AWS EKS ManagedNodegroups tags because we automatically add them to every managed nodegroup. The DescribeNodegroup API is already in the AWS SDK that Cluster Autoscaler uses. ([Current EKS interface][])

When Cluster Autoscaler finds the ManagedNodegroups tags, it will call functions on a ManagedNodegroupCache struct to get the labels, taints, and some other values. The cache struct will hold the response data from the EKS [DescribeNodegroup API][]. When the cache is accessed (the cache struct methods are called), it will check if there's data cached. If data exists, it will then check if the TTL has expired (1 minute). If there's no cached data or the TTL has expired, the DescribeNodegroup API will be called. Cluster Autoscaler will include both the ASG tag values and the values from the EKS API in its decisions. If an ASG tag contains the same resource as the EKS API (ex: a label key is used in both), Cluster Autoscaler will choose the value from the ASG tag so customers can override any value they want.

This is what a DescribeNodegroup API response looks like (also see [here][]):
```json
HTTP/1.1 200
Content-type: application/json
{
   "nodegroup": { 
      "amiType": "*string*",
      "capacityType": "*string*",
      "clusterName": "*string*",
      "createdAt": *number*,
      "diskSize": *number*,
      "health": { 
         "issues": [ 
            { 
               "code": "*string*",
               "message": "*string*",
               "resourceIds": [ "*string*" ]
            }
         ]
      },
      "instanceTypes": [ "*string*" ],
      "labels": { 
         "*string*" : "*string*" 
      },
      "launchTemplate": { 
         "id": "*string*",
         "name": "*string*",
         "version": "*string*"
      },
      "modifiedAt": *number*,
      "nodegroupArn": "*string*",
      "nodegroupName": "*string*",
      "nodeRole": "*string*",
      "releaseVersion": "*string*",
      "remoteAccess": { 
         "ec2SshKey": "*string*",
         "sourceSecurityGroups": [ "*string*" ]
      },
      "resources": { 
         "autoScalingGroups": [ 
            { 
               "name": "*string*"
            }
         ],
         "remoteAccessSecurityGroup": "*string*"
      },
      "scalingConfig": { 
         "desiredSize": *number*,
         "maxSize": *number*,
         "minSize": *number*
      },
      "status": "*string*",
      "subnets": [ "*string*" ],
      "tags": { 
         "*string*" : "*string*" 
      },
      "version": "*string*"
   }
}
```

[Current EKS interface]: https://github.com/aws/aws-sdk-go/blob/82e096143fdfb8f52fbeb4ef78d400eea6381ccd/service/eks/eksiface/interface.go#L96
[DescribeNodegroup API]: https://docs.aws.amazon.com/eks/latest/APIReference/API_DescribeNodegroup.html
[here]: https://docs.aws.amazon.com/eks/latest/APIReference/API_DescribeNodegroup.html#API_DescribeNodegroup_ResponseSyntax

## Implementation Proposal

1. In the autoscaling loop, Cluster Autoscaler finds no nodes, so it will build the nodeTemplate from other information.
1. Cluster Autoscaler will first get some common allocatable resource values. ([allocatable resources code][])
1. Cluster Autoscaler will then check the ASG tags for the EKS ManagedNodegroups tags. ([ASG tags code][])
1. If Cluster Autoscaler finds the EKS ManagedNodegroups tags, it will call methods on the ManagedNodegroupCache to get labels, taints, and some other values.
   1. If the cache has data for the given ManagedNodegroup, it will return that data.
   1. If the data’s TTL has expired, the DescribeNodegroup API will be called and the data will be added to the cache and returned.
   1. If there is no data cached for the ManagedNodegroup, the DescribeNodegroup API will be called and the data will be added to the cache and returned.
1. Then Cluster Autoscaler will pull information from the ASG tags like it already does. If duplicate keys are found, the values from the ASG tags are kept.
1. Cluster Autoscaler will then continue as usual.

[allocatable resources code]: https://github.com/kubernetes/autoscaler/blob/99936f3f3971c91b30def913c1c807c42c4b87a7/cluster-autoscaler/cloudprovider/aws/aws_manager.go#L366
[ASG tags code]: https://github.com/kubernetes/autoscaler/blob/99936f3f3971c91b30def913c1c807c42c4b87a7/cluster-autoscaler/cloudprovider/aws/aws_manager.go#L371

## Testing

- [TC1] Test creating a DescribeNodegroup request
- [TC2] Test pulling values from cache when data exists for the ManagedNodegroups and TTL has not expired
- [TC3] Test pulling values from cache and calling DescribeNodegroup API if no data exists
- [TC4] Test pulling values from cache and calling DescribeNodegroup API if TTL has expired
- [TC5] Test missing permissions and falling back to just ASG tags
- [TC6] Test duplicate label keys - selecting the value from ASG tags response

## Risks

Latency and throttling:

By default, Cluster Autoscaler runs every 10 seconds. Our best practices documentation notes that this short interval can cause throttling because Cluster Autoscaler already makes AWS API calls during each loop. Our documentation recommends that customers increase the interval, so adding this API shouldn’t cause latency problems for customers. ([EKS Best Practices][]) Also, the API call will only happen for EKS ManagedNodegroups. If this increase in latency is too much for even one run of the loop, we will look into moving the API calls into parallel goroutines. 

EKS also throttles describe API calls by default. To mitigate this issue, Cluster Autoscaler will keep a cache of DescribeNodegroup responses: ManagedNodegroupCache. Each cache bucket will have a TTL and, when the TTL expires, DescribeNodegroup will be called again. If there are errors during the call to DescribeNodegroup, Cluster Autoscaler will move on and just look at the ASG tags and existing default allocatable resource values.

[EKS Best Practices]: https://aws.github.io/aws-eks-best-practices/cluster-autoscaling/cluster-autoscaling/#reducing-the-scan-interval

## Other Solutions Considered

The AWS EKS ManagedNodegroups service creating a configMap in the cluster that saves all of the node information for ManagedNodegroups was considered.
- Pros
  - This is a more general approach that could be adopted by others or repurposed
- Cons
  - This will move the solution more into ManagedNodegroups so the ManagedNodegroups implementation would be closely tied with Cluster Autoscaler’s implementation. We don’t want to tie them together because ManagedNodegroups is a nodegroup implementation and should work with various autoscalers rather than be specifically built for one autoscaler.
