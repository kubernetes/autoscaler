# v1.57.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.57.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.57.0 (2025-05-15)

* **Feature**: Added the new AlwaysOn running mode for WorkSpaces Pools. Customers can now choose between AlwaysOn (for instant access, with hourly usage billing regardless of connection status), or AutoStop (to optimize cost, with a brief startup delay) for their pools.

# v1.56.0 (2025-05-09)

* **Feature**: Remove parameter EnableWorkDocs from WorkSpacesServiceModel due to end of support of Amazon WorkDocs service.

# v1.55.3 (2025-05-08)

* No change notes available for this release.

# v1.55.2 (2025-04-03)

* No change notes available for this release.

# v1.55.1 (2025-03-11)

* No change notes available for this release.

# v1.55.0 (2025-03-06)

* **Feature**: Added a new ModifyEndpointEncryptionMode API for managing endpoint encryption settings.

# v1.54.0 (2025-03-05)

* **Feature**: Added DeviceTypeWorkSpacesThinClient type to allow users to access their WorkSpaces through a WorkSpaces Thin Client.

# v1.53.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.53.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.6 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.5 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.4 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.3 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.2 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.52.1 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.52.0 (2025-01-15)

* **Feature**: Added GeneralPurpose.4xlarge & GeneralPurpose.8xlarge ComputeTypes.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2024-12-19)

* **Feature**: Added AWS Global Accelerator (AGA) support for WorkSpaces Personal.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.3 (2024-12-09)

* **Documentation**: Added text to clarify case-sensitivity

# v1.50.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.1 (2024-11-22)

* **Documentation**: While integrating WSP-DCV rebrand, a few mentions were erroneously renamed from WSP to DCV. This release reverts those mentions back to WSP.

# v1.50.0 (2024-11-20)

* **Feature**: Added support for Rocky Linux 8 on Amazon WorkSpaces Personal.

# v1.49.0 (2024-11-19)

* **Feature**: Releasing new ErrorCodes for Image Validation failure during CreateWorkspaceImage process

# v1.48.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.48.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.5 (2024-10-03)

* No change notes available for this release.

# v1.47.4 (2024-10-02)

* **Documentation**: WSP is being rebranded to become DCV.

# v1.47.3 (2024-09-27)

* No change notes available for this release.

# v1.47.2 (2024-09-25)

* No change notes available for this release.

# v1.47.1 (2024-09-23)

* No change notes available for this release.

# v1.47.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Feature**: Releasing new ErrorCodes for SysPrep failures during ImageImport and CreateImage process
* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.4 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.46.3 (2024-09-04)

* No change notes available for this release.

# v1.46.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2024-08-28)

* **Documentation**: Documentation-only update that clarifies the StartWorkspaces and StopWorkspaces actions, and a few other minor edits.

# v1.46.0 (2024-08-26)

* **Feature**: This release adds support for creating and managing directories that use AWS IAM Identity Center as user identity source. Such directories can be used to create non-Active Directory domain joined WorkSpaces Personal.Updated RegisterWorkspaceDirectory and DescribeWorkspaceDirectories APIs.

# v1.45.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-08-06)

* **Feature**: Added support for BYOL_GRAPHICS_G4DN_WSP IngestionProcess

# v1.44.3 (2024-07-30)

* **Documentation**: Removing multi-session as it isn't supported for pools

# v1.44.2 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-07-03)

* **Feature**: Fix create workspace bundle RootStorage/UserStorage to accept non null values

# v1.43.0 (2024-06-28)

* **Feature**: Added support for Red Hat Enterprise Linux 8 on Amazon WorkSpaces Personal.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-06-27)

* **Feature**: Added support for WorkSpaces Pools.

# v1.41.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.40.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.4 (2024-05-23)

* No change notes available for this release.

# v1.39.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.39.0 (2024-04-18)

* **Feature**: Adds new APIs for managing and sharing WorkSpaces BYOL configuration across accounts.

# v1.38.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Documentation**: Added note for user decoupling
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.37.3 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.37.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.37.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2024-02-08)

* **Feature**: This release introduces User-Decoupling feature. This feature allows Workspaces Core customers to provision workspaces without providing users. CreateWorkspaces and DescribeWorkspaces APIs will now take a new optional parameter "WorkspaceName".

# v1.35.9 (2024-02-05)

* **Documentation**: Added definitions of various WorkSpace states

# v1.35.8 (2024-01-11)

* **Documentation**: Added AWS Workspaces RebootWorkspaces API - Extended Reboot documentation update

# v1.35.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.6 (2023-12-14)

* **Documentation**: Updated note to ensure customers understand running modes.

# v1.35.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.35.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.35.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.34.0 (2023-11-27)

