# v1.13.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2025-05-12)

* **Feature**: Launch new AWS Supply Chain public APIs for DataIntegrationEvent, DataIntegrationFlowExecution and DatasetNamespace. Also add more capabilities to existing public APIs to support direct dataset event publish, data deduplication in DataIntegrationFlow, partition specification of custom datasets.

# v1.12.2 (2025-04-03)

* No change notes available for this release.

# v1.12.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.12.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.8 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.7 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.6 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.5 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.4 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.11.3 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.11.2 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2025-01-06)

* **Feature**: Allow vanity DNS domain when creating a new ASC instance

# v1.10.6 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.5 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.10.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-10-25)

* **Feature**: API doc updates, and also support showing error message on a failed instance

# v1.9.0 (2024-10-14)

* **Feature**: This release adds AWS Supply Chain instance management functionality. Specifically adding CreateInstance, DeleteInstance, GetInstance, ListInstances, and UpdateInstance APIs.

# v1.8.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-10-03)

* No change notes available for this release.

# v1.7.0 (2024-09-30)

* **Feature**: Release DataLakeDataset, DataIntegrationFlow and ResourceTagging APIs for AWS Supply Chain

# v1.6.3 (2024-09-27)

* No change notes available for this release.

# v1.6.2 (2024-09-25)

* No change notes available for this release.

# v1.6.1 (2024-09-23)

* No change notes available for this release.

# v1.6.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.8 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.5.7 (2024-09-04)

* No change notes available for this release.

# v1.5.6 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.5 (2024-08-23)

* **Documentation**: Update API documentation to clarify the event SLA as well as the data model expectations

# v1.5.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.4.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.4 (2024-05-23)

* No change notes available for this release.

# v1.3.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.3.0 (2024-04-10)

* **Feature**: This release includes API SendDataIntegrationEvent for AWS Supply Chain

# v1.2.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.1.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.1.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2024-01-12)

* **Release**: New AWS service client module
* **Feature**: This release includes APIs CreateBillOfMaterialsImportJob and GetBillOfMaterialsImportJob.

