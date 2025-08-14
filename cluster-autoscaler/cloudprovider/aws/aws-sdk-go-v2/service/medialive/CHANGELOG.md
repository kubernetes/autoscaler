# v1.76.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.76.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.76.0 (2025-05-12)

* **Feature**: Add support to the AV1 rate control mode

# v1.75.0 (2025-05-07)

* **Feature**: Enables Updating Anywhere Settings on a MediaLive Anywhere Channel.

# v1.74.0 (2025-04-10)

* **Feature**: AWS Elemental MediaLive / Features : Add support for CMAF Ingest CaptionLanguageMappings, TimedMetadataId3 settings, and Link InputResolution.

# v1.73.0 (2025-04-07)

* **Feature**: AWS Elemental MediaLive now supports SDI inputs to MediaLive Anywhere Channels in workflows that use AWS SDKs.

# v1.72.1 (2025-04-03)

* No change notes available for this release.

# v1.72.0 (2025-04-02)

* **Feature**: Added support for SMPTE 2110 inputs when running a channel in a MediaLive Anywhere cluster. This feature enables ingestion of SMPTE 2110-compliant video, audio, and ancillary streams by reading SDP files that AWS Elemental MediaLive can retrieve from a network source.

# v1.71.0 (2025-03-11)

* **Feature**: Add an enum option DISABLED for Output Locking Mode under Global Configuration.

# v1.70.0 (2025-03-10)

* **Feature**: Adds defaultFontSize and defaultLineHeight as options in the EbuTtDDestinationSettings within the caption descriptions for an output stream.

# v1.69.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.69.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.68.1 (2025-02-25)

* No change notes available for this release.

# v1.68.0 (2025-02-18)

* **Feature**: Adds support for creating CloudWatchAlarmTemplates for AWS Elemental MediaTailor Playback Configuration resources.
* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.67.0 (2025-02-12)

* **Feature**: Adds a RequestId parameter to all MediaLive Workflow Monitor create operations.  The RequestId parameter allows idempotent operations.

# v1.66.4 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.66.3 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.66.2 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.66.1 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.66.0 (2025-01-22)

* **Feature**: AWS Elemental MediaLive adds a new feature, ID3 segment tagging, in CMAF Ingest output groups. It allows customers to insert ID3 tags into every output segment, controlled by a newly added channel schedule action Id3SegmentTagging.

# v1.65.4 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.65.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.65.2 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.65.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.65.0 (2024-12-19)

* **Feature**: MediaLive is releasing ListVersions api
* **Dependency Update**: Updated to the latest SDK module versions

# v1.64.0 (2024-12-16)

* **Feature**: AWS Elemental MediaLive adds three new features: MediaPackage v2 endpoint support for live stream delivery, KLV metadata passthrough in CMAF Ingest output groups, and Metadata Name Modifier in CMAF Ingest output groups for customizing metadata track names in output streams.

# v1.63.0 (2024-12-09)

* **Feature**: H265 outputs now support disabling the deblocking filter.

# v1.62.7 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.62.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.62.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.61.4 (2024-10-03)

* No change notes available for this release.

# v1.61.3 (2024-09-27)

* No change notes available for this release.

# v1.61.2 (2024-09-25)

* No change notes available for this release.

# v1.61.1 (2024-09-23)

* No change notes available for this release.

# v1.61.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.60.0 (2024-09-19)

* **Feature**: Adds Bandwidth Reduction Filtering for HD AVC and HEVC encodes, multiplex container settings.

# v1.59.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.59.0 (2024-09-16)

* **Feature**: Removing the ON_PREMISE enum from the input settings field.

# v1.58.0 (2024-09-11)

* **Feature**: Adds AV1 Codec support, SRT ouputs, and MediaLive Anywhere support.

# v1.57.1 (2024-09-04)

* No change notes available for this release.

# v1.57.0 (2024-09-03)

* **Feature**: Added MinQP as a Rate Control option for H264 and H265 encodes.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.0 (2024-08-12)

* **Feature**: AWS Elemental MediaLive now supports now supports editing the PID values for a Multiplex.

# v1.55.0 (2024-07-18)

* **Feature**: AWS Elemental MediaLive now supports the SRT protocol via the new SRT Caller input type.

# v1.54.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.53.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.4 (2024-05-23)

* No change notes available for this release.

# v1.52.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.52.0 (2024-05-06)

* **Feature**: AWS Elemental MediaLive now supports configuring how SCTE 35 passthrough triggers segment breaks in HLS and MediaPackage output groups. Previously, messages triggered breaks in all these output groups. The new option is to trigger segment breaks only in groups that have SCTE 35 passthrough enabled.

# v1.51.0 (2024-04-11)

