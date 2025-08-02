# v1.51.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2025-06-02)

* **Feature**: Add support for the managed query result in the workgroup APIs. The managed query result configuration enables users to store query results to Athena owned storage.

# v1.50.5 (2025-05-09)

* **Documentation**: Minor API documentation updates

# v1.50.4 (2025-04-16)

* No change notes available for this release.

# v1.50.3 (2025-04-10)

* No change notes available for this release.

# v1.50.2 (2025-04-03)

* No change notes available for this release.

# v1.50.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.50.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.11 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.10 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.9 (2025-02-04)

* No change notes available for this release.

# v1.49.8 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.7 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.6 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.49.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.49.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2024-12-17)

* No change notes available for this release.

# v1.49.0 (2024-12-03.2)

* **Feature**: Add FEDERATED type to CreateDataCatalog. This creates Athena Data Catalog, AWS Lambda connector, and AWS Glue connection. Create/DeleteDataCatalog returns DataCatalog. Add Status, ConnectionType, and Error to DataCatalog and DataCatalogSummary. Add DeleteCatalogOnly to delete Athena Catalog only.

# v1.48.5 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.48.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2024-10-18)

* **Feature**: **BREAKING CHANGE**: Remove DataCatalog from Create/DeleteDataCatalog. Remove Status, ConnectionType, and Error from DataCatalog and DataCatalogSummary. These were released inadvertently with no functionality. They were not populated or populated with a default value. Code related to these changes can be safely removed.

# v1.47.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.3 (2024-10-03)

* No change notes available for this release.

# v1.46.2 (2024-09-27)

* No change notes available for this release.

# v1.46.1 (2024-09-25)

* No change notes available for this release.

# v1.46.0 (2024-09-23)

* **Feature**: List/Get/Update/Delete/CreateDataCatalog now integrate with AWS Glue connections. Users can create a Glue connection through Athena or use a Glue connection to define their Athena federated parameters.

# v1.45.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.8 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.44.7 (2024-09-04)

* No change notes available for this release.

# v1.44.6 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.5 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.4 (2024-08-09)

* No change notes available for this release.

# v1.44.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.43.0 (2024-06-19)

* **Feature**: Add v2 smoke tests and smithy smokeTests trait for SDK testing.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.3 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-05-29)

* **Feature**: Throwing validation errors on CreateNotebook with Name containing `/`,`:`,`\`

# v1.40.8 (2024-05-23)

* No change notes available for this release.

# v1.40.7 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.6 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.5 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.40.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.39.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.39.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.39.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-01-19)

* **Feature**: Introducing new NotebookS3LocationUri parameter to Athena ImportNotebook API. Payload is no longer required and either Payload or NotebookS3LocationUri needs to be provided (not both) for a successful ImportNotebook API call. If both are provided, an InvalidRequestException will be thrown.

# v1.37.4 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.3 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.37.2 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.37.0 (2023-12-05)

* **Feature**: Adding IdentityCenter enabled request for interactive query

# v1.36.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.3 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.2 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.35.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2023-11-17)

* **Feature**: Adding SerivicePreProcessing time metric

# v1.34.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.31.8 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.7 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.6 (2023-08-24)

* No change notes available for this release.

# v1.31.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2023-08-01)

* No change notes available for this release.

# v1.31.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.5 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.4 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.3 (2023-07-03)

* No change notes available for this release.

# v1.30.2 (2023-06-15)

* No change notes available for this release.

# v1.30.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-06-08)

* **Feature**: You can now define custom spark properties at start of the session for use cases like cluster encryption, table formats, and general Spark tuning.

# v1.29.0 (2023-06-02)

* **Feature**: This release introduces the DeleteCapacityReservation API and the ability to manage capacity reservations using CloudFormation

# v1.28.0 (2023-05-18)

* **Feature**: Removing SparkProperties from EngineConfiguration object for StartSession API call

# v1.27.0 (2023-05-15)

* **Feature**: You can now define custom spark properties at start of the session for use cases like cluster encryption, table formats, and general Spark tuning.

# v1.26.1 (2023-05-04)

* No change notes available for this release.

# v1.26.0 (2023-04-28)

* **Feature**: You can now use capacity reservations on Amazon Athena to run SQL queries on fully-managed compute capacity.

# v1.25.4 (2023-04-27)

* No change notes available for this release.

# v1.25.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-04-10)

* No change notes available for this release.

# v1.25.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-03-30)

* **Feature**: Make DefaultExecutorDpuSize and CoordinatorDpuSize  fields optional  in StartSession

# v1.24.0 (2023-03-27)

* **Feature**: Enforces a minimal level of encryption for the workgroup for query and calculation results that are written to Amazon S3. When enabled, workgroup users can set encryption only to the minimum level set by the administrator or higher when they submit queries.

# v1.23.2 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-03-08)

* **Feature**: A new field SubstatementType is added to GetQueryExecution API, so customers have an error free way to detect the query type and interpret the result.

# v1.22.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.22.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.22.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.21.0 (2022-12-19)

* **Feature**: Add missed InvalidRequestException in GetCalculationExecutionCode,StopCalculationExecution APIs. Correct required parameters (Payload and Type) in UpdateNotebook API. Change Notebook size from 15 Mb to 10 Mb.

# v1.20.3 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2022-12-08)

* No change notes available for this release.

# v1.20.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-11-30)

* **Feature**: This release includes support for using Apache Spark in Amazon Athena.

# v1.19.1 (2022-11-08)

* No change notes available for this release.

# v1.19.0 (2022-11-07)

* **Feature**: Adds support for using Query Result Reuse

# v1.18.12 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.11 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.10 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.9 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.8 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.7 (2022-09-01)

* No change notes available for this release.

# v1.18.6 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.3 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-07-21)

* **Feature**: This feature allows customers to retrieve runtime statistics for completed queries

# v1.17.0 (2022-07-14)

* **Feature**: This release updates data types that contain either QueryExecutionId, NamedQueryId or ExpectedBucketOwner. Ids must be between 1 and 128 characters and contain only non-whitespace characters. ExpectedBucketOwner must be 12-digit string.

# v1.16.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-06-30)

* **Feature**: This feature introduces the API support for Athena's parameterized query and BatchGetPreparedStatement API.

# v1.15.4 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.3 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.2 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-04-15)

* **Feature**: This release adds subfields, ErrorMessage, Retryable, to the AthenaError response object in the GetQueryExecution API when a query fails.

# v1.14.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service client model to latest release.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.9.2 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.8.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-08-12)

* **Feature**: API client updated

# v1.4.2 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

