// for updating kubernetes dependencies, use `autoscaler/vertical-pod-autoscaler/hack/update-kubernetes-deps-in-e2e.sh`
// for any other update, use standard `go mod` commands

module k8s.io/autoscaler/vertical-pod-autoscaler/e2e

go 1.12

require (
	github.com/auth0/go-jwt-middleware v0.0.0-20190805220309-36081240882b // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/heketi/heketi v9.0.1-0.20191108182613-7340ee406820+incompatible // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/procfs v0.0.6 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20191112182307-2180aed22343 // indirect
	golang.org/x/sys v0.0.0-20191112214154-59a1497f0cea // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/autoscaler/vertical-pod-autoscaler v0.0.0-20200605154545-936eea18fb1d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.0.0
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.16.2
	k8s.io/utils v0.0.0-20200109141947-94aeca20bf09
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/autoscaler => ../../
	k8s.io/autoscaler/vertical-pod-autoscaler => ../
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/code-generator => k8s.io/code-generator v0.16.5-beta.1
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.16.5-beta.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	k8s.io/node-api => k8s.io/node-api v0.0.0-20191016115955-b0b11a2622b0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20191016114214-d25a4244b17f
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20191016113152-0c2dd40eec0c
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)
