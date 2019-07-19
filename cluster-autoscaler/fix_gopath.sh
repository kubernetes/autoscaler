#!/bin/bash

echo Updating $GOPATH

to_fix=`ls $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io`
for item in $to_fix; do
   echo Updating staging dep $item
   rm -rf $GOPATH/src/k8s.io/$item
   mkdir $GOPATH/src/k8s.io/$item
   cd $GOPATH/src/k8s.io/$item
   git init
   cp -R $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io/$item/* ./
   git add .
   git commit -a -m from_staging
done

with_vendor=`find $GOPATH/src/ -type d | grep vendor | grep -v 'vendor/'`
for item in $with_vendor; do
    echo Removing vendor from $item
    cd $item
    cd ..
    rm -rf $item
    git commit -a -m no_vendor
done

echo Overriding GKE API
mkdir -p $GOPATH/src/google.golang.org/api/container/v1alpha1
cp $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/_override/google.golang.org/api/container/v1alpha1/*  $GOPATH/src/google.golang.org/api/container/v1alpha1
cp $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/_override/google.golang.org/api/container/v1beta1/*  $GOPATH/src/google.golang.org/api/container/v1beta1
cd $GOPATH/src/google.golang.org/api/
git add .
git commit -a -m "Api override for NAP"

echo Overriding AKS API
mkdir -p $GOPATH/src/github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice
cp $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/_override/github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice/* $GOPATH/src/github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice/
cp -r $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/_override/github.com/Azure/azure-sdk-for-go/version $GOPATH/src/github.com/Azure/azure-sdk-for-go/
cd $GOPATH/src/github.com/Azure/azure-sdk-for-go/
git add .
git commit -a -m "Api override for AKS"

echo Overriding Azure autorest
cp -r $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/_override/github.com/Azure/go-autorest/* $GOPATH/src/github.com/Azure/go-autorest/
cd $GOPATH/src/github.com/Azure/go-autorest/
git add .
git commit -a -m "Api override for autorest"
