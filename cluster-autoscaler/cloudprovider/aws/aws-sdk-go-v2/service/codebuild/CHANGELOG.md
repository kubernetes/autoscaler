# v1.61.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.0 (2025-05-15)

* **Feature**: AWS CodeBuild now supports Docker Server capability

# v1.60.0 (2025-04-23)

* **Feature**: Add support for custom instance type for reserved capacity fleets

# v1.59.0 (2025-04-07)

* **Feature**: AWS CodeBuild now offers an enhanced debugging experience.

# v1.58.1 (2025-04-03)

* No change notes available for this release.

# v1.58.0 (2025-04-02)

* **Feature**: This release adds support for environment type WINDOWS_SERVER_2022_CONTAINER in ProjectEnvironment

# v1.57.0 (2025-03-28)

* **Feature**: This release adds support for cacheNamespace in ProjectCache

# v1.56.0 (2025-03-13)

* **Feature**: AWS CodeBuild now supports webhook filtering by organization name

# v1.55.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.55.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2025-02-25)

* **Feature**: Adding "reportArns" field in output of BatchGetBuildBatches API. "reportArns" is an array that contains the ARNs of reports created by merging reports from builds associated with the batch build.

# v1.53.0 (2025-02-20)

* **Feature**: Add webhook status and status message to AWS CodeBuild webhooks

# v1.52.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2025-02-14)

* **Feature**: Added test suite names to test case metadata

# v1.51.3 (2025-02-12)

* **Documentation**: Add note for the RUNNER_BUILDKITE_BUILD buildType.

# v1.51.2 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.1 (2025-02-04)

* No change notes available for this release.

# v1.51.0 (2025-01-31)

* **Feature**: Added support for CodeBuild self-hosted Buildkite runner builds
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.4 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.3 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.50.2 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.50.1 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2025-01-09)

* **Feature**: AWS CodeBuild Now Supports BuildBatch in Reserved Capacity and Lambda
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.4 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.3 (2024-12-13)

* No change notes available for this release.

# v1.49.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-11-12)

* **Feature**: AWS CodeBuild now supports non-containerized Linux and Windows builds on Reserved Capacity.

# v1.48.1 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.48.0 (2024-11-06)

* **Feature**: AWS CodeBuild now adds additional compute types for reserved capacity fleet.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2024-10-25)

* **Feature**: AWS CodeBuild now supports automatically retrying failed builds

# v1.46.0 (2024-10-15)

* **Feature**: Enable proxy for reserved capacity fleet.

# v1.45.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.4 (2024-10-03)

* No change notes available for this release.

# v1.44.3 (2024-09-27)

* No change notes available for this release.

# v1.44.2 (2024-09-25)

* No change notes available for this release.

# v1.44.1 (2024-09-23)

* No change notes available for this release.

# v1.44.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-09-17)

* **Feature**: GitLab Enhancements - Add support for Self-Hosted GitLab runners in CodeBuild. Add group webhooks
* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.42.3 (2024-09-04)

* No change notes available for this release.

# v1.42.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2024-08-23)

* **Documentation**: Added support for the MAC_ARM environment type for CodeBuild fleets.

# v1.42.0 (2024-08-19)

* **Feature**: AWS CodeBuild now supports creating fleets with macOS platform for running builds.

# v1.41.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-08-14)

* **Feature**: AWS CodeBuild now supports using Secrets Manager to store git credentials and using multiple source credentials in a single project.

# v1.40.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.39.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-06-17)

* **Feature**: AWS CodeBuild now supports global and organization GitHub webhooks
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.3 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-05-31)

* **Documentation**: AWS CodeBuild now supports Self-hosted GitHub Actions runners for Github Enterprise

# v1.37.0 (2024-05-29)

* **Feature**: AWS CodeBuild now supports manually creating GitHub webhooks

# v1.36.1 (2024-05-23)

* No change notes available for this release.

# v1.36.0 (2024-05-17)

* **Feature**: Aws CodeBuild now supports 36 hours build timeout

# v1.35.1 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2024-05-15)

* **Feature**: CodeBuild Reserved Capacity VPC Support
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.34.1 (2024-04-11)

* **Documentation**: Support access tokens for Bitbucket sources

# v1.34.0 (2024-04-09)

* **Feature**: Add new webhook filter types for GitHub webhooks

# v1.33.0 (2024-03-29)

* **Feature**: Add new fleet status code for Reserved Capacity.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2024-03-25)

* **Feature**: Supporting GitLab and GitLab Self Managed as source types in AWS CodeBuild.

# v1.31.2 (2024-03-20)

* **Documentation**: This release adds support for new webhook events (RELEASED and PRERELEASED) and filter types (TAG_NAME and RELEASE_NAME).

# v1.31.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2024-03-15)

* **Feature**: AWS CodeBuild now supports overflow behavior on Reserved Capacity.

# v1.30.3 (2024-03-08)

* **Documentation**: This release adds support for a new webhook event: PULL_REQUEST_CLOSED.

# v1.30.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.29.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.29.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.29.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2024-01-19)

* **Feature**: Release CodeBuild Reserved Capacity feature

# v1.27.0 (2024-01-08)

* **Feature**: Aws CodeBuild now supports new compute type BUILD_GENERAL1_XLARGE

# v1.26.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.26.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.26.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.25.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-11-06)

* **Feature**: AWS CodeBuild now supports AWS Lambda compute.

# v1.24.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2023-09-18)

* **Announcement**: [BREAKFIX] Change in MaxResults datatype from value to pointer type in cognito-sync service.
* **Feature**: Adds several endpoint ruleset changes across all models: smaller rulesets, removed non-unique regional endpoints, fixes FIPS and DualStack endpoints, and make region not required in SDK::Endpoint. Additional breakfix to cognito-sync field.

# v1.21.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2023-08-01)

* No change notes available for this release.

# v1.21.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.17 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.16 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.15 (2023-06-15)

* No change notes available for this release.

# v1.20.14 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.13 (2023-05-25)

* No change notes available for this release.

# v1.20.12 (2023-05-04)

* No change notes available for this release.

# v1.20.11 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.10 (2023-04-10)

* No change notes available for this release.

# v1.20.9 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.8 (2023-04-05)

* No change notes available for this release.

# v1.20.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.20.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.20.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2023-01-20)

* No change notes available for this release.

# v1.20.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.19.21 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.20 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.19 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.18 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.17 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.16 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.15 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.14 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.13 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.12 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.11 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.10 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.9 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.8 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.7 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.6 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.5 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.14.2 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.13.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-09-02)

* **Feature**: API client updated

# v1.9.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-08-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-08-12)

* **Feature**: API client updated

# v1.6.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

