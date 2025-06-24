# v1.196.1 (2025-06-17)

* **Dependency Update**: Update to smithy-go v1.22.4.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.196.0 (2025-06-16)

* **Feature**: This release 1) adds a new S3DataType Converse for SageMaker training 2)adds C8g R7gd M8g C6in P6 P6e instance type for SageMaker endpoint 3) adds m7i, r7i, c7i instance type for SageMaker Training and Processing.

# v1.195.1 (2025-06-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.195.0 (2025-06-04)

* **Feature**: Added support for p6-b200 instance type in SageMaker Training Jobs and Training Plans.

# v1.194.0 (2025-05-30)

* **Feature**: Release new parameter CapacityReservationConfig in ProductionVariant

# v1.193.0 (2025-05-29)

* **Feature**: Add maintenance status field to DescribeMlflowTrackingServer API response

# v1.192.0 (2025-05-12)

* **Feature**: No API changes from previous release. This release migrated the model to Smithy keeping all features unchanged.

# v1.191.0 (2025-05-07)

* **Feature**: SageMaker AI Studio users can now migrate to SageMaker Unified Studio, which offers a unified web-based development experience that integrates AWS data, analytics, artificial intelligence (AI), and machine learning (ML) services, as well as additional tools and resource

# v1.190.0 (2025-05-01)

* **Feature**: Feature - Adding support for Scheduled and Rolling Update Software in Sagemaker Hyperpod.

# v1.189.0 (2025-04-29)

* **Feature**: Introduced support for P5en instance types on SageMaker Studio for JupyterLab and CodeEditor applications.

# v1.188.0 (2025-04-18)

* **Feature**: This release adds a new Neuron driver option in InferenceAmiVersion parameter for ProductionVariant. Additionally, it adds support for fetching model lifecycle status in the ListModelPackages API. Users can now use this API to view the lifecycle stage of models that have been shared with them.

# v1.187.0 (2025-04-03)

* **Feature**: Adds support for i3en, m7i, r7i instance types for SageMaker Hyperpod

# v1.186.0 (2025-04-01)

* **Feature**: Added tagging support for SageMaker notebook instance lifecycle configurations

# v1.185.1 (2025-03-31)

* No change notes available for this release.

# v1.185.0 (2025-03-28)

* **Feature**: TransformAmiVersion for Batch Transform and SageMaker Search Service Aggregate Search API Extension

# v1.184.0 (2025-03-27)

* **Feature**: add: recovery mode for SageMaker Studio apps

# v1.183.0 (2025-03-25)

* **Feature**: This release adds support for customer-managed KMS keys in Amazon SageMaker Partner AI Apps

# v1.182.0 (2025-03-21)

* **Feature**: This release does the following: 1.) Adds DurationHours as a required field to the SearchTrainingPlanOfferings action in the SageMaker AI API; 2.) Adds support for G6e instance types for SageMaker AI inference optimization jobs.

# v1.181.0 (2025-03-19)

* **Feature**: Added support for g6, g6e, m6i, c6i instance types in SageMaker Processing Jobs.

# v1.180.2 (2025-03-11)

* No change notes available for this release.

# v1.180.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.180.0 (2025-03-04)

* **Feature**: Add DomainId to CreateDomainResponse

# v1.179.0 (2025-02-27)

* **Feature**: SageMaker HubService is introducing support for creating Training Jobs in Curated Hub (Private Hub). Additionally, it is introducing two new APIs: UpdateHubContent and UpdateHubContentReference.
* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.178.0 (2025-02-26)

* **Feature**: AWS SageMaker InferenceComponents now support rolling update deployments for Inference Components.

# v1.177.0 (2025-02-20)

* **Feature**: Added new capability in the UpdateCluster operation to remove instance groups from your SageMaker HyperPod cluster.

# v1.176.0 (2025-02-19)

* **Feature**: Adds r8g instance type support to SageMaker Realtime Endpoints

# v1.175.1 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.175.0 (2025-02-13)

* **Feature**: Adds additional values to the InferenceAmiVersion parameter in the ProductionVariant data type.

# v1.174.3 (2025-02-06)

* No change notes available for this release.

# v1.174.2 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.174.1 (2025-02-04)

* **Documentation**: IPv6 support for Hyperpod clusters

# v1.174.0 (2025-01-31)

