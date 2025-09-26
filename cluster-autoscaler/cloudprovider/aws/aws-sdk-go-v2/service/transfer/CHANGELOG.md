# v1.60.3 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.60.2 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.60.1 (2025-04-16)

* No change notes available for this release.

# v1.60.0 (2025-04-09)

* **Feature**: This launch includes 2 enhancements to SFTP connectors user-experience: 1) Customers can self-serve concurrent connections setting for their connectors, and 2) Customers can discover the public host key of remote servers using their SFTP connectors.

# v1.59.0 (2025-04-07)

* **Feature**: This launch enables customers to manage contents of their remote directories, by deleting old files or moving files to archive folders in remote servers once they have been retrieved. Customers will be able to automate the process using event-driven architecture.

# v1.58.1 (2025-04-03)

* No change notes available for this release.

# v1.58.0 (2025-03-31)

* **Feature**: Add WebAppEndpointPolicy support for WebApps

# v1.57.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.57.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.4 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.3 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.2 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.0 (2025-01-24)

* **Feature**: Added CustomDirectories as a new directory option for storing inbound AS2 messages, MDN files and Status files.
* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.55.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.55.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.3 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.55.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.0 (2024-12-18)

* **Feature**: Added AS2 agreement configurations to control filename preservation and message signing enforcement. Added AS2 connector configuration to preserve content type from S3 objects.

# v1.54.0 (2024-12-02)

* **Feature**: AWS Transfer Family now offers Web apps that enables simple and secure access to data stored in Amazon S3.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.5 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.4 (2024-11-15.2)

* No change notes available for this release.

# v1.53.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.53.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.0 (2024-10-14)

* **Feature**: This release enables customers using SFTP connectors to query the transfer status of their files to meet their monitoring needs as well as orchestrate post transfer actions.

# v1.52.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.4 (2024-10-03)

* No change notes available for this release.

# v1.51.3 (2024-09-27)

* No change notes available for this release.

# v1.51.2 (2024-09-25)

* No change notes available for this release.

# v1.51.1 (2024-09-23)

* No change notes available for this release.

# v1.51.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.50.6 (2024-09-04)

* No change notes available for this release.

# v1.50.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.49.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.4 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.3 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.2 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-05-23)

* No change notes available for this release.

# v1.48.0 (2024-05-17)

* **Feature**: Enable use of CloudFormation traits in Smithy model to improve generated CloudFormation schema from the Smithy API model.

# v1.47.5 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.4 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.3 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.47.2 (2024-05-06)

* No change notes available for this release.

# v1.47.1 (2024-05-03)

* No change notes available for this release.

# v1.47.0 (2024-04-22)

* **Feature**: Adding new API to support remote directory listing using SFTP connector

# v1.46.0 (2024-04-12)

* **Feature**: This change releases support for importing self signed certificates to the Transfer Family for sending outbound file transfers over TLS/HTTPS.

# v1.45.0 (2024-04-03)

* **Feature**: Add ability to specify Security Policies for SFTP Connectors

# v1.44.2 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-03-08)

* **Feature**: Added DES_EDE3_CBC to the list of supported encryption algorithms for messages sent with an AS2 connector.

# v1.43.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.42.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.42.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.41.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.41.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-01-12)

* **Feature**: AWS Transfer Family now supports static IP addresses for SFTP & AS2 connectors and for async MDNs on AS2 servers.

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

# v1.38.3 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.38.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2023-11-16)

* **Feature**: Introduced S3StorageOptions for servers to enable directory listing optimizations and added Type fields to logical directory mappings.

# v1.37.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2023-10-26)

* **Feature**: No API changes from previous release. This release migrated the model to Smithy keeping all features unchanged.

# v1.34.2 (2023-10-16)

* **Documentation**: Documentation updates for AWS Transfer Family

# v1.34.1 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-10-06)

* **Feature**: This release updates the max character limit of PreAuthenticationLoginBanner and PostAuthenticationLoginBanner to 4096 characters
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.9 (2023-10-02)

* **Documentation**: Documentation updates for AWS Transfer Family

# v1.33.8 (2023-09-05)

* No change notes available for this release.

