# v1.13.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2025-05-27)

* **Feature**: AWS Deadline Cloud service-managed fleets now support storage profiles. With storage profiles, you can map file paths between a workstation and the worker hosts running the job.

# v1.12.0 (2025-05-12)

* **Feature**: AWS Deadline Cloud service-managed fleets now support configuration scripts. Configuration scripts make it easy to install additional software, like plugins and packages, onto a worker.

# v1.11.0 (2025-04-30)

* **Feature**: Adds support for tag management on workers and tag inheritance from fleets to their associated workers.

# v1.10.1 (2025-04-03)

* No change notes available for this release.

# v1.10.0 (2025-03-31)

* **Feature**: With this release you can use a new field to specify the search term match type. Search term match types currently support fuzzy and contains matching.

# v1.9.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.9.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.4 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.3 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2025-01-28)

* **Feature**: feature: Deadline: Add support for limiting the concurrent usage of external resources, like floating licenses, using limits and the ability to constrain the maximum number of workers that work on a job

# v1.7.8 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.7.7 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.7.6 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.5 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.7.4 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-11-14)

* **Feature**: Adds support for select GPU accelerated instance types when creating new service-managed fleets.

# v1.6.4 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.6.3 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-10-07)

* **Feature**: Add support for using the template from a previous job during job creation and listing parameter definitions for a job.
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

# v1.3.3 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.3.2 (2024-09-04)

* No change notes available for this release.

# v1.3.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2024-08-19)

* **Feature**: This release adds additional search fields and provides sorting by multiple fields.

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

# v1.0.0 (2024-04-01)

* **Release**: New AWS service client module
* **Feature**: AWS Deadline Cloud is a new fully managed service that helps customers set up, deploy, and scale rendering projects in minutes, so they can improve the efficiency of their rendering pipelines and take on more projects.