* **Feature**: This release introduces a new valid value in InstanceType parameter: p5en.48xlarge, in ProductionVariant.
* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.173.3 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.173.2 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.173.1 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.
* **Documentation**: Correction of docs for  "Added support for ml.trn1.32xlarge instance type in Reserved Capacity Offering"

# v1.173.0 (2025-01-16)

* **Feature**: Added support for ml.trn1.32xlarge instance type in Reserved Capacity Offering

# v1.172.3 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.172.2 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.172.1 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.172.0 (2025-01-08)

* **Feature**: Adds support for IPv6 for SageMaker HyperPod cluster nodes.

# v1.171.0 (2025-01-02)

* **Feature**: Adding ETag information with Model Artifacts for Model Registry

# v1.170.0 (2024-12-20)

* **Feature**: This release adds support for c6i, m6i and r6i instance on SageMaker Hyperpod and trn1 instances in batch

# v1.169.1 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.169.0 (2024-12-04)

* **Feature**: Amazon SageMaker HyperPod launched task governance to help customers maximize accelerator utilization for model development and flexible training plans to meet training timelines and budget while reducing weeks of training time. AI apps from AWS partner is now available in SageMaker.

# v1.168.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.168.0 (2024-11-22)

* **Feature**: This release adds APIs for new features for SageMaker endpoint to scale down to zero instances, native support for multi-adapter inference, and endpoint scaling improvements.

# v1.167.1 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.167.0 (2024-11-14)

* **Feature**: Add support for Neuron instance types [ trn1/trn1n/inf2 ] on SageMaker Notebook Instances Platform.

# v1.166.2 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.166.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.166.0 (2024-10-31)

* **Feature**: SageMaker HyperPod adds scale-down at instance level via BatchDeleteClusterNodes API and group level via UpdateCluster API. SageMaker Training exposes secondary job status in TrainingJobSummary from ListTrainingJobs API. SageMaker now supports G6, G6e, P5e instances for HyperPod and Training.

# v1.165.0 (2024-10-30)

* **Feature**: Added support for Model Registry Staging construct. Users can define series of stages that models can progress through for model workflows and lifecycle. This simplifies tracking and managing models as they transition through development, testing, and production stages.

# v1.164.0 (2024-10-29)

* **Feature**: Adding `notebook-al2-v3` as allowed value to SageMaker NotebookInstance PlatformIdentifier attribute

# v1.163.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.163.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.163.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.163.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.162.1 (2024-10-03)

* No change notes available for this release.

# v1.162.0 (2024-10-02)

* **Feature**: releasing builtinlcc to public

# v1.161.1 (2024-09-27)

* No change notes available for this release.

# v1.161.0 (2024-09-26)

* **Feature**: Adding `TagPropagation` attribute to Sagemaker API

# v1.160.1 (2024-09-25)

* No change notes available for this release.

# v1.160.0 (2024-09-24)

* **Feature**: Adding `HiddenInstanceTypes` and `HiddenSageMakerImageVersionAliases` attribute to SageMaker API

# v1.159.1 (2024-09-23)

* No change notes available for this release.

# v1.159.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Feature**: Amazon SageMaker now supports using manifest files to specify the location of uncompressed model artifacts within Model Packages
* **Dependency Update**: Updated to the latest SDK module versions

# v1.158.0 (2024-09-19)

* **Feature**: Introduced support for G6e instance types on SageMaker Studio for JupyterLab and CodeEditor applications.

# v1.157.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.157.0 (2024-09-09)

* **Feature**: Amazon Sagemaker supports orchestrating SageMaker HyperPod clusters with Amazon EKS

# v1.156.0 (2024-09-05)

* **Feature**: Amazon SageMaker now supports idle shutdown of JupyterLab and CodeEditor applications on SageMaker Studio.

# v1.155.1 (2024-09-04)

* No change notes available for this release.

# v1.155.0 (2024-09-03)

* **Feature**: Amazon SageMaker now supports automatic mounting of a user's home folder in the Amazon Elastic File System (EFS) associated with the SageMaker Studio domain to their Studio Spaces to enable users to share data between their own private spaces.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.154.0 (2024-08-16)

* **Feature**: Introduce Endpoint and EndpointConfig Arns in sagemaker:ListPipelineExecutionSteps API response

