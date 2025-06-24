# v1.20.5 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.4 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.3 (2025-04-03)

* No change notes available for this release.

# v1.20.2 (2025-03-27)

* No change notes available for this release.

# v1.20.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.20.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.15 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.14 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.13 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.12 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.11 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.19.10 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.19.9 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.8 (2025-01-10)

* **Documentation**: Doc only update for ServiceName that fixes several customer-reported issues

# v1.19.7 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.6 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.5 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.19.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2024-10-14)

* **Feature**: This release updates request validation regex for resource ARNs.

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

# v1.16.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.16.6 (2024-09-04)

* No change notes available for this release.

# v1.16.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.15.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-06-12)

* **Feature**: This release updates request validation regex to account for non-commercial aws partitions.

# v1.13.10 (2024-06-11)

* No change notes available for this release.

# v1.13.9 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.8 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.7 (2024-05-23)

* No change notes available for this release.

# v1.13.6 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.5 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.4 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.13.3 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.2 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-02-29)

* **Feature**: Add capability to update the Data Lake's MetaStoreManager Role in order to perform required data lake updates to use Iceberg table format in their data lake or update the role for any other reason.

# v1.12.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.11.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.11.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Documentation**: Documentation updates for Security Lake
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.7 (2024-01-10)

* No change notes available for this release.

# v1.10.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.10.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.10.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.9.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2023-10-27)

* No change notes available for this release.

# v1.7.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2023-08-28)

* **Feature**: Remove incorrect regex enforcement on pagination tokens.

# v1.6.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2023-08-01)

* No change notes available for this release.

# v1.6.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2023-07-20.2)

* **Feature**: Adding support for Tags on Create and Resource Tagging API.

# v1.4.4 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.3 (2023-06-15)

* No change notes available for this release.

# v1.4.2 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2023-06-01)

* No change notes available for this release.

# v1.4.0 (2023-05-30)

* **Feature**: Log sources are now versioned. AWS log sources and custom sources will now come with a version identifier that enables producers to vend multiple schema versions to subscribers. Security Lake API have been refactored to more closely align with AWS API conventions.

# v1.3.6 (2023-05-04)

* No change notes available for this release.

# v1.3.5 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.4 (2023-04-10)

* No change notes available for this release.

# v1.3.3 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.2 (2023-03-22)

* No change notes available for this release.

# v1.3.1 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2023-03-15)

* **Feature**: Make Create/Get/ListSubscribers APIs return resource share ARN and name so they can be used to validate the RAM resource share to accept. GetDatalake can be used to track status of UpdateDatalake and DeleteDatalake requests.

# v1.2.5 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.2.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.2.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.1.0 (2023-01-03)

* **Feature**: Allow CreateSubscriber API to take string input that allows setting more descriptive SubscriberDescription field. Make souceTypes field required in model level for UpdateSubscriberRequest as it is required for every API call on the backend. Allow ListSubscribers take any String as nextToken param.

# v1.0.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2022-11-29.2)

* **Release**: New AWS service client module
* **Feature**: Amazon Security Lake automatically centralizes security data from cloud, on-premises, and custom sources into a purpose-built data lake stored in your account. Security Lake makes it easier to analyze security data, so you can improve the protection of your workloads, applications, and data

