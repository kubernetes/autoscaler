#!/usr/bin/env bash
#
# This script is intended for maintainers only, for vendoring egoscale
# module files in this repository.
#
# The following modules have been vendored manually and are not managed
# by this script:
#   - github.com/gofrs/uuid
#   - github.com/deepmap/oapi-codegen/{runtime,types}
#   - k8s.io/klog

if [[ $# -ne 1 ]]; then 
    echo "usage: $0 <path to local egoscale Git checkout>"
    exit 1
fi

EGOSCALE_DIR=$(readlink -f "$1")

rm -rf ./internal/github.com/exoscale/egoscale/*
cp -rf $EGOSCALE_DIR/* ./internal/github.com/exoscale/egoscale/
rm -rf ./internal/github.com/exoscale/egoscale/{*_test.go,doc.go,api/v2/*_test.go,internal/v2/*_test.go,internal/v2/mock.go,go.*,gopher.png,*.md,admin,cmd,generate,test,website}

find ./internal -name '*.go' | while read f; do
    sed -i -r \
        -e 's#"github.com/exoscale/egoscale#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale#g' \
        -e 's#"github.com/gofrs/uuid#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/gofrs/uuid#g' \
        -e 's#"github.com/deepmap/oapi-codegen#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/deepmap/oapi-codegen#g' \
        -e 's#"k8s.io/klog#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/k8s.io/klog#g' \
        -e "/[Cc]ode generated/d" \
        "$f"
    cat <<EOF > "$f.tmp"
/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

EOF
    cat "$f" >> "$f.tmp"
    mv -f "$f.tmp" "$f"
    gofmt -w "$f"
done
