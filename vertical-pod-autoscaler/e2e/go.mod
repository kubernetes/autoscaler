// for updating kubernetes dependencies, use `autoscaler/vertical-pod-autoscaler/hack/update-kubernetes-deps-in-e2e.sh`
// for any other update, use standard `go mod` commands

module k8s.io/autoscaler/vertical-pod-autoscaler/e2e

go 1.21

require (
	github.com/onsi/ginkgo/v2 v2.9.4
	github.com/onsi/gomega v1.27.6
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/autoscaler/vertical-pod-autoscaler v0.0.0-20200605154545-936eea18fb1d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.28.3
	k8s.io/klog/v2 v2.100.1
	k8s.io/kubernetes v1.28.3
	k8s.io/pod-security-admission v0.0.0
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
)

require (
	cloud.google.com/go/compute v1.19.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.18.1-0.20220218231025-f11817397a1b // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.8.25 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/container-storage-interface/spec v1.8.0 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/cadvisor v0.47.3 // indirect
	github.com/google/cel-go v0.16.1 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.7.1 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/moby/ipvs v1.1.0 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mrunalp/fileutils v0.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.1.7 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/opencontainers/selinux v1.10.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.9 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.9 // indirect
	go.etcd.io/etcd/client/v3 v3.5.9 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.35.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.35.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.35.1 // indirect
	go.opentelemetry.io/otel v1.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.10.0 // indirect
	go.opentelemetry.io/otel/metric v0.31.0 // indirect
	go.opentelemetry.io/otel/sdk v1.10.0 // indirect
	go.opentelemetry.io/otel/trace v1.10.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.4.0 // indirect
	golang.org/x/tools v0.8.0 // indirect
	google.golang.org/api v0.114.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230526161137-0005af68ea54 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230525234035-dd9d682886f9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.0.0 // indirect
	k8s.io/apiserver v0.28.3 // indirect
	k8s.io/cloud-provider v0.28.3 // indirect
	k8s.io/component-helpers v0.28.3 // indirect
	k8s.io/controller-manager v0.28.3 // indirect
	k8s.io/cri-api v0.28.3 // indirect
	k8s.io/csi-translation-lib v0.26.1 // indirect
	k8s.io/dynamic-resource-allocation v0.0.0 // indirect
	k8s.io/kms v0.28.3 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubectl v0.0.0 // indirect
	k8s.io/kubelet v0.28.3 // indirect
	k8s.io/legacy-cloud-providers v0.0.0 // indirect
	k8s.io/mount-utils v0.25.0 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.1.2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.28.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.28.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.3
	k8s.io/apiserver => k8s.io/apiserver v0.28.3
	k8s.io/autoscaler => ../../
	k8s.io/autoscaler/vertical-pod-autoscaler => ../
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.28.3
	k8s.io/client-go => k8s.io/client-go v0.28.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.28.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.28.3
	k8s.io/code-generator => k8s.io/code-generator v0.28.3
	k8s.io/component-base => k8s.io/component-base v0.28.3
	k8s.io/cri-api => k8s.io/cri-api v0.28.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.28.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.28.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.28.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.28.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.28.3
	k8s.io/kubectl => k8s.io/kubectl v0.28.3
	k8s.io/kubelet => k8s.io/kubelet v0.28.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.28.3
	k8s.io/metrics => k8s.io/metrics v0.28.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.28.3
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.28.3
	k8s.io/sample-controller => k8s.io/sample-controller v0.28.3
)

replace k8s.io/component-helpers => k8s.io/component-helpers v0.28.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.28.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.28.3

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.28.3

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.28.3

replace k8s.io/kms => k8s.io/kms v0.28.3

replace k8s.io/endpointslice => k8s.io/endpointslice v0.28.3
