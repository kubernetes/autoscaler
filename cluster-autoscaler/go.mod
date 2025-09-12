module k8s.io/autoscaler/cluster-autoscaler

go 1.24.0

require (
	cloud.google.com/go/compute/metadata v0.6.0
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go-extensions v0.1.6
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.13.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.7.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5 v5.1.0-beta.2
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/adal v0.9.24
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/skewer v0.0.19
	github.com/aws/aws-sdk-go v1.44.241
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/digitalocean/godo v1.27.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.7.0
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.6.0
	github.com/jmattheis/goverter v1.4.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.22.0
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
	github.com/vburenin/ifacemaker v1.2.1
	go.uber.org/mock v0.4.0
	golang.org/x/crypto v0.36.0
	golang.org/x/net v0.38.0
	golang.org/x/oauth2 v0.27.0
	golang.org/x/sys v0.31.0
	google.golang.org/api v0.151.0
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.5
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.34.1
	k8s.io/apimachinery v0.34.1
	k8s.io/apiserver v0.34.1
	k8s.io/autoscaler/cluster-autoscaler/apis v0.0.0-20240627115740-d52e4b9665d7
	k8s.io/client-go v0.34.1
	k8s.io/cloud-provider v0.30.1
	k8s.io/cloud-provider-aws v1.27.0
	k8s.io/cloud-provider-gcp/providers v0.28.2
	k8s.io/component-base v0.34.1
	k8s.io/component-helpers v0.34.1
	k8s.io/dynamic-resource-allocation v0.0.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubelet v0.34.1
	k8s.io/kubernetes v1.34.1
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397
	sigs.k8s.io/cloud-provider-azure v1.29.4
	sigs.k8s.io/cloud-provider-azure/pkg/azclient v0.0.13
	sigs.k8s.io/yaml v1.6.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.12.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4 v4.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault v1.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 v4.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage v1.5.0 // indirect
	github.com/Azure/go-armbalancer v0.0.2 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.25.0 // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hnslib v0.1.1 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/armon/circbuf v0.0.0-20190214190532-5111143e8da2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/container-storage-interface/spec v1.9.0 // indirect
	github.com/containerd/containerd/api v1.8.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/ttrpc v1.2.6 // indirect
	github.com/containerd/typeurl/v2 v2.2.2 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/dave/jennifer v1.6.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cadvisor v0.52.1 // indirect
	github.com/google/cel-go v0.26.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jessevdk/go-flags v1.4.1-0.20181029123624-5de817a9aa20 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/onsi/ginkgo/v2 v2.21.0 // indirect
	github.com/onsi/gomega v1.35.1 // indirect
	github.com/opencontainers/cgroups v0.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.etcd.io/etcd/api/v3 v3.6.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.4 // indirect
	go.etcd.io/etcd/client/v3 v3.6.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.44.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.0.0 // indirect
	k8s.io/code-generator v0.34.1 // indirect
	k8s.io/controller-manager v0.34.1 // indirect
	k8s.io/cri-api v0.34.1 // indirect
	k8s.io/cri-client v0.0.0 // indirect
	k8s.io/csi-translation-lib v0.27.0 // indirect
	k8s.io/gengo/v2 v2.0.0-20250604051438-85fd79dbfd9f // indirect
	k8s.io/kms v0.34.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubectl v0.28.0 // indirect
	k8s.io/mount-utils v0.26.0-alpha.0 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.2 // indirect
	sigs.k8s.io/cloud-provider-azure/pkg/azclient/configloader v0.0.4 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

replace github.com/aws/aws-sdk-go/service/eks => github.com/aws/aws-sdk-go/service/eks v1.38.49

replace github.com/digitalocean/godo => github.com/digitalocean/godo v1.27.0

replace github.com/rancher/go-rancher => github.com/rancher/go-rancher v0.1.0

replace k8s.io/api => k8s.io/api v0.34.1

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.34.1

replace k8s.io/apimachinery => k8s.io/apimachinery v0.34.1

replace k8s.io/apiserver => k8s.io/apiserver v0.34.1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.34.1

replace k8s.io/client-go => k8s.io/client-go v0.34.1

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.34.1

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.34.1

replace k8s.io/code-generator => k8s.io/code-generator v0.34.1

replace k8s.io/component-base => k8s.io/component-base v0.34.1

replace k8s.io/component-helpers => k8s.io/component-helpers v0.34.1

replace k8s.io/controller-manager => k8s.io/controller-manager v0.34.1

replace k8s.io/cri-api => k8s.io/cri-api v0.34.1

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.34.1

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.34.1

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.34.1

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.34.1

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.34.1

replace k8s.io/kubectl => k8s.io/kubectl v0.34.1

replace k8s.io/kubelet => k8s.io/kubelet v0.34.1

replace k8s.io/metrics => k8s.io/metrics v0.34.1

replace k8s.io/mount-utils => k8s.io/mount-utils v0.34.1

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.34.1

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.34.1

replace k8s.io/sample-controller => k8s.io/sample-controller v0.34.1

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.34.1

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.34.1

replace k8s.io/kms => k8s.io/kms v0.34.1

replace k8s.io/endpointslice => k8s.io/endpointslice v0.34.1

replace k8s.io/autoscaler/cluster-autoscaler/apis => ./apis

replace k8s.io/cri-client => k8s.io/cri-client v0.34.1

replace k8s.io/externaljwt => k8s.io/externaljwt v0.34.1
