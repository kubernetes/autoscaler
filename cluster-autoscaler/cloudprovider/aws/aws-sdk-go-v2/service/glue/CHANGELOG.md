# v1.113.3 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.113.2 (2025-06-11)

* No change notes available for this release.

# v1.113.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.113.0 (2025-05-22)

* **Feature**: This release supports additional ConversionSpec parameter as part of IntegrationPartition Structure in CreateIntegrationTableProperty API. This parameter is referred to apply appropriate column transformation for columns that are used for timestamp based partitioning

# v1.112.0 (2025-05-20)

* **Feature**: Enhanced AWS Glue ListConnectionTypes API Model with additional metadata fields.

# v1.111.0 (2025-05-16)

* **Feature**: Changes include (1) Excel as S3 Source type and XML and Tableau's Hyper as S3 Sink types, (2) targeted number of partitions parameter in S3 sinks and (3) new compression types in CSV/JSON and Parquet S3 sinks.

# v1.110.0 (2025-05-08)

* **Feature**: This new release supports customizable RefreshInterval for all Saas ZETL integrations from 15 minutes to 6 days.

# v1.109.2 (2025-04-23)

* No change notes available for this release.

# v1.109.1 (2025-04-10)

* No change notes available for this release.

# v1.109.0 (2025-04-09)

* **Feature**: The TableOptimizer APIs in AWS Glue now return the DpuHours field in each TableOptimizerRun, providing clients visibility to the DPU-hours used for billing in managed Apache Iceberg table compaction optimization.

# v1.108.0 (2025-04-07)

* **Feature**: Add input validations for multiple Glue APIs

# v1.107.1 (2025-04-03)

* No change notes available for this release.

# v1.107.0 (2025-03-14)

* **Feature**: This release added AllowFullTableExternalDataAccess to glue catalog resource.

# v1.106.2 (2025-03-13)

* No change notes available for this release.

# v1.106.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.106.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.10 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.9 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.8 (2025-02-04)

* No change notes available for this release.

# v1.105.7 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.6 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.5 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.105.4 (2025-01-22)

* **Documentation**: Docs Update for timeout changes

# v1.105.3 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.105.2 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.105.0 (2024-12-23)

* **Feature**: Add IncludeRoot parameters to GetCatalogs API to return root catalog.

# v1.104.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.104.0 (2024-12-12)

* **Feature**: To support customer-managed encryption in Data Quality to allow customers encrypt data with their own KMS key, we will add a DataQualityEncryption field to the SecurityConfiguration API where customers can provide their KMS keys.

# v1.103.0 (2024-12-03.2)

* **Feature**: This release includes(1)Zero-ETL integration to ingest data from 3P SaaS and DynamoDB to Redshift/Redlake (2)new properties on Connections to enable reuse; new connection APIs for retrieve/preview metadata (3)support of CRUD operations for Multi-catalog (4)support of automatic statistics collections

# v1.102.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.102.0 (2024-11-19)

* **Feature**: AWS Glue Data Catalog now enhances managed table optimizations of Apache Iceberg tables that can be accessed only from a specific Amazon Virtual Private Cloud (VPC) environment.

# v1.101.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.101.3 (2024-11-13)

* No change notes available for this release.

# v1.101.2 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.101.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.101.0 (2024-10-31)

* **Feature**: Add schedule support for AWS Glue column statistics

# v1.100.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.100.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.100.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.100.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.99.3 (2024-10-03)

* No change notes available for this release.

# v1.99.2 (2024-09-27)

* No change notes available for this release.

# v1.99.1 (2024-09-25)

* No change notes available for this release.

# v1.99.0 (2024-09-23)

* **Feature**: Added AthenaProperties parameter to Glue Connections, allowing Athena to store service specific properties on Glue Connections.

# v1.98.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.97.0 (2024-09-19)

* **Feature**: This change is for releasing TestConnection api SDK model

# v1.96.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.96.0 (2024-09-12)

* **Feature**: AWS Glue is introducing two new optimizers for Apache Iceberg tables: snapshot retention and orphan file deletion. Customers can enable these optimizers and customize their configurations to perform daily maintenance tasks on their Iceberg tables based on their specific requirements.

