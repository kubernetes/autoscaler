#!/bin/bash

# : ${DOCKER_USER:? required}

export GO111MODULE=on 
export GOPROXY=https://goproxy.cn
# build webhook
#CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/app_linux
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/vpa-informer
# build docker image
# docker build --no-cache -t resources-controller:v1 .
# rm -rf admission-webhook-example

# docker push ${DOCKER_USER}/admission-webhook-example:v1
