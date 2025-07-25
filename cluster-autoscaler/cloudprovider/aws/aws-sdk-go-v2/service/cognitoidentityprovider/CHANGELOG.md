# v1.53.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.0 (2025-05-14)

* **Feature**: Add exceptions to WebAuthn operations.

# v1.52.0 (2025-04-22)

* **Feature**: This release adds refresh token rotation.

# v1.51.4 (2025-04-03)

* No change notes available for this release.

# v1.51.3 (2025-03-14)

* **Documentation**: Minor description updates to API parameters

# v1.51.2 (2025-03-07)

* No change notes available for this release.

# v1.51.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.51.0 (2025-03-04)

* **Feature**: Added the capacity to return available challenges in admin authentication and to set version 3 of the pre token generation event for M2M ATC.

# v1.50.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.5 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.4 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.3 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.49.0 (2025-01-21)

* **Feature**: corrects the dual-stack endpoint configuration for cognitoidp

# v1.48.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.48.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.4 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.3 (2024-12-20)

* No change notes available for this release.

# v1.48.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-12-11)

* **Documentation**: Updated descriptions for some API operations and parameters, corrected some errors in Cognito user pools

# v1.48.0 (2024-12-09)

* **Feature**: Change `CustomDomainConfig` from a required to an optional parameter for the `UpdateUserPoolDomain` operation.

# v1.47.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.0 (2024-11-22)

* **Feature**: Add support for users to sign up and sign in without passwords, using email and SMS OTPs and Passkeys. Add support for Passkeys based on WebAuthn. Add support for enhanced branding customization for hosted authentication pages with Amazon Cognito Managed Login. Add feature tiers with new pricing.

# v1.46.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.46.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.4 (2024-10-03)

* No change notes available for this release.

# v1.45.3 (2024-09-27)

* No change notes available for this release.

# v1.45.2 (2024-09-25)

* No change notes available for this release.

# v1.45.1 (2024-09-23)

* No change notes available for this release.

# v1.45.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.44.0 (2024-09-12)

* **Feature**: Added email MFA option to user pools with advanced security features.

# v1.43.4 (2024-09-04)

* No change notes available for this release.

# v1.43.3 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-08-09)

* **Documentation**: Fixed a description of AdvancedSecurityAdditionalFlows in Amazon Cognito user pool configuration.

# v1.43.0 (2024-08-08)

* **Feature**: Added support for threat protection for custom authentication in Amazon Cognito user pools.

# v1.42.0 (2024-08-06)

* **Feature**: Advanced security feature updates to include password history and log export for Cognito user pools.

# v1.41.4 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.3 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2024-07-05)

* No change notes available for this release.

# v1.41.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.40.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2024-06-17)

* **Feature**: Add v2 smoke tests and smithy smokeTests trait for SDK testing.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.5 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.4 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.3 (2024-05-23)

* No change notes available for this release.

# v1.38.2 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-05-08)

* **Feature**: Add EXTERNAL_PROVIDER enum value to UserStatusType.
* **Bug Fix**: GoDoc improvement

# v1.37.0 (2024-04-26)

* **Feature**: Add LimitExceededException to SignUp errors

# v1.36.5 (2024-04-16)

* No change notes available for this release.

# v1.36.4 (2024-04-11)

* No change notes available for this release.

# v1.36.3 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.2 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2024-03-15)

* No change notes available for this release.

# v1.36.0 (2024-03-08)

* **Feature**: Add ConcurrentModificationException to SetUserPoolMfaConfig

# v1.35.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.34.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.34.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.34.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2024-02-01)

* **Feature**: Added CreateIdentityProvider and UpdateIdentityProvider details for new SAML IdP features

# v1.32.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-12-18)

* **Feature**: Amazon Cognito now supports trigger versions that define the fields in the request sent to pre token generation Lambda triggers.

# v1.31.6 (2023-12-15)

* No change notes available for this release.

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

# v1.30.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.30.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.27.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-09-27)

* **Feature**: The UserPoolType Status field is no longer used.

# v1.26.1 (2023-08-31)

* No change notes available for this release.

# v1.26.0 (2023-08-29)

* **Feature**: Added API example requests and responses for several operations. Fixed the validation regex for user pools Identity Provider name.

# v1.25.4 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.3 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-08-02)

* **Feature**: New feature that logs Cognito user pool error messages to CloudWatch logs.

# v1.24.1 (2023-08-01)

* No change notes available for this release.

# v1.24.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-07-13)

* **Feature**: API model updated in Amazon Cognito
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.12 (2023-06-15)

* No change notes available for this release.

# v1.22.11 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.10 (2023-05-04)

* No change notes available for this release.

# v1.22.9 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.8 (2023-04-10)

* No change notes available for this release.

# v1.22.7 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.6 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.5 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.21.4 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2022-12-07)

* No change notes available for this release.

# v1.21.2 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-10-21)

* **Feature**: This release adds a new "DeletionProtection" field to the UserPool in Cognito. Application admins can configure this value with either ACTIVE or INACTIVE value. Setting this field to ACTIVE will prevent a user pool from accidental deletion.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-09-02)

* **Feature**: This release adds a new "AuthSessionValidity" field to the UserPoolClient in Cognito. Application admins can configure this value for their users' authentication duration, which is currently fixed at 3 minutes, up to 15 minutes. Setting this field will also apply to the SMS MFA authentication flow.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.6 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-08-18)

* **Documentation**: This change is being made simply to fix the public documentation based on the models. We have included the PasswordChange and ResendCode events, along with the Pass, Fail and InProgress status. We have removed the Success and Failure status which are never returned by our APIs.

# v1.18.3 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-08-03)

* **Feature**: Add a new exception type, ForbiddenException, that is returned when request is not allowed

# v1.17.4 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.2 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-05-31)

* **Feature**: Amazon Cognito now supports IP Address propagation for all unauthenticated APIs (e.g. SignUp, ForgotPassword).

# v1.16.0 (2022-05-24)

* **Feature**: Amazon Cognito now supports requiring attribute verification (ex. email and phone number) before update.

# v1.15.5 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-03-15)

* **Feature**: Updated EmailConfigurationType and SmsConfigurationType to reflect that you can now choose Amazon SES and Amazon SNS resources in the same Region.

# v1.14.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
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

# v1.5.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.3 (2021-06-11)

* **Documentation**: Updated to latest API model.

# v1.3.2 (2021-06-04)

* No change notes available for this release.

# v1.3.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

