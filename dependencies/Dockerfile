FROM golang:1.11

RUN git config --global user.email "autoscaler@kubernetes.io" && git config --global user.name "Cluster Autoscaler"
RUN apt-get update && apt-get --yes install libseccomp-dev

RUN go get \
	"github.com/tools/godep" \
	"github.com/tinylib/msgp" \
	"github.com/opentracing/opentracing-go" \
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace" \
	"cloud.google.com/go/profiler"

WORKDIR "${GOPATH}/src/k8s.io"

# https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#how-can-i-update-ca-dependencies-particularly-k8siokubernetes
RUN git clone "https://github.com/kubernetes/kubernetes" && \
	cd "kubernetes" && \
	git checkout release-1.14 && \
	godep restore

RUN git clone "https://github.com/kubernetes/autoscaler" && \
	cd "autoscaler" && \
	git checkout cluster-autoscaler-1.14.2 && \
	cd "cluster-autoscaler" && \
	rm -rf vendor Godeps && \
	./fix_gopath.sh

# kubernetes requires the v0.1.0 google APIs, which do not have the profiler package.
# autoscaler ./fix_gopath.sh makes commits in many google API depedencies.
# said dependencies must be removed and updated.
RUN cd "${GOPATH}/src" && rm -rf \
		"cloud.google.com/go" \
		"github.com/golang/protobuf" \
		"github.com/hashicorp/golang-lru" \
		"google.golang.org/genproto" \
		"google.golang.org/api" \
		"google.golang.org/grpc" \
		"golang.org/x/oauth2" \
		"golang.org/x/net" \
		"golang.org/x/sys" \
		"golang.org/x/text"

RUN go get -u \
		"cloud.google.com/go" \
		"cloud.google.com/go/profiler" \
		"github.com/golang/protobuf/proto" \
		"google.golang.org/grpc" \
		"golang.org/x/oauth2" \
		"golang.org/x/net/context"

WORKDIR "${GOPATH}/src/k8s.io/autoscaler/cluster-autoscaler"
COPY "cluster-autoscaler/main.go" "main.go"
RUN godep save ./...

# .dockerignore must prevent host Godeps and vendor from being copied
ADD "cluster-autoscaler" "."
RUN make build-binary
