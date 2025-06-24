# v1.66.1 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.66.0 (2025-06-11)

* **Feature**: Release for EKS Pod Identity Cross Account feature and disableSessionTags flag.

# v1.65.2 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.65.1 (2025-06-06)

* No change notes available for this release.

# v1.65.0 (2025-06-02)

* **Feature**: Add support for filtering ListInsights API calls on MISCONFIGURATION insight category

# v1.64.0 (2025-04-16)

* **Feature**: Added support for new AL2023 ARM64 NVIDIA AMIs to the supported AMITypes.

# v1.63.2 (2025-04-10)

* No change notes available for this release.

# v1.63.1 (2025-04-03)

* No change notes available for this release.

# v1.63.0 (2025-03-31)

* **Feature**: Add support for updating RemoteNetworkConfig for hybrid nodes on EKS UpdateClusterConfig API

# v1.62.0 (2025-03-27)

* **Feature**: Added support for BOTTLEROCKET FIPS AMIs to AMI types in US regions.

# v1.61.0 (2025-03-25)

* **Feature**: Added support to override upgrade-blocking readiness checks via force flag when updating a cluster.

# v1.60.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.60.0 (2025-02-28)

* **Feature**: Adding licenses to EKS Anywhere Subscription operations response.

# v1.59.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.58.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.58.0 (2025-02-07)

* **Feature**: Introduce versionStatus field to take place of status field in EKS DescribeClusterVersions API

# v1.57.4 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.57.3 (2025-02-04)

* No change notes available for this release.

# v1.57.2 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.57.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.57.0 (2025-01-24)

* **Feature**: Adds support for UpdateStrategies in EKS Managed Node Groups.
* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.56.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.56.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.3 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.56.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.1 (2025-01-08)

* No change notes available for this release.

# v1.56.0 (2024-12-23)

* **Feature**: This release adds support for DescribeClusterVersions API that provides important information about Kubernetes versions along with end of support dates

# v1.55.0 (2024-12-20)

* **Feature**: This release expands the catalog of upgrade insight checks

# v1.54.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.0 (2024-12-13)

* **Feature**: Add NodeRepairConfig in CreateNodegroupRequest and UpdateNodegroupConfigRequest

# v1.53.0 (2024-12-02)

* **Feature**: Added support for Auto Mode Clusters, Hybrid Nodes, and specifying computeTypes in the DescribeAddonVersions API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2024-11-08)

* **Feature**: Adds new error code `Ec2InstanceTypeDoesNotExist` for Amazon EKS managed node groups

# v1.51.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.51.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2024-10-21)

* **Feature**: This release adds support for Amazon Application Recovery Controller (ARC) zonal shift and zonal autoshift with EKS that enhances the resiliency of multi-AZ cluster environments

# v1.50.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.4 (2024-10-03)

* No change notes available for this release.

# v1.49.3 (2024-09-27)

* No change notes available for this release.

# v1.49.2 (2024-09-25)

* No change notes available for this release.

# v1.49.1 (2024-09-23)

* No change notes available for this release.

# v1.49.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.5 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.48.4 (2024-09-04)

* No change notes available for this release.

# v1.48.3 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.2 (2024-08-22)

* No change notes available for this release.

# v1.48.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2024-08-12)

* **Feature**: Added support for new AL2023 GPU AMIs to the supported AMITypes.

# v1.47.0 (2024-07-25)

* **Feature**: This release adds support for EKS cluster to manage extended support.

# v1.46.2 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.1 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.46.0 (2024-07-01)

* **Feature**: Updates EKS managed node groups to support EC2 Capacity Blocks for ML

# v1.45.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-06-26)

* **Feature**: Added support for disabling unmanaged addons during cluster creation.
* **Feature**: Support list-of-string endpoint parameter.

# v1.44.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2024-06-18)

* **Feature**: This release adds support to surface async fargate customer errors from async path to customer through describe-fargate-profile API response.
* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-06-03)

* **Feature**: Adds support for EKS add-ons pod identity associations integration
* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.5 (2024-05-23)

* No change notes available for this release.

# v1.42.4 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.3 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.2 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.42.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2024-03-28)

* **Feature**: Add multiple customer error code to handle customer caused failure when managing EKS node groups

# v1.41.2 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-02-29)

* **Feature**: Added support for new AL2023 AMIs to the supported AMITypes.

# v1.40.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.39.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.39.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.38.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2023-12-20)

* **Feature**: Add support for cluster insights, new EKS capability that surfaces potentially upgrade impacting issues.

# v1.36.0 (2023-12-18)

* **Feature**: Add support for EKS Cluster Access Management.

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

* **Feature**: This release adds support for EKS Pod Identity feature. EKS Pod Identity makes it easy for customers to obtain IAM permissions for the applications running in their EKS clusters.

# v1.33.2 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2023-11-09.2)

* **Feature**: Adding EKS Anywhere subscription related operations.

# v1.32.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2023-10-24)

* **Feature**: Added support for Cluster Subnet and Security Group mutability.

# v1.29.7 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.6 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.28.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2023-07-27)

* **Feature**: Add multiple customer error code to handle customer caused failure when managing EKS node groups

# v1.27.15 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.14 (2023-06-15)

* No change notes available for this release.

# v1.27.13 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.12 (2023-05-04)

* No change notes available for this release.

# v1.27.11 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.10 (2023-04-10)

* No change notes available for this release.

# v1.27.9 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.8 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.7 (2023-03-14)

* No change notes available for this release.

# v1.27.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.27.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.27.2 (2023-02-08)

* No change notes available for this release.

# v1.27.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.26.0 (2022-12-15)

* **Feature**: Add support for Windows managed nodes groups.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-12-07)

* **Feature**: Adds support for EKS add-ons configurationValues fields and DescribeAddonConfiguration function

# v1.24.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-11-29)

* **Feature**: Adds support for additional EKS add-ons metadata and filtering fields

# v1.23.0 (2022-11-16)

* **Feature**: Adds support for customer-provided placement groups for Kubernetes control plane instances when creating local EKS clusters on Outposts

# v1.22.4 (2022-11-07)

* No change notes available for this release.

# v1.22.3 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.2 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-09-14)

* **Feature**: Adding support for local Amazon EKS clusters on Outposts
* **Feature**: Adds support for EKS Addons ResolveConflicts "preserve" flag. Also adds new update failed status for EKS Addons.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.11 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.10 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.9 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.8 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.7 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.6 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-05-10)

* **Feature**: Adds BOTTLEROCKET_ARM_64_NVIDIA and BOTTLEROCKET_x86_64_NVIDIA AMI types to EKS managed nodegroups

# v1.20.7 (2022-05-03)

* No change notes available for this release.

# v1.20.6 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.5 (2022-04-12)

* No change notes available for this release.

# v1.20.4 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.3 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2022-03-08.3)

* No change notes available for this release.

# v1.20.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-01-07)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.

# v1.15.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-11-30)

* **Feature**: API client updated

# v1.14.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-12)

* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.12.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service to latest API model.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.2 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-09-10)

* **Feature**: API client updated

# v1.9.0 (2021-08-27)

* **Feature**: Updated API model to latest revision.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-07-15)

* **Feature**: Updated service model to latest version.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-11)

* **Feature**: Updated to latest API model.

# v1.5.0 (2021-05-20)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Feature**: Updated to latest service API model.
* **Dependency Update**: Updated to the latest SDK module versions

