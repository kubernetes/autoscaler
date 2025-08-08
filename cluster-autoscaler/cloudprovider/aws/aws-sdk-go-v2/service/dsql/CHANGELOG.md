# v1.5.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2025-05-22)

* **Feature**: Features: support for customer managed encryption keys

# v1.4.0 (2025-05-19)

* **Feature**: CreateMultiRegionCluster and DeleteMultiRegionCluster APIs removed

# v1.3.0 (2025-05-13)

* **Feature**: CreateMultiRegionClusters and DeleteMultiRegionClusters APIs marked as deprecated. Introduced new multi-Region clusters creation experience through multiRegionProperties parameter in CreateCluster API.

# v1.2.0 (2025-04-16)

* **Feature**: Added GetClusterEndpointService API. The new API allows retrieving endpoint service name specific to a cluster.

# v1.1.2 (2025-04-03)

* No change notes available for this release.

# v1.1.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.1.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.10 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.9 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.8 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.7 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.6 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.0.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.0.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.3 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.0.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2024-12-03.2)

* **Release**: New AWS service client module
* **Feature**: Add new API operations for Amazon Aurora DSQL. Amazon Aurora DSQL is a serverless, distributed SQL database with virtually unlimited scale, highest availability, and zero infrastructure management.

