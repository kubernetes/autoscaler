# v1.43.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2025-06-02)

* **Feature**: This release enables AWS Compute Optimizer to analyze Amazon Aurora database clusters and generate Aurora I/O-Optimized recommendations.

# v1.42.3 (2025-04-10)

* No change notes available for this release.

# v1.42.2 (2025-04-03)

* No change notes available for this release.

# v1.42.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.42.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.8 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.7 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.6 (2025-02-04)

* No change notes available for this release.

# v1.41.5 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.4 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.3 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.41.2 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.41.1 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2025-01-09)

* **Feature**: This release expands AWS Compute Optimizer rightsizing recommendation support for Amazon EC2 Auto Scaling groups to include those with scaling policies and multiple instance types.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-11-20)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate optimization recommendations for Amazon Aurora database instances. It also enables Compute Optimizer to identify idle Amazon EC2 instances, Amazon EBS volumes, Amazon ECS services running on Fargate, and Amazon RDS databases.

# v1.39.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.39.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.4 (2024-10-03)

* No change notes available for this release.

# v1.38.3 (2024-09-27)

* No change notes available for this release.

# v1.38.2 (2024-09-25)

* No change notes available for this release.

# v1.38.1 (2024-09-23)

* No change notes available for this release.

# v1.38.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.8 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.37.7 (2024-09-04)

* No change notes available for this release.

# v1.37.6 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.5 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.4 (2024-08-12)

* **Documentation**: Doc only update for Compute Optimizer that fixes several customer-reported issues relating to ECS finding classifications

# v1.37.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.36.0 (2024-06-20)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate optimization recommendations for Amazon RDS MySQL and RDS PostgreSQL.

# v1.35.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.8 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.7 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.6 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.5 (2024-05-23)

* No change notes available for this release.

# v1.34.4 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.3 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.34.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2024-03-28)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate recommendations with a new customization preference, Memory Utilization.

# v1.33.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.32.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.32.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.32.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.31.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.31.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.30.0 (2023-11-27)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate recommendations with customization and discounts preferences.

# v1.29.4 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.3 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.2 (2023-11-14)

* No change notes available for this release.

# v1.29.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-09-05)

* **Feature**: This release adds support to provide recommendations for G4dn and P3 instances that use NVIDIA GPUs.

# v1.26.0 (2023-08-28)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate licensing optimization recommendations for sql server running on EC2 instances.

# v1.25.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-08-01)

* No change notes available for this release.

# v1.25.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.4 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.3 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2023-06-15)

* No change notes available for this release.

# v1.24.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-05-18)

* **Feature**: In this launch, we add support for showing integration status with external metric providers such as Instana, Datadog ...etc in GetEC2InstanceRecommendations and ExportEC2InstanceRecommendations apis

# v1.23.1 (2023-05-04)

* No change notes available for this release.

# v1.23.0 (2023-05-01)

* **Feature**: support for tag filtering within compute optimizer. ability to filter recommendation results by tag and tag key value pairs. ability to filter by inferred workload type added.

# v1.22.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2023-04-10)

* No change notes available for this release.

# v1.22.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2023-03-30)

* **Feature**: This release adds support for HDD EBS volume types and io2 Block Express. We are also adding support for 61 new instance types and instances that have non consecutive runtime.

# v1.21.5 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.21.2 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.21.0 (2023-02-06)

* **Feature**: AWS Compute optimizer can now infer if Kafka is running on an instance.

# v1.20.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.19.0 (2022-12-22)

* **Feature**: This release enables AWS Compute Optimizer to analyze and generate optimization recommendations for ecs services running on Fargate.

# v1.18.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-11-29)

* **Feature**: Adds support for a new recommendation preference that makes it possible for customers to optimize their EC2 recommendations by utilizing an external metrics ingestion service to provide metrics.

# v1.17.21 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.20 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.19 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.18 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.17 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.16 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.15 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.14 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.13 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.12 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.11 (2022-08-04)

* No change notes available for this release.

# v1.17.10 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.9 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.8 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.7 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.6 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.5 (2022-05-10)

* **Documentation**: Documentation updates for Compute Optimizer

# v1.17.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-30)

* **Feature**: API client updated

# v1.12.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.
* **Feature**: Updated service to latest API model.

# v1.11.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.2 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-09-02)

* **Feature**: API client updated

# v1.8.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-05-25)

* **Feature**: API client updated

# v1.4.0 (2021-05-20)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

