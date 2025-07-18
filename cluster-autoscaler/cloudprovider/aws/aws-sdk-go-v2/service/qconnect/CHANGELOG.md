# v1.17.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2025-04-18)

* **Feature**: This release adds support for the following capabilities: Chunking generative answer replies from Amazon Q in Connect. Integration support for the use of additional LLM models with Amazon Q in Connect.

# v1.16.3 (2025-04-03)

* No change notes available for this release.

# v1.16.2 (2025-03-24)

* **Documentation**: Provides the correct value for supported model ID.

# v1.16.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.16.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.8 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.7 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.6 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.5 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.4 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.15.3 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.15.2 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-12-19)

* **Feature**: Amazon Q in Connect enables agents to ask Q for assistance in multiple languages and Q will provide answers and recommended step-by-step guides in those languages. Qs default language is English (United States) and you can switch this by setting the locale configuration on the AI Agent.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-12-02)

* **Feature**: This release adds following capabilities: Configuring safeguards via AIGuardrails for Q in Connect inferencing, and APIs to support Q&A self-service use cases
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-11-18)

* **Feature**: This release introduces MessageTemplate as a resource in Amazon Q in Connect, along with APIs to create, read, search, update, and delete MessageTemplate resources.
* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.12.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.2 (2024-10-09)

* No change notes available for this release.

# v1.12.1 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-10-07)

* **Feature**: This release adds support for the following capabilities: Configuration of the Gen AI system via AIAgent and AIPrompts. Integration support for Bedrock Knowledge Base.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.4 (2024-10-03)

* No change notes available for this release.

# v1.10.3 (2024-09-27)

* No change notes available for this release.

# v1.10.2 (2024-09-25)

* No change notes available for this release.

# v1.10.1 (2024-09-23)

* No change notes available for this release.

# v1.10.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.7 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.9.6 (2024-09-04)

* No change notes available for this release.

# v1.9.5 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.4 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2024-06-27)

* **Feature**: Adds CreateContentAssociation, ListContentAssociations, GetContentAssociation, and DeleteContentAssociation APIs.

# v1.8.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.7.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.4 (2024-05-23)

* No change notes available for this release.

# v1.6.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.6.0 (2024-04-10)

* **Feature**: This release adds a new QiC public API updateSession and updates an existing QiC public API createSession

# v1.5.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.4.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.4.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2024-01-29)

* No change notes available for this release.

# v1.3.0 (2024-01-10)

* **Feature**: QueryAssistant and GetRecommendations will be discontinued starting June 1, 2024. To receive generative responses after March 1, 2024 you will need to create a new Assistant in the Connect console and integrate the Amazon Q in Connect JavaScript library (amazon-q-connectjs) into your applications.

# v1.2.4 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.3 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.2.2 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.2.0 (2023-12-01)

* **Feature**: This release adds the PutFeedback API and allows providing feedback against the specified assistant for the specified target.
* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.0.0 (2023-11-28.2)

* **Release**: New AWS service client module
* **Feature**: Amazon Q in Connect, an LLM-enhanced evolution of Amazon Connect Wisdom. This release adds generative AI support to Amazon Q Connect QueryAssistant and GetRecommendations APIs.
* **Dependency Update**: Updated to the latest SDK module versions

