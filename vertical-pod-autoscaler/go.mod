module k8s.io/autoscaler/vertical-pod-autoscaler

go 1.24.0

toolchain go1.24.2

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/golang/mock v1.6.0
	github.com/prometheus/client_golang v1.22.0
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.63.0
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
	golang.org/x/time v0.11.0
	k8s.io/api v0.33.0
	k8s.io/apimachinery v0.33.0
	k8s.io/client-go v0.33.0
	k8s.io/code-generator v0.33.0
	k8s.io/component-base v0.33.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/metrics v0.33.0
	k8s.io/utils v0.0.0-20250502105355-0f33e8f1c979
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/oauth2 v0.29.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/gengo/v2 v2.0.0-20250207200755-1244d31929d7 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.7.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.33.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.33.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.33.0
	k8s.io/apiserver => k8s.io/apiserver v0.33.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.33.0
	k8s.io/client-go => k8s.io/client-go v0.33.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.33.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.33.0
	k8s.io/code-generator => k8s.io/code-generator v0.33.0
	k8s.io/component-base => k8s.io/component-base v0.33.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.33.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.33.0
	k8s.io/cri-api => k8s.io/cri-api v0.33.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.33.0
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.33.0
	k8s.io/endpointslice => k8s.io/endpointslice v0.33.0
	k8s.io/kms => k8s.io/kms v0.33.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.33.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.33.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.33.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.33.0
	k8s.io/kubectl => k8s.io/kubectl v0.33.0
	k8s.io/kubelet => k8s.io/kubelet v0.33.0
	k8s.io/metrics => k8s.io/metrics v0.33.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.33.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.33.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.33.0
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.33.0
	k8s.io/sample-controller => k8s.io/sample-controller v0.33.0
)

replace k8s.io/cri-client => k8s.io/cri-client v0.33.0

replace k8s.io/externaljwt => k8s.io/externaljwt v0.33.0
