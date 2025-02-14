#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

echo "> Installing gosec"

TOOLS_BIN_DIR=${TOOLS_BIN_DIR:-$(dirname $0)/bin}
if [[ ! -d $TOOLS_BIN_DIR ]]; then
  mkdir -p $TOOLS_BIN_DIR
fi

platform=$(uname -s | tr '[:upper:]' '[:lower:]')
version=$GOSEC_VERSION
echo "gosec version:$GOSEC_VERSION"
case $(uname -m) in
  aarch64 | arm64)
    arch="arm64"
    ;;
  x86_64)
    arch="amd64"
    ;;
  *)
    echo "Unknown architecture"
    exit 1
    ;;
esac

archive_name="gosec_${version#v}_${platform}_${arch}"
file_name="${archive_name}.tar.gz"

temp_dir="$(mktemp -d)"
function cleanup {
  rm -rf "${temp_dir}"
}
trap cleanup EXIT ERR INT TERM
echo "Downloading from: https://github.com/securego/gosec/releases/download/${version}/${file_name}"
curl -L -o ${temp_dir}/${file_name} "https://github.com/securego/gosec/releases/download/${version}/${file_name}"

tar -xzm -C "${temp_dir}" -f "${temp_dir}/${file_name}"
mv "${temp_dir}/gosec" $TOOLS_BIN_DIR
chmod +x $TOOLS_BIN_DIR/gosec
