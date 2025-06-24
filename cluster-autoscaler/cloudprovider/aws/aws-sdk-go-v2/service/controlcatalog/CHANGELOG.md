# v1.9.1 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2025-06-11)

* **Feature**: Introduced ListControlMappings API that retrieves control mappings. Added control aliases and governed resources fields in GetControl and ListControls APIs. New filtering capability in ListControls API, with implementation identifiers and implementation types.

# v1.8.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2025-04-09)

* **Feature**: The GetControl API now surfaces a control's Severity, CreateTime, and Identifier for a control's Implementation. The ListControls API now surfaces a control's Behavior, Severity, CreateTime, and Identifier for a control's Implementation.

# v1.7.3 (2025-04-03)

* No change notes available for this release.

# v1.7.2 (2025-03-20)

* **Documentation**: Add ExemptAssumeRoot parameter to adapt for new AWS AssumeRoot capability.

# v1.7.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.7.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.12 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.11 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.10 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.9 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.8 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.6.7 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.6.6 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.5 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.4 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.3 (2024-12-11)

* **Documentation**: Minor documentation updates to the content of ImplementationDetails object part of the Control Catalog GetControl API

# v1.6.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-11-08)

* **Feature**: AWS Control Catalog GetControl public API returns additional data in output, including Implementation and Parameters

# v1.5.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.5.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.4 (2024-10-03)

* No change notes available for this release.

# v1.4.3 (2024-09-27)

* No change notes available for this release.

# v1.4.2 (2024-09-25)

* No change notes available for this release.

# v1.4.1 (2024-09-23)

* No change notes available for this release.

# v1.4.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.4 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.3.3 (2024-09-04)

* No change notes available for this release.

# v1.3.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2024-08-01)

* **Feature**: AWS Control Tower provides two new public APIs controlcatalog:ListControls and controlcatalog:GetControl under controlcatalog service namespace, which enable customers to programmatically retrieve control metadata of available controls.

# v1.2.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.1.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.4 (2024-05-23)

* No change notes available for this release.

# v1.0.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.0.0 (2024-04-08)

* **Release**: New AWS service client module
* **Feature**: This is the initial SDK release for AWS Control Catalog, a central catalog for AWS managed controls. This release includes 3 new APIs - ListDomains, ListObjectives, and ListCommonControls - that vend high-level data to categorize controls across the AWS platform.