# v1.153.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.153.0 (2024-08-12)

* **Feature**: Releasing large data support as part of CreateAutoMLJobV2 in SageMaker Autopilot and CreateDomain API for SageMaker Canvas.

# v1.152.0 (2024-08-01)

* **Feature**: This release adds support for Amazon EMR Serverless applications in SageMaker Studio for running data processing jobs.

# v1.151.0 (2024-07-18)

* **Feature**: SageMaker Training supports R5, T3 and R5D instances family. And SageMaker Processing supports G5 and R5D instances family.

# v1.150.2 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.150.1 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.150.0 (2024-07-09)

* **Feature**: This release 1/ enables optimization jobs that allows customers to perform Ahead-of-time compilation and quantization. 2/ allows customers to control access to Amazon Q integration in SageMaker Studio. 3/ enables AdditionalModelDataSources for CreateModel action.

# v1.149.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.149.0 (2024-06-27)

* **Feature**: Add capability for Admins to customize Studio experience for the user by showing or hiding Apps and MLTools.

# v1.148.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.147.0 (2024-06-20)

* **Feature**: Adds support for model references in Hub service, and adds support for cross-account access of Hubs

# v1.146.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.146.0 (2024-06-18)

* **Feature**: Launched a new feature in SageMaker to provide managed MLflow Tracking Servers for customers to track ML experiments. This release also adds a new capability of attaching additional storage to SageMaker HyperPod cluster instances.
* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.145.1 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.145.0 (2024-06-11)

* **Feature**: Introduced Scope and AuthenticationRequestExtraParams to SageMaker Workforce OIDC configuration; this allows customers to modify these options for their private Workforce IdP integration. Model Registry Cross-account model package groups are discoverable.

# v1.144.0 (2024-06-07)

* **Feature**: This release introduces a new optional parameter: InferenceAmiVersion, in ProductionVariant.
* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.143.0 (2024-06-04)

* **Feature**: Extend DescribeClusterNode response with private DNS hostname and IP address, and placement information about availability zone and availability zone ID.

# v1.142.1 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.142.0 (2024-05-30)

* **Feature**: Adds Model Card information as a new component to Model Package. Autopilot launches algorithm selection for TimeSeries modality to generate AutoML candidates per algorithm.

# v1.141.1 (2024-05-23)

* No change notes available for this release.

# v1.141.0 (2024-05-16)

* **Feature**: Introduced WorkerAccessConfiguration to SageMaker Workteam. This allows customers to configure resource access for workers in a workteam.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.140.1 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.140.0 (2024-05-10)

* **Feature**: Introduced support for G6 instance types on Sagemaker Notebook Instances and on SageMaker Studio for JupyterLab and CodeEditor applications.

# v1.139.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.139.0 (2024-05-03)

* **Feature**: Amazon SageMaker Inference now supports m6i, c6i, r6i, m7i, c7i, r7i and g5 instance types for Batch Transform Jobs

# v1.138.0 (2024-04-30)

* **Feature**: Amazon SageMaker Training now supports the use of attribute-based access control (ABAC) roles for training job execution roles. Amazon SageMaker Inference now supports G6 instance types.

# v1.137.0 (2024-04-22)

* **Feature**: This release adds support for Real-Time Collaboration and Shared Space for JupyterLab App on SageMaker Studio.

# v1.136.0 (2024-04-18)

* **Feature**: Removed deprecated enum values and updated API documentation.

# v1.135.0 (2024-03-29)

* **Feature**: This release adds support for custom images for the CodeEditor App on SageMaker Studio
* **Dependency Update**: Updated to the latest SDK module versions

# v1.134.0 (2024-03-25)

* **Feature**: Introduced support for the following new instance types on SageMaker Studio for JupyterLab and CodeEditor applications: m6i, m6id, m7i, c6i, c6id, c7i, r6i, r6id, r7i, and p5

# v1.133.1 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.133.0 (2024-03-15)

* **Feature**: Adds m6i, m6id, m7i, c6i, c6id, c7i, r6i r6id, r7i, p5 instance type support to Sagemaker Notebook Instances and miscellaneous wording fixes for previous Sagemaker documentation.

# v1.132.1 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.132.0 (2024-02-29)