# v1.95.2 (2024-09-04)

* No change notes available for this release.

# v1.95.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.95.0 (2024-08-21)

* **Feature**: Add optional field JobRunQueuingEnabled to CreateJob and UpdateJob APIs.

# v1.94.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.94.0 (2024-08-13)

* **Feature**: Add AttributesToGet parameter support for Glue GetTables

# v1.93.0 (2024-08-08)

* **Feature**: This release adds support to retrieve the validation status when creating or updating Glue Data Catalog Views. Also added is support for BasicCatalogTarget partition keys.

# v1.92.0 (2024-08-07)

* **Feature**: Introducing AWS Glue Data Quality anomaly detection, a new functionality that uses ML-based solutions to detect data anomalies users have not explicitly defined rules for.

# v1.91.0 (2024-07-10.2)

* **Feature**: Add recipe step support for recipe node
* **Dependency Update**: Updated to the latest SDK module versions

# v1.90.0 (2024-07-10)

* **Feature**: Add recipe step support for recipe node
* **Dependency Update**: Updated to the latest SDK module versions

# v1.89.0 (2024-06-28)

* **Feature**: Added AttributesToGet parameter to Glue GetDatabases, allowing caller to limit output to include only the database name.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.88.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.87.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.87.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.86.0 (2024-06-17)

* **Feature**: This release introduces a new feature, Usage profiles. Usage profiles allow the AWS Glue admin to create different profiles for various classes of users within the account, enforcing limits and defaults for jobs and sessions.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.85.0 (2024-06-13)

* **Feature**: This release adds support for configuration of evaluation method for composite rules in Glue Data Quality rulesets.

# v1.84.1 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.84.0 (2024-06-06)

* **Feature**: This release adds support for creating and updating Glue Data Catalog Views.

# v1.83.0 (2024-06-05)

* **Feature**: AWS Glue now supports native SaaS connectivity: Salesforce connector available now

# v1.82.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.82.0 (2024-05-29)

* **Feature**: Add optional field JobMode to CreateJob and UpdateJob APIs.

# v1.81.1 (2024-05-23)

* No change notes available for this release.

# v1.81.0 (2024-05-21)

* **Feature**: Add Maintenance window to CreateJob and UpdateJob APIs and JobRun response. Add a new Job Run State for EXPIRED.

# v1.80.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.80.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.80.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.80.0 (2024-04-19)

* **Feature**: Adding RowFilter in the response for GetUnfilteredTableMetadata API

# v1.79.0 (2024-04-12)

* **Feature**: Modifying request for GetUnfilteredTableMetadata for view-related fields.

# v1.78.0 (2024-04-02)

* **Feature**: Adding View related fields to responses of read-only Table APIs.

# v1.77.5 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.77.4 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.77.3 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.77.2 (2024-02-29)

* No change notes available for this release.

# v1.77.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.77.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.76.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.76.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.76.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.76.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.75.0 (2024-02-05)

* **Feature**: Introduce Catalog Encryption Role within Glue Data Catalog Settings. Introduce SASL/PLAIN as an authentication method for Glue Kafka connections

# v1.74.0 (2024-01-31)

* **Feature**: Update page size limits for GetJobRuns and GetTriggers APIs.

# v1.73.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.73.0 (2023-12-22)

* **Feature**: This release adds additional configurations for Query Session Context on the following APIs: GetUnfilteredTableMetadata, GetUnfilteredPartitionMetadata, GetUnfilteredPartitionsMetadata.

# v1.72.4 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.72.3 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.72.2 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.72.1 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.72.0 (2023-11-30.2)

* **Feature**: Adds observation and analyzer support to the GetDataQualityResult and BatchGetDataQualityResult APIs.

# v1.71.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.71.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.70.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.70.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.70.0 (2023-11-27.2)

* **Feature**: add observations support to DQ CodeGen config model + update document for connectiontypes supported by ConnectorData entities

# v1.69.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.69.0 (2023-11-16)

* **Feature**: Introduces new column statistics APIs to support statistics generation for tables within the Glue Data Catalog.

