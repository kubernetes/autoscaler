#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

root_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
TOOLS_BIN_DIR="${root_dir}/hack/tools/bin"

gosec_report="false"
gosec_report_parse_flags=""
dir_to_exclude=""

parse_flags() {
  while test $# -gt 1; do
    case "$1" in
      --gosec-report)
        shift; gosec_report="$1"
        ;;
      *)
        echo "Unknown argument: $1"
        exit 1
        ;;
    esac
    shift
  done
}

parse_flags "$@"

echo "> Running gosec"
${TOOLS_BIN_DIR}/gosec --version
if [[ "$gosec_report" != "false" ]]; then
  echo "Exporting report to $root_dir/gosec-report.sarif"
  gosec_report_parse_flags="-track-suppressions -fmt=sarif -out=gosec-report.sarif -stdout"
fi

dir_to_exclude="-exclude-dir=cloudprovider/alicloud
-exclude-dir=cloudprovider/aws
-exclude-dir=cloudprovider/azure
-exclude-dir=cloudprovider/baiducloud
-exclude-dir=cloudprovider/bizflycloud
-exclude-dir=cloudprovider/brightbox
-exclude-dir=cloudprovider/builder
-exclude-dir=cloudprovider/cherryservers
-exclude-dir=cloudprovider/civo
-exclude-dir=cloudprovider/cloudstack
-exclude-dir=cloudprovider/clusterapi
-exclude-dir=cloudprovider/digitalocean
-exclude-dir=cloudprovider/equinixmetal
-exclude-dir=cloudprovider/exoscale
-exclude-dir=cloudprovider/externalgrpc
-exclude-dir=cloudprovider/gce
-exclude-dir=cloudprovider/hetzner
-exclude-dir=cloudprovider/huaweicloud
-exclude-dir=cloudprovider/ionoscloud
-exclude-dir=cloudprovider/kamatera
-exclude-dir=cloudprovider/kubemark
-exclude-dir=cloudprovider/kwok
-exclude-dir=cloudprovider/linode
-exclude-dir=cloudprovider/magnum
-exclude-dir=cloudprovider/oci
-exclude-dir=cloudprovider/ovhcloud
-exclude-dir=cloudprovider/rancher
-exclude-dir=cloudprovider/scaleway
-exclude-dir=cloudprovider/tencentcloud
-exclude-dir=cloudprovider/volcengine
-exclude-dir=cloudprovider/vultr
-exclude-dir=apis
-exclude-dir=cluster-state
-exclude-dir=config
-exclude-dir=context
-exclude-dir=core
-exclude-dir=debuggingsnapshot
-exclude-dir=estimator
-exclude-dir=expander
-exclude-dir=hack
-exclude-dir=loop
-exclude-dir=metrics
-exclude-dir=observers
-exclude-dir=processors
-exclude-dir=proposals
-exclude-dir=provisioningrequest
-exclude-dir=simulator
-exclude-dir=util
-exclude-dir=version
"

${TOOLS_BIN_DIR}/gosec -exclude-generated $dir_to_exclude $gosec_report_parse_flags ./...
