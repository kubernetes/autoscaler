# Cluster Autoscaler Deployment/Configuration
This documnet describes how to deploy/change ClusterAutascaler
In the document cloud_init described simple script responsible to install required packages, generate config files and join new node to cluster
For new deployment there is necessary to change some values inside this cloud_init script:
First need to change data related to new cluser join credentials (token and cacert hashes):
* get the join data from the master:
```bash
#create and fetch token from master
kubeadm token create --ttl 0 --print-join-command
#as result will be the output like: 
kubeadm join 127.0.0.1:6443 --token sv0xn7.pyl6kakhon296itl --discovery-token-ca-cert-hash sha256:d094f629fd1249c6fa78c2eca9fc2a82f38aab648d55c3962e8e30312d0bf1ee
```
* example create secret for HC for CCM:
```yaml
kubectl -n kube-system create secret generic hcloud --from-literal=token=<hcloud API token> --from-literal=network=<hcloud Network ID>
```
* examle setting nodeSelector, it`s help start scaling and scheduling on affinity poll
```yaml
nodeSelector:
  hcloud/node-group: pool2
```
or for first pooll
```yaml
nodeSelector:
  hcloud/node-group: pool1
```
* replace values for token and caCertHashes, ip control_plane and ssh_name in autoscaler.yaml
* encode file with base64 encoding:
```bash
cat cloud_init |base64 > init
```
* create configmap for ClusterAutoscaler:
```bash
kubectl create cm cloud-init --namespace kube-system --from-file init --dry-run=client -o yaml | kubectl apply -f -
```
* apply manifest for Cluster Autoscaler:
```bash
kubectl apply -f autoscaler.yaml
```

* ```$TRIMNAME``` set lable for nodes depends on pool name(its help set nodeSelector)


