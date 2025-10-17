# v1.23.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2025-05-19)

* **Feature**: This release adds support for DVB-DASH, EBU-TT-D subtitle format, and non-compacted manifests for DASH in MediaPackage v2 Origin Endpoints.

# v1.22.2 (2025-04-03)

* No change notes available for this release.

# v1.22.1 (2025-03-28)

* No change notes available for this release.

# v1.22.0 (2025-03-13)

* **Feature**: This release adds the ResetChannelState and ResetOriginEndpointState operation to reset MediaPackage V2 channel and origin endpoint. This release also adds a new field, UrlEncodeChildManifest, for HLS/LL-HLS to allow URL-encoding child manifest query string based on the requirements of AWS SigV4.

# v1.21.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.21.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.11 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.10 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.9 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.8 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.7 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.20.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.20.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.4 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.20.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2024-11-20)

* **Feature**: MediaPackage v2 now supports the Media Quality Confidence Score (MQCS) published from MediaLive. Customers can control input switching based on the MQCS and publishing HTTP Headers for the MQCS via the API.

# v1.19.3 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.2 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.19.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2024-10-28)

* **Feature**: MediaPackage V2 Live to VOD Harvester is a MediaPackage V2 feature, which is used to export content from an origin endpoint to a S3 bucket.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2024-10-03)

* **Feature**: Added support for ClipStartTime on the FilterConfiguration object on OriginEndpoint manifest settings objects. Added support for EXT-X-START tags on produced HLS child playlists.

# v1.16.3 (2024-09-27)

* No change notes available for this release.

# v1.16.2 (2024-09-25)

* No change notes available for this release.

# v1.16.1 (2024-09-23)

* No change notes available for this release.

# v1.16.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.5 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.15.4 (2024-09-10)

* No change notes available for this release.

# v1.15.3 (2024-09-04)

* No change notes available for this release.

# v1.15.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-07-24)

* **Feature**: This release adds support for Irdeto DRM encryption in DASH manifests.

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

# v1.12.1 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-06-13)

* **Feature**: This release adds support for CMAF ingest (DASH-IF live media ingest protocol interface 1)

# v1.11.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.4 (2024-05-23)

* No change notes available for this release.

# v1.11.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.11.0 (2024-04-16)

* **Feature**: Dash v2 is a MediaPackage V2 feature to support egressing on DASH manifest format.

# v1.10.2 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-03-11)

* **Feature**: This release enables customers to safely update their MediaPackage v2 channel groups, channels and origin endpoints using entity tags.

# v1.9.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.8.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.8.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.6 (2023-12-28)

* No change notes available for this release.

# v1.7.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.7.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.7.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.6.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2023-10-30)

* **Feature**: This feature allows customers to create a combination of manifest filtering, startover and time delay configuration that applies to all egress requests by default.

# v1.3.1 (2023-10-26)

* No change notes available for this release.

# v1.3.0 (2023-10-16)

* **Feature**: This release allows customers to manage MediaPackage v2 resource using CloudFormation.

# v1.2.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2023-09-18)

* **Announcement**: [BREAKFIX] Change in MaxResults datatype from value to pointer type in cognito-sync service.
* **Feature**: Adds several endpoint ruleset changes across all models: smaller rulesets, removed non-unique regional endpoints, fixes FIPS and DualStack endpoints, and make region not required in SDK::Endpoint. Additional breakfix to cognito-sync field.

# v1.1.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.1 (2023-08-01)

* No change notes available for this release.

# v1.1.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.6 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.5 (2023-07-18)

* No change notes available for this release.

# v1.0.4 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.3 (2023-06-15)

* No change notes available for this release.

# v1.0.2 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2023-05-23)

* No change notes available for this release.

# v1.0.0 (2023-05-19)

* **Release**: New AWS service client module
* **Feature**: Adds support for the MediaPackage Live v2 API

