# v1.25.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2025-04-30)

* **Feature**: This release adds support for ProtectedQuery results to be delivered to more than one collaboration member via the new distribute output configuration in StartProtectedQuery.

# v1.24.1 (2025-04-03)

* No change notes available for this release.

# v1.24.0 (2025-04-01)

* **Feature**: This release adds support for updating the analytics engine of a collaboration.

# v1.23.2 (2025-03-21)

* No change notes available for this release.

# v1.23.1 (2025-03-19)

* No change notes available for this release.

# v1.23.0 (2025-03-18)

* **Feature**: This release adds support for PySpark jobs. Customers can now analyze data by running jobs using approved PySpark analysis templates.

# v1.22.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.22.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.9 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.8 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.7 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.6 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.21.4 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.21.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2024-12-02)

* **Feature**: This release allows customers and their partners to easily collaborate with data stored in Snowflake and Amazon Athena, without having to move or share their underlying data among collaborators.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2024-11-07)

* **Feature**: This release introduces support for Custom Models in AWS Clean Rooms ML.
* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.19.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2024-10-29)

* **Feature**: This release adds the option for customers to configure analytics engine when creating a collaboration, and introduces the new SPARK analytics engine type in addition to maintaining the legacy CLEAN_ROOMS_SQL engine type.

# v1.18.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.4 (2024-10-03)

* No change notes available for this release.

# v1.17.3 (2024-09-27)

* No change notes available for this release.

# v1.17.2 (2024-09-25)

* No change notes available for this release.

# v1.17.1 (2024-09-23)

* No change notes available for this release.

# v1.17.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.5 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.16.4 (2024-09-10)

* No change notes available for this release.

# v1.16.3 (2024-09-04)

* No change notes available for this release.

# v1.16.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-07-24)

* **Feature**: Three enhancements to the AWS Clean Rooms: Disallowed Output Columns, Flexible Result Receivers, SQL as a Seed

# v1.15.0 (2024-07-23)

* **Feature**: This release adds AWS Entity Resolution integration to associate ID namespaces & ID mapping workflow resources as part of ID namespace association and  ID mapping table  in AWS Clean Rooms. It also introduces a new ID_MAPPING_TABLE analysis rule to manage the protection on ID mapping table.

# v1.14.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.13.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.4 (2024-05-23)

* No change notes available for this release.

# v1.12.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.12.0 (2024-04-10)

* **Feature**: AWS Clean Rooms Differential Privacy is now fully available. Differential privacy protects against user-identification attempts.

# v1.11.0 (2024-04-04)

* **Feature**: Feature: New schemaStatusDetails field to the existing Schema object that displays a status on Schema API responses to show whether a schema is queryable or not. New BatchGetSchemaAnalysisRule API to retrieve multiple schemaAnalysisRules using a single API call.

# v1.10.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.9.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.9.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.8.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.8.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2023-11-29)

* **Feature**: AWS Clean Rooms now provides differential privacy to protect against user-identification attempts and machine learning modeling to allow two parties to identify similar users in their data.
* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.4 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.7.2 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2023-11-14)

* **Feature**: This feature provides the ability for the collaboration creator to configure either the member who can run queries or a different member in the collaboration to be billed for query compute costs.

# v1.6.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2023-08-30)

* **Feature**: This release decouples member abilities in a collaboration. With this change, the member who can run queries no longer needs to be the same as the member who can receive results.

# v1.3.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2023-08-01)

* No change notes available for this release.

# v1.3.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Feature**: This release introduces custom SQL queries - an expanded set of SQL you can run. This release adds analysis templates, a new resource for storing pre-defined custom SQL queries ahead of time. This release also adds the Custom analysis rule, which lets you approve analysis templates for querying.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2023-06-29)

* **Feature**: This release adds support for the OR operator in RSQL join match conditions and the ability to control which operators (AND, OR) are allowed in a join match condition.

# v1.1.6 (2023-06-15)

* No change notes available for this release.

# v1.1.5 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.4 (2023-05-04)

* No change notes available for this release.

# v1.1.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.2 (2023-04-10)

* No change notes available for this release.

# v1.1.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2023-03-21)

* **Feature**: GA Release of AWS Clean Rooms, Added Tagging Functionality
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.5 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.0.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.0.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2023-01-12)

* **Release**: New AWS service client module
* **Feature**: Initial release of AWS Clean Rooms

