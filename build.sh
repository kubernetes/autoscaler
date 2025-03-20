#!/bin/bash

IMAGE_REPO=${IMAGE_REPO:-"125608480246.dkr.ecr.eu-west-3.amazonaws.com"}
IMAGE_TAG=${REPO:-"dev"}
PLATFORMS=${PLATFORMS:-"linux/amd64"}

docker buildx build --platform $PLATFORMS --tag $REPO/vpa-updater:$IMAGE_TAG -f vertical-pod-autoscaler/pkg/updater/Dockerfile vertical-pod-autoscaler/
docker buildx build --platform $PLATFORMS --tag $REPO/vpa-admission-controller:$IMAGE_TAG -f vertical-pod-autoscaler/pkg/admission-controller/Dockerfile vertical-pod-autoscaler/
