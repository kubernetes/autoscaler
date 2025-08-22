# v1.10.5 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.4 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.3 (2025-04-03)

* No change notes available for this release.

# v1.10.2 (2025-03-10)

* **Documentation**: This release updates the default value of pprof-disabled from false to true.

# v1.10.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.10.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2025-02-17)

* **Feature**: This release introduces APIs to manage DbClusters and adds support for read replicas

# v1.8.3 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2025-01-28)

* **Feature**: Adds 'allocatedStorage' parameter to UpdateDbInstance API that allows increasing the database instance storage size and 'dbStorageType' parameter to UpdateDbInstance API that allows changing the storage type of the database instance

# v1.7.5 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.7.4 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.7.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-12-11)

* **Feature**: Adds networkType parameter to CreateDbInstance API which allows IPv6 support to the InfluxDB endpoint

# v1.6.7 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.5.1 (2024-10-03)

* No change notes available for this release.

# v1.5.0 (2024-09-30)

* **Feature**: Timestream for InfluxDB now supports port configuration and additional customer-modifiable InfluxDB v2 parameters. This release adds Port to the CreateDbInstance and UpdateDbInstance API, and additional InfluxDB v2 parameters to the CreateDbParameterGroup API.

# v1.4.3 (2024-09-27)

* No change notes available for this release.

# v1.4.2 (2024-09-25)

* No change notes available for this release.

# v1.4.1 (2024-09-23)

* No change notes available for this release.

# v1.4.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.2 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.3.1 (2024-09-04)

* No change notes available for this release.

# v1.3.0 (2024-09-03)

* **Feature**: Timestream for InfluxDB now supports compute scaling and deployment type conversion. This release adds the DbInstanceType and DeploymentType parameters to the UpdateDbInstance API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

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

# v1.0.9 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.8 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.7 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.6 (2024-05-23)

* No change notes available for this release.

# v1.0.5 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.4 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.3 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.0.2 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2024-03-14)

* **Release**: New AWS service client module
* **Feature**: This is the initial SDK release for Amazon Timestream for InfluxDB. Amazon Timestream for InfluxDB is a new time-series database engine that makes it easy for application developers and DevOps teams to run InfluxDB databases on AWS for near real-time time-series applications using open source APIs.