# v1.68.1 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.68.0 (2023-11-14)

* **Feature**: Introduces new storage optimization APIs to support automatic compaction of Apache Iceberg tables.

# v1.67.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.67.0 (2023-11-02)

* **Feature**: This release introduces Google BigQuery Source and Target in AWS Glue CodeGenConfigurationNode.

# v1.66.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.65.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.64.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.63.0 (2023-10-12)

* **Feature**: Extending version control support to GitLab and Bitbucket from AWSGlue
* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.0 (2023-08-24)

* **Feature**: Added API attributes that help in the monitoring of sessions.

# v1.61.3 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.2 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.1 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.0 (2023-08-15)

* **Feature**: AWS Glue Crawlers can now accept SerDe overrides from a custom csv classifier. The two SerDe options are LazySimpleSerDe and OpenCSVSerDe. In case, the user wants crawler to do the selection, "None" can be selected for this purpose.

# v1.60.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.60.0 (2023-08-02)

* **Feature**: This release includes additional Glue Streaming KAKFA SASL property types.

# v1.59.1 (2023-08-01)

* No change notes available for this release.

# v1.59.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.58.2 (2023-07-28.2)

* No change notes available for this release.

# v1.58.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.58.0 (2023-07-26)

* **Feature**: Release Glue Studio Snowflake Connector Node for SDK/CLI

# v1.57.0 (2023-07-24)

* **Feature**: Added support for Data Preparation Recipe node in Glue Studio jobs

# v1.56.0 (2023-07-21)

* **Feature**: This release adds support for AWS Glue Crawler with Apache Hudi Tables, allowing Crawlers to discover Hudi Tables in S3 and register them in Glue Data Catalog for query engines to query against.

# v1.55.0 (2023-07-17)

* **Feature**: Adding new supported permission type flags to get-unfiltered endpoints that callers may pass to indicate support for enforcing Lake Formation fine-grained access control on nested column attributes.

# v1.54.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2023-07-07)

* **Feature**: This release enables customers to create new Apache Iceberg tables and associated metadata in Amazon S3 by using native AWS Glue CreateTable operation.

# v1.53.0 (2023-06-29)

* **Feature**: This release adds support for AWS Glue Crawler with Iceberg Tables, allowing Crawlers to discover Iceberg Tables in S3 and register them in Glue Data Catalog for query engines to query against.

# v1.52.0 (2023-06-26)

* **Feature**: Timestamp Starting Position For Kinesis and Kafka Data Sources in a Glue Streaming Job

# v1.51.0 (2023-06-19)

* **Feature**: This release adds support for creating cross region table/database resource links

# v1.50.2 (2023-06-15)

* No change notes available for this release.

# v1.50.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2023-05-30)

* **Feature**: Added Runtime parameter to allow selection of Ray Runtime

# v1.49.0 (2023-05-25)

* **Feature**: Added ability to create data quality rulesets for shared, cross-account Glue Data Catalog tables. Added support for dataset comparison rules through a new parameter called AdditionalDataSources. Enhanced the data quality results with a map containing profiled metric values.

# v1.48.0 (2023-05-16)

* **Feature**: Add Support for Tags for Custom Entity Types

# v1.47.0 (2023-05-09)

* **Feature**: This release adds AmazonRedshift Source and Target nodes in addition to DynamicTransform OutputSchemas

# v1.46.0 (2023-05-08)

* **Feature**: Support large worker types G.4x and G.8x for Glue Spark

# v1.45.5 (2023-05-04)

* No change notes available for this release.

# v1.45.4 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.3 (2023-04-10)

* No change notes available for this release.

# v1.45.2 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.1 (2023-04-06)

* No change notes available for this release.

# v1.45.0 (2023-04-03)

* **Feature**: Add support for database-level federation

# v1.44.0 (2023-03-30)

* **Feature**: This release adds support for AWS Glue Data Quality, which helps you evaluate and monitor the quality of your data and includes the API for creating, deleting, or updating data quality rulesets, runs and evaluations.

# v1.43.4 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.3 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.43.1 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2023-02-17)