* **Feature**: Adds support for ModelDataSource in Model Packages to support unzipped models. Adds support to specify SourceUri for models which allows registration of models without mandating a container for hosting. Using SourceUri, customers can decouple the model from hosting information during registration.

# v1.131.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.131.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.130.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.130.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.130.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.129.0 (2024-02-15)

* **Feature**: This release adds a new API UpdateClusterSoftware for SageMaker HyperPod. This API allows users to patch HyperPod clusters with latest platform softwares.
* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.128.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.127.0 (2024-02-02)

* **Feature**: Amazon SageMaker Canvas adds GenerativeAiSettings support for CanvasAppSettings.

# v1.126.0 (2024-01-26)

* **Feature**: Amazon SageMaker Automatic Model Tuning now provides an API to programmatically delete tuning jobs.

# v1.125.0 (2024-01-14)

* **Feature**: This release will have ValidationException thrown if certain invalid app types are provided. The release will also throw ValidationException if more than 10 account ids are provided in VpcOnlyTrustedAccounts.

# v1.124.0 (2024-01-04)

* **Feature**: Adding support for provisioned throughput mode for SageMaker Feature Groups
* **Dependency Update**: Updated to the latest SDK module versions

# v1.123.0 (2023-12-28)

* **Feature**: Amazon SageMaker Studio now supports Docker access from within app container

# v1.122.1 (2023-12-26)

* No change notes available for this release.

# v1.122.0 (2023-12-21)

* **Feature**: Amazon SageMaker Training now provides model training container access for debugging purposes. Amazon SageMaker Search now provides the ability to use visibility conditions to limit resource access to a single domain or multiple domains.

# v1.121.0 (2023-12-15)

* **Feature**: This release 1) introduces a new API: DeleteCompilationJob , and 2) adds InfraCheckConfig for Create/Describe training job API

# v1.120.4 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.120.3 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.120.2 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.120.1 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.120.0 (2023-11-30.2)

* **Feature**: This release adds support for 1/ Code Editor, based on Code-OSS, Visual Studio Code Open Source, a new fully managed IDE option in SageMaker Studio  2/ JupyterLab, a new fully managed JupyterLab IDE experience in SageMaker Studio

# v1.119.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.119.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Feature**: This release adds following support 1/ Improved SDK tooling for model deployment. 2/ New Inference Component based features to lower inference costs and latency 3/ SageMaker HyperPod management. 4/ Additional parameters for FM Fine Tuning in Autopilot
* **Dependency Update**: Updated to the latest SDK module versions

# v1.118.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.118.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.118.0 (2023-11-22)

* **Feature**: This feature adds the end user license agreement status as a model access configuration parameter.

# v1.117.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.117.0 (2023-11-16)

* **Feature**: Amazon SageMaker Studio now supports Trainium instance types - trn1.2xlarge, trn1.32xlarge, trn1n.32xlarge.

# v1.116.1 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.116.0 (2023-11-14)

* **Feature**: This release makes Model Registry Inference Specification fields as not required.

# v1.115.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.115.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Feature**: Support for batch transform input in Model dashboard
* **Dependency Update**: Updated to the latest SDK module versions

# v1.114.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.113.0 (2023-10-26)

* **Feature**: Amazon Sagemaker Autopilot now supports Text Generation jobs.

# v1.112.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.111.1 (2023-10-20)

* No change notes available for this release.

# v1.111.0 (2023-10-12)

* **Feature**: Amazon SageMaker Canvas adds KendraSettings and DirectDeploySettings support for CanvasAppSettings
* **Dependency Update**: Updated to the latest SDK module versions

# v1.110.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.110.0 (2023-10-04)

* **Feature**: Adding support for AdditionalS3DataSource, a data source used for training or inference that is in addition to the input dataset or model data.

# v1.109.0 (2023-10-03)

* **Feature**: This release allows users to run Selective Execution in SageMaker Pipelines without SourcePipelineExecutionArn if selected steps do not have any dependent steps.

# v1.108.0 (2023-09-28)

* **Feature**: Online store feature groups supports Standard and InMemory tier storage types for low latency storage for real-time data retrieval. The InMemory tier supports collection types List, Set, and Vector.

# v1.107.0 (2023-09-19)

* **Feature**: This release adds support for one-time model monitoring schedules that are executed immediately without delay, explicit data analysis windows for model monitoring schedules and exclude features attributes to remove features from model monitor analysis.

