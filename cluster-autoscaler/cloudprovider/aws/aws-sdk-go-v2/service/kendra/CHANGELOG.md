# v1.56.4 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.3 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.2 (2025-04-03)

* No change notes available for this release.

# v1.56.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.56.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.9 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.8 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.7 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.6 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.5 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.55.4 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.55.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.0 (2024-12-04)

* **Feature**: This release adds GenAI Index in Amazon Kendra for Retrieval Augmented Generation (RAG) and intelligent search. With the Kendra GenAI Index, customers get high retrieval accuracy powered by the latest information retrieval technologies and semantic models.

# v1.54.7 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.54.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.4 (2024-10-03)

* No change notes available for this release.

# v1.53.3 (2024-09-27)

* No change notes available for this release.

# v1.53.2 (2024-09-25)

* No change notes available for this release.

# v1.53.1 (2024-09-23)

* No change notes available for this release.

# v1.53.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.52.6 (2024-09-04)

* No change notes available for this release.

# v1.52.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.51.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.8 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.7 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.6 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.5 (2024-05-23)

* No change notes available for this release.

# v1.50.4 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.3 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.2 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.50.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2024-03-22)

* **Feature**: Documentation update, March 2024. Corrects some docs for Amazon Kendra.

# v1.49.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.48.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.48.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.48.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.47.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.47.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.46.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2023-10-18)

* **Feature**: Changes for a new feature in Amazon Kendra's Query API to Collapse/Expand query results

# v1.43.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2023-09-12)

* **Feature**: Amazon Kendra now supports confidence score buckets for retrieved passage results using the Retrieve API.

# v1.42.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2023-08-01)

* No change notes available for this release.

# v1.42.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2023-06-22)

* **Feature**: Introducing Amazon Kendra Retrieve API that can be used to retrieve relevant passages or text excerpts given an input query.

# v1.40.4 (2023-06-15)

* No change notes available for this release.

# v1.40.3 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.2 (2023-06-06)

* No change notes available for this release.

# v1.40.1 (2023-05-04)

* No change notes available for this release.

# v1.40.0 (2023-05-02)

* **Feature**: AWS Kendra now supports configuring document fields/attributes via the GetQuerySuggestions API. You can now base query suggestions on the contents of document fields.

# v1.39.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.2 (2023-04-10)

* No change notes available for this release.

# v1.39.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2023-03-30)

* **Feature**: AWS Kendra now supports featured results for a query.

# v1.38.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.38.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.38.2 (2023-02-08)

* No change notes available for this release.

# v1.38.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2023-01-11)

* **Feature**: This release adds support to new document types - RTF, XML, XSLT, MS_EXCEL, CSV, JSON, MD

# v1.37.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.36.3 (2022-12-30)

* No change notes available for this release.

# v1.36.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2022-11-28)

* **Feature**: Amazon Kendra now supports preview of table information from HTML tables in the search results. The most relevant cells with their corresponding rows, columns are displayed as a preview in the search result. The most relevant table cell or cells are also highlighted in table preview.

# v1.35.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2022-09-27)

* **Feature**: My AWS Service (placeholder) - Amazon Kendra now provides a data source connector for DropBox. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-dropbox.html

# v1.34.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2022-09-14)

* **Feature**: This release enables our customer to choose the option of Sharepoint 2019 for the on-premise Sharepoint connector.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.3 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2022-08-19)

* **Feature**: This release adds support for a new authentication type - Personal Access Token (PAT) for confluence server.

# v1.32.0 (2022-08-17)

* **Feature**: This release adds Zendesk connector (which allows you to specify Zendesk SAAS platform as data source), Proxy Support for Sharepoint and Confluence Server (which allows you to specify the proxy configuration if proxy is required to connect to your Sharepoint/Confluence Server as data source).

# v1.31.4 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2022-07-21)

* **Feature**: Amazon Kendra now provides Oauth2 support for SharePoint Online. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-sharepoint.html

# v1.30.0 (2022-07-14)

* **Feature**: This release adds AccessControlConfigurations which allow you to redefine your document level access control without the need for content re-indexing.

# v1.29.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2022-06-30)

* **Feature**: Amazon Kendra now provides a data source connector for alfresco

# v1.28.2 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2022-06-02)

* **Feature**: Amazon Kendra now provides a data source connector for GitHub. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-github.html

# v1.27.1 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2022-05-12)

* **Feature**: Amazon Kendra now provides a data source connector for Jira. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-jira.html

# v1.26.0 (2022-05-05)

* **Feature**: AWS Kendra now supports hierarchical facets for a query. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/filtering.html

# v1.25.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-04-19)

* **Feature**: Amazon Kendra now provides a data source connector for Quip. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-quip.html

# v1.24.0 (2022-04-06)

* **Feature**: Amazon Kendra now provides a data source connector for Box. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-box.html

# v1.23.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-03-14)

* **Feature**: Amazon Kendra now provides a data source connector for Slack. For more information, see https://docs.aws.amazon.com/kendra/latest/dg/data-source-slack.html

# v1.22.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service client model to latest release.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.17.0 (2021-12-02)

* **Feature**: API client updated
* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2021-11-19)

* **Announcement**: Fix API modeling bug incorrectly generating `DocumentAttributeValue` type as a union instead of a structure. This update corrects this bug by correcting the `DocumentAttributeValue` type to be a `struct` instead of an `interface`. This change also removes the `DocumentAttributeValueMember` types. To migrate to this change your application using service/kendra will need to be updated to use struct members in `DocumentAttributeValue` instead of `DocumentAttributeValueMember` types.
* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.14.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service to latest API model.
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

* **Feature**: API client updated
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

