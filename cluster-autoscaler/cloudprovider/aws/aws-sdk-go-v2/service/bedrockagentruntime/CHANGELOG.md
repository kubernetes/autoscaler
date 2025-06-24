# v1.45.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2025-06-06)

* **Feature**: This release introduces the `PromptCreationConfigurations` input parameter, which includes fields to control prompt population for `InvokeAgent` or `InvokeInlineAgent` requests.

# v1.44.0 (2025-05-21)

* **Feature**: Amazon Bedrock introduces asynchronous flows (in preview), which let you run flows for longer durations and yield control so that your application can perform other tasks and you don't have to actively monitor the flow's progress.

# v1.43.0 (2025-05-13)

* **Feature**: Changes for enhanced metadata in trace

# v1.42.0 (2025-04-30)

* **Feature**: Support for Custom Orchestration within InlineAgents

# v1.41.1 (2025-04-03)

* No change notes available for this release.

# v1.41.0 (2025-03-27)

* **Feature**: bedrock flow now support node action trace.

# v1.40.0 (2025-03-10)

* **Feature**: Add support for computer use tools

# v1.39.0 (2025-03-07)

* **Feature**: Support Multi Agent Collaboration within Inline Agents

# v1.38.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.38.0 (2025-02-27)

* **Feature**: Introduces Sessions (preview) to enable stateful conversations in GenAI applications.
* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.0 (2025-02-24)

* **Feature**: Adding support for ReasoningContent fields in Pre-Processing, Post-Processing and Orchestration Trace outputs.

# v1.36.2 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.1 (2025-02-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2025-02-12)

* **Feature**: This releases adds the additionalModelRequestFields field to the InvokeInlineAgent operation. Use additionalModelRequestFields to specify  additional inference parameters for a model beyond the base inference parameters.

# v1.35.1 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2025-01-31)

* **Feature**: This change is to deprecate the existing citation field under RetrieveAndGenerateStream API response in lieu of GeneratedResponsePart and RetrievedReferences
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2025-01-30)

* **Feature**: Add a 'reason' field to InternalServerException
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.33.0 (2025-01-22)

* **Feature**: Adds multi-turn input support for an Agent node in an Amazon Bedrock Flow

# v1.32.1 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.32.0 (2025-01-15)

* **Feature**: Now supports streaming for inline agents.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.0 (2024-12-20)

* **Feature**: bedrock agents now supports long term memory and performance configs. Invokeflow supports performance configs. RetrieveAndGenerate performance configs

# v1.30.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2024-12-04)

* **Feature**: This release introduces the ability to generate SQL using natural language, through a new GenerateQuery API (with native integration into Knowledge Bases); ability to ingest and retrieve images through Bedrock Data Automation; and ability to create a Knowledge Base backed by Kendra GenAI Index.

# v1.29.0 (2024-12-03.2)

* **Feature**: Releasing SDK for multi agent collaboration

# v1.28.0 (2024-12-02)

* **Feature**: This release introduces a new Rerank API to leverage reranking models (with integration into Knowledge Bases); APIs to upload documents directly into Knowledge Base; RetrieveAndGenerateStream API for streaming response; Guardrails on Retrieve API; and ability to automatically generate filters
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2024-11-26)

* **Feature**: Custom Orchestration and Streaming configurations API release for AWSBedrockAgents.

# v1.26.0 (2024-11-22)

* **Feature**: InvokeInlineAgent API release to help invoke runtime agents without any dependency on preconfigured agents.

# v1.25.0 (2024-11-20)

* **Feature**: Releasing new Prompt Optimization to enhance your prompts for improved performance

# v1.24.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2024-11-08)

* **Feature**: This release adds trace functionality to Bedrock Prompt Flows

# v1.23.3 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.23.2 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2024-10-21)

* **Feature**: Knowledge Bases for Amazon Bedrock now supports custom prompts and model parameters in the orchestrationConfiguration of the RetrieveAndGenerate API. The modelArn field accepts Custom Models and Imported Models ARNs.

# v1.22.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2024-10-03)

* No change notes available for this release.

# v1.21.0 (2024-10-02)

* **Feature**: Added raw model response and usage metrics to PreProcessing and PostProcessing Trace

# v1.20.3 (2024-09-27)

* No change notes available for this release.

# v1.20.2 (2024-09-25)

* No change notes available for this release.

# v1.20.1 (2024-09-23)

* No change notes available for this release.

# v1.20.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.19.0 (2024-09-11)

* **Feature**: Amazon Bedrock Knowledge Bases now supports using inference profiles to increase throughput and improve resilience.

# v1.18.2 (2024-09-04)

* No change notes available for this release.

# v1.18.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2024-08-29)

* **Feature**: Lifting the maximum length on Bedrock KnowledgeBase RetrievalFilter array

# v1.17.0 (2024-08-23)

* **Feature**: Releasing the support for Action User Confirmation.

# v1.16.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-08-06)

* **Feature**: Introduce model invocation output traces for orchestration traces, which contain the model's raw response and usage.

# v1.15.0 (2024-07-10.2)

* **Feature**: Introduces query decomposition, enhanced Agents integration with Knowledge bases, session summary generation, and code interpretation (preview) for Claude V3 Sonnet and Haiku models. Also introduces Prompt Flows (preview) to link prompts, foundational models, and resources for end-to-end solutions.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2024-07-10)

* **Feature**: Introduces query decomposition, enhanced Agents integration with Knowledge bases, session summary generation, and code interpretation (preview) for Claude V3 Sonnet and Haiku models. Also introduces Prompt Flows (preview) to link prompts, foundational models, and resources for end-to-end solutions.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.12.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.4 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.3 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.2 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.1 (2024-05-23)

* No change notes available for this release.

# v1.11.0 (2024-05-20)

* **Feature**: This release adds support for using Guardrails with Bedrock Agents.

# v1.10.1 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-05-15)

* **Feature**: Updating Bedrock Knowledge Base Metadata & Filters feature with two new filters listContains and stringContains
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2024-05-09)

* **Feature**: This release adds support to provide guardrail configuration and modify inference parameters that are then used in RetrieveAndGenerate API in Agents for Amazon Bedrock.

# v1.8.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.8.0 (2024-04-23)

* **Feature**: This release introduces zero-setup file upload support for the RetrieveAndGenerate API. This allows you to chat with your data without setting up a Knowledge Base.

# v1.7.0 (2024-04-22)

* **Feature**: Releasing the support for simplified configuration and return of control

# v1.6.1 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2024-03-27)

* **Feature**: This release introduces filtering support on Retrieve and RetrieveAndGenerate APIs.

# v1.5.0 (2024-03-26)

* **Feature**: This release adds support to customize prompts sent through the RetrieveAndGenerate API in Agents for Amazon Bedrock.

# v1.4.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.2 (2024-03-08)

* **Documentation**: Documentation update for Bedrock Runtime Agent

# v1.4.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2024-02-28)

* **Feature**: This release adds support to override search strategy performed by the Retrieve and RetrieveAndGenerate APIs for Amazon Bedrock Agents

# v1.3.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.2.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.2.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

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
* **Feature**: This release introduces Agents for Amazon Bedrock Runtime
* **Dependency Update**: Updated to the latest SDK module versions