# v1.106.0 (2023-09-15)

* **Feature**: This release introduces Skip Model Validation for Model Packages

# v1.105.0 (2023-09-08)

* **Feature**: Autopilot APIs will now support holiday featurization for Timeseries models. The models will now hold holiday metadata and should be able to accommodate holiday effect during inference.

# v1.104.0 (2023-09-05)

* **Feature**: SageMaker Neo now supports data input shape derivation for Pytorch 2.0  and XGBoost compilation job for cloud instance targets. You can skip DataInputConfig field during compilation job creation. You can also access derived information from model in DescribeCompilationJob response.

# v1.103.0 (2023-08-30)

* **Feature**: Amazon SageMaker Canvas adds IdentityProviderOAuthSettings support for CanvasAppSettings

# v1.102.3 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.102.2 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.102.1 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.102.0 (2023-08-15)

* **Feature**: SageMaker Inference Recommender now provides SupportedResponseMIMETypes from DescribeInferenceRecommendationsJob response

# v1.101.0 (2023-08-09)

* **Feature**: This release adds support for cross account access for SageMaker Model Cards through AWS RAM.

# v1.100.1 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.100.0 (2023-08-04)

* **Feature**: Including DataCaptureConfig key in the Amazon Sagemaker Search's transform job object

# v1.99.0 (2023-08-03)

* **Feature**: Amazon SageMaker now supports running training jobs on p5.48xlarge instance types.

# v1.98.0 (2023-08-02)

* **Feature**: SageMaker Inference Recommender introduces a new API GetScalingConfigurationRecommendation to recommend auto scaling policies based on completed Inference Recommender jobs.

# v1.97.0 (2023-08-01)

* **Feature**: Add Stairs TrafficPattern and FlatInvocations to RecommendationJobStoppingConditions

# v1.96.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.95.1 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.95.0 (2023-07-27)

* **Feature**: Expose ProfilerConfig attribute in SageMaker Search API response.

# v1.94.0 (2023-07-25)

* **Feature**: Mark ContentColumn and TargetLabelColumn as required Targets in TextClassificationJobConfig in CreateAutoMLJobV2API

# v1.93.0 (2023-07-20.2)

* **Feature**: Cross account support for SageMaker Feature Store

# v1.92.0 (2023-07-13)

* **Feature**: Amazon SageMaker Canvas adds WorkspeceSettings support for CanvasAppSettings
* **Dependency Update**: Updated to the latest SDK module versions

# v1.91.0 (2023-07-03)

* **Feature**: SageMaker Inference Recommender now accepts new fields SupportedEndpointType and ServerlessConfiguration to support serverless endpoints.

# v1.90.0 (2023-06-30)

* **Feature**: This release adds support for rolling deployment in SageMaker Inference.

# v1.89.0 (2023-06-29)

* **Feature**: Adding support for timeseries forecasting in the CreateAutoMLJobV2 API.

# v1.88.0 (2023-06-28)

* **Feature**: This release adds support for Model Cards Model Registry integration.

# v1.87.0 (2023-06-27)

* **Feature**: Introducing TTL for online store records in feature groups.

# v1.86.0 (2023-06-21)

* **Feature**: This release provides support in SageMaker for output files in training jobs to be uploaded without compression and enable customer to deploy uncompressed model from S3 to real-time inference Endpoints. In addition, ml.trn1n.32xlarge is added to supported instance type list in training job.

# v1.85.0 (2023-06-19)

* **Feature**: Amazon Sagemaker Autopilot releases CreateAutoMLJobV2 and DescribeAutoMLJobV2 for Autopilot customers with ImageClassification, TextClassification and Tabular problem type config support.

# v1.84.2 (2023-06-15)

* No change notes available for this release.

# v1.84.1 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.84.0 (2023-06-12)

* **Feature**: Sagemaker Neo now supports compilation for inferentia2 (ML_INF2) and Trainium1 (ML_TRN1) as available targets. With these devices, you can run your workloads at highest performance with lowest cost. inferentia2 (ML_INF2) is available in CMH and Trainium1 (ML_TRN1) is available in IAD currently

# v1.83.0 (2023-06-02)

* **Feature**: This release adds Selective Execution feature that allows SageMaker Pipelines users to run selected steps in a pipeline.

