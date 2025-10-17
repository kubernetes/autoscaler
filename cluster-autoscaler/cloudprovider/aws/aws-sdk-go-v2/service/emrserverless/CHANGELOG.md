# v1.31.1 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2025-06-12)

* **Feature**: This release adds support for retrieval of the optional executionIamPolicy field in the GetJobRun API response.

# v1.30.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2025-06-03)

* **Feature**: AWS EMR Serverless: Adds a new option in the CancelJobRun API in EMR 7.9.0+, to cancel a job with grace period. This feature is enabled by default with a 120-second grace period for streaming jobs and is not enabled by default for batch jobs.

# v1.29.0 (2025-05-30)

* **Feature**: This release adds the capability for users to specify an optional Execution IAM policy in the StartJobRun action. The resulting permissions assumed by the job run is the intersection of the permissions in the Execution Role and the specified Execution IAM Policy.

# v1.28.4 (2025-04-10)

* No change notes available for this release.

# v1.28.3 (2025-04-09)

* No change notes available for this release.

# v1.28.2 (2025-04-03)

* No change notes available for this release.

# v1.28.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.28.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.9 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.8 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.7 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.6 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.5 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.27.4 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.27.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2024-12-11)

* **Feature**: This release adds support for accessing system profile logs in Lake Formation-enabled jobs.

# v1.26.7 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.26.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.3 (2024-10-03)

* No change notes available for this release.

# v1.25.2 (2024-09-27)

* No change notes available for this release.

# v1.25.1 (2024-09-25)

* No change notes available for this release.

# v1.25.0 (2024-09-23)

* **Feature**: This release adds support for job concurrency and queuing configuration at Application level.

# v1.24.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.23.6 (2024-09-04)

* No change notes available for this release.

# v1.23.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.22.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2024-05-30)

* **Feature**: The release adds support for spark structured streaming.

# v1.20.0 (2024-05-23)

* **Feature**: This release adds the capability to run interactive workloads using Apache Livy Endpoint.

# v1.19.4 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.2 (2024-05-09)

* No change notes available for this release.

# v1.19.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.19.0 (2024-04-18)

* **Feature**: This release adds the capability to publish detailed Spark engine metrics to Amazon Managed Service for Prometheus (AMP) for  enhanced monitoring for Spark jobs.

# v1.18.0 (2024-04-16)

* **Feature**: This release adds support for shuffle optimized disks that allow larger disk sizes and higher IOPS to efficiently run shuffle heavy workloads.

# v1.17.5 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.4 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.2 (2024-02-29)

* No change notes available for this release.

# v1.17.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.16.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.16.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2024-02-09)

* No change notes available for this release.

# v1.15.0 (2024-01-18)

* **Feature**: **BREAKFIX**: Correct nullability of InitialCapacityConfig's WorkerCount field. The type of this value has changed from int64 to *int64. Due to this field being marked required, with an enforced minimum of 1, but a default of 0, the former type would result in automatic failure behavior without caller intervention. Calling code will need to be updated accordingly.

# v1.14.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.6 (2023-12-13)

* No change notes available for this release.

# v1.14.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.14.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.14.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.6 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.5 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.13.4 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.3 (2023-11-16)

* No change notes available for this release.

# v1.13.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2023-09-25)

* **Feature**: This release adds support for application-wide default job configurations.

# v1.10.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2023-08-01)

* No change notes available for this release.

# v1.10.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2023-07-25)

* **Feature**: This release adds support for publishing application logs to CloudWatch.

# v1.8.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2023-06-27)

* **Feature**: This release adds support to update the release label of an EMR Serverless application to upgrade it to a different version of Amazon EMR via UpdateApplication API.

# v1.7.6 (2023-06-15)

* No change notes available for this release.

# v1.7.5 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.4 (2023-05-11)

* No change notes available for this release.

# v1.7.3 (2023-05-04)

* No change notes available for this release.

# v1.7.2 (2023-05-01)

* No change notes available for this release.

# v1.7.1 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2023-04-17)

* **Feature**: The GetJobRun API has been updated to include the job's billed resource utilization. This utilization shows the aggregate vCPU, memory and storage that AWS has billed for the job run. The billed resources include a 1-minute minimum usage for workers, plus additional storage over 20 GB per worker.

# v1.6.0 (2023-04-11)

* **Feature**: This release extends GetJobRun API to return job run timeout (executionTimeoutMinutes) specified during StartJobRun call (or default timeout of 720 minutes if none was specified).

# v1.5.8 (2023-04-10)

* No change notes available for this release.

# v1.5.7 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.6 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.5 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.5.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.5.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).
* **Feature**: Adds support for customized images. You can now provide runtime images when creating or updating EMR Serverless Applications.

# v1.4.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2022-11-17)

* **Feature**: Adds support for AWS Graviton2 based applications. You can now select CPU architecture when creating new applications or updating existing ones.

# v1.3.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2022-09-29)

* **Feature**: This release adds API support to debug Amazon EMR Serverless jobs in real-time with live application UIs

# v1.2.4 (2022-09-27)

* No change notes available for this release.

# v1.2.3 (2022-09-23)

* No change notes available for this release.

# v1.2.2 (2022-09-21)

* No change notes available for this release.

# v1.2.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.10 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.9 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.8 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.7 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.6 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.5 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.4 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.3 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.2 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2022-05-27)

* **Feature**: This release adds support for Amazon EMR Serverless, a serverless runtime environment that simplifies running analytics applications using the latest open source frameworks such as Apache Spark and Apache Hive.

# v1.0.0 (2022-05-26)

* **Release**: New AWS service client module
* **Feature**: This release adds support for Amazon EMR Serverless, a serverless runtime environment that simplifies running analytics applications using the latest open source frameworks such as Apache Spark and Apache Hive.

