#!/bin/bash
export DEBIAN_FRONTEND=noninteractive
export HOME=/root/
apt-get update && apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
add-apt-repository   "deb [arch=amd64] https://download.docker.com/linux/ubuntu   $(lsb_release -cs)   stable"
apt-get update
apt-get upgrade -y
apt-get install -y kubelet=1.20.1-00 kubeadm=1.20.1-00 kubectl=1.20.1-00
apt-mark hold kubelet kubeadm kubectl
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable"
apt update
apt install -y docker-ce
cat <<EOF | tee /etc/default/kubelet
KUBELET_EXTRA_ARGS=--cloud-provider=external
EOF
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
mkdir -p /etc/systemd/system/docker.service.d
systemctl daemon-reload
systemctl restart docker
kubeadm join <master-ip> --token <token>  --discovery-token-ca-cert-hash sha256:<hash>