* **Feature**: AWS Elemental MediaLive introduces workflow monitor, a new feature that enables the visualization and monitoring of your media workflows. Create signal maps of your existing workflows and monitor them by creating notification and monitoring template groups.

# v1.50.0 (2024-04-03)

* **Feature**: Cmaf Ingest outputs are now supported in Media Live

# v1.49.3 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.2 (2024-03-28)

* No change notes available for this release.

# v1.49.1 (2024-03-27)

* No change notes available for this release.

# v1.49.0 (2024-03-25)

* **Feature**: Exposing TileMedia H265 options

# v1.48.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.47.0 (2024-02-21)

* **Feature**: MediaLive now supports the ability to restart pipelines in a running channel.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.46.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.45.1 (2024-02-14)

* No change notes available for this release.

# v1.45.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2023-12-21)

* **Feature**: MediaLive now supports the ability to configure the audio that an AWS Elemental Link UHD device produces, when the device is configured as the source for a flow in AWS Elemental MediaConnect.

# v1.43.3 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.43.2 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.43.0 (2023-12-04)

* **Feature**: Adds support for custom color correction on channels using 3D LUT files.

# v1.42.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.3 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.2 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.41.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2023-11-17)

* **Feature**: MediaLive has now added support for per-output static image overlay.

# v1.40.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).
* **Feature**: **BREAKFIX**: Correct nullability representation of APIGateway-based services.

# v1.37.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2023-09-11)

* **Feature**: AWS Elemental Link now supports attaching a Link UHD device to a MediaConnect flow.

# v1.36.0 (2023-09-06)

* **Feature**: Adds advanced Output Locking options for Epoch Locking: Custom Epoch and Jam Sync Time

# v1.35.0 (2023-08-24)

* **Feature**: MediaLive now supports passthrough of KLV data to a HLS output group with a TS container. MediaLive now supports setting an attenuation mode for AC3 audio when the coding mode is 3/2 LFE. MediaLive now supports specifying whether to include filler NAL units in RTMP output group settings.

# v1.34.4 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.3 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2023-08-01)

* **Feature**: AWS Elemental Link devices now report their Availability Zone. Link devices now support the ability to change their Availability Zone.

# v1.33.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.2 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.1 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-07-07)

* **Feature**: This release enables the use of Thumbnails in AWS Elemental MediaLive.

# v1.31.6 (2023-06-15)

* No change notes available for this release.

# v1.31.5 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.4 (2023-05-04)

* No change notes available for this release.

# v1.31.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2023-04-10)

* No change notes available for this release.

# v1.31.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-03-27)

* **Feature**: AWS Elemental MediaLive now supports ID3 tag insertion for audio only HLS output groups. AWS Elemental Link devices now support tagging.

# v1.30.2 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-03-03)

* **Feature**: AWS Elemental MediaLive adds support for Nielsen watermark timezones.

# v1.29.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.29.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.2 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.29.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2023-01-19)

* **Feature**: AWS Elemental MediaLive adds support for SCTE 35 preRollMilliSeconds.

# v1.28.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.27.0 (2022-12-20)

* **Feature**: This release adds support for two new features to AWS Elemental MediaLive. First, you can now burn-in timecodes to your MediaLive outputs. Second, we now now support the ability to decode Dolby E audio when it comes in on an input.

# v1.26.1 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2022-12-09)

* **Feature**: Link devices now support buffer size (latency) configuration. A higher latency value means a longer delay in transmitting from the device to MediaLive, but improved resiliency. A lower latency value means a shorter delay, but less resiliency.

# v1.25.0 (2022-12-02)

* **Feature**: Updates to Event Signaling and Management (ESAM) API and documentation.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-10-13)

* **Feature**: AWS Elemental MediaLive now supports forwarding SCTE-35 messages through the Event Signaling and Management (ESAM) API, and can read those SCTE-35 messages from an inactive source.

# v1.23.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-09-14)

* **Feature**: This change exposes API settings which allow Dolby Atmos and Dolby Vision to be used when running a channel using Elemental Media Live
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.7 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.6 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.5 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.4 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.3 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-07-22)

* **Feature**: Link devices now support remote rebooting. Link devices now support maintenance windows. Maintenance windows allow a Link device to install software updates without stopping the MediaLive channel. The channel will experience a brief loss of input from the device while updates are installed.

# v1.21.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-06-29)

* **Feature**: This release adds support for automatic renewal of MediaLive reservations at the end of each reservation term. Automatic renewal is optional. This release also adds support for labelling accessibility-focused audio and caption tracks in HLS outputs.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.4 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.3 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-03-28)

* **Feature**: This release adds support for selecting a maintenance window.

# v1.19.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.14.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-12)

* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.12.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-07-15)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-11)

* **Feature**: Updated to latest API model.

# v1.5.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

