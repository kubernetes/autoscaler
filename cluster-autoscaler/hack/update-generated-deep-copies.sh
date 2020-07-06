#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

go build -o "${OS_OUTPUT_BINPATH}/deepcopy-gen" "k8s.io/code-generator/cmd/deepcopy-gen"

${OS_OUTPUT_BINPATH}/deepcopy-gen \
                --go-header-file "hack/boilerplate/boilerplate.go.txt" \
                --input-dirs "${PROJECT_PREFIX}/cloudprovider/clusterapi" \
                --output-file-base zz_generated.deepcopy
