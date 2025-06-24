# v1.11.3 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.2 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2025-04-03)

* No change notes available for this release.

# v1.11.0 (2025-04-02)

* **Feature**: Application Signals now supports creating Service Level Objectives on service dependencies. Users can now create or update SLOs on discovered service dependencies to monitor their standard application metrics.

# v1.10.0 (2025-03-17)

* **Feature**: This release adds support for adding, removing, and listing SLO time exclusion windows with the BatchUpdateExclusionWindows and ListServiceLevelObjectiveExclusionWindows APIs.

# v1.9.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.9.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2025-02-26)

* **Feature**: This release adds API support for reading Service Level Objectives and Services from monitoring accounts, from SLO and Service-scoped operations, including ListServices and ListServiceLevelObjectives.

# v1.7.11 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.10 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.9 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.8 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.7 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.7.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.7.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.4 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-11-13)

* **Feature**: Amazon CloudWatch Application Signals now supports creating Service Level Objectives with burn rates. Users can now create or update SLOs with burn rate configurations to meet their specific business requirements.

# v1.6.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.6.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.4 (2024-10-03)

* No change notes available for this release.

# v1.5.3 (2024-09-27)

* No change notes available for this release.

# v1.5.2 (2024-09-25)

* No change notes available for this release.

# v1.5.1 (2024-09-23)

* No change notes available for this release.

# v1.5.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.4.0 (2024-09-05)

* **Feature**: Amazon CloudWatch Application Signals now supports creating Service Level Objectives using a new calculation type. Users can now create SLOs which are configured with request-based SLIs to help meet their specific business requirements.

# v1.3.3 (2024-09-04)

* No change notes available for this release.

# v1.3.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2024-07-25)

* **Feature**: CloudWatch Application Signals now supports application logs correlation with traces and operational health metrics of applications running on EC2 instances. Users can view the most relevant telemetry to troubleshoot application health anomalies such as spikes in latency, errors, and availability.

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

# v1.0.1 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2024-06-10)

* **Release**: New AWS service client module
* **Feature**: This is the initial SDK release for Amazon CloudWatch Application Signals. Amazon CloudWatch Application Signals provides curated application performance monitoring for developers to monitor and troubleshoot application health using pre-built dashboards and Service Level Objectives.

