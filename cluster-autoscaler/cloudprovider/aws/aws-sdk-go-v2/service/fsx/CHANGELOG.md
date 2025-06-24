# v1.54.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2025-05-29)

* **Feature**: FSx API changes to support the public launch of new Intelligent Tiering storage class on Amazon FSx for Lustre

# v1.53.4 (2025-05-09)

* No change notes available for this release.

# v1.53.3 (2025-04-17)

* No change notes available for this release.

# v1.53.2 (2025-04-03)

* No change notes available for this release.

# v1.53.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.53.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2025-02-12)

* **Feature**: Support for in-place Lustre version upgrades

# v1.51.10 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.9 (2025-02-04)

* No change notes available for this release.

# v1.51.8 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.7 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.6 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.51.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.51.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.2 (2025-01-03)

* No change notes available for this release.

# v1.51.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2024-12-02)

* **Feature**: FSx API changes to support the public launch of the Amazon FSx Intelligent Tiering for OpenZFS storage class.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2024-11-27)

* **Feature**: This release adds EFA support to increase FSx for Lustre file systems' throughput performance to a single client instance. This can be done by specifying EfaEnabled=true at the time of creation of Persistent_2 file systems.

# v1.49.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.49.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.4 (2024-10-03)

* No change notes available for this release.

# v1.48.3 (2024-09-27)

* No change notes available for this release.

# v1.48.2 (2024-09-25)

* **Documentation**: Doc-only update to address Lustre S3 hard-coded names.

# v1.48.1 (2024-09-23)

* No change notes available for this release.

# v1.48.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.6 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.47.5 (2024-09-04)

* No change notes available for this release.

# v1.47.4 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.3 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.2 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.1 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2024-07-09)

* **Feature**: Adds support for FSx for NetApp ONTAP 2nd Generation file systems, and FSx for OpenZFS Single AZ HA file systems.

# v1.46.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.45.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.2 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-06-06)

* **Feature**: This release adds support to increase metadata performance on FSx for Lustre file systems beyond the default level provisioned when a file system is created. This can be done by specifying MetadataConfiguration during the creation of Persistent_2 file systems or by updating it on demand.

# v1.43.10 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.9 (2024-05-23)

* No change notes available for this release.

# v1.43.8 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.7 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.6 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.43.5 (2024-05-03)

* No change notes available for this release.

# v1.43.4 (2024-04-05)

* No change notes available for this release.

# v1.43.3 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-03-04)

* **Feature**: Added support for creating FSx for NetApp ONTAP file systems with up to 12 HA pairs, delivering up to 72 GB/s of read throughput and 12 GB/s of write throughput.

# v1.42.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.41.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.41.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.41.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2023-12-19)

* **Feature**: Added support for FSx for OpenZFS on-demand data replication across AWS accounts and/or regions.Added the IncludeShared attribute for DescribeSnapshots.Added the CopyStrategy attribute for OpenZFSVolumeConfiguration.

# v1.39.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.39.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.39.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.38.0 (2023-11-27)

* **Feature**: Added support for FSx for ONTAP scale-out file systems and FlexGroup volumes. Added the HAPairs field and ThroughputCapacityPerHAPair for filesystem. Added AggregateConfiguration (containing Aggregates and ConstituentsPerAggregate) and SizeInBytes for volume.

# v1.37.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2023-11-16)

* **Feature**: Enables customers to update their PerUnitStorageThroughput on their Lustre file systems.

# v1.36.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.33.1 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2023-10-06)

* **Feature**: After performing steps to repair the Active Directory configuration of a file system, use this action to initiate the process of attempting to recover to the file system.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.6 (2023-09-08)

* **Documentation**: Amazon FSx documentation fixes

# v1.32.5 (2023-08-29)

* **Documentation**: Documentation updates for project quotas.

# v1.32.4 (2023-08-25)

* No change notes available for this release.

# v1.32.3 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.2 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-08-09)

* **Feature**: For FSx for Lustre, add new data repository task type, RELEASE_DATA_FROM_FILESYSTEM, to release files that have been archived to S3. For FSx for Windows, enable support for configuring and updating SSD IOPS, and for updating storage type. For FSx for OpenZFS, add new deployment type, MULTI_AZ_1.

# v1.31.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2023-08-01)

* No change notes available for this release.

# v1.31.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-07-13)

* **Feature**: Amazon FSx for NetApp ONTAP now supports SnapLock, an ONTAP feature that enables you to protect your files in a volume by transitioning them to a write once, read many (WORM) state.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.3 (2023-06-23)

* **Documentation**: Update to Amazon FSx documentation.

# v1.29.2 (2023-06-15)

* No change notes available for this release.

# v1.29.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2023-06-12)

* **Feature**: Amazon FSx for NetApp ONTAP now supports joining a storage virtual machine (SVM) to Active Directory after the SVM has been created.

# v1.28.13 (2023-05-25)

* No change notes available for this release.

# v1.28.12 (2023-05-04)

* No change notes available for this release.

# v1.28.11 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.10 (2023-04-14)

* No change notes available for this release.

# v1.28.9 (2023-04-10)

* No change notes available for this release.

# v1.28.8 (2023-04-07)

* **Documentation**: Amazon FSx for Lustre now supports creating data repository associations on Persistent_1 and Scratch_2 file systems.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.28.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.28.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.1 (2023-01-23)

* No change notes available for this release.

# v1.28.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.27.0 (2022-12-23)

* **Feature**: Fix a bug where a recent release might break certain existing SDKs.

# v1.26.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2022-11-29)

* **Feature**: This release adds support for 4GB/s / 160K PIOPS FSx for ONTAP file systems and 10GB/s / 350K PIOPS FSx for OpenZFS file systems (Single_AZ_2). For FSx for ONTAP, this also adds support for DP volumes, snapshot policy, copy tags to backups, and Multi-AZ route table updates.

# v1.25.4 (2022-10-31)

* No change notes available for this release.

# v1.25.3 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2022-10-19)

* No change notes available for this release.

# v1.25.0 (2022-09-29)

* **Feature**: This release adds support for Amazon File Cache.

# v1.24.14 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.13 (2022-09-19)

* No change notes available for this release.

# v1.24.12 (2022-09-14)

* **Documentation**: Documentation update for Amazon FSx.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.11 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.10 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.9 (2022-08-29)

* **Documentation**: Documentation updates for Amazon FSx for NetApp ONTAP.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.8 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.7 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.6 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.5 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.4 (2022-07-29)

* **Documentation**: Documentation updates for Amazon FSx

# v1.24.3 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-05-25)

* **Feature**: This release adds root squash support to FSx for Lustre to restrict root level access from clients by mapping root users to a less-privileged user/group with limited permissions.

# v1.23.2 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-04-13)

* **Feature**: This release adds support for deploying FSx for ONTAP file systems in a single Availability Zone.

# v1.22.0 (2022-04-05)

* **Feature**: Provide customers more visibility into file system status by adding new "Misconfigured Unavailable" status for Amazon FSx for Windows File Server.

# v1.21.0 (2022-03-30)

* **Feature**: This release adds support for modifying throughput capacity for FSx for ONTAP file systems.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service client model to latest release.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-01-28)

* **Feature**: Updated to latest API model.

# v1.17.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.14.0 (2021-12-02)

* **Feature**: API client updated
* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.12.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service to latest API model.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-09-02)

* **Feature**: API client updated

# v1.8.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-11)

* **Feature**: Updated to latest API model.

# v1.5.0 (2021-06-04)

* **Feature**: Updated service client to latest API model.

# v1.4.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

