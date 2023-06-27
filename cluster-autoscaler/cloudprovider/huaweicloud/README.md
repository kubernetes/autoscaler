# Cluster Autoscaler on Huawei Cloud 

## Overview
The cluster autoscaler works with self-built Kubernetes cluster on [Huaweicloud ECS](https://www.huaweicloud.com/intl/en-us/product/ecs.html) and
specified [Huaweicloud Auto Scaling Groups](https://www.huaweicloud.com/intl/en-us/product/as.html) 
It runs as a Deployment on a worker node in the cluster. This README will go over some of the necessary steps required 
to get the cluster autoscaler up and running.

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
  

## Build Kubernetes Cluster on ECS   

### 1. Install kubelet, kubeadm and kubectl   

Please see installation [here](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

For example:
- OS: CentOS 8
- Note: The following example should be run on ECS that has access to the Google Container Registry (GCR)
    ```bash
    cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
    [kubernetes]
    name=Kubernetes
    baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
    enabled=1
    gpgcheck=1
    repo_gpgcheck=1
    gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/ doc/rpm-package-key.gpg
    exclude=kubelet kubeadm kubectl
    EOF
    ```

    ```
    sudo setenforce 0
    sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

    sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes

    sudo systemctl enable --now kubelet
    ```
### 2. Install Docker
Please see installation [here](https://docs.docker.com/engine/install/)

For example:
- OS: CentOS 8
- Note: The following example should be run on ECS that has access to the Google Container Registry (GCR)

    ```bash
    sudo yum install -y yum-utils

    sudo yum-config-manager \
        --add-repo \
        https://download.docker.com/linux/centos/docker-ce.repo

    sudo yum install docker-ce docker-ce-cli containerd.io

    sudo systemctl start docker
    ```

### 3. Initialize Cluster
Create a Kubeadm.yaml file with the following content:
```yaml
apiVersion: kubeadm.k8s.io/v1beta3
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.0123456789abcdef
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: 1.2.3.4
  bindPort: 6443
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  imagePullPolicy: IfNotPresent
  name: node
  taints: null
---
apiServer:
  timeoutForControlPlane: 4m0s
apiVersion: kubeadm.k8s.io/v1beta3
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.k8s.io
kind: ClusterConfiguration
kubernetesVersion: 1.22.0
networking:
  dnsDomain: cluster.local
  podSubnet: 10.244.0.0/16
  serviceSubnet: 10.96.0.0/12
scheduler: {}
---
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
cgroupDriver: cgroupfs

```

**note: replace the advertiseAddress to your ECS ip address**

```bash
kubeadm init --config kubeadm.yaml
```

Modify these two files, Comment out line  `- --port=0`:
```sh
sudo vim /etc/kubernetes/manifests/kube-controller-manager.yaml
sudo vim /etc/kubernetes/manifests/kube-scheduler.yaml
```

Restart Kubelet service
```sh
systemctl restart kubelet.service
```

```sh
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

### 4. Install Flannel Network
```bash 
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```
### 5. Generate Token
```bash
kubeadm token create --ttl 0
```
Generate a token that never expires. Remember this token since it will be used later.

Get hash key. Remember the key since it will be used later.
```
openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
```

### 6. Create OS Image with K8S Tools
- Launch a new ECS instance, and install Kubeadm, Kubectl and docker.
    ```sh
    cat <<EOF > /etc/yum.repos.d/kubernetes.repo
    [kubernetes]
    name=Kubernetes
    baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
    enabled=1
    gpgcheck=1
    repo_gpgcheck=1
    gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
    EOF

    sudo yum install -y kubeadm kubectl --disableexcludes=kubernetes

    sudo yum install -y yum-utils

    sudo yum-config-manager \
        --add-repo \
        http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo

    sudo yum install docker-ce docker-ce-cli containerd.io
    ```
- Create a script to join the new instance into the k8s cluster.
    ```bash
    cat <<EOF >/etc/rc.d/init.d/init-k8s.sh
    #!bin/bash
    #chkconfig: 2345 80 90
    setenforce 0
    swapoff -a
    sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
    yum install -y kubelet
    sudo systemctl enable --now kubelet

    systemctl start docker
    systemctl enable docker.service
    kubeadm join --token $TOKEN $API_Server_EndPoint --discovery-token-ca-cert-hash sha256:$HASHKEY
    EOF
    ```
    Replace the $TOKEN with the one created above.

    Replace the $API_Server_EndPoint, this could be find in the context file.
    ```sh
    cat ~./.kube/config

    # server: https://192.168.0.239:6443
    # the API_Server_EndPoint is 192.168.0.239:6443
    ```

- Add this script into chkconfig, to let it run automatically after the instance is started.
    ```
    chmod +x /etc/rc.d/init.d/init-k8s.sh
    chkconfig --add /etc/rc.d/init.d/init-k8s.sh
    ```
<!--TODO: Remove "previously referred to as master" references from this doc once this terminology is fully removed from k8s-->
- Copy `~/.kube/config` from a control plane (previously referred to as master) node to this ECS `~./kube/config` to setup kubectl on this instance.

- Go to Huawei Cloud `Image Management` Service and click on `Create Image`. Select type `System disk image`, select your ECS instance as `Source`, then give it a name and then create.

- Remember this ECS instance ID since it will be used later.

### 7. Create AS Group
- Follow the Huawei cloud instruction to create an AS Group.
- Create an AS Configuration, and select private image which we just created. Make sure the AS Configuration with EIP automatically assign.
- While creating the `AS Configuration`, add the following script into `Advanced Settings`.
    ```bash
    #!bin/bash

    IDS=$(ls /var/lib/cloud/instances/)
    while true
    do
        for ID in $IDS
        do
            if [ $ID != $ECS_INSTANCE_ID ]; then
                /usr/bin/kubectl --kubeconfig ~/.kube/config patch node $HOSTNAME -p "{\"spec\":{\"providerID\":\"$ID\"}}"
            fi
        done
    sleep 30
    done
    ```
    Replace the $ECS_INSTANCE_ID
 - Bind the AS Group with this AS Configuration

### Deploy Cluster Autoscaler
#### Configure credentials
The autoscaler needs a `ServiceAccount` which is granted permissions to the cluster's resources and a `Secret` which 
stores credential (AK/SK in this case) information for authenticating with Huawei cloud.
    
Examples of `ServiceAccount` and `Secret` are provided in [examples/cluster-autoscaler-svcaccount.yaml](examples/cluster-autoscaler-svcaccount.yaml)
and [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml). Modify the Secret 
object yaml file with your credentials.

The following parameters are required in the Secret object yaml file:

- `as-endpoint`

    Find the as endpoint for different regions [here](https://developer.huaweicloud.com/endpoint?AS), 
        
    For example, for region `cn-north-4`, the endpoint is
    ```
    as.cn-north-4.myhuaweicloud.com
    ```

- `ecs-endpoint`

    Find the ecs endpoint for different regions [here](https://developer.huaweicloud.com/endpoint?ECS), 
        
    For example, for region `cn-north-4`, the endpoint is 
    ```
    ecs.cn-north-4.myhuaweicloud.com
    ```

- `project-id`
    
    Follow this link to find the project-id: [Obtaining a Project ID](https://support.huaweicloud.com/en-us/api-servicestage/servicestage_api_0023.html)

- `access-key` and `secret-key`

    Create and find the Huawei cloud access-key and secret-key
required by the Secret object yaml file by referring to [Access Keys](https://support.huaweicloud.com/en-us/usermanual-ca/ca_01_0003.html)
and [My Credentials](https://support.huaweicloud.com/en-us/usermanual-ca/ca_01_0001.html).


#### Configure deployment
   An example deployment file is provided at [examples/cluster-autoscaler-deployment.yaml](examples/cluster-autoscaler-deployment.yaml). 
   Change the `image` to the image you just pushed, the `cluster-name` to the cluster's id and `nodes` to your
   own configurations of the node pool with format
   ```
   {Minimum number of nodes}:{Maximum number of nodes}:{Node pool name}
   ```
   The above parameters should match the parameters of the AS Group you created.
   
   More configuration options can be added to the cluster autoscaler, such as `scale-down-delay-after-add`, `scale-down-unneeded-time`, etc.
   See available configuration options [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-are-the-parameters-to-ca).

#### Deploy cluster autoscaler on the cluster
1. Log in to a machine which can manage the cluster with `kubectl`.

    Make sure the machine has kubectl access to the cluster.

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
Service. Cluster autoscaler should be able to autoscale the AS Group to accommodate the load.

A simple testing method is like this:
- Create a Service: listening for http request

- Create HPA policy for pods to be autoscaled
    * Install [metrics server](https://github.com/kubernetes-sigs/metrics-server) by yourself and create an HPA policy
    by executing something like this:
        ```
        kubectl autoscale deployment [Deployment name] --cpu-percent=10 --min=1 --max=20
        ```  
        The above command creates an HPA policy on the deployment with target average cpu usage of 10%. The number of 
        pods will grow if average cpu usage is above 10%, and will shrink otherwise. The `min` and `max` parameters set
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
    
- Wait for pods to be added: as load increases, more pods will be added by HPA

- Wait for nodes to be added: when there's insufficient resource for additional pods, new nodes will be added to the 
cluster by the cluster autoscaler

- Stop the load

- Wait for pods to be removed: as load decreases, pods will be removed by HPA

- Wait for nodes to be removed: as pods being removed from nodes, several nodes will become underutilized or empty, 
and will be removed by the cluster autoscaler

    
## Support & Contact Info

Interested in Cluster Autoscaler on Huawei Cloud? Want to talk? Have questions, concerns or great ideas?

Please reach out to us at `shiqi.wang1@huawei.com`.