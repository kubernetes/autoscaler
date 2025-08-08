# v1.44.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2025-06-02)

* **Feature**: This release adds the Agent Lifecycle Paused State feature to Amazon Bedrock agents. By using an agent's alias, you can temporarily suspend agent operations during maintenance, updates, or other situations.

# v1.43.0 (2025-05-15)

* **Feature**: Amazon Bedrock Flows introduces DoWhile loops nodes, parallel node executions, and enhancements to knowledge base nodes.

# v1.42.0 (2025-04-30)

* **Feature**: Features:    Add inline code node to prompt flow

# v1.41.0 (2025-04-03)

* **Feature**: Added optional "customMetadataField" for Amazon Aurora knowledge bases, allowing single-column metadata. Also added optional "textIndexName" for MongoDB Atlas knowledge bases, enabling hybrid search support.

# v1.40.0 (2025-03-25)

* **Feature**: Adding support for Amazon OpenSearch Managed clusters as a vector database in Knowledge Bases for Amazon Bedrock

# v1.39.0 (2025-03-10)

* **Feature**: Add support for computer use tools

# v1.38.0 (2025-03-07)

* **Feature**: Introduces support for Neptune Analytics as a vector data store and adds Context Enrichment Configurations, enabling use cases such as GraphRAG.

# v1.37.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.37.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2025-02-21)

* **Feature**: Introduce a new parameter which represents the user-agent header value used by the Bedrock Knowledge Base Web Connector.

# v1.35.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2025-02-12)

* **Feature**: This releases adds the additionalModelRequestFields field to the CreateAgent and UpdateAgent operations. Use additionalModelRequestFields to specify  additional inference parameters for a model beyond the base inference parameters.

# v1.34.3 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.1 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2025-01-27)

* **Feature**: Add support for the prompt caching feature for Bedrock Prompt Management

# v1.33.4 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.33.3 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.33.2 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2024-12-20)

* **Feature**: Support for custom user agent and max web pages crawled for web connector. Support app only credentials for SharePoint connector. Increase agents memory duration limit to 365 days. Support to specify max number of session summaries to include in agent invocation context.

# v1.32.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2024-12-04)

* **Feature**: This release introduces the ability to generate SQL using natural language, through a new GenerateQuery API (with native integration into Knowledge Bases); ability to ingest and retrieve images through Bedrock Data Automation; and ability to create a Knowledge Base backed by Kendra GenAI Index.

# v1.31.0 (2024-12-03.2)

* **Feature**: Releasing SDK for Multi-Agent Collaboration.

# v1.30.0 (2024-12-02)

* **Feature**: This release introduces APIs to upload documents directly into a Knowledge Base
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2024-11-27)

* **Feature**: Add support for specifying embeddingDataType, either FLOAT32 or BINARY

# v1.28.0 (2024-11-26)

* **Feature**: Custom Orchestration API release for AWSBedrockAgents.

# v1.27.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2024-11-07)

* **Feature**: Add prompt support for chat template configuration and agent generative AI resource. Add support for configuring an optional guardrail in Prompt and Knowledge Base nodes in Prompt Flows. Add API to validate flow definition
* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.26.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2024-11-01)

* **Feature**: Amazon Bedrock Knowledge Bases now supports using application inference profiles to increase throughput and improve resilience.

# v1.25.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2024-10-25)

* **Feature**: Add support of new model types for Bedrock Agents, Adding inference profile support for Flows and Prompt Management, Adding new field to configure additional inference configurations for Flows and Prompt Management

# v1.24.0 (2024-10-17)

* **Feature**: Removing support for topK property in PromptModelInferenceConfiguration object, Making PromptTemplateConfiguration property as required, Limiting the maximum PromptVariant to 1

# v1.23.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2024-10-03)

* No change notes available for this release.

# v1.22.0 (2024-10-01)

* **Feature**: This release adds support to stop an ongoing ingestion job using the StopIngestionJob API in Agents for Amazon Bedrock.

# v1.21.2 (2024-09-27)

* No change notes available for this release.

# v1.21.1 (2024-09-25)

* No change notes available for this release.

# v1.21.0 (2024-09-23)

* **Feature**: Amazon Bedrock Prompt Flows and Prompt Management now supports using inference profiles to increase throughput and improve resilience.

# v1.20.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.19.0 (2024-09-11)

* **Feature**: Amazon Bedrock Knowledge Bases now supports using inference profiles to increase throughput and improve resilience.

# v1.18.0 (2024-09-04)

* **Feature**: Add support for user metadata inside PromptVariant.

# v1.17.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2024-08-23)

* **Feature**: Releasing the support for Action User Confirmation.

# v1.16.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-07-10.2)

* **Feature**: Introduces new data sources and chunking strategies for Knowledge bases, advanced parsing logic using FMs, session summary generation, and code interpretation (preview) for Claude V3 Sonnet and Haiku models. Also introduces Prompt Flows (preview) to link prompts, foundational models, and resources.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-07-10)

* **Feature**: Introduces new data sources and chunking strategies for Knowledge bases, advanced parsing logic using FMs, session summary generation, and code interpretation (preview) for Claude V3 Sonnet and Haiku models. Also introduces Prompt Flows (preview) to link prompts, foundational models, and resources.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.13.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.3 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.2 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-05-30)

* **Feature**: With this release, Knowledge bases for Bedrock adds support for Titan Text Embedding v2.

# v1.11.1 (2024-05-23)

* No change notes available for this release.

# v1.11.0 (2024-05-20)

* **Feature**: This release adds support for using Guardrails with Bedrock Agents.

# v1.10.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.10.0 (2024-05-03)

* **Feature**: This release adds support for using Provisioned Throughput with Bedrock Agents.

# v1.9.0 (2024-05-01)

* **Feature**: This release adds support for using MongoDB Atlas as a vector store when creating a knowledge base.

# v1.8.0 (2024-04-23)

* **Feature**: Introducing the ability to create multiple data sources per knowledge base, specify S3 buckets as data sources from external accounts, and exposing levers to define the deletion behavior of the underlying vector store data.

# v1.7.0 (2024-04-22)

* **Feature**: Releasing the support for simplified configuration and return of control

# v1.6.0 (2024-04-16)

* **Feature**: For Create Agent API, the agentResourceRoleArn parameter is no longer required.

# v1.5.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2024-03-27)

* **Feature**: This changes introduces metadata documents statistics and also updates the documentation for bedrock agent.

# v1.4.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.3.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.3.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2023-12-21)

* **Feature**: This release introduces Amazon Aurora as a vector store on Knowledge Bases for Amazon Bedrock

# v1.1.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.1.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.1.2 (2023-12-01)

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
* **Feature**: This release introduces Agents for Amazon Bedrock
* **Dependency Update**: Updated to the latest SDK module versions

