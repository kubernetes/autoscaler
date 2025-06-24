# v1.35.1 (2025-06-17)

* **Documentation**: This release of the SDK has the API and documentation for the createcustommodel API. This feature lets you copy a trained model into Amazon Bedrock for inference.
* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.0 (2025-06-16)

* **Feature**: This release of the SDK has the API and documentation for the createcustommodel API. This feature lets you copy a Amazon SageMaker trained Amazon Nova model into Amazon Bedrock for inference.

# v1.34.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2025-05-13)

* **Feature**: Enable cross-Region inference for Amazon Bedrock Guardrails by using the crossRegionConfig parameter when calling the CreateGuardrail or UpdateGuardrail operation.

# v1.33.0 (2025-04-30)

* **Feature**: You can now specify a cross region inference profile as a teacher model for the CreateModelCustomizationJob API. Additionally, the GetModelCustomizationJob API has been enhanced to return the sub-task statuses of a customization job within the StatusDetails response field.

# v1.32.0 (2025-04-17)

* **Feature**: With this release, Bedrock Evaluation will now support custom metrics for evaluation.

# v1.31.0 (2025-04-07)

* **Feature**: New options for how to handle harmful content detected by Amazon Bedrock Guardrails.

# v1.30.1 (2025-04-03)

* No change notes available for this release.

# v1.30.0 (2025-03-21)

* **Feature**: A CustomModelUnit(CMU) is an abstract view of the hardware utilization that Bedrock needs to host a a single copy of your custom imported model. Bedrock determines the number of CMUs that a model copy needs when you import the custom model. You can use CMUs to estimate the cost of Inference's.

# v1.29.0 (2025-03-20)

* **Feature**: With this release, Bedrock Evaluation will now support bring your own inference responses.

# v1.28.1 (2025-03-14)

* No change notes available for this release.

# v1.28.0 (2025-03-06)

* **Feature**: This releases adds support for Custom Prompt Router

# v1.27.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.27.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.8 (2025-02-20)

* No change notes available for this release.

# v1.26.7 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.6 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.5 (2025-01-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.4 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.26.2 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.26.1 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2025-01-13)

* **Feature**: With this release, Bedrock Evaluation will now support latency-optimized inference for foundation models.

# v1.25.3 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2024-12-09)

* No change notes available for this release.

# v1.25.0 (2024-12-04)

* **Feature**: Introduced two APIs ListPromptRouters and GetPromptRouter for Intelligent Prompt Router feature. Add support for Bedrock Guardrails image content filter. New Bedrock Marketplace feature enabling a wider range of bedrock compatible models with self-hosted capability.

# v1.24.0 (2024-12-03.2)

* **Feature**: Tagging support for Async Invoke resources. Added support for Distillation in CreateModelCustomizationJob API. Support for videoDataDeliveryEnabled flag in invocation logging.

# v1.23.0 (2024-12-02)

* **Feature**: Add support for Knowledge Base Evaluations & LLM as a judge
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.4 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.3 (2024-11-11)

* No change notes available for this release.

# v1.22.2 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.22.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2024-10-29)

* **Feature**: Update Application Inference Profile

# v1.21.2 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2024-10-23)

* **Documentation**: Doc updates for supporting converse

# v1.21.0 (2024-10-18)

* **Feature**: Adding converse support to CMI API's

# v1.20.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.4 (2024-10-03)

* No change notes available for this release.

# v1.19.3 (2024-10-02)

* No change notes available for this release.

# v1.19.2 (2024-09-27)

* No change notes available for this release.

# v1.19.1 (2024-09-25)

* No change notes available for this release.

# v1.19.0 (2024-09-24)

* **Feature**: Add support for Cross Region Inference in Bedrock Model Evaluations.

# v1.18.1 (2024-09-23)

* No change notes available for this release.

# v1.18.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.17.0 (2024-09-16)

* **Feature**: This feature adds cross account s3 bucket and VPC support to ModelInvocation jobs. To use a cross account bucket, pass in the accountId of the bucket to s3BucketOwner in the ModelInvocationJobInputDataConfig or ModelInvocationJobOutputDataConfig.

# v1.16.2 (2024-09-04)

* No change notes available for this release.

# v1.16.1 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2024-08-27)

* **Feature**: Amazon Bedrock SDK updates for Inference Profile.

# v1.15.0 (2024-08-22)

* **Feature**: Amazon Bedrock Evaluation BatchDeleteEvaluationJob API allows customers to delete evaluation jobs under terminated evaluation job statuses - Stopped, Failed, or Completed. Customers can submit a batch of 25 evaluation jobs to be deleted at once.

# v1.14.0 (2024-08-19)

* **Feature**: Amazon Bedrock Batch Inference/ Model Invocation is a feature which allows customers to asynchronously run inference on a large set of records/files stored in S3.

# v1.13.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2024-08-01)

* **Feature**: API and Documentation for Bedrock Model Copy feature. This feature lets you share and copy a custom model from one region to another or one account to another.

# v1.12.0 (2024-07-10.2)

* **Feature**: Add support for contextual grounding check for Guardrails for Amazon Bedrock.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2024-07-10)

* **Feature**: Add support for contextual grounding check for Guardrails for Amazon Bedrock.
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

# v1.8.9 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.8 (2024-06-11)

* No change notes available for this release.

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

# v1.8.1 (2024-05-01)

* No change notes available for this release.

# v1.8.0 (2024-04-23)

* **Feature**: This release introduces Model Evaluation and Guardrails for Amazon Bedrock.

# v1.7.7 (2024-04-10)

* No change notes available for this release.

# v1.7.6 (2024-04-04)

* No change notes available for this release.

# v1.7.5 (2024-04-02)

* No change notes available for this release.

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

# v1.5.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.6 (2023-12-22)

* No change notes available for this release.

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

* **Feature**: This release adds support for customization types, model life cycle status and minor versions/aliases for model identifiers.
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

# v1.1.5 (2023-10-23)

* No change notes available for this release.

# v1.1.4 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.3 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.1.2 (2023-10-05)

* No change notes available for this release.

# v1.1.1 (2023-10-04)

* No change notes available for this release.

# v1.1.0 (2023-10-02)

* **Feature**: Provisioned throughput feature with Amazon and third-party base models, and update validators for model identifier and taggable resource ARNs.

# v1.0.0 (2023-09-28)

* **Release**: New AWS service client module
* **Feature**: Model Invocation logging added to enable or disable logs in customer account. Model listing and description support added. Provisioned Throughput feature added. Custom model support added for creating custom models. Also includes list, and delete functions for custom model.

