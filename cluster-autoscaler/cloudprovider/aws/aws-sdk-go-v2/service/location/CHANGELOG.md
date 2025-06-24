# v1.44.4 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.3 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.2 (2025-04-03)

* No change notes available for this release.

# v1.44.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.44.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2025-02-19)

* **Feature**: Adds support for larger property maps for tracking and geofence positions changes. It increases the maximum number of items from 3 to 4, and the maximum value length from 40 to 150.

# v1.42.17 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.16 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.15 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.14 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.13 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.42.12 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.42.11 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.10 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.9 (2025-01-02)

* No change notes available for this release.

# v1.42.8 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.7 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.42.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.5 (2024-10-03)

* No change notes available for this release.

# v1.41.4 (2024-10-01)

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

# v1.40.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.40.6 (2024-09-04)

* No change notes available for this release.

# v1.40.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.38.2 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-06-06)

* **Feature**: Added two new APIs, VerifyDevicePosition and ForecastGeofenceEvents. Added support for putting larger geofences up to 100,000 vertices with Geobuf fields.

# v1.37.9 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.8 (2024-05-23)

* No change notes available for this release.

# v1.37.7 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.6 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.5 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.37.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.36.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.36.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2024-01-12)

* **Documentation**: Location SDK documentation update. Added missing fonts to the MapConfiguration data type. Updated note for the SubMunicipality property in the place data type.

# v1.35.0 (2024-01-10)

* **Feature**: This release adds API support for custom layers for the maps service APIs: CreateMap, UpdateMap, DescribeMap.

# v1.34.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-12-29)

* **Feature**: This release introduces a new parameter to bypasses an API key's expiry conditions and delete the key.

# v1.33.0 (2023-12-12)

* **Feature**: This release 1)  adds sub-municipality field in Places API for searching and getting places information, and 2) allows optimizing route calculation based on expected arrival time.

# v1.32.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.32.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.32.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.31.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-11-17)

* **Feature**: Remove default value and allow nullable for request parameters having minimum value larger than zero.

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

# v1.28.0 (2023-10-12)

* **Feature**: This release adds endpoint updates for all AWS Location resource operations.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-10-03)

* **Feature**: Amazon Location Service adds support for bounding polygon queries. Additionally, the GeofenceCount field has been added to the DescribeGeofenceCollection API response.

# v1.26.6 (2023-09-06)

* No change notes available for this release.

# v1.26.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2023-08-01)

* No change notes available for this release.

# v1.26.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-07-06)

* **Feature**: This release adds support for authenticating with Amazon Location Service's Places & Routes APIs with an API Key. Also, with this release developers can publish tracked device position updates to Amazon EventBridge.

# v1.24.0 (2023-06-15)

* **Feature**: Amazon Location Service adds categories to places, including filtering on those categories in searches. Also, you can now add metadata properties to your geofences.

# v1.23.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-05-30)

* **Feature**: This release adds API support for political views for the maps service APIs: CreateMap, UpdateMap, DescribeMap.

# v1.22.7 (2023-05-04)

* No change notes available for this release.

# v1.22.6 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.5 (2023-04-10)

* No change notes available for this release.

# v1.22.4 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.3 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2023-03-07)

* **Documentation**: Documentation update for the release of 3 additional map styles for use with Open Data Maps: Open Data Standard Dark, Open Data Visualization Light & Open Data Visualization Dark.

# v1.22.0 (2023-02-23)

* **Feature**: This release adds support for using Maps APIs with an API Key in addition to AWS Cognito. This includes support for adding, listing, updating and deleting API Keys.

# v1.21.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.21.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.21.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2023-01-10)

* **Feature**: This release adds support for two new route travel models, Bicycle and Motorcycle which can be used with Grab data source.

# v1.20.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.19.5 (2022-12-15)

* **Documentation**: This release adds support for a new style, "VectorOpenDataStandardLight" which can be used with the new data source, "Open Data Maps (Preview)".
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.4 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2022-10-25)

* **Documentation**: Added new map styles with satellite imagery for map resources using HERE as a data provider.

# v1.19.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-09-27)

* **Feature**: This release adds place IDs, which are unique identifiers of places, along with a new GetPlace operation, which can be used with place IDs to find a place again later. UnitNumber and UnitType are also added as new properties of places.

# v1.18.6 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.3 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-08-09)

* **Feature**: Amazon Location Service now allows circular geofences in BatchPutGeofence, PutGeofence, and GetGeofence  APIs.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.6 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.5 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.4 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.2 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-05-06)

* **Feature**: Amazon Location Service now includes a MaxResults parameter for ListGeofences requests.

# v1.16.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-03-22)

* **Feature**: Amazon Location Service now includes a MaxResults parameter for GetDevicePositionHistory requests.

# v1.15.1 (2022-03-15)

* **Documentation**: New HERE style "VectorHereExplore" and "VectorHereExploreTruck".

# v1.15.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2022-01-28)

* **Feature**: Updated to latest API model.
* **Bug Fix**: Updates SDK API client deserialization to pre-allocate byte slice and string response payloads, [#1565](https://github.com/aws/aws-sdk-go-v2/pull/1565). Thanks to [Tyson Mote](https://github.com/tysonmote) for submitting this PR.

# v1.12.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: API client updated

# v1.9.2 (2021-12-03)

* **Bug Fix**: Fixed a bug that prevented the resolution of the correct endpoint for some API operations.
* **Bug Fix**: Fixed an issue that caused some operations to not be signed using sigv4, resulting in authentication failures.

# v1.9.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Bug Fix**: Fixed an issue that caused one or more API operations to fail when attempting to resolve the service endpoint. ([#1349](https://github.com/aws/aws-sdk-go-v2/pull/1349))
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2021-06-04)

* **Feature**: Updated service client to latest API model.

# v1.1.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