# v1.82.1 (2023-06-01)

* **Documentation**: Amazon Sagemaker Autopilot adds support for Parquet file input to NLP text classification jobs.

# v1.82.0 (2023-05-26)

* **Feature**: Added ml.p4d and ml.inf1 as supported instance type families for SageMaker Notebook Instances.

# v1.81.0 (2023-05-25)

* **Feature**: Amazon SageMaker Automatic Model Tuning now supports enabling Autotune for tuning jobs which can choose tuning job configurations.

# v1.80.0 (2023-05-24)

* **Feature**: SageMaker now provides an instantaneous deployment recommendation through the DescribeModel API

# v1.79.0 (2023-05-23)

* **Feature**: Added ModelNameEquals, ModelPackageVersionArnEquals in request and ModelName, SamplePayloadUrl, ModelPackageVersionArn in response of ListInferenceRecommendationsJobs API. Added Invocation timestamps in response of DescribeInferenceRecommendationsJob API & ListInferenceRecommendationsJobSteps API.

# v1.78.0 (2023-05-09)

* **Feature**: This release includes support for (1) Provisioned Concurrency for Amazon SageMaker Serverless Inference and (2) UpdateEndpointWeightsAndCapacities API for Serverless endpoints.

# v1.77.0 (2023-05-04)

* **Feature**: We added support for ml.inf2 and ml.trn1 family of instances on Amazon SageMaker for deploying machine learning (ML) models for Real-time and Asynchronous inference. You can use these instances to achieve high performance at a low cost for generative artificial intelligence (AI) models.

# v1.76.0 (2023-05-02)

* **Feature**: Amazon Sagemaker Autopilot supports training models with sample weights and additional objective metrics.

# v1.75.0 (2023-04-27)

* **Feature**: Added ml.p4d.24xlarge and ml.p4de.24xlarge as supported instances for SageMaker Studio

# v1.74.1 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.74.0 (2023-04-20)

* **Feature**: Amazon SageMaker Canvas adds ModelRegisterSettings support for CanvasAppSettings.

# v1.73.3 (2023-04-10)

* No change notes available for this release.

# v1.73.2 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.73.1 (2023-04-06)

* No change notes available for this release.

# v1.73.0 (2023-04-04)

* **Feature**: Amazon SageMaker Asynchronous Inference now allows customer's to receive failure model responses in S3 and receive success/failure model responses in SNS notifications.

# v1.72.2 (2023-03-30)

* No change notes available for this release.

# v1.72.1 (2023-03-27)

* **Documentation**: Fixed some improperly rendered links in SDK documentation.

# v1.72.0 (2023-03-23)

* **Feature**: Amazon SageMaker Autopilot adds two new APIs - CreateAutoMLJobV2 and DescribeAutoMLJobV2. Amazon SageMaker Notebook Instances now supports the ml.geospatial.interactive instance type.

# v1.71.2 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.71.1 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.71.0 (2023-03-09)

* **Feature**: Amazon SageMaker Inference now allows SSM access to customer's model container by setting the "EnableSSMAccess" parameter for a ProductionVariant in CreateEndpointConfig API.

# v1.70.0 (2023-03-08)

* **Feature**: There needs to be a user identity to specify the SageMaker user who perform each action regarding the entity. However, these is a not a unified concept of user identity across SageMaker service that could be used today.

# v1.69.0 (2023-03-02)

* **Feature**: Add a new field "EndpointMetrics" in SageMaker Inference Recommender "ListInferenceRecommendationsJobSteps" API response.

# v1.68.3 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.68.2 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.68.1 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.68.0 (2023-02-10)

* **Feature**: Amazon SageMaker Autopilot adds support for selecting algorithms in CreateAutoMLJob API.

# v1.67.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.67.0 (2023-01-31)

* **Feature**: Amazon SageMaker Automatic Model Tuning now supports more completion criteria for Hyperparameter Optimization.

# v1.66.0 (2023-01-27)

* **Feature**: This release supports running SageMaker Training jobs with container images that are in a private Docker registry.

# v1.65.0 (2023-01-25)

* **Feature**: SageMaker Inference Recommender now decouples from Model Registry and could accept Model Name to invoke inference recommendations job; Inference Recommender now provides CPU/Memory Utilization metrics data in recommendation output.

