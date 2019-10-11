module k8s.io/autoscaler/vertical-pod-autoscaler

go 1.13

replace k8s.io/api => k8s.io/api v0.0.0-20191003000013-35e20aa79eb8

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191003002041-49e3d608220c

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191003001037-3c8b233e046c

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191003002408-6e42c232ac7d

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191003000419-f68efa97b39e

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191003003426-b4b1f434fead

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191003003255-c493acd9e2ff

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190927045949-f81bca4f5e85

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20191003000551-f573d376509c

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191003003551-0eecdcdcc049

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191003001317-a019a9d85a86

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191003003129-09316795c0dd

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191003002707-f6b7b0f55cc0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191003003001-314f0beee0a9

replace k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191003004222-1f3c0cd90ca9

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191003002833-e367e4712542

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191003003732-7d49cdad1c12

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20191003002233-837aead57baf

replace k8s.io/node-api => k8s.io/node-api v0.0.0-20191003003902-772c6d2244f3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191003001538-80f33ca02582

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20191003002540-40951731b79f

replace k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20191003001734-27680fba8268

require (
	github.com/golang/mock v1.2.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_golang v0.9.2
	github.com/stretchr/testify v1.3.0
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v0.4.0
	k8s.io/kubernetes v1.16.1
	k8s.io/metrics v0.0.0
)
