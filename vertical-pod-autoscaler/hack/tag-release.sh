#!/bin/bash

# Copyright 2025 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# Convenience script that tags the VPA automatically built images with a
# version tag in preparation for release.

# The following environment variables must be set for this script:
# BUILD_TAG: corresponds to the tag of the automatic build (eg. v20250512-cluster-autoscaler-chart-9.46.6-151-g2b33c4c79)
# TAG: corresponds to the version of the release (eg. 1.4.0)

# See VPA release instructions (in RELEASE.md) before using this file.

if [[ -z "$BUILD_TAG" ]]; then
    echo "BUILD_TAG must be set to the existing image build tag (eg. BUILD_TAG=v20250512-cluster-autoscaler-chart-9.46.6-151-g2b33c4c79)"
    exit 1
fi
if [[ -z "$TAG" ]]; then
    echo "TAG must be set to the VPA release number (eg. TAG=1.4.0)"
    exit 1
fi

echo "Adding tag $TAG based on built tag $BUILD_TAG"

read -p "Press enter to continue"

gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller-amd64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller-amd64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller-arm:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller-arm:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller-arm64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller-arm64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller-ppc64le:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller-ppc64le:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-admission-controller-s390x:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-admission-controller-s390x:$TAG --project=k8s-staging-autoscaling

gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender-amd64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender-amd64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender-arm:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender-arm:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender-arm64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender-arm64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender-ppc64le:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender-ppc64le:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-recommender-s390x:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-recommender-s390x:$TAG --project=k8s-staging-autoscaling

gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater-amd64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater-amd64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater-arm:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater-arm:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater-arm64:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater-arm64:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater-ppc64le:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater-ppc64le:$TAG --project=k8s-staging-autoscaling
gcloud container images add-tag -q gcr.io/k8s-staging-autoscaling/vpa-updater-s390x:$BUILD_TAG gcr.io/k8s-staging-autoscaling/vpa-updater-s390x:$TAG --project=k8s-staging-autoscaling
