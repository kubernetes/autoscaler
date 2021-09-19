# Cluster Autoscaler on CloudStack
The cluster autoscaler on CloudStack scales worker nodes within any specified cluster. It runs as a `Deployment` in your cluster.
This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Requirements
Cluster Autoscaler requires Apache CloudStack 4.16 onward as well as Kubernetes v1.16.0 or greater.

## Deployment
The CloudStack kubernetes cluster can be autoscaled via the `scaleKubernetesCluster` by passing `autoscalingenabled=true` along with
the `minsize` and `maxsize` parameters API from 4.16 onwards. eg:
```
scaleKubernetesCluster id=<cluster-id> autoscalingenabled=true minsize=<minsize> maxsize=<maxsize>
```
Autoscaling on the cluster can be disabled by passing `autoscalingenabled=false`. This will delete the deployment and leave the cluster
at its current size. eg:
```
scaleKubernetesCluster id=<cluster-id> autoscalingenabled=false
```

To manually deploy the cluster autoscaler, please follow the guide below :

To configure API access to your CloudStack management server, you need to create a secret containing a `cloud-config`
that is suitable for your environment.

`cloud-config` should look like this:
```ini
[Global]
api-url = <CloudStack API URL>
api-key = <CloudStack API Key>
secret-key = <CloudStack API Secret>
```
The access token needs to be able to execute the `listKubernetesClusters` and `scaleKubernetesCluster` APIs.

To create the secret, use the following command:
```bash
kubectl -n kube-system create secret generic cloudstack-secret --from-file=cloud-config
```

Finally, to deploy the autoscaler, modify the `cluster-autoscaler-standard.yaml` with the cluster id, minsize and maxsize located
[here](./examples/cluster-autoscaler-standard.yaml) and execute it

```
kubectl apply -f cluster-autoscaler-standard.yaml
```

## Common Notes and Gotchas:
- The automated deployment of the autoscaler will run with the defaults configured [here](./examples/cluster-autoscaler-standard.yaml). To change it, alter the file and deploy it again.
- By default, cluster autoscaler will not terminate nodes running pods in the kube-system namespace. You can override this default behaviour by passing in the `--skip-nodes-with-system-pods=false` flag.
- By default, cluster autoscaler will wait 10 minutes between scale down operations, you can adjust this using the `--scale-down-delay` flag. E.g. `--scale-down-delay=5m` to decrease the scale down delay to 5 minutes.
