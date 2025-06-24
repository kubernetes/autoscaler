# v1.30.2 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2025-04-25)

* **Feature**: You can now reference images and documents stored in Amazon S3 when using InvokeModel and Converse APIs with Amazon Nova Lite and Nova Pro. This enables direct integration of S3-stored multimedia assets in your model requests without manual downloading or base64 encoding.

# v1.29.0 (2025-04-07)

* **Feature**: New options for how to handle harmful content detected by Amazon Bedrock Guardrails.

# v1.28.1 (2025-04-03)

* No change notes available for this release.

# v1.28.0 (2025-03-31)

* **Feature**: Add Prompt Caching support to Converse and ConverseStream APIs

# v1.27.0 (2025-03-28)

* **Feature**: Launching Multi-modality Content Filter for Amazon Bedrock Guardrails.

# v1.26.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.26.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2025-02-24)

* **Feature**: This release adds Reasoning Content support to Converse and ConverseStream APIs

# v1.24.6 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.5 (2025-02-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.4 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.3 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.1 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.24.0 (2025-01-17)

* **Feature**: Allow hyphens in tool name for Converse and ConverseStream APIs
* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.23.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.2 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2024-12-04)

* **Feature**: Added support for Intelligent Prompt Router in Invoke, InvokeStream, Converse and ConverseStream. Add support for Bedrock Guardrails image content filter. New Bedrock Marketplace feature enabling a wider range of bedrock compatible models with self-hosted capability.

# v1.22.0 (2024-12-03.2)

* **Feature**: Added support for Async Invoke Operations Start, List and Get. Support for invocation logs with `requestMetadata` field in Converse, ConverseStream, Invoke and InvokeStream. Video content blocks in Converse/ConverseStream accept raw bytes or S3 URI.

# v1.21.0 (2024-12-03)

* **Feature**: Add an API parameter that allows customers to set performance configuration for invoking a model.

# v1.20.2 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2024-11-07)

* **Feature**: Add Prompt management support to Bedrock runtime APIs: Converse, ConverseStream, InvokeModel, InvokeModelWithStreamingResponse
* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.19.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2024-10-03)

* No change notes available for this release.

# v1.18.0 (2024-10-02)

* **Feature**: Added new fields to Amazon Bedrock Guardrails trace

# v1.17.3 (2024-09-27)

* No change notes available for this release.

# v1.17.2 (2024-09-25)

* No change notes available for this release.

# v1.17.1 (2024-09-23)

* No change notes available for this release.

# v1.17.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.3 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.16.2 (2024-09-04)

* No change notes available for this release.

# v1.16.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-08-29)

* **Feature**: Add support for imported-model in invokeModel and InvokeModelWithResponseStream.

# v1.15.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2024-07-25)

* **Feature**: Provides ServiceUnavailableException error message

# v1.14.0 (2024-07-10.2)

* **Feature**: Add support for contextual grounding check and ApplyGuardrail API for Guardrails for Amazon Bedrock.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-07-10)

* **Feature**: Add support for contextual grounding check and ApplyGuardrail API for Guardrails for Amazon Bedrock.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.11.0 (2024-06-20)

* **Feature**: This release adds document support to Converse and ConverseStream APIs

# v1.10.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2024-06-18)

* **Feature**: This release adds support for using Guardrails with the Converse and ConverseStream APIs.
* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.3 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.2 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2024-05-30)

* **Feature**: This release adds Converse and ConverseStream APIs to Bedrock Runtime

# v1.8.4 (2024-05-23)

* No change notes available for this release.

# v1.8.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.8.0 (2024-04-23)

* **Feature**: This release introduces Guardrails for Amazon Bedrock.

# v1.7.4 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.3 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.6.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.6.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.6 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.5.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.5.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2023-11-28.2)

* **Feature**: This release adds support for minor versions/aliases for invoke model identifier.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.3.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.2.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.0 (2023-10-02)

* **Feature**: Add model timeout exception for InvokeModelWithResponseStream API and update validator for invoke model identifier.

# v1.0.0 (2023-09-28)

* **Release**: New AWS service client module
* **Feature**: Run Inference: Added support to run the inference on models.  Includes set of APIs for running inference in streaming and non-streaming mode.

