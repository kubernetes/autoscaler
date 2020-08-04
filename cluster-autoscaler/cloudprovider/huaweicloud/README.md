# Cluster Autoscaler on Huawei Cloud 

## Overview
The cluster autoscaler for [Huawei ServiceStage](https://www.huaweicloud.com/intl/en-us/product/servicestage.html) scales worker nodes within any
specified container cluster's node pool where the `Autoscaler` label is on. 
It runs as a Deployment on a worker node in the cluster. This README will go over some of the necessary steps required 
to get the cluster autoscaler up and running.

Note: 

1. Cluster autoscaler must be run on a version of Huawei container engine 1.15.6 or later.
2. Node pool attached to the cluster must have the `Autoscaler` flag turned on, and minimum number of nodes and maximum
number of nodes being set. Node pools can be managed by `Resource Management` from Huawei container engine console.
3. If warnings about installing `autoscaler addon` are encountered after creating a node pool with `Autoscaler` flag on, 
just ignore this warning and DO NOT install the addon.
4. Do not build your image in a Huawei Cloud ECS. Build the image in a machine that has access to the Google Container Registry (GCR).

## Deployment Steps
### Build Image
#### Environment
1. Download Project

    Get the latest `autoscaler` project and download it to `${GOPATH}/src/k8s.io`. 
    
    This is used for building your image, so the machine you use here should be able to access GCR. Do not use a Huawei 
    Cloud ECS.

2. Go environment

    Make sure you have Go installed in the above machine.
    
3. Docker environment

    Make sure you have Docker installed in the above machine.
    
#### Build and push the image
Execute the following commands in the directory of `autoscaler/cluster-autoscaler` of the autoscaler project downloaded previously.
The following steps use Huawei SoftWare Repository for Container (SWR) as an example registry.

1. Build the `cluster-autoscaler` binary:
    ```
    make build-in-docker
    ```
2. Build the docker image:
    ```
   docker build -t {Image repository address}/{Organization name}/{Image name:tag} .
   ```
    For example:
    ```
    docker build -t swr.cn-north-4.myhuaweicloud.com/{Organization name}/cluster-autoscaler:dev .
    ```
   Follow the `Pull/Push Image` section of `Interactive Walkthroughs` under the SWR console to find the image repository address and organization name,
   and also refer to `My Images` -> `Upload Through Docker Client` in SWR console.
    
3. Login to SWR:
    ```
    docker login -u {Encoded username} -p {Encoded password} {SWR endpoint}
    ```
    
    For example:
    ```
    docker login -u cn-north-4@ABCD1EFGH2IJ34KLMN -p 1a23bc45678def9g01hi23jk4l56m789nop01q2r3s4t567u89v0w1x23y4z5678 swr.cn-north-4.myhuaweicloud.com
    ```
   Follow the `Pull/Push Image` section of `Interactive Walkthroughs` under the SWR console to find the encoded username, encoded password and swr endpoint,
   and also refer to `My Images` -> `Upload Through Docker Client` in SWR console.
   
4. Push the docker image to SWR:
    ```
    docker push {Image repository address}/{Organization name}/{Image name:tag}
    ```
   
    For example:
    ```
    docker push swr.cn-north-4.myhuaweicloud.com/{Organization name}/cluster-autoscaler:dev
    ```
   
5. For the cluster autoscaler to function normally, make sure the `Sharing Type` of the image is `Public`.
    If the cluster has trouble pulling the image, go to SWR console and check whether the `Sharing Type` of the image is 
    `Private`. If it is, click `Edit` button on top right and set the `Sharing Type` to `Public`.

### Deploy Cluster Autoscaler
#### Configure credentials
The autoscaler needs a `ServiceAccount` which is granted permissions to the cluster's resources and a `Secret` which 
stores credential (AK/SK in this case) information for authenticating with Huawei cloud.
    
Examples of `ServiceAccount` and `Secret` are provided in [examples/cluster-autoscaler-svcaccount.yaml](examples/cluster-autoscaler-svcaccount.yaml)
and [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml). Modify the Secret 
object yaml file with your credentials.

The following parameters are required in the Secret object yaml file:

- `identity-endpoint`

    Find the identity endpoint for different regions [here](https://support.huaweicloud.com/en-us/api-iam/iam_01_0001.html), 
    and fill in this field with `https://{Identity Endpoint}/v3.0`.
        
    For example, for region `cn-north-4`, fill in the `identity-endpoint` as
    ```
    identity-endpoint=https://iam.cn-north-4.myhuaweicloud.com/v3.0
    ```

- `project-id`
    
    Follow this link to find the project-id: [Obtaining a Project ID](https://support.huaweicloud.com/en-us/api-servicestage/servicestage_api_0023.html)

- `access-key` and `secret-key`

    Create and find the Huawei cloud access-key and secret-key
required by the Secret object yaml file by referring to [Access Keys](https://support.huaweicloud.com/en-us/usermanual-ca/ca_01_0003.html)
and [My Credentials](https://support.huaweicloud.com/en-us/usermanual-ca/ca_01_0001.html).

- `region`

    Fill in the region of the cluster here. For example, for region `Beijing4`:
    ```
    region=cn-north-4
    ```

- `domain-id`

    The required domain-id is the Huawei cloud [Account ID](https://support.huaweicloud.com/en-us/api-servicestage/servicestage_api_0048.html).

#### Configure deployment
   An example deployment file is provided at [examples/cluster-autoscaler-deployment.yaml](examples/cluster-autoscaler-deployment.yaml). 
   Change the `image` to the image you just pushed, the `cluster-name` to the cluster's id and `nodes` to your
   own configurations of the node pool with format
   ```
   {Minimum number of nodes}:{Maximum number of nodes}:{Node pool name}
   ```
   The above parameters should match the parameters of the node pool you created. Currently, Huawei ServiceStage only provides 
   autoscaling against a single node pool.
   
   More configuration options can be added to the cluster autoscaler, such as `scale-down-delay-after-add`, `scale-down-unneeded-time`, etc.
   See available configuration options [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-are-the-parameters-to-ca).

#### Deploy cluster autoscaler on the cluster
1. Log in to a machine which can manage the cluster with `kubectl`.

    Make sure the machine has kubectl access to the cluster. We recommend using a worker node to manage the cluster. Follow
    the instructions for 
[Connecting to a Kubernetes Cluster Using kubectl](https://support.huaweicloud.com/intl/en-us/usermanual-cce/cce_01_0107.html)
to set up kubectl access to the cluster if you cannot execute `kubectl` on your machine.

2. Create the Service Account:
    ```
    kubectl create -f cluster-autoscaler-svcaccount.yaml
    ```

3. Create the Secret:
    ```
    kubectl create -f cluster-autoscaler-secret.yaml
    ```

4. Create the cluster autoscaler deployment:
    ```
    kubectl create -f cluster-autoscaler-deployment.yaml
    ```

### Testing
Now the cluster autoscaler should be successfully deployed on the cluster. Check it by executing
```
kubectl get pods -n kube-system
```

To see whether it functions correctly, deploy a Service to the cluster, and increase and decrease workload to the
Service. Cluster autoscaler should be able to autoscale the node pool with `Autoscaler` on to accommodate the load.

A simple testing method is like this:
- Create a Service: listening for http request

- Create HPA or AOM policy for pods to be autoscaled
    * AOM policy: To create an AOM policy, go into the deployment, click `Scaling` tag and click `Add Scaling Policy`
    button on Huawei Cloud UI.
    * HPA policy: There're two ways to create an HPA policy.
        * Follow this instruction to create an HPA policy through UI: 
        [Scaling a Workload](https://support.huaweicloud.com/intl/en-us/usermanual-cce/cce_01_0208.html)
        * Install [metrics server](https://github.com/kubernetes-sigs/metrics-server) by yourself and create an HPA policy
        by executing something like this:
            ```
            kubectl autoscale deployment [Deployment name] --cpu-percent=50 --min=1 --max=10
            ```  
            The above command creates an HPA policy on the deployment with target average cpu usage of 50%. The number of 
            pods will grow if average cpu usage is above 50%, and will shrink otherwise. The `min` and `max` parameters set
            the minimum and maximum number of pods of this deployment.
- Generate load to the above service

    Example tools for generating workload to an http service are:
    * [Use `hey` command](https://github.com/rakyll/hey) 
    * Use `busybox` image:
        ```
        kubectl run --generator=run-pod/v1 -it --rm load-generator --image=busybox /bin/sh
  
        # send an infinite loop of queries to the service
        while true; do wget -q -O- {Service access address}; done
        ```
    
    Feel free to use other tools which have a similar function.
    
- Wait for pods to be added: as load increases, more pods will be added by HPA or AOM

- Wait for nodes to be added: when there's insufficient resource for additional pods, new nodes will be added to the 
cluster by the cluster autoscaler

- Stop the load

- Wait for pods to be removed: as load decreases, pods will be removed by HPA or AOM

- Wait for nodes to be removed: as pods being removed from nodes, several nodes will become underutilized or empty, 
and will be removed by the cluster autoscaler


## Notes

1. Huawei ServiceStage does not yet support autoscaling against multiple node pools within a single cluster, but 
this is currently under development. For now, make sure that there's only one node pool with `Autoscaler` label
on in the cluster.
2. If the version of the cluster is v1.15.6 or older, log statements similar to the following may be present in the
autoscaler pod logs:
    ```
    E0402 13:25:05.472999       1 reflector.go:178] k8s.io/client-go/informers/factory.go:135: Failed to list *v1.CSINode: the server could not find the requested resource
    ```
    This is normal and will be fixed by a future version of cluster.
    
## Support & Contact Info

Interested in Cluster Autoscaler on Huawei Cloud? Want to talk? Have questions, concerns or great ideas?

Please reach out to us at `shiyu.yuan@huawei.com`.