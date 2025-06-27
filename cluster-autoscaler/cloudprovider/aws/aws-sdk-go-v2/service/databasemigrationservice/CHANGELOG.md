# v1.53.0 (2025-06-17)

* **Feature**: Add "Virtual" field to Data Provider as well as "S3Path" and "S3AccessRoleArn" fields to DataProvider settings
* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.2 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2025-06-06)

* No change notes available for this release.

# v1.52.0 (2025-05-15)

* **Feature**: Introduces Data Resync feature to describe-table-statistics and IAM database authentication for MariaDB, MySQL, and PostgreSQL.

# v1.51.3 (2025-04-16)

* No change notes available for this release.

# v1.51.2 (2025-04-03)

* No change notes available for this release.

# v1.51.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.51.0 (2025-02-28)

* **Feature**: Add skipped status to the Result Statistics of an Assessment Run

# v1.50.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2025-02-17)

* **Feature**: Support replicationConfigArn in DMS DescribeApplicableIndividualAssessments API.

# v1.48.0 (2025-02-14)

* **Feature**: Introduces premigration assessment feature to DMS Serverless API for start-replication and describe-replications

# v1.47.0 (2025-02-10)

* **Feature**: New vendors for DMS Data Providers: DB2 LUW and DB2 for z/OS

# v1.46.1 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2025-02-04)

* **Feature**: Introduces TargetDataSettings with the TablePreparationMode option available for data migrations.

# v1.45.9 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.8 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.7 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.45.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.45.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.4 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.45.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.2 (2025-01-08)

* No change notes available for this release.

# v1.45.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-12-12)

* **Feature**: Add parameters to support for kerberos authentication. Add parameter for disabling the Unicode source filter with PostgreSQL settings. Add parameter to use large integer value with Kinesis/Kafka settings.

# v1.44.5 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.44.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-10-21)

* **Feature**: Added support for tagging in StartReplicationTaskAssessmentRun API and introduced IsLatestTaskAssessmentRun and ResultStatistic fields for enhanced tracking and assessment result statistics.

# v1.43.0 (2024-10-10)

* **Feature**: Introduces DescribeDataMigrations, CreateDataMigration, ModifyDataMigration, DeleteDataMigration, StartDataMigration, StopDataMigration operations to SDK. Provides FailedDependencyFault error message.

# v1.42.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.4 (2024-10-03)

* No change notes available for this release.

# v1.41.3 (2024-09-27)

* No change notes available for this release.

# v1.41.2 (2024-09-25)

* No change notes available for this release.

# v1.41.1 (2024-09-23)

* No change notes available for this release.

# v1.41.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.8 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.40.7 (2024-09-04)

* No change notes available for this release.

# v1.40.6 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.5 (2024-08-22)

* No change notes available for this release.

# v1.40.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.38.11 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.10 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.9 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.8 (2024-05-23)

* No change notes available for this release.

# v1.38.7 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.6 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.5 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.38.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.37.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.37.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.36.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.36.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.6 (2023-12-20)

* No change notes available for this release.

# v1.35.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.35.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.35.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.4 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.3 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.34.2 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.1 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-11-13)

* **Feature**: Added new Db2 LUW Target endpoint with related endpoint settings. New executeTimeout endpoint setting for mysql endpoint. New ReplicationDeprovisionTime field for serverless describe-replications.

# v1.33.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-09-22)

* **Feature**: new vendors for DMS CSF: MongoDB, MariaDB, DocumentDb and Redshift

# v1.30.4 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.3 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.2 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-08-03)

* **Feature**: The release makes public API for DMS Schema Conversion feature.

# v1.29.0 (2023-08-01)

* **Feature**: Adding new API describe-engine-versions which provides information about the lifecycle of a replication instance's version.

# v1.28.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-07-13)

* **Feature**: Enhanced PostgreSQL target endpoint settings for providing Babelfish support.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-07-07)

* **Feature**: Releasing DMS Serverless. Adding support for PostgreSQL 15.x as source and target endpoint. Adding support for DocDB Elastic Clusters with sharded collections, PostgreSQL datatype mapping customization and disabling hostname validation of the certificate authority in Kafka endpoint settings

# v1.25.7 (2023-06-15)

* No change notes available for this release.

# v1.25.6 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.5 (2023-05-04)

* No change notes available for this release.

# v1.25.4 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.3 (2023-04-10)

* No change notes available for this release.

# v1.25.2 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-03-17)

* **Feature**: S3 setting to create AWS Glue Data Catalog. Oracle setting to control conversion of timestamp column. Support for Kafka SASL Plain authentication. Setting to map boolean from PostgreSQL to Redshift. SQL Server settings to force lob lookup on inline LOBs and to control access of database logs.

# v1.24.1 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-03-07)

* **Feature**: This release adds DMS Fleet Advisor Target Recommendation APIs and exposes functionality for DMS Fleet Advisor. It adds functionality to start Target Recommendation calculation.

# v1.23.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.23.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.23.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2023-01-23)

* No change notes available for this release.

# v1.23.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.22.3 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2022-11-22)

* No change notes available for this release.

# v1.22.0 (2022-11-17)

* **Feature**: Adds support for Internet Protocol Version 6 (IPv6) on DMS Replication Instances

# v1.21.16 (2022-11-16)

* No change notes available for this release.

# v1.21.15 (2022-11-10)

* No change notes available for this release.

# v1.21.14 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.13 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.12 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.11 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.10 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.9 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.8 (2022-08-30)

* No change notes available for this release.

# v1.21.7 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.6 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2022-08-04)

* **Documentation**: Documentation updates for Database Migration Service (DMS).

# v1.21.2 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2022-07-21)

* **Documentation**: Documentation updates for Database Migration Service (DMS).

# v1.21.0 (2022-07-07)

* **Feature**: New api to migrate event subscriptions to event bridge rules

# v1.20.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-07-01)

* **Feature**: Added new features for AWS DMS version 3.4.7 that includes new endpoint settings for S3, OpenSearch, Postgres, SQLServer and Oracle.

# v1.19.1 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-06-08)

* **Feature**: This release adds DMS Fleet Advisor APIs and exposes functionality for DMS Fleet Advisor. It adds functionality to create and modify fleet advisor instances, and to collect and analyze information about the local data infrastructure.

# v1.18.6 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Documentation**: Updated API models
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: Updated to latest service endpoints

# v1.13.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-30)

* **Feature**: API client updated

# v1.12.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.
* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.10.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service to latest API model.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-09-24)

* **Feature**: API client updated

# v1.7.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-07-15)

* **Feature**: Updated service model to latest version.
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

