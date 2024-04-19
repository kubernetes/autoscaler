module k8s.io/autoscaler/cluster-autoscaler

go 1.22

toolchain go1.22.2

require (
	cloud.google.com/go/compute/metadata v0.2.3
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/adal v0.9.23
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/skewer v0.0.14
	github.com/aws/aws-sdk-go v1.44.241
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/digitalocean/godo v1.27.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.6.0
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.6.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.16.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.19.0
	golang.org/x/net v0.20.0
	golang.org/x/oauth2 v0.10.0
	golang.org/x/sys v0.17.0
	google.golang.org/api v0.126.0
	google.golang.org/grpc v1.58.3
	google.golang.org/protobuf v1.31.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.30.0-alpha.3
	k8s.io/apimachinery v0.30.0-alpha.3
	k8s.io/apiserver v0.30.0-alpha.3
	k8s.io/autoscaler/cluster-autoscaler/apis v0.0.0-00010101000000-000000000000
	k8s.io/client-go v0.30.0-alpha.3
	k8s.io/cloud-provider v0.30.0-alpha.3
	k8s.io/cloud-provider-aws v1.27.0
	k8s.io/component-base v0.30.0-alpha.3
	k8s.io/component-helpers v0.30.0-alpha.3
	k8s.io/klog/v2 v2.120.1
	k8s.io/kubelet v0.30.0-alpha.3
	k8s.io/kubernetes v1.30.0-alpha.3
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/cloud-provider-azure v1.28.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.18.1-0.20220218231025-f11817397a1b // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.8.25 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230321174746-8dcc6526cfb1 // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/container-storage-interface/spec v1.8.0 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/dnaeon/go-vcr v1.2.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/cadvisor v0.48.1 // indirect
	github.com/google/cel-go v0.17.7 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/onsi/ginkgo/v2 v2.16.0 // indirect
	github.com/onsi/gomega v1.31.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.1.12 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.10 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.10 // indirect
	go.etcd.io/etcd/client/v3 v3.5.10 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0 // indirect
	go.opentelemetry.io/otel v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.19.0 // indirect
	go.opentelemetry.io/otel/sdk v1.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230726155614-23370e0ffb3e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.0.0 // indirect
	k8s.io/code-generator v0.30.0-alpha.3 // indirect
	k8s.io/controller-manager v0.30.0-alpha.3 // indirect
	k8s.io/cri-api v0.30.0-alpha.3 // indirect
	k8s.io/csi-translation-lib v0.27.0 // indirect
	k8s.io/dynamic-resource-allocation v0.0.0 // indirect
	k8s.io/gengo v0.0.0-20230829151522-9cce18d56c01 // indirect
	k8s.io/kms v0.30.0-alpha.3 // indirect
	k8s.io/kube-openapi v0.0.0-20231113174909-778a5567bc1e // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubectl v0.28.0 // indirect
	k8s.io/mount-utils v0.26.0-alpha.0 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.29.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace github.com/aws/aws-sdk-go/service/eks => github.com/aws/aws-sdk-go/service/eks v1.38.49

replace github.com/digitalocean/godo => github.com/digitalocean/godo v1.27.0

replace github.com/rancher/go-rancher => github.com/rancher/go-rancher v0.1.0

replace k8s.io/api => k8s.io/api v0.30.0-alpha.3

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.30.0-alpha.3

replace k8s.io/apimachinery => k8s.io/apimachinery v0.30.0-alpha.3

replace k8s.io/apiserver => k8s.io/apiserver v0.30.0-alpha.3

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.30.0-alpha.3

replace k8s.io/client-go => k8s.io/client-go v0.30.0-alpha.3

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.30.0-alpha.3

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.30.0-alpha.3

replace k8s.io/code-generator => k8s.io/code-generator v0.30.0-alpha.3

replace k8s.io/component-base => k8s.io/component-base v0.30.0-alpha.3

replace k8s.io/component-helpers => k8s.io/component-helpers v0.30.0-alpha.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.30.0-alpha.3

replace k8s.io/cri-api => k8s.io/cri-api v0.30.0-alpha.3

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.30.0-alpha.3

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.30.0-alpha.3

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.30.0-alpha.3

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.30.0-alpha.3

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.30.0-alpha.3

replace k8s.io/kubectl => k8s.io/kubectl v0.30.0-alpha.3

replace k8s.io/kubelet => k8s.io/kubelet v0.30.0-alpha.3

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.30.0-alpha.3

replace k8s.io/metrics => k8s.io/metrics v0.30.0-alpha.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.30.0-alpha.3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.30.0-alpha.3

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.30.0-alpha.3

replace k8s.io/sample-controller => k8s.io/sample-controller v0.30.0-alpha.3

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.30.0-alpha.3

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.30.0-alpha.3

replace k8s.io/kms => k8s.io/kms v0.30.0-alpha.3

replace k8s.io/endpointslice => k8s.io/endpointslice v0.30.0-alpha.3

replace k8s.io/autoscaler/cluster-autoscaler/apis => ./apis