# v1.33.7 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.6 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.5 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.4 (2023-08-14)

* **Documentation**: Documentation updates for AWS Transfer Family

# v1.33.3 (2023-08-10)

* **Documentation**: Documentation updates for AW Transfer Family

# v1.33.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2023-08-01)

* No change notes available for this release.

# v1.33.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-07-25)

* **Feature**: This release adds support for SFTP Connectors.

# v1.31.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-06-30)

* **Feature**: Add outbound Basic authentication support to AS2 connectors

# v1.30.0 (2023-06-21)

* **Feature**: This release adds a new parameter StructuredLogDestinations to CreateServer, UpdateServer APIs.

# v1.29.3 (2023-06-15)

* No change notes available for this release.

# v1.29.2 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.1 (2023-05-30)

* No change notes available for this release.

# v1.29.0 (2023-05-15)

* **Feature**: This release introduces the ability to require both password and SSH key when users authenticate to your Transfer Family servers that use the SFTP protocol.

# v1.28.12 (2023-05-04)

* No change notes available for this release.

# v1.28.11 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.10 (2023-04-10)

* No change notes available for this release.

# v1.28.9 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.8 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.7 (2023-03-20)

* No change notes available for this release.

# v1.28.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.28.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.28.2 (2023-02-07)

* **Documentation**: Updated the documentation for the ImportCertificate API call, and added examples.

# v1.28.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.27.0 (2022-12-27)

* **Feature**: Add additional operations to throw ThrottlingExceptions

# v1.26.0 (2022-12-21)

* **Feature**: This release adds support for Decrypt as a workflow step type.

# v1.25.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-11-18)

* **Feature**: Adds a NONE encryption algorithm type to AS2 connectors, providing support for skipping encryption of the AS2 message body when a HTTPS URL is also specified.

# v1.24.0 (2022-11-16)

* **Feature**: Allow additional operations to throw ThrottlingException

# v1.23.3 (2022-11-08)

* No change notes available for this release.

# v1.23.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-10-13)

* **Feature**: This release adds an option for customers to configure workflows that are triggered when files are only partially received from a client due to premature session disconnect.

# v1.22.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-09-14)

* **Feature**: This release introduces the ability to have multiple server host keys for any of your Transfer Family servers that use the SFTP protocol.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.8 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.7 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.6 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2022-08-24)

* **Documentation**: Documentation updates for AWS Transfer Family

# v1.21.4 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-07-26)

* **Feature**: AWS Transfer Family now supports Applicability Statement 2 (AS2), a network protocol used for the secure and reliable transfer of critical Business-to-Business (B2B) data over the public internet using HTTP/HTTPS as the transport mechanism.

# v1.20.2 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-06-22)

* **Feature**: Until today, the service supported only RSA host keys and user keys. Now with this launch, Transfer Family has expanded the support for ECDSA and ED25519 host keys and user keys, enabling customers to support a broader set of clients by choosing RSA, ECDSA, and ED25519 host and user keys.

# v1.19.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-05-18)

* **Feature**: AWS Transfer Family now supports SetStat server configuration option, which provides the ability to ignore SetStat command issued by file transfer clients, enabling customers to upload files without any errors.

# v1.18.7 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.6 (2022-05-12)

* **Documentation**: AWS Transfer Family now accepts ECDSA keys for server host keys

# v1.18.5 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-04-19)

* **Documentation**: This release contains corrected HomeDirectoryMappings examples for several API functions: CreateAccess, UpdateAccess, CreateUser, and UpdateUser,.

# v1.18.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-03-23)

* **Documentation**: Documentation updates for AWS Transfer Family to describe how to remove an associated workflow from a server.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-03-10)

* **Feature**: Adding more descriptive error types for managed workflows

# v1.17.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service client model to latest release.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: API client updated

# v1.12.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.10.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-09-30)

* **Feature**: API client updated

# v1.7.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-09-02)

* **Feature**: API client updated

# v1.6.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.3 (2021-08-19)

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

# v1.4.1 (2021-06-11)

* **Documentation**: Updated to latest API model.

# v1.4.0 (2021-05-25)

* **Feature**: API client updated

# v1.3.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

