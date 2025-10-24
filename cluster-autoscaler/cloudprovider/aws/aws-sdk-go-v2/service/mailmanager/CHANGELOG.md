# v1.13.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2025-04-30)

* **Feature**: Introducing new RuleSet rule PublishToSns action, which allows customers to publish email notifications to an Amazon SNS topic. New PublishToSns action enables customers to easily integrate their email workflows via Amazon SNS, allowing them to notify other systems about important email events.

# v1.12.0 (2025-04-03)

* **Feature**: Add support for Dual_Stack and PrivateLink types of IngressPoint. For configuration requests, SES Mail Manager will now accept both IPv4/IPv6 dual-stack endpoints and AWS PrivateLink VPC endpoints for email receiving.

# v1.11.0 (2025-03-20)

* **Feature**: Amazon SES Mail Manager. Extended rule string and boolean expressions to support analysis in condition evaluation. Extended ingress point string expression to support analysis in condition evaluation

# v1.10.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.10.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2025-02-19)

* **Feature**: This release adds additional metadata fields in Mail Manager archive searches to show email source and details about emails that were archived when being sent with SES.

# v1.8.4 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.3 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2025-01-29)

* **Feature**: This release includes a new feature for Amazon SES Mail Manager which allows customers to specify known addresses and domains and make use of those in traffic policies and rules actions to distinguish between known and unknown entries.

# v1.7.6 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.7.5 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.7.4 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-11-22)

* **Feature**: Added new "DeliverToQBusiness" rule action to MailManager RulesSet for ingesting email data into Amazon Q Business customer applications

# v1.6.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.6.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-10-14)

* **Feature**: Mail Manager support for viewing and exporting metadata of archived messages.

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

# v1.3.0 (2024-09-18)

* **Feature**: Introduce a new RuleSet condition evaluation, where customers can set up a StringExpression with a MimeHeader condition. This condition will perform the necessary validation based on the X-header provided by customers.

# v1.2.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.2.6 (2024-09-04)

* No change notes available for this release.

# v1.2.5 (2024-09-03)

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

# v1.0.4 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.3 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.2 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.1 (2024-05-23)

* No change notes available for this release.

# v1.0.0 (2024-05-21)

* **Release**: New AWS service client module
* **Feature**: This release includes a new Amazon SES feature called Mail Manager, which is a set of email gateway capabilities designed to help customers strengthen their organization's email infrastructure, simplify email workflow management, and streamline email compliance control.

