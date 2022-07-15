module k8s.io/autoscaler/cluster-autoscaler

go 1.16

require (
	cloud.google.com/go v0.81.0
	github.com/Azure/azure-sdk-for-go v62.3.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/adal v0.9.14
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/aws/aws-sdk-go v1.38.49
	github.com/digitalocean/godo v1.27.0
	github.com/gardener/machine-controller-manager v0.45.0
	github.com/gardener/machine-controller-manager-provider-aws v0.11.0
	github.com/gardener/machine-controller-manager-provider-azure v0.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.1.2
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	google.golang.org/api v0.46.0
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/apiserver v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/cloud-provider v0.23.0
	k8s.io/component-base v0.23.0
	k8s.io/component-helpers v0.23.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.30.0
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.23.0
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/cloud-provider-azure v1.1.3
)

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v55.8.0+incompatible

replace github.com/digitalocean/godo => github.com/digitalocean/godo v1.27.0

replace github.com/rancher/go-rancher => github.com/rancher/go-rancher v0.1.0

replace k8s.io/api => k8s.io/api v0.23.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.1-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.23.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.0

replace k8s.io/client-go => k8s.io/client-go v0.23.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.0

replace k8s.io/code-generator => k8s.io/code-generator v0.23.1-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.23.0

replace k8s.io/component-helpers => k8s.io/component-helpers v0.23.0

replace k8s.io/controller-manager => k8s.io/controller-manager v0.23.0

replace k8s.io/cri-api => k8s.io/cri-api v0.23.1-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.0

replace k8s.io/kubectl => k8s.io/kubectl v0.23.0

replace k8s.io/kubelet => k8s.io/kubelet v0.23.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.0

replace k8s.io/metrics => k8s.io/metrics v0.23.0

replace k8s.io/mount-utils => k8s.io/mount-utils v0.23.2-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.0

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.23.0

replace k8s.io/sample-controller => k8s.io/sample-controller v0.23.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.0
