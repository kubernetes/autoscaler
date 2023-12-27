#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#This is a template of cloud-init script for cluster-autoscaler/cloudprovider/equinixmetal
#This gets base64'd and put into the example cluster-autoscaler-secret.yaml file
sleep 10
sed -ri '/\sswap\s/s/^#?/#/' /etc/fstab
swapoff -a
mount -a
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
modprobe overlay
modprobe br_netfilter
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sysctl --system
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get remove docker docker-engine containerd runc -y
apt-get install apt-transport-https ca-certificates curl gnupg lsb-release jq kubetail -y
mkdir -m 0755 -p /etc/apt/keyrings
curl -fsSL https://dl.k8s.io/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-archive-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y containerd.io kubelet=1.28.1-00 kubeadm=1.28.1-00 kubectl=1.28.1-00
apt-mark hold kubelet kubeadm kubectl
containerd config default > /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sed -i "s,sandbox_image.*$,sandbox_image = \"$(kubeadm config images list | grep pause | sort -r | head -n1)\"," /etc/containerd/config.toml
systemctl restart containerd
systemctl enable kubelet.service
echo "source <(kubectl completion bash)" >> /root/.bashrc
echo "alias k=kubectl" >> /root/.bashrc
echo "complete -o default -F __start_kubectl k" >> /root/.bashrc
cat <<EOF | tee /etc/default/kubelet
KUBELET_EXTRA_ARGS=--cloud-provider=external
EOF
kubeadm join --discovery-token-unsafe-skip-ca-verification --token {{.BootstrapTokenID}}.{{.BootstrapTokenSecret}} {{.APIServerEndpoint}}
