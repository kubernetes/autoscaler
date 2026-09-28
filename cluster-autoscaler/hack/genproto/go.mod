module k8s.io/autoscaler/cluster-autoscaler/genproto

go 1.26.0

tool (
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
	google.golang.org/protobuf/cmd/protoc-gen-go
	k8s.io/api/core/v1
	k8s.io/apimachinery/pkg/apis/meta/v1
)

require (
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.5.1 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.37.1 // indirect
	k8s.io/apimachinery v0.37.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad // indirect
	k8s.io/utils v0.0.0-20260626114624-be93311217bd // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
)

replace github.com/stretchr/objx => github.com/stretchr/objx v0.5.3

replace go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.44.0

replace go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.44.0

replace go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.44.0

replace go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.44.0

replace google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa

replace k8s.io/gengo/v2 => k8s.io/gengo/v2 v2.0.0-20260408192533-25e2208e0dc3