* **Feature**: The release introduces Multi-Region Resilience one-way data replication that allows you to replicate data from your primary WorkSpace to a standby WorkSpace in another AWS Region. DescribeWorkspaces now returns the status of data replication.

# v1.33.4 (2023-11-21)

* No change notes available for this release.

# v1.33.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2023-10-19)

* **Documentation**: Documentation updates for WorkSpaces

# v1.31.2 (2023-10-12)

* **Documentation**: Updated the CreateWorkspaces action documentation to clarify that the PCoIP protocol is only available for Windows bundles.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-10-05)

* **Feature**: This release introduces Manage applications. This feature allows users to manage their WorkSpaces applications by associating or disassociating their WorkSpaces with applications. The DescribeWorkspaces API will now additionally return OperatingSystemName in its responses.

# v1.30.0 (2023-09-08)

* **Feature**: A new field "ErrorDetails" will be added to the output of "DescribeWorkspaceImages" API call. This field provides in-depth details about the error occurred during image import process. These details include the possible causes of the errors and troubleshooting information.

# v1.29.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.1 (2023-08-01)

* No change notes available for this release.

# v1.29.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.18 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.17 (2023-07-21)

* **Documentation**: Fixed VolumeEncryptionKey descriptions

# v1.28.16 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.15 (2023-06-15)

* No change notes available for this release.

# v1.28.14 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.13 (2023-06-05)

* No change notes available for this release.

# v1.28.12 (2023-05-04)

* No change notes available for this release.

# v1.28.11 (2023-04-28)

* **Documentation**: Added Windows 11 to support Microsoft_Office_2019

# v1.28.10 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.9 (2023-04-10)

* No change notes available for this release.

# v1.28.8 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.28.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.28.2 (2023-02-09)

* **Documentation**: Removed Windows Server 2016 BYOL and made changes based on IAM campaign.

# v1.28.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.27.2 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2022-11-17)

* **Feature**: The release introduces CreateStandbyWorkspaces, an API that allows you to create standby WorkSpaces associated with a primary WorkSpace in another Region. DescribeWorkspaces now includes related WorkSpaces properties. DescribeWorkspaceBundles and CreateWorkspaceBundle now return more bundle details.

# v1.26.0 (2022-11-15)

* **Feature**: This release introduces ModifyCertificateBasedAuthProperties, a new API that allows control of certificate-based auth properties associated with a WorkSpaces directory. The DescribeWorkspaceDirectories API will now additionally return certificate-based auth properties in its responses.

# v1.25.0 (2022-11-07)

* **Feature**: This release adds protocols attribute to workspaces properties data type. This enables customers to migrate workspaces from PC over IP (PCoIP) to WorkSpaces Streaming Protocol (WSP) using create and modify workspaces public APIs.

# v1.24.0 (2022-10-25)

* **Feature**: This release adds new enums for supporting Workspaces Core features, including creating Manual running mode workspaces, importing regular Workspaces Core images and importing g4dn Workspaces Core images.

# v1.23.2 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-09-29)

* **Feature**: This release includes diagnostic log uploading feature. If it is enabled, the log files of WorkSpaces Windows client will be sent to Amazon WorkSpaces automatically for troubleshooting. You can use modifyClientProperty api to enable/disable this feature.

# v1.22.9 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.8 (2022-09-15)

* No change notes available for this release.

# v1.22.7 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.6 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.5 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.4 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.3 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-08-01)

* **Feature**: This release introduces ModifySamlProperties, a new API that allows control of SAML properties associated with a WorkSpaces directory. The DescribeWorkspaceDirectories API will now additionally return SAML properties in its responses.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-07-27)

* **Feature**: Added CreateWorkspaceImage API to create a new WorkSpace image from an existing WorkSpace.

# v1.20.0 (2022-07-19)

* **Feature**: Increased the character limit of the login message from 850 to 2000 characters.

# v1.19.2 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-06-15)

* **Feature**: Added new field "reason" to OperationNotSupportedException. Receiving this exception in the DeregisterWorkspaceDirectory API will now return a reason giving more context on the failure.

# v1.18.4 (2022-06-10)

* No change notes available for this release.

# v1.18.3 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-04-11)

* **Feature**: Added API support that allows customers to create GPU-enabled WorkSpaces using EC2 G4dn instances.

# v1.17.0 (2022-03-31)

* **Feature**: Added APIs that allow you to customize the logo, login message, and help links in the WorkSpaces client login page. To learn more, visit https://docs.aws.amazon.com/workspaces/latest/adminguide/customize-branding.html

# v1.16.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.11.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-11-30)

* **Feature**: API client updated

# v1.10.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.9.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-09-30)

* **Feature**: API client updated

# v1.6.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-08-27)

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

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-25)

* **Feature**: API client updated

# v1.3.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

