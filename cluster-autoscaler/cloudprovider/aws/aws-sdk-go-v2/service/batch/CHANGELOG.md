# v1.52.6 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.5 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.4 (2025-06-06)

* No change notes available for this release.

# v1.52.3 (2025-04-25)

* No change notes available for this release.

# v1.52.2 (2025-04-10)

* No change notes available for this release.

# v1.52.1 (2025-04-03)

* No change notes available for this release.

# v1.52.0 (2025-03-27)

* **Feature**: This release will enable two features: Firelens log driver, and Execute Command on Batch jobs on ECS. Both features will be passed through to ECS.

# v1.51.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.51.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2025-02-26)

* **Feature**: AWS Batch: Resource Aware Scheduling feature support

# v1.49.13 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Documentation**: This documentation-only update corrects some typos.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.12 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.11 (2025-02-04)

* No change notes available for this release.

# v1.49.10 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.9 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.8 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.49.7 (2025-01-21)

* **Documentation**: Documentation-only update: clarified the description of the shareDecaySeconds parameter of the FairsharePolicy data type, clarified the description of the priority parameter of the JobQueueDetail data type.

# v1.49.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.49.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.4 (2025-01-14)

* No change notes available for this release.

# v1.49.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2025-01-08)

* No change notes available for this release.

# v1.49.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-12-17)

* **Feature**: This feature allows AWS Batch on Amazon EKS to support configuration of Pod Annotations, overriding Namespace on which the Batch job's Pod runs on, and allows Subpath and Persistent Volume claim to be set for AWS Batch on Amazon EKS jobs.

# v1.48.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2024-11-08)

* **Feature**: This feature allows override LaunchTemplates to be specified in an AWS Batch Compute Environment.

# v1.47.2 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.47.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2024-10-31)

* **Feature**: Add `podNamespace` to `EksAttemptDetail` and `containerID` to `EksAttemptContainerDetail`.

# v1.46.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.4 (2024-10-03)

* No change notes available for this release.

# v1.45.3 (2024-09-27)

* No change notes available for this release.

# v1.45.2 (2024-09-25)

* No change notes available for this release.

# v1.45.1 (2024-09-23)

* No change notes available for this release.

# v1.45.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.4 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.44.3 (2024-09-04)

* No change notes available for this release.

# v1.44.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-08-22)

* No change notes available for this release.

# v1.44.0 (2024-08-16)

* **Feature**: Improvements of integration between AWS Batch and EC2.

# v1.43.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-07-10.2)

* **Feature**: This feature allows AWS Batch Jobs with EKS container orchestration type to be run as Multi-Node Parallel Jobs.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-07-10)

* **Feature**: This feature allows AWS Batch Jobs with EKS container orchestration type to be run as Multi-Node Parallel Jobs.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.40.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2024-06-17)

* **Feature**: Add v2 smoke tests and smithy smokeTests trait for SDK testing.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-06-03)

* **Feature**: This release adds support for the AWS Batch GetJobQueueSnapshot API operation.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.4 (2024-05-23)

* No change notes available for this release.

# v1.37.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.37.0 (2024-04-11)

* **Feature**: This release adds the task properties field to attempt details and the name field on EKS container detail.

# v1.36.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2024-03-27)

* **Feature**: This feature allows AWS Batch to support configuration of imagePullSecrets and allowPrivilegeEscalation for jobs running on EKS

# v1.35.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2024-03-08)

* **Feature**: This release adds JobStateTimeLimitActions setting to the Job Queue API. It allows you to configure an action Batch can take for a blocking job in front of the queue after the defined period of time. The new parameter applies for ECS, EKS, and FARGATE Job Queues.

# v1.34.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2024-02-28)

* **Feature**: This release adds Batch support for configuration of multicontainer jobs in ECS, Fargate, and EKS. This support is available for all types of jobs, including both array jobs and multi-node parallel jobs.

# v1.33.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.32.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.32.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2024-02-09)

* **Feature**: This feature allows Batch to support configuration of repository credentials for jobs running on ECS

# v1.30.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.6 (2023-12-20)

* No change notes available for this release.

# v1.30.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.30.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.30.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.29.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.26.7 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.6 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.5 (2023-08-23)

* No change notes available for this release.

# v1.26.4 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.2 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-08-01)

* **Feature**: This release adds support for price capacity optimized allocation strategy for Spot Instances.

# v1.25.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-07-03)

* **Feature**: This feature allows customers to use AWS Batch with Linux with ARM64 CPU Architecture and X86_64 CPU Architecture with Windows OS on Fargate Platform.

# v1.23.8 (2023-06-15)

* No change notes available for this release.

# v1.23.7 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.6 (2023-06-01)

* No change notes available for this release.

# v1.23.5 (2023-05-04)

* No change notes available for this release.

# v1.23.4 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.3 (2023-04-14)

* No change notes available for this release.

# v1.23.2 (2023-04-10)

* No change notes available for this release.

# v1.23.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-03-30)

* **Feature**: This feature allows Batch on EKS to support configuration of Pod Labels through Metadata for Batch on EKS Jobs.

# v1.22.0 (2023-03-23)

* **Feature**: This feature allows Batch to support configuration of ephemeral storage size for jobs running on FARGATE

# v1.21.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.21.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.21.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2023-01-09)

* No change notes available for this release.

# v1.21.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.20.0 (2022-12-20)

* **Feature**: Adds isCancelled and isTerminated to DescribeJobs response.

# v1.19.3 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.2 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2022-11-16)

* **Documentation**: Documentation updates related to Batch on EKS

# v1.19.0 (2022-10-24)

* **Feature**: This release adds support for AWS Batch on Amazon EKS.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.17 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.16 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.15 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.14 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.13 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.12 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.11 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.10 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.9 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.8 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.7 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.6 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-05-27)

* No change notes available for this release.

# v1.18.3 (2022-05-18)

* **Documentation**: Documentation updates for AWS Batch.

# v1.18.2 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-04-14)

* **Feature**: Enables configuration updates for compute environments with BEST_FIT_PROGRESSIVE and SPOT_CAPACITY_OPTIMIZED allocation strategies.

# v1.17.1 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-03-25)

* **Feature**: Bug Fix: Fixed a bug where shapes were marked as unboxed and were not serialized and sent over the wire, causing an API error from the service.
* This is a breaking change, and has been accepted due to the API operation not being usable due to the members modeled as unboxed (aka value) types. The update changes the members to boxed (aka pointer) types so that the zero value of the members can be handled correctly by the SDK and service. Your application will fail to compile with the updated module. To workaround this you'll need to update your application to use pointer types for the members impacted.

# v1.16.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.11.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-11-12)

* **Feature**: Updated service to latest API model.

# v1.9.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

