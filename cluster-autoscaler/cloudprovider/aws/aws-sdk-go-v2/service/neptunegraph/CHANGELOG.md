# v1.17.5 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.4 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2025-04-03)

* No change notes available for this release.

# v1.17.2 (2025-03-07)

* **Documentation**: Several small updates to resolve customer requests.

# v1.17.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.17.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.2 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.1 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2025-02-04)

* **Feature**: Added argument to `list-export` to filter by graph ID

# v1.15.9 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.8 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.7 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.15.6 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.15.5 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.4 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.15.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-11-22)

* **Feature**: Add 4 new APIs to support new Export features, allowing Parquet and CSV formats. Add new arguments in Import APIs to support Parquet import. Add a new query "neptune.read" to run algorithms without loading data into database

# v1.14.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.14.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-10-10)

* **Feature**: Support for 16 m-NCU graphs available through account allowlisting

# v1.13.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.4 (2024-10-03)

* No change notes available for this release.

# v1.12.3 (2024-09-27)

* No change notes available for this release.

# v1.12.2 (2024-09-25)

* No change notes available for this release.

# v1.12.1 (2024-09-23)

* No change notes available for this release.

# v1.12.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.4 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.11.3 (2024-09-04)

* No change notes available for this release.

# v1.11.2 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2024-08-13)

* **Feature**: Amazon Neptune Analytics provides a new option for customers to load data into a graph using the RDF (Resource Description Framework) NTRIPLES format. When loading NTRIPLES files, use the value `convertToIri` for the `blankNodeHandling` parameter.

# v1.10.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.9.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.8 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.7 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.6 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.5 (2024-05-23)

* No change notes available for this release.

# v1.8.4 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.3 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.8.1 (2024-04-12)

* **Documentation**: Update to API documentation to resolve customer reported issues.

# v1.8.0 (2024-03-29)

* **Feature**: Add the new API Start-Import-Task for Amazon Neptune Analytics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-03-28)

* **Feature**: Update ImportTaskCancelled waiter to evaluate task state correctly and minor documentation changes.

# v1.6.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.5.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.5.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.4.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2024-02-12)

* **Feature**: Adding a new option "parameters" for data plane api ExecuteQuery to support running parameterized query via SDK.

# v1.2.0 (2024-02-01)

* **Feature**: Adding new APIs in SDK for Amazon Neptune Analytics. These APIs include operations to execute, cancel, list queries and get the graph summary.

# v1.1.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2023-12-21)

* **Feature**: Adds Waiters for successful creation and deletion of Graph, Graph Snapshot, Import Task and Private Endpoints for Neptune Analytics

# v1.0.0 (2023-12-14)

* **Release**: New AWS service client module
* **Feature**: This is the initial SDK release for Amazon Neptune Analytics