* **Feature**: Release of Delta Lake Data Lake Format for Glue Studio Service

# v1.42.0 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Feature**: Fix DirectJDBCSource not showing up in CLI code gen
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.41.0 (2023-02-08)

* **Feature**: DirectJDBCSource + Glue 4.0 streaming options

# v1.40.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2023-01-31)

* No change notes available for this release.

# v1.40.0 (2023-01-19)

* **Feature**: Release Glue Studio Hudi Data Lake Format for SDK/CLI

# v1.39.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.38.1 (2022-12-19)

* No change notes available for this release.

# v1.38.0 (2022-12-15)

* **Feature**: This release adds support for AWS Glue Crawler with native DeltaLake tables, allowing Crawlers to classify Delta Lake format tables and catalog them for query engines to query against.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2022-11-30)

* **Feature**: This release adds support for AWS Glue Data Quality, which helps you evaluate and monitor the quality of your data and includes the API for creating, deleting, or updating data quality rulesets, runs and evaluations.

# v1.36.0 (2022-11-29)

* **Feature**: This release allows the creation of Custom Visual Transforms (Dynamic Transforms) to be created via AWS Glue CLI/SDK.

# v1.35.0 (2022-11-18)

* **Feature**: AWSGlue Crawler - Adding support for Table and Column level Comments with database level datatypes for JDBC based crawler.

# v1.34.1 (2022-11-11)

* **Documentation**: Added links related to enabling job bookmarks.

# v1.34.0 (2022-10-27)

* **Feature**: Added support for custom datatypes when using custom csv classifier.

# v1.33.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2022-10-05)

* **Feature**: This SDK release adds support to sync glue jobs with source control provider. Additionally, a new parameter called SourceControlDetails will be added to Job model.

# v1.32.0 (2022-09-22)

* **Feature**: Added support for S3 Event Notifications for Catalog Target Crawlers.

# v1.31.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.4 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.3 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.2 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2022-08-25)

* No change notes available for this release.

# v1.30.0 (2022-08-11)

* **Feature**: Add support for Python 3.9 AWS Glue Python Shell jobs
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.1 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2022-08-08)

* **Feature**: Add an option to run non-urgent or non-time sensitive Glue Jobs on spare capacity
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.2 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.1 (2022-07-19)

* **Documentation**: Documentation updates for AWS Glue Job Timeout and Autoscaling

# v1.28.0 (2022-07-14)

* **Feature**: This release adds an additional worker type for Glue Streaming jobs.

# v1.27.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2022-06-30)

* **Feature**: This release adds tag as an input of CreateDatabase

# v1.26.1 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-05-17)

* **Feature**: This release adds a new optional parameter called codeGenNodeConfiguration to CRUD job APIs that allows users to manage visual jobs via APIs. The updated CreateJob and UpdateJob will create jobs that can be viewed in Glue Studio as a visual graph. GetJob can be used to get codeGenNodeConfiguration.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2022-04-26)

* **Documentation**: This release adds documentation for the APIs to create, read, delete, list, and batch read of AWS Glue custom patterns, and for Lake Formation configuration settings in the AWS Glue crawler.

# v1.24.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-04-21)

* **Feature**: This release adds APIs to create, read, delete, list, and batch read of Glue custom entity types

# v1.23.0 (2022-04-14)

* **Feature**: Auto Scaling for Glue version 3.0 and later jobs to dynamically scale compute resources. This SDK change provides customers with the auto-scaled DPU usage

# v1.22.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-03-18)

* **Feature**: Added 9 new APIs for AWS Glue Interactive Sessions: ListSessions, StopSession, CreateSession, GetSession, DeleteSession, RunStatement, GetStatement, ListStatements, CancelStatement

# v1.21.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-01-07)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.16.0 (2021-12-02)

* **Feature**: API client updated
* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.14.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-10-21)

* **Feature**: API client updated
* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-07-15)

* **Feature**: Updated service model to latest version.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-07-01)

* **Feature**: API client updated

# v1.7.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-11)

* **Feature**: Updated to latest API model.

# v1.5.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

