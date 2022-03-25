// for updating kubernetes dependencies, use `autoscaler/vertical-pod-autoscaler/hack/update-kubernetes-deps-in-e2e.sh`
// for any other update, use standard `go mod` commands

module k8s.io/autoscaler/vertical-pod-autoscaler/e2e

go 1.14

require (
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/autoscaler/vertical-pod-autoscaler v0.0.0-20200605154545-936eea18fb1d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.23.5
	k8s.io/klog/v2 v2.30.0
	k8s.io/kubernetes v1.23.5
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
)

replace (
	k8s.io/api => k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.6-rc.0
	k8s.io/apiserver => k8s.io/apiserver v0.23.5
	k8s.io/autoscaler => ../../
	k8s.io/autoscaler/vertical-pod-autoscaler => ../
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.5
	k8s.io/client-go => k8s.io/client-go v0.23.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.5
	k8s.io/code-generator => k8s.io/code-generator v0.23.6-rc.0
	k8s.io/component-base => k8s.io/component-base v0.23.5
	k8s.io/cri-api => k8s.io/cri-api v0.23.6-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.5
	k8s.io/kubectl => k8s.io/kubectl v0.23.5
	k8s.io/kubelet => k8s.io/kubelet v0.23.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.5
	k8s.io/metrics => k8s.io/metrics v0.23.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.23.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.23.5
)

replace k8s.io/component-helpers => k8s.io/component-helpers v0.23.5

replace k8s.io/controller-manager => k8s.io/controller-manager v0.23.5

replace k8s.io/mount-utils => k8s.io/mount-utils v0.23.6-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.5
