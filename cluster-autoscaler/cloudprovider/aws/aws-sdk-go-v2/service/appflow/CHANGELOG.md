# v1.46.4 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.3 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.2 (2025-04-03)

* No change notes available for this release.

# v1.46.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.46.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.17 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.16 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.15 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.14 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.13 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.45.12 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.45.11 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.10 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.9 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.8 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.7 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.6 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.45.5 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.4 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.3 (2024-10-11)

* **Documentation**: Doc only updates for clarification around OAuth2GrantType for Salesforce.

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

# v1.43.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.43.6 (2024-09-04)

* No change notes available for this release.

# v1.43.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.42.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.11 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.10 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.9 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.8 (2024-05-23)

* No change notes available for this release.

# v1.41.7 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.6 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.5 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.41.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.40.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.40.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.38.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.38.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.35.3 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.2 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2023-09-27)

* No change notes available for this release.

# v1.35.0 (2023-09-06)

* **Feature**: Adding OAuth2.0 support for servicenow connector.

# v1.34.0 (2023-08-30)

* **Feature**: Add SAP source connector parallel and pagination feature

# v1.33.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2023-08-01)

* No change notes available for this release.

# v1.33.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-06-27)

* **Feature**: This release adds support to bypass SSO with the SAPOData connector when connecting to an SAP instance.

# v1.31.0 (2023-06-20)

* **Feature**: This release adds new API to reset connector metadata cache

# v1.30.2 (2023-06-15)

* No change notes available for this release.

# v1.30.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-06-01)

* **Feature**: Added ability to select DataTransferApiType for DescribeConnector and CreateFlow requests when using Async supported connectors. Added supportedDataTransferType to DescribeConnector/DescribeConnectors/ListConnector response.

# v1.29.1 (2023-05-04)

* No change notes available for this release.

# v1.29.0 (2023-05-02)

* **Feature**: This release adds new API to cancel flow executions.

# v1.28.0 (2023-04-28)

* **Feature**: Adds Jwt Support for Salesforce Credentials.

# v1.27.0 (2023-04-24)

* **Feature**: Increased the max length for RefreshToken and AuthCode from 2048 to 4096.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-04-17)

* **Feature**: This release adds a Client Token parameter to the following AppFlow APIs: Create/Update Connector Profile, Create/Update Flow, Start Flow, Register Connector, Update Connector Registration. The Client Token parameter allows idempotent operations for these APIs.

# v1.25.4 (2023-04-10)

* No change notes available for this release.

# v1.25.3 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-02-23)

* **Feature**: This release enables the customers to choose whether to use Private Link for Metadata and Authorization call when using a private Salesforce connections

# v1.24.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.24.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.24.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-01-19)

* **Feature**: Adding support for Salesforce Pardot connector in Amazon AppFlow.

# v1.23.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.22.0 (2022-12-16)

* **Feature**: This release updates the ListConnectorEntities API action so that it returns paginated responses that customers can retrieve with next tokens.

# v1.21.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-11-22)

* **Feature**: Adding support for Amazon AppFlow to transfer the data to Amazon Redshift databases through Amazon Redshift Data API service. This feature will support the Redshift destination connector on both public and private accessible Amazon Redshift Clusters and Amazon Redshift Serverless.

# v1.20.0 (2022-11-18)

* **Feature**: AppFlow provides a new API called UpdateConnectorRegistration to update a custom connector that customers have previously registered. With this API, customers no longer need to unregister and then register a connector to make an update.

# v1.19.0 (2022-11-17)

* **Feature**: AppFlow simplifies the preparation and cataloging of SaaS data into the AWS Glue Data Catalog where your data can be discovered and accessed by AWS analytics and ML services. AppFlow now also supports data field partitioning and file size optimization to improve query performance and reduce cost.

# v1.18.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-10-13)

* **Feature**: With this update, you can choose which Salesforce API is used by Amazon AppFlow to transfer data to or from your Salesforce account. You can choose the Salesforce REST API or Bulk API 2.0. You can also choose for Amazon AppFlow to pick the API automatically.

# v1.17.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.10 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.9 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.8 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.7 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.6 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.5 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.4 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.3 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.2 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-05-27)

* **Feature**: Adding the following features/changes: Parquet output that preserves typing from the source connector, Failed executions threshold before deactivation for scheduled flows, increasing max size of access and refresh token from 2048 to 4096

# v1.15.2 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-04-14)

* **Feature**: Enables users to pass custom token URL parameters for Oauth2 authentication during create connector profile

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

# v1.9.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-10-21)

* **Feature**: API client updated
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

# v1.6.0 (2021-08-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.2 (2021-08-04)

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

