// for updating kubernetes dependencies, use `autoscaler/vertical-pod-autoscaler/hack/update-kubernetes-deps-in-e2e.sh`
// for any other update, use standard `go mod` commands

module k8s.io/autoscaler/vertical-pod-autoscaler/e2e

go 1.14

require (
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/autoscaler/vertical-pod-autoscaler v0.0.0-20200605154545-936eea18fb1d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.20.4
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.20.4
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
)

replace (
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0-alpha.0
	k8s.io/apiserver => k8s.io/apiserver v0.20.4
	k8s.io/autoscaler => ../../
	k8s.io/autoscaler/vertical-pod-autoscaler => ../
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator => k8s.io/code-generator v0.20.5-rc.0
	k8s.io/component-base => k8s.io/component-base v0.20.4
	k8s.io/cri-api => k8s.io/cri-api v0.20.5-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/kubelet => k8s.io/kubelet v0.20.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics => k8s.io/metrics v0.20.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.4
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.4
)

replace k8s.io/component-helpers => k8s.io/component-helpers v0.20.4

replace k8s.io/controller-manager => k8s.io/controller-manager v0.20.4

replace k8s.io/mount-utils => k8s.io/mount-utils v0.20.5-rc.0
