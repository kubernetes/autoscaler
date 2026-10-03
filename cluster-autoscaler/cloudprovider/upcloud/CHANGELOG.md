# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

### Changed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with `master` branch (`ff19d41`)
- adopt new cloud provider builder pattern

## [1.2.0]

### Added

- added support for using the autoscaler with an UpCloud API token, via `UPCLOUD_TOKEN` env var

### Fixed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with upstream `master` branch (CA 1.35.0-rc.0)
  - fixed regression with `*coreoptions.AutoscalerOptions` in UpCloud provider builder
  - fixed regression in `upCloudNodeGroup`: added `ForceDeleteNodes()` and delegating it to the existing `DeleteNodes()` flow
  - fixed regression in `upCloudNodeGroup`: updated `TemplateNodeInfo()` to match the current upstream `cloudprovider.NodeGroup` signature
- pinned the version of UpCloud autoscaler Docker image version in examples

### Images

| Kubernetes Version | Image                                 |
| ------------------ | ------------------------------------- |
| 1.29.x             | ghcr.io/upcloudltd/autoscaler:v1.29.5 |

## [1.1.0]

### Added

- set custom limits for node groups using `--nodes` parameter

### Changed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with `master` branch (CA 1.31.0-beta.0)

### Images

| Kubernetes Version | Image                                 |
| ------------------ | ------------------------------------- |
| 1.29.x             | ghcr.io/upcloudltd/autoscaler:v1.29.4 |
| 1.28.x             | ghcr.io/upcloudltd/autoscaler:v1.28.6 |
| 1.27.X             | ghcr.io/upcloudltd/autoscaler:v1.27.8 |

## [1.0.0]

First stable release

[Unreleased]: https://github.com/UpCloudLtd/autoscaler/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/UpCloudLtd/autoscaler/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/UpCloudLtd/autoscaler/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/UpCloudLtd/autoscaler/releases/tag/v1.0.0