# v1.64.0 (2023-01-23)

* **Feature**: Amazon SageMaker Inference now supports P4de instance types.

# v1.63.0 (2023-01-19)

* **Feature**: HyperParameterTuningJobs now allow passing environment variables into the corresponding TrainingJobs

# v1.62.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.61.0 (2022-12-21)

* **Feature**: This release enables adding RStudio Workbench support to an existing Amazon SageMaker Studio domain. It allows setting your RStudio on SageMaker environment configuration parameters and also updating the RStudioConnectUrl and RStudioPackageManagerUrl parameters for existing domains

# v1.60.0 (2022-12-20)

* **Feature**: Amazon SageMaker Autopilot adds support for new objective metrics in CreateAutoMLJob API.

# v1.59.0 (2022-12-19)

* **Feature**: AWS Sagemaker - Sagemaker Images now supports Aliases as secondary identifiers for ImageVersions. SageMaker Images now supports additional metadata for ImageVersions for better images management.

# v1.58.0 (2022-12-16)

* **Feature**: AWS sagemaker - Features: This release adds support for random seed, it's an integer value used to initialize a pseudo-random number generator. Setting a random seed will allow the hyperparameter tuning search strategies to produce more consistent configurations for the same tuning job.

# v1.57.0 (2022-12-15)

* **Feature**: SageMaker Inference Recommender now allows customers to load tests their models on various instance types using private VPC.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.1 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.0 (2022-11-30)

* **Feature**: Added Models as part of the Search API. Added Model shadow deployments in realtime inference, and shadow testing in managed inference. Added support for shared spaces, geospatial APIs, Model Cards, AutoMLJobStep in pipelines, Git repositories on user profiles and domains, Model sharing in Jumpstart.

# v1.55.0 (2022-11-18)

* **Feature**: Added DisableProfiler flag as a new field in ProfilerConfig

# v1.54.0 (2022-11-03)

* **Feature**: Amazon SageMaker now supports running training jobs on ml.trn1 instance types.

# v1.53.0 (2022-11-02)

* **Feature**: This release updates Framework model regex for ModelPackage to support new Framework version xgboost, sklearn.

# v1.52.0 (2022-10-27)

* **Feature**: This change allows customers to provide a custom entrypoint script for the docker container to be run while executing training jobs, and provide custom arguments to the entrypoint script.

# v1.51.0 (2022-10-26)

* **Feature**: Amazon SageMaker Automatic Model Tuning now supports specifying Grid Search strategy for tuning jobs, which evaluates all hyperparameter combinations exhaustively based on the categorical hyperparameters provided.

# v1.50.0 (2022-10-24)

* **Feature**: SageMaker Inference Recommender now supports a new API ListInferenceRecommendationJobSteps to return the details of all the benchmark we create for an inference recommendation job.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2022-10-21)

* **Feature**: CreateInferenceRecommenderjob API now supports passing endpoint details directly, that will help customers to identify the max invocation and max latency they can achieve for their model and the associated endpoint along with getting recommendations on other instances.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.48.0 (2022-10-18)

* **Feature**: This change allows customers to enable data capturing while running a batch transform job, and configure monitoring schedule to monitoring the captured data.

# v1.47.0 (2022-10-17)

* **Feature**: This release adds support for C7g, C6g, C6gd, C6gn, M6g, M6gd, R6g, and R6gn Graviton instance types in Amazon SageMaker Inference.

# v1.46.0 (2022-09-30)

* **Feature**: A new parameter called ExplainerConfig is added to CreateEndpointConfig API to enable SageMaker Clarify online explainability feature.

# v1.45.0 (2022-09-29)

* **Feature**: SageMaker Training Managed Warm Pools let you retain provisioned infrastructure to reduce latency for repetitive training workloads.

# v1.44.0 (2022-09-21)

* **Feature**: SageMaker now allows customization on Canvas Application settings, including enabling/disabling time-series forecasting and specifying an Amazon Forecast execution role at both the Domain and UserProfile levels.

# v1.43.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2022-09-15)

* **Feature**: Amazon SageMaker Automatic Model Tuning now supports specifying Hyperband strategy for tuning jobs, which uses a multi-fidelity based tuning strategy to stop underperforming hyperparameter configurations early.

