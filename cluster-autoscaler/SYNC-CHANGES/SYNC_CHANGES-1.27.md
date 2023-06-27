<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.27.0](#v1270)
  - [Synced with which upstream CA](#synced-with-which-upstream-ca)
  - [Changes made](#changes-made)
    - [During merging](#during-merging)
    - [During vendoring k8s](#during-vendoring-k8s)
    - [Others](#others)


# v1.27.0


## Synced with which upstream CA
Somehow in this release CA has done some goof-up with 2 release done for v1.27 - v1.27.0/v1.27.1.
For Gardener we have used v1.27.1 to be on the safer side for future patch release. 
[v1.27.1](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.27.1/cluster-autoscaler)

## Changes made
 - See general release notes of 1.27.1: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.27.1
 - The cloudProvider interface has added a new method [`GetNodeGpuConfig`](https://github.com/kubernetes/autoscaler/blob/41a957b7e8b8630cb4c3e7811bdcecf85c510f09/cluster-autoscaler/cloudprovider/cloud_provider.go#L139) which needs implementation. For now we have implemented this in our mcm_cloud_provider.go by returning nil. However, we need to investigate its usage and provide relevant implementation if required/desired with subsequent release.
 - `Go-Lang` version updated from v1.19.5 to v1.20.5 in - 
   -  `.ci/pipeline_definitions` 
   - `cluster-autoscaler/Dockerfile`   
   (*this is required otherwise the concourse build job will continue to fail as this release is build in upstream with go v1.20)*


### During merging
  - CA `Readme` was merged with upstream changes capturing `PATCH RELEASE` schedule for coming months. 
  - 3 new regions were added in defaults.go namely:
     - 	ApSouth2RegionID     = "ap-south-2"     // Asia Pacific (Hyderabad).
	   -  EuCentral2RegionID   = "eu-central-2"   // Europe (Zurich).
	   -  EuSouth2RegionID     = "eu-south-2"     // Europe (Spain).
  - A new controller `balancer` is added to the `autoscaler` repo at a peer level to `cluster-autoscaler` and `vertical-pod-autoscaler`. We will need further investigation to see if this component is useful for gardener. 
  - The types for fields in `Readiness` struct in `clusterstate/clusterstate.go` are changed now to []string from int. 

### During vendoring k8s
- mcm v0.47.0 -> 0.49.0
- mcm-provider-aws v0.16.0 -> v0.17.0
- mcm-provider-azure v0.9.0 -> v0.10.0 -- no change here

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.
- With Kubernetes 1.27, the feature gates `CSIMigration` and `CSIMigrationAWS` will not be supported anymore. See [this](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/#descriptions-for-removed-feature-gates).
CSIMigration PR on gardener, which mutates CA feature gates by gardener extensions -> https://github.com/gardener/gardener/pull/6047/. 