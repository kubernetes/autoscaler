#!/bin/bash
set -eo pipefail

if [[ -n $(git status -s) ]]; then
    echo "Clean git working tree required"
    exit 1
fi

CA_VERSION=${1:?"usage: update-chart-version-readme.sh <cluster-autoscaler-version> (must be a row present in README.md)"}

BASE=$(dirname $0)

REF=$(git branch --show-current | grep . || git rev-parse HEAD)

function cleanup {
    echo 'Going back to $REF'
    git checkout "$REF" &>/dev/null
    trap - EXIT
}

trap cleanup EXIT

# Build version map
VERSIONS=$(
    git tag | grep cluster-autoscaler-chart | sort -V | while read ver; do
        echo "Checking chart release: $ver" >&2
        git checkout $ver &>/dev/null

        (
            set -eo pipefail
            cat $BASE/../charts/cluster-autoscaler/Chart.yaml \
                | grep -e version -e appVersion
        ) \
            | sed -E -e 's/^([^:]+): (.*)/"\1": "\2"/g' \
            | tr '\n' ',' \
            | sed -E -e 's/(.*),/{\1}/g'
    done | jq -s '{"cluster-autoscaler": group_by(.appVersion) | map({version: .[0].appVersion, charts: . | map(.version)})}'
        )

# Get min version where cluster-autoscaler v$CA_VERSION is used
CA_VERSION=${CA_VERSION%%.X}
MIN_COMPATIBLE_VERSION=$(echo "$VERSIONS" | jq -r --arg ver $CA_VERSION '.["cluster-autoscaler"] | map(select(.version | startswith("\($ver).")) | .charts) | flatten | .[]' | sort -V | head -n1)

if [[ -z "$MIN_COMPATIBLE_VERSION" ]]; then
    echo "No chart versions using cluster-autoscaler v$CA_VERSION detected" >&2
    exit 0
fi

echo "Detected min compatible version for cluster-autoscaler v$CA_VERSION"

cleanup

# Replace README info with parsed data
awk "!/^\|/ { versions=0 }
/$CA_VERSION/ { if (versions) print(\"\", \$2, \$3, \"$MIN_COMPATIBLE_VERSION+\" \"\", \"\"); else print(\$0) }
!/$CA_VERSION/ { print }
/Kubernetes Version/ { versions=1; FS=\"|\"; OFS=\"|\"; }" $BASE/README.md > README.md.tmp

mv README.md.tmp $BASE/README.md
