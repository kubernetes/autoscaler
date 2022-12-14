#!/bin/bash
export DEBIAN_FRONTEND=noninteractive
export HOME=/root/
apt-get update && apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository   "deb [arch=amd64] https://download.docker.com/linux/ubuntu   $(lsb_release -cs)   stable"
apt-get update
apt install -y docker-ce
sudo docker run -d --privileged --restart=unless-stopped --net=host -v /etc/kubernetes:/etc/kubernetes -v /var/run:/var/run \
  rancher/rancher-agent:v2.6.4 \
  --server https://rancher.domain \
  --token aaa \
  --ca-checksum bbb
