## Machine Controller Manager.  
Copyright (c) 2017-2018 SAP SE or an SAP affiliate company. All rights reserved.

## Seed Source

The source code of this component was seeded based on a copy of the following files from kubernetes/kubernetes. 

Kubernetes.  
https://github.com/kubernetes/kubernetes/tree/release-1.8.  
Copyright 2017 The Kubernetes Authors.   
Apache 2 license (https://github.com/kubernetes/kubernetes/blob/release-1.8/LICENSE )

Release: 1.8.   
Commit-ID: 682da6ea1fd7a8b471d84c83b17c5239ded056d5.    
Commit-Message:  Add/Update CHANGELOG-1.8.md for v1.8.6.     
To the left are the list of copied files -> and to the right the current location they are at.  

	cmd/kube-controller-manager/app/controllermanager.go -> cmd/machine-controller-manager/app/controllermanager.go
	cmd/kube-controller-manager/app/options/options.go -> cmd/machine-controller-manager/app/options/options.go
	cmd/kube-controller-manager/controller_manager.go -> cmd/machine-controller-manager/controller_manager.go
	pkg/controller/deployment/deployment_controller.go -> pkg/controller/deployment_controller.go
	pkg/controller/deployment/util/replicaset_util.go -> pkg/controller/deployment_machineset_util.go
	pkg/controller/deployment/progress.go -> pkg/controller/deployment_progress.go
	pkg/controller/deployment/recreate.go -> pkg/controller/deployment_recreate.go
	pkg/controller/deployment/rollback.go -> pkg/controller/deployment_rollback.go
	pkg/controller/deployment/rolling.go -> pkg/controller/deployment_rolling.go
	pkg/controller/deployment/sync.go -> pkg/controller/deployment_sync.go
	pkg/controller/deployment/util/deployment_util.go -> pkg/controller/deployment_util.go
	pkg/controller/deployment/util/hash_test.go -> pkg/controller/hasttest.go
	pkg/controller/deployment/util/pod_util.go -> pkg/controller/machine_util.go
	pkg/controller/replicaset/replica_set.go -> pkg/controller/machineset.go
	pkg/controller/deployment/util/replicaset_util.go -> pkg/controller/machineset_util.go