# v1.42.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Feature**: SageMaker Hosting now allows customization on ML instance storage volume size, model data download timeout and inference container startup ping health check timeout for each ProductionVariant in CreateEndpointConfig API.
* **Feature**: This release adds HyperParameterTuningJob type in Search API.
* **Feature**: This release adds Mode to AutoMLJobConfig.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2022-09-02)

* **Feature**: This release enables administrators to attribute user activity and API calls from Studio notebooks, Data Wrangler and Canvas to specific users even when users share the same execution IAM role.  ExecutionRoleIdentityConfig at Sagemaker domain level enables this feature.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2022-08-31)

* **Feature**: SageMaker Inference Recommender now accepts Inference Recommender fields: Domain, Task, Framework, SamplePayloadUrl, SupportedContentTypes, SupportedInstanceTypes, directly in our CreateInferenceRecommendationsJob API through ContainerConfig
* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.3 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.2 (2022-08-22)

* No change notes available for this release.

# v1.39.1 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2022-08-09)

* **Feature**: Amazon SageMaker Automatic Model Tuning now supports specifying multiple alternate EC2 instance types to make tuning jobs more robust when the preferred instance type is not available due to insufficient capacity.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.2 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2022-07-19)

* **Feature**: Fixed an issue with cross account QueryLineage

# v1.37.0 (2022-07-18)

* **Feature**: Amazon SageMaker Edge Manager provides lightweight model deployment feature to deploy machine learning models on requested devices.

# v1.36.0 (2022-07-14)

* **Feature**: This release adds support for G5, P4d, and C6i instance types in Amazon SageMaker Inference and increases the number of hyperparameters that can be searched from 20 to 30 in Amazon SageMaker Automatic Model Tuning

# v1.35.0 (2022-07-07)

* **Feature**: Heterogeneous clusters: the ability to launch training jobs with multiple instance types. This enables running component of the training job on the instance type that is most suitable for it. e.g. doing data processing and augmentation on CPU instances and neural network training on GPU instances

# v1.34.1 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2022-06-29)

* **Feature**: This release adds: UpdateFeatureGroup, UpdateFeatureMetadata, DescribeFeatureMetadata APIs; FeatureMetadata type in Search API; LastModifiedTime, LastUpdateStatus, OnlineStoreTotalSizeBytes in DescribeFeatureGroup API.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2022-06-23)

* **Feature**: SageMaker Ground Truth now supports Virtual Private Cloud. Customers can launch labeling jobs and access to their private workforce in VPC mode.

# v1.32.1 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2022-05-27)

* **Feature**: Amazon SageMaker Notebook Instances now allows configuration of Instance Metadata Service version and Amazon SageMaker Studio now supports G5 instance types.

# v1.31.0 (2022-05-25)

* **Feature**: Amazon SageMaker Autopilot adds support for manually selecting features from the input dataset using the CreateAutoMLJob API.

# v1.30.1 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2022-05-03)

* **Feature**: SageMaker Autopilot adds new metrics for all candidate models generated by Autopilot experiments; RStudio on SageMaker now allows users to bring your own development environment in a custom image.

# v1.29.0 (2022-04-27)

* **Feature**: Amazon SageMaker Autopilot adds support for custom validation dataset and validation ratio through the CreateAutoMLJob and DescribeAutoMLJob APIs.

# v1.28.0 (2022-04-26)

* **Feature**: SageMaker Inference Recommender now accepts customer KMS key ID for encryption of endpoints and compilation outputs created during inference recommendation.

# v1.27.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2022-04-07)

* **Feature**: Amazon Sagemaker Notebook Instances now supports G5 instance types

# v1.26.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-01-28)

* **Feature**: Updated to latest API model.

# v1.23.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-01-07)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: API client updated

# v1.20.0 (2021-12-02)

* **Feature**: API client updated
* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.
* **Feature**: Updated service to latest API model.
* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.18.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Feature**: Updated service to latest API model.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2021-10-21)

* **Feature**: API client updated
* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-09-17)

* **Feature**: Updated API client and endpoints to latest revision.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2021-09-10)

* **Feature**: API client updated

# v1.13.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-08-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-07-15)

* **Feature**: Updated service model to latest version.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-07-01)

* **Feature**: API client updated

# v1.8.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-06-11)

* **Feature**: Updated to latest API model.

# v1.6.0 (2021-05-20)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

