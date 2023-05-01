# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)

## 65.36.1 - 2023-04-25
### Added
- Support for enabling mTLS authentication with Listener and for providing custom value for TLS port and Non-TLS Port during AVM Cluster Creation in Database service
- Support for usedDataStorageSizeInGbs property for autonomous database in the Database service
- Support for csiNumber organization in Tenant Manager Control Plane service
- Support for creating and updating an infrastructure with LACP support in Database service
- Support for changePrivateEndpointOutboundConnection operation in Integration Cloud service
- Support for Enable Process in Integration Cloud service
- Support for Disaster Recovery, DR enablement, switchover, and failover feature in Fusion Apps service
- Support for discovery and monitoring of External Exadata infrastructure in Database Management Service

## 65.36.0 - 2023-04-18
### Added
- Support for private endpoints in the Digital Assistant service
- Support for canceling backups in the Database service
- Support for improved labeling of key/value pairs in the Data Labeling service  
 
### Breaking Changes
- Support for retries by default on operations of the Digital Assistant service
- The property `LifetimeLogicalClock` was removed from the models `Record`, `Dataset` and `Annotation` in the Data Labeling service
- The property `OpcRetryToken` was removed from the models `ConfigureDigitalAssistantParametersRequest`, `RotateChannelKeysRequest`, `StartChannelRequest`, `StopChannelRequest` in the Data Labeling service
- The property `DigitalAssistantId` was renamed to `Id` in the `ListDigitalAssistantsRequest` model in the Data Labeling service
- The property `IsLatestSkillOnly` was renamed to `IsLatestVersionOnly` in the `ListPackagesRequest` model in the Data Labeling service
- The property `IsLatestSkillOnly` was renamed to `IsLatestVersionOnly` in the `ListPackagesRequest` model in the Data Labeling service
- The property `SkillId` was renamed to `Id` in the `ListSkillsRequest` model in the Data Labeling service
- The properties `AuthorizationEndpointUrl` and `SubjectClaim` were made optional in the `AuthenticationProvider` model in the Data Labeling service

## 65.35.0 - 2023-04-11
### Added
- Support for rotation of certificates on autonomous VM clusters on Exadata Cloud at Customer in the Database service
- Support for ACD and OKV wallet naming for autonomous databases and dedicated autonomous databases on Exadata Cloud at Customer in the Database service
- Support for Exadata cloud service application virtual IPs (VIPs) in the Database service
- Support for additional manageability features for large sensitive data models and masking policies in the Data Safe service
- Support for getting user profile details and assignments for databases and fleets in the Data Safe service
- Support for enabling ADDM spotlight for databases in the Operations Insights service
 
### Breaking Changes
- The property `AdditionalDatabaseStatus` was removed from the models `AutonomousDatabase`, `AutonomousDatabaseSummary`, `AutonomousDataWarehouse`and `AutonomousDataWarehouseSummary` in the Database service


## 65.34.0 - 2023-04-04
### Added
- Support for pre-emptible worker nodes in the Container Engine for Kubernetes service
- Support for larger data storage (now up to 128TB) in the MySQL Database service
- Support for HTTP health checks for HTTPS backend sets in the Load Balancer service  
     
### Breaking Changes
- The property `BackendSetName` was made required in the `ForwardToBackendSet` model in the Load Balancer service
- Support for the Data Connectivity Management service was removed


## 65.33.1 - 2023-03-28
### Added
- Support for ACD and OKV wallet naming for autonomous databases and dedicated autonomous databases on Exadata Cloud at Customer in the Database service
- Support for validating the credentials of a connection in the DevOps service
- Support for GoldenGate Replicat performance profiles when creating a migration in the Database Migration service
- Support for connection diagnostics on registered databases in the Database Migration service
- Support for launching bare metal instances in an RDMA network in the Compute service


## 65.33.0 - 2023-03-21
### Added
- Support for backup automation integration with the Database Recovery service in the Database service
- Support for changing the disaster recovery configuration of an autonomous database in remote regions of its disaster recovery association in the Database service
- Support for creating a remote disaster recovery association clone of an autonomous database in the Database service
- Support for managed build stages to be configured to use custom shape build runners in the DevOps service
- Support for listing pre-built functions and creating functions from pre-built functions in the Functions service
- Support for connections types for database resources of type Amazon S3, HDFS, SQL Server, Java Messaging service, Mongo DB, Oracle NoSQL, and Snowflake in the GoldenGate service

### Changes
- Upgraded golang.org/x/sys package version to v0.6.0

### Breaking Changes
- The enum value `LAKE_HOUSE_CONNECTION` was renamed to `LAKE_CONNECTION` in the enum ModelTypeEnum in the Connection, ConnectionDetails, ConnectionSummary, CreateConnectionDetails and UpdateConnectionDetails models in the Data Integration Service
- The enum value `LAKE_HOUSE_DATA_ASSET` was renamed to `LAKE_DATA_ASSET` in the enum ModelTypeEnum in the DataAsset, CreateDataAssetDetails, DataAssetSummary, and UpdateDataAssetDetails models in the Data Integration Service
- `DefaultValue` is now a required param in `BuildPipelineParameter` model in the Devops service


## 65.32.1 - 2023-03-14
### Added
- Support for the Identity Domains service
- Support for long-term backups for autonomous databases on Exadata Cloud at Customer in the Database service
- Support for database OS patching in the Database service
- Support for managing enhanced clusters, cluster add-ons, and serverless virtual node pools in the Container Engine for Kubernetes service
- Support for templates and copy object requests in the Data Integration service
- Support for maintenance features in the GoldenGate service
- Support for `AMD_MILAN_BM_GPU` configuration type on instances in the Compute service
- Support for host storage metrics and network metrics as part of host capacity planning in the Operations Insights service


## 65.32.0 - 2023-03-07
### Added
- Support for creating and updating autonomous database long-term backup schedules in the Database service
- Support for creating, updating, and deleting autonomous database long-term backups in the Database service
- Support for model deployment resources to use customized container images containing runtime dependencies of ML models and custom web servers to handle inference requests in the Data Science service
- Support for using the compartmentIdInSubtree parameter when summarizing management agent counts in the Management Agent Cloud service
- Support for getting agent property details in the Management Agent Cloud service
- Support for filtering by gateway ID when listing agents in the Management Agent Cloud service
- Support for the Hebrew and Greek languages during AI language text translation in the AI Language service
- Support for auto-detection when analyzing text with pre-trained models in the AI Language service
- Support for specifying update operation constraints when updating an instance in the Compute Service
- Support for disaster recovery in the Content Management service
- Support for advanced autonomous databases insights in the Operations Insights service
     
### Breaking Changes
- Support for retries by default on operations of the Analytics Cloud service
- The enum member `ACTIVE` was removed from the enum LifecycleDetailsActive is removed from OCE service


## 65.31.1 - 2023-02-28
### Added
- Support for calling Oracle Cloud Infrastructure services in the eu-dcc-rating-1, eu-dcc-rating-2, eu-dcc-dublin-1, eu-dcc-dublin-2, and eu-dcc-milan-2 regions
- Support for on-demand bootstrap script execution in the Big Data Service


## 65.31.0 - 2023-02-21
### Added
- Support for async jobs in the AI Anomaly Detection service
- Support for specifying algorithm hints and windows sizes during model training in the AI Anomaly Detection service
- Support for specifying a sensitivity value during model detection in the AI Anomaly Detection service
- Support for discovery and monitoring of external Oracle database infrastructure components in the Database Management service   
 
### Breaking Changes
- The type for property `SystemTags` was changed from `map[string]map[string]interface{}` to `map[string]interface{}` for `ProjectSummary`, `Project`, `ModelSummary`, `Model`, `DataAssetSummary`, `DataAsset`, `AiPrivateEndpointSummary`, `AiPrivateEndpoint` models in the AI Anomaly Detection service
- Support for retries by default on operations of the AI Anomaly Detection service


## 65.30.0 - 2023-02-14
### Added
- Support for the Visual Builder Studio service
- Support for the Autonomous Recovery service
- Support for selecting specific database servers when creating autonomous VM clusters in the Database service
- Support for creating autonomous VMs during the creation of autonomous VM clusters in the Database service

### Breaking Changes
- Support for retries by default on operations of the Compute service


## 65.29.0 - 2023-02-07
### Added
- Support for changing Data Guard role of a database instance within the Database service
- Support for listing autonomous container database versions in the Database service
- Support for specifying a database version when creating or updating an autonomous container database in the Database service
- Support for specifying an eCPU count when creating or updating autonomous shared databases in the Database service
- Support for Helm attestation and Helm arguments on deploy operations in the DevOps service
- Support for uploading master key wallets for deployments in the GoldenGate service
- Support for custom configurations in the Operations Insights service
- Support for refreshing the session token in SessionTokenAuthenticationDetailsProvider
     
### Breaking Changes
- The property `CpuCoreCount` has been made optional in `AutonomousDatabase` and `AutonomousDatabaseSummary` model in the Database service


## 65.28.3 - 2023-01-31
### Added
- Support for ECPU billing for autonomous databases and dedicated autonomous databases on Exadata Cloud at Customer in the Database service
- Support for providing a vault secret ID when creating or updating autonomous shared databases in the Database service
- Support for including ORDS and database transform URLs as autonomous database connections in the Database service
- Support for role-based access control on OpenSearch clusters in the Search service
- Support for managed shell stages on deployments in the DevOps service
- Support for memory encryption on confidential VMs in the Compute service
- Support for configuration items, and reporting ownership of configuration items, in the Application Performance Monitoring service


## 65.28.2 - 2023-01-24
### Added
- Support for the Cloud Migrations service
- Support for setting up custom private IPs while creating private endpoints in the Database service
- Support for machine learning pipelines in the Data Science service
- Support for personally identifiable information detection in the AI Language service


## 65.28.1 - 2023-01-17
### Added
- Support for calling Oracle Cloud Infrastructure services in the us-chicago-1 region
- Support for cross-region replication in the File Storage service
- Support for setting up private DNS on ExaCS systems during provisioning in the Database service
- Support for elastic storage expansion on infrastructure resources for Exadata Cloud at Customer in the Database service
- Support for target versions during infrastructure patching on Cloud Exadata infrastructure in the Database service
- Support for creating model version sets in the model catalog in the Data Science service
- Support for associating a model with a model version set in the Data Science service
- Support for custom key/value annotations on documents in the Data Labeling service
- Support for configurable timeouts in the Service Mesh service


## 65.28.0 - 2022-12-13
### Added
- Support for the Queue service
- Support for Intel X9 shapes when launching VM database systems in the Database service
- Support for enabling, disabling, and editing Database Management service connections on pluggable databases in the Database service
- Support for availability configurations and maintenance window schedules on synthetic monitors in the Application Performance Monitoring service
- Support for scheduling cascading deletes on a project in the DevOps service
- Support for cancelling a scheduled cascading delete on a project in the DevOps service
- Support for issue and action fields on job phases of validation and migration processes in the Database Migration service
- Support for cluster profiles in the Big Data service
- Support for egress-only services in the Service Mesh service
- Support for optional listeners and service discovery metadata on virtual deployments in the Service Mesh service
- Support for canceling work requests in the accepted state in the Service Mesh service
- Support for filtering work requests on associated resource id and operation status in the Service Mesh service
- Support for sorting while listing work requests, listing work request logs, and listing work request errors in the Service Mesh service  
 
 
### Breaking Changes
- The type for property `RouteRules` was changed from a List of `VirtualServiceTrafficRouteRule` to a List of `VirtualServiceTrafficRouteRuleDetails` in the models `UpdateVirtualServiceRouteTableDetails` and `CreateVirtualServiceRouteTableDetails` in the Service Mesh service
- The type for property `Mtls` was changed from `CreateMutualTransportLayerSecurityDetails` to `VirtualServiceMutualTransportLayerSecurityDetails` in the models `UpdateVirtualServiceDetails` and `CreateVirtualServiceDetails.` in the Service Mesh service
- The type for property `RouteRules` was changed from a List of `IngressGatewayTrafficRouteRule` to a List of `IngressGatewayTrafficRouteRuleDetails` in the models `UpdateIngressGatewayRouteTableDetails` and `CreateIngressGatewayRouteTableDetails` in the Service Mesh service
- The type for property `Mtls` was changed from `CreateIngressGatewayMutualTransportLayerSecurityDetails` to `IngressGatewayMutualTransportLayerSecurityDetails` in the models `UpdateIngressGatewayDetails` and `CreateIngressGatewayDetails` in the Service Mesh service
- The type for property `Rules` was changed from a List of `AccessPolicyRule` to a List of `AccessPolicyRuleDetails` in the models `UpdateAccessPolicyDetails` and `CreateAccessPolicyDetails` in the Service Mesh service
- Support for default retries on operations of the Service Mesh service
- Support for default retries on operations of the Database Migration service
- Support for default retries on operations of the Fusion Apps as a Service service


## 65.27.0 - 2022-12-06
### Added
- Support for the Container Instances service
- Support for the Document Understanding service
- Support for creating stacks from OCI DevOps service and Bitbucket Cloud/Server as source control management in the Resource Manager service
- Support for deployment stage level parameters in the DevOps service
- Support for PeopleSoft discovery in the Stack Monitoring service
- Support for Apache Tomcat discovery in the Stack Monitoring service
- Support for SQL Server discovery in the Stack Monitoring service
- Support for OpenId Connect in the API Gateway service
- Support for returning compartment ids when listing backups in the MySQL Database service
- Support for adding a load balancer endpoint to a DB system in the MySQL Database service
- Support for managed read replicas in the MySQL Database service
- Support for setting replication filters on channels in the MySQL Database service
- Support for replicating from a source configured without global transaction identifiers into a channel in the MySQL Database service
- Support for time zone and language preferences in the Announcements service
- Support for adding report schedules for activity auditing and alerts reports in the Data Safe service
- Support for bulk operations on alerts in the Data Safe service
- Support for Java server usage reporting in the Java Management service
- Support for Java library usage reporting in the Java Management service
- Support for cryptographic roadmap impact analysis in the Java Management service
- Support for Java Flight Recorder recordings in the Java Management service
- Support for post-installation steps in the Java Management service
- Support for restricting management of advanced functionality in the Java Management service
- Support for plugin improvements in the Java Management service
- Support for collecting diagnostics on deployments in the GoldenGate service
- Support for onboarding Exadata Public Cloud (ExaCS) targets to the Operations Insights service  
 
### Breaking Changes
- A required property `CompartmentId` was added to `PatchAlertsDetails` model in the Data Safe service
- The property `items` is changed from optional to required in `PatchAlertsDetails` model in the Data Safe service
- The property `datasafePrivateEndpointId` is changed from optional to required in the `PrivateEndpoint` model in the Data Safe service
- The properties `ListenerPort` and `ServiceName` were made required in `InstalledDatabaseDetails` model in the Data Safe service
- The property `serviceName` is is changed from optional to required in `DatabaseCloudServiceDetails` model in the Data Safe service
- The property `AutonomousDatabaseId` was made required in `AutonomousDatabaseDetails` model in the Data Safe service
- The property `OnPremConnectorId` was made required in `OnPremiseConnector` model in the Data Safe service


## 65.26.1 - 2022-11-15
### Added
- Support for mTLS authentication with listeners during Autonomous VM Cluster creation on Exadata Cloud at Customer in the Database service
- Support for providing custom values for TLS and non-TLS ports during Autonomous VM Cluster creation on Exadata Cloud at Customer in the Database service
- Support for creating multiple Autonomous VM Clusters in the same Exadata infrastructure in the Database service
- Support for listing resources associated with a job in the Resource Manager service
- Support for listing resources associated with a stack in the Resource Manager service
- Support for listing outputs associated with a job in the Resource Manager service
- Support for the Oracle distribution of Apache Hadoop 2.0 in the Big Data service


## 65.26.0 - 2022-11-08
### Added
- Support for listing local and cross-region refreshable clones in the Database service
- Support for adding multiple cloud VM clusters in the Database service
- Support for creating rollback jobs in the Resource Manager service
- Support for edge nodes in the Big Data service
- Support for Single Client Access Name (SCAN) in the Data Flow service
- Support for additional filters when listing application dependencies in the Application Dependency Management service
- Support for additional properties when reading Vulnerability Audit resources in the Application Dependency Management service
- Support for optionally passing compartment IDs when creating Vulnerability Audit resources in the Application Dependency Management service

### Breaking Changes
- The property `CertificateId` was made required in `PrivateServerConfigDetails` model in the Resource Manager service


## 65.25.0 - 2022-11-01
### Added
- Support for cloning from a backup from the last available timestamp in the Database service
- Support for third-party scanning using Qualys in the Vulnerability Scanning service
- Support for customer-provided encryption keys in the Logging Analytics service
- Support for connections for database resources in the GoldenGate service

### Breaking Changes
- HostScanAgentConfigurationVendorEnum is removed from vulnerability scanning service
- CompartmentId, TimeDataEnded and DataType properties are no longer a mandatory field in StorageWorkRequestSummary model


## 65.24.0 - 2022-10-25
### Added
- Support for the Disaster Recovery service
- Support for running code interactively with session applications using statements in the Data Flow service
- Support for language custom models and language translation in the AI Language service

### Breaking Changes
- `TextClassificationDocument` and `KeyPhraseDocument` modles are removed from  ailanguage service
- Document parameter in `BatchDetectLanguage` related models changed to `TextDocument` type

## 65.23.0 - 2022-10-04
### Added
- Support for calling Oracle Cloud Infrastructure services in the eu-dcc-milan-1 region
- Support for target host identification and SOCKS support on dynamic port forwarding sessions in the Bastion service
- Support for viewing top processes running at a particular point of time in the Operations Insights service
- Support for filtering top processes by a single process to view that process's trend over time in the Operations Insights service
- Support for creating Enterprise Manager-based Windows host targets in the Operations Insights service
- Support for creating Management Agent Cloud-based Windows and Solaris host targets in the Operations Insights service

### Breaking Changes
- The property `TargetResourcePort` was removed from the models `TargetResourceDetails` and `CreateSessionTargetResourceDetails` in the Bastion service

## 65.22.0 - 2022-09-27
### Added
- Support for search capabilities for monitored resources in the Stack Monitoring service
- Support for deleting monitored resources with their members in the Stack Monitoring service
- Support for creating host-type monitored resources in the Stack Monitoring service
- Support for associating external resources during creation of monitored resources in the Stack Monitoring service
- Support for uploading bulk data in the NoSQL Database Cloud service
- Support for examining query execution plans in the NoSQL Database Cloud service
- Support for starting and stopping clusters in the Big Data service
- Support for additional compute shapes in the Big Data service
- Support for backwards pagination in the Search service
- Support for elastic compute for Exadata Cloud at Customer in the Database service
 
### Breaking Changes
- Support for default retries on operations of the NoSQL Database Cloud service

## 65.21.0 - 2022-09-20
### Added Support for the Cloud Bridge service
- Support for the Cloud Migrations service
- Support for display banners, trails, and sizes in the GoldenGate service
- Support for generic REST data assets, flattening of data in Data Flow, and runtime information on pipelines in the Data Integration service
- Support for expanded search functionality in the Threat Intelligence service
- Support for ingest-time rules and specifying logsets and query strings during recalls in the Logging Analytics service
- Support for repository mirroring from Visual Builder Studio in the DevOps service
- Support for running a managed build stage with the source code hosted in a Visual Builder Studio repository in the DevOps service
- Support for triggering a build run based on an event in a Visual Builder Studio repository in the DevOps service
- Support for additional parameters during cost management scheduling in the Usage service
 
### Breaking Changes
- Support for retries by default on operations of the GoldenGate service
- Support for retries by default on operations of the Threat Intelligence service
- PreviousDeploymentId and DeployStageId are now mandatory parameters in operations in devops service


## 65.20.0 - 2022-09-13
### Added
- Support for calling Oracle Cloud Infrastructure services in the eu-madrid-1 region
- Support for exporting and importing larger model artifacts in the model catalog in the Data Science service
- Support for Request Based Authorization in the API Gateway service
- Support for Dynamic Authentication in the API Gateway service
- Support for Dynamic Routing Backend in the API Gateway service

### Breaking Changes
- Support for retries by default on some operations of the Data Science service

## 65.19.0 - 2022-09-06
### Added
- Support for generic REST, OCI Streaming service, and Lake House connectors in the Data Connectivity Management service
- Support for connecting to the Data Catalog service in the Data Connectivity Management service
- Support for Kerberos and SSL for HDFS operations in the Data Connectivity Management service
- Support for excel-formatted data and default columns in the Data Connectivity Management service
- Support for reporting connector usage in the Data Connectivity Management service
- Support for preferred credentials for performing privileged operations in the Database Management service
- Support for passing a content encoding when posting metrics in the Monitoring service
- Support for Session Token authentication
  
### Breaking Changes
- The operations `DeleteConnectionValidation` and `ListConnectionValidations` were removed from `DataConnectivityManagementClient` in the Data Connectivity Management service
- The operation `ListConnectionValidationsResponseEnumerator` was removed from `DataConnectivityManagementPaginators` in the Data Connectivity Management service
- The models `ListConnectionValidationsResponse`, `ListConnectionValidationsRequest` and `DeleteConnectionValidationRequest` were removed in the Data Connectivity Management service
- The return type of property `LifecycleState` was changed to `LifecycleStateEnum` from `Registry.LifecycleStateEnum` in `ListRegistriesRequest` model in the Data Connectivity Management service

## 65.18.1 - 2022-08-30
### Added
- Support for opting out of guest VM event collection, health metrics, diagnostics logs, and traces in the Database service
- Support for in-place upgrades for software-defined data centers in the VMWare Solution service
- Support for single-client access name protocol as a data source for private access channels in the Analytics Cloud service
- Support for network security groups, egress control on public datasources, and GitHub access in the Analytics Cloud service
- Support for performance-based autotuning of block and boot volumes in the Block Storage service

## 65.18.0 - 2022-08-23
### Added
- Support for the Enterprise Manager Warehouse service
- Support for additional configuration variables in the MySQL Database service
- Support for file filters in the DevOps service
- Support for support rewards redemption summaries in the Usage service
- Support for the parent tenancy of an organization to view child tenancy categories, recommendations, and resource actions in the Optimizer service
- Support for choosing prior versions during infrastructure maintenance on Exadata Cloud at Customer in the Database service

### Breaking Changes
- The property `parameters` has its object value type changed from `string` to `any`
- EmDataLakeClient is renamed to EmWarehouseClient for the EM Warehouse service.

### Changed
- Compute Out of Capacity error improvement for Go SDK

## 65.17.0 - 2022-08-16
### Added
- Support for Logging Analytics as a streaming source target in the Service Connector Hub service
- Support for data sources for logging query registration in the Cloud Guard service
- Support for custom detector rules on insight detector recipes in the Cloud Guard service
- Support for fetching data source events and problem entities in the Cloud Guard service
- Support for E3, E4, Standard3, and Optimized3 flexible compute shapes on notebooks, model deployment, and jobs in the Data Science service
- Support for streaming application logs to the Logging service in the Data Flow service
 
### Breaking Changes
- Support for retries by default on operations of the Dataflow service

## 65.16.0 - 2022-08-09
### Added
- Support for single-host software-defined data centers in the VMWare Solution service
- Support for Java download and installation in the Java Management service
- Support for lifecycle management for Windows in the Java Management service
- Support for installation scripts in the Java Management service
- Support for unlimited-installation keys in the Java Management service
- Support for configuring automatic usage tracking in the Java Management service
- Support for STANDARDX and ENTERPRISEX instance types in Integration service
- Support for additional languages and multimedia formats in transcription jobs in the AI Speech service
- Support for maintenance run history for Exadata Cloud at Customer in the Database service
- Support for Optimizer statistics monitoring and management on various database administration operations in the Database Management service
- Support for OCI Compute instances in the Operations Insights service
- Support for moving resources in the Console Dashboard service
- Support for round-robin alerting in the Application Performance Monitoring service
- Support for aggregated network data of synthetic monitors in the Application Performance Monitoring service
- Support for etags on operations in the Load Balancing service
 
 
### Breaking Changes
- The enum `UsageUnit` was replaced by `UsageUnitEnum` in the Operations Insights service
- Property `inventoryLog` changed from optional to required in the model `CreateFleetDetails` in Java Management Service

## 65.15.0 - 2022-08-02
### Added
- Support for OpenSearch in the Search service
- Support for child tables in the NoSQL Database Cloud service
- Support for private repositories in the DevOps service

### Breaking Changes
- Support for retries by default on operations of the Quotas service

## 65.14.0 - 2022-07-26
### Added
- Support for the Fusion Apps as a Service service
- Support for the Digital Media service
- Support for accessing all Terraform providers from Hashicorp Registry, as well as bringing your own providers, in the Resource Manager service
- Support for runtime configurations in notebook sessions in the Data Science service
- Support for compartmentIdInSubtree and accessLevel filters when listing management agents in the Management Agent Cloud service
- Support for filtering by type when listing work requests in the Management Agent Cloud service
- Support for filtering by agent id when listing management agent plugins in the Management Agent Cloud service
- Support for specifying size preference when requesting a data transfer appliance in the Data Transfer service
- Support for encryption of boot and block volumes associated with a cluster using customer-specified KMS keys in the Big Data service
- Support for the VM.Standard.E4.Flex shape for Cloud SQL (CSQL) nodes in the Big Data service
- Support for listing block and boot volumes, as well as block and boot volume replicas, within a volume group in the Block Volume service
- Support for dedicated autonomous databases in the Operator Access Control service
- Support for viewing automatic workload repository (AWR) data for databases added to AWRHub in the Operations Insights service
- Support for ports, protocols, roles, and SSL secrets when enabling or modifying database management in the Database service
- Support for monthly security maintenance runs in the Database service
- Support for monthly infrastructure patching for Exadata Cloud at Customer resources in the Database service
  
### Breaking Changes
- `DataMaskingActivityClient`,`FusionEnvironmentClient`, `FusionEnvironmentFamilyClient`, `RefreshActivityClient`,`ScheduledActivityClient`, and `ServiceAttachmentClient` clients were merged into a single client `FusionApplicationsClient` for the Fusion Apps as a Service service
- Properties `addressee`, `address1`, `cityOrLocality`, `stateOrRegion`, `zipcode`, `country` are changed from optional to required for ShippingAddress model in Data Transfer Service.

## 65.13.1 - 2022-07-19
### Added
- Support for calling Oracle Cloud Infrastructure services in the `mx-queretaro-1` region
- Support for the Process Automation service
- Support for the Managed Access service
- Support for extending maintenance reboot due dates on virtual machines in the Compute service
- Support for ingress routing tables on NAT gateways and internet gateways in the Networking service
- Support for container database and pluggable database discovery in the Stack Monitoring service
- Support for displaying rack serial numbers for Exadata infrastructure resources in the Database service
- Support for grace periods for wallet rotation on autonomous databases in the Database service
- Support for hosting models on flexible compute shapes with customizable OCPUs and memory in the Data Science service

## 65.13.0 - 2022-07-12
### Added
- Support for DBCS databases in the Operations Insights service
- Support for point-in-time recovery for non-highly-available database systems in the MySQL Database service
- Support for triggering reboot migration on instances with pending maintenance in the Compute service
- Support for native pod networking in the Container Engine for Kubernetes service
- Support for creating Data Guard associations with new database systems in the Database service
- Fix for double encoding in URL
  
### Breaking Changes
- The data type of the property `HostType` was changed from a List of `string` to a List of `HostTypeEnum` in ListHostInsightsRequest in the Operations Insights service
- The property `PreserveDataVolumes` was removed from the TerminateInstanceRequest in the Compute service

## 65.12.0 - 2022-07-05
### Added
- Support for backup policies returned as part of the database system list operation in the MySQL Database service
 
### Breaking Changes
- Support for retries by default on some operations of the Bastion service

## 65.11.0 - 2022-06-27
### Added
- Support for the Network Monitoring service
- Support for specifying application scan settings when creating or updating host scan recipes in the Vulnerability Scanning service
- Support for moving data into an autonomous data warehouse in the Operations Insights service
- Support for shared infrastructure autonomous database character sets in the Database service
- Support for data collection logging events on Exadata instances in the Database service
- Support for specifying boot volume VPUs when launching instances from images in the Compute service
- Support for safe-deleting nodes in the Container Engine for Kubernetes service

### Breaking Changes
- Support for retries by default on operations of the Logging Analytics service

## 65.10.0 - 2022-06-21
### Added
- Support for the Network Firewall service
- Support for smaller and larger HeatWave cluster nodes in the MySQL Database service
- Support for CSV file type datasets for text labeling and JSONL in the Data Labeling service
- Support for diagnostics in the Database Management service
  
### Breaking Changes
- Support for retries by default on operations of the Network Firewall service
- Support for retries by default on the createAnnotation operation of the Data Labeling service

## 65.9.0 - 2022-06-14
### Added
- Support for the Web Application Acceleration (WAA) service
- Support for the Governance Rules service
- Support for the OneSubscription service
- Support for resource locking in the Identity service
- Support for quota resource locking in the Limits service
- Support for returning the backup with the requested changes in the MySQL Database service
- Support for time zone in Cloud Autonomous VM (CAVM) clusters in the Database service
- Support for configuration options in the Application Performance Monitoring service
- Support for MySQL connections in the Database Tools service
 
### Breaking Changes
- Support for retries by default on operations in the Database Tools service
- Model `DatabaseToolsAllowedNetworkSources`, `DatabaseToolsVirtualSource` and `ServiceCapability`  removed in Database Tools service
- `SecretId` is a required property in `DatabaseToolsUserPasswordSecretIdDetails` model in Databasetools service

## 65.8.1 - 2022-06-07
### Added
- Support for calling Oracle Cloud Infrastructure services in the eu-paris-1 region
- Support for private endpoints in Resource Manager service
- Support downloading generated Terraform plan output in JSON or binary format in Resource Manager service
- Support for querying OPSI Data Objects in the Operations Insights service
- Added 400-ResourceDisabled to EventualConsistency retry
### Changed
- Network security groups (NSGs) are now optional for autonomous databases on private endpoints in the Database service
- Changed the scope of `DefaultSDKLogger`, `SetSDKLogger` and `NewSDKLogger` functions to public for enabling logs by code

## 65.8.0 - 2022-05-31
### Added
- Support for in-depth monitoring, diagnostics capabilities, and advanced management functionality for on-premise Oracle databases in the Database Management service
- Support for using Oracle Cloud Agent to perform iSCSI login and logout for non-multipath-enabled iSCSI attachments in the Container Engine for Kubernetes service
- Support for Fault Domain placement in the Container Engine for Kubernetes service
- Support for worker node images in the Container Engine for Kubernetes service
- Support for flexible shapes using the driverShapeConfig and executorShapeConfig properties in the Data Flow service

### Breaking Changes
- Support for retries by default on operations in the Application Dependency Management service

## 65.7.0 - 2022-05-24
### Added
- Support for the License Manager service
- Support for usage plans in the API Gateway service
- Support for packaged skill and instance metadata management, role-based access options on instance creation, and assigned ownership in the Digital Assistant service
- Support for compute capacity reservations in the VMWare Solution service
- Support for Oracle Linux 8 application streams in the OS Management service
  
### Breaking Changes
- Support for retries by default on operations in the API Gateway service
- Property `specification` is changed from optional to required from model `Deployment` and `CreateDeploymentDetails` in the API Gateway service

## 65.6.0 - 2022-05-17
### Added
- Support for information requests in the Operator Access Control service
- Support for Helm charts and repositories on deployments in the DevOps service
- Support for Application Dependency Management service scan results on builds in the DevOps service
- Support for build resources to use Bitbucket Cloud repositories for source code in the DevOps service
- Support for character set selection on autonomous dedicated databases in the Database service
- Support for listing autonomous dedicated database supported character sets in the Database service
- Support for AMD E4 flex shapes on virtual machine database systems in the Database service
- Support for terraform and improvements for cross-region ADGs in the Database service

### Breaking Changes
- Support for retries by default on GET and LIST operations in the Visual Builder service


## 65.5.0 - 2022-05-10
### Added
- Support for getting usage information for autonomous databases and Cloud at Customer autonomous databases in the Database service
- Support for the `standby` lifecycle state on autonomous databases in the Database service
- Support for BIP connections and dataflow operators in the Data Integration service

### Breaking Changes
- Support for retries by default on WAF Edge Policy GET / LIST operations in the Web Application Acceleration and Security service
- Support for retries by default on some operations in the Stack Monitoring service
- Support for retries by default on some resource discovery and monitoring operations in the Application Management service
- Support for retries by default on some operations in the MySQL Database service

## 65.4.0 - 2022-05-03
### Added
- Support for the Application Dependency Management service
- Support for platform configuration options on some bare metal shapes in the Compute service
- Support for shielded instances for BM.Standard.E4.128 and BM.Standard3.64 shapes in the Compute service
- Support for E4 dense VMs on launch and update instance operations in the Compute service
- Support for reboot migration on DenseIO shapes in the Compute service
- Support for an increased database name maximum length, from 14 to 30 characters, in the Database service
- Support for provisioned concurrency in the Functions service

### Breaking
- Support for retries by default on operations in the Vault service
- Support for retries by default on operations in the DNS service
- Support for retries by default on operations in the Content Management service
- Support for retries by default on operations in the Console Dashboard service
- Support for retries by default on Web Application Firewall operations in the Web Application Acceleration and Security service
- Support for retries by default on operations in the Data Science service

## 65.3.0 - 2022-04-26
### Added
- Support for the Service Mesh service
- Support for security zones in the Cloud Guard service
- Support for virtual test access points (VTAPs) in the Networking service
- Support for monitoring as a source in the Service Connector Hub service
- Support for creating budgets that target subscriptions and child tenancies in the Budgets service
- Support for listing shapes and specifying a shape during creation of a node in the Roving Edge Infrastructure service
- Support for bringing your own key in the Roving Edge Infrastructure service
- Support for enabling inspection of HTTP request bodies in the Web Application Acceleration and Security
- Support for cost management schedules in the Usage service
- Support for TCPS on external containers as well as non-container and pluggable databases in the Database service
- Support for autoscaling on Open Data Hub (ODH) clusters in the Big Data service
- Support for creating Open Data Hub (ODH) 0.9 clusters in the Big Data service
- Support for Open Data Hub (ODH) patch management in the Big Data service
- Support for customizable Kerberos realm names in the Big Data service
- Support for dedicated vantage points in the Application Performance Monitoring service
- Support for reactivating child tenancies in the Organizations service
- Support for punctuation and the SRT transcription format in the AI Speech service
 
### Breaking Changes
- Support for default retries on some operations in the Networking service
- Support for default retries on all operations in the Data Safe service
- Support for default retries on some additional operations in the Application Performance Monitoring service
- The deprecated parameter `RiskScore` was removed in the sighting model in the Cloud Guard service

## 65.2.0 - 2022-04-19
### Added
- Support for the Stack Monitoring service
- Support for stack monitoring on external databases in the Database service
- Support for upgrading VM database systems in place in the Database service
- Support for viewing supported VMWare software versions when listing host shapes in the VMWare Solution service
- Support for choosing compute shapes when creating SDDCs and ESXi hosts in the VMWare Solution service
- Support for work requests on delete operations in the Vulnerability Scanning service
- Support for additional scan metadata in reports, including CVE descriptions, in the Vulnerability  Scanning service
- Support for redemption codes in the Usage service

### Breaking Changes
- The property `Etag` was removed from ListRedeemableUsersResponse model in the Usage service


## 65.1.0 - 2022-04-12
### Added
- Support for bringing your own IPv6 addresses in the Networking service
- Support for specifying database edition and maximum CPU core count when creating or updating an autonomous database in the Database service
- Support for enabling and disabling data collection options when creating or updating Exadata Cloud at Customer VM clusters in the Database service
 
### Breaking Changes
- Support for retries by default on operations in the Identity service
- Support for retries by default on operations in the Operations Insights service

## 65.0.0 - 2022-04-05
### Added
- Support for content length and content type response headers when downloading PDFs in the Account Management service
- Support for creating Enterprise Manager-based zLinux host targets, creating alarms, and viewing top process analytics in the Operations Insights service
- Support for diagnostic reboots on VM instances in the Compute service

### Breaking Changes
- The return type of property LifecycleState was changed from `LifecycleState` to `TargetDatabaseLifecycleState` for TargetDatabase and TargetDatabaseSummary model in the Data Safe service

## 64.0.0 - 2022-03-29
### Added
- Support for returning the number of network ports as part of listing shapes in the Compute service
- Support for Java runtime removal and custom logs in the Java Management service
- Support for new parameters for BGP admin state and enabling/disabling BFD in the Networking service
- Support for private OKE clusters and blue-green deployments in the DevOps service
- Support for international customers to consume and launch third-party paid listings in the Marketplace service
- Support for additional fields on entities, attributes, and folders in the Data Catalog service
 
### Breaking Changes
- Support for retries by default on operations in the Marketplace service

## 63.0.0 - 2022-03-22
### Added
- Support for getting the storage utilization of a deployment on deployment list and get operations in the GoldenGate service
- Support for virtual machines, bare metal machines, and Exadata databases with private endpoints in the Operations Insights service
- Support for setting deletion policies on database systems in the MySQL Database service

### Breaking Changes
- Support for retries by default on operations in the Data Labeling service (data plane and control plane)

## 62.0.0 - 2022-03-15
### Added
- Support for Ubuntu platforms and unlimited installation keys in the Management Agent Cloud service
- Support for shielded instances in the VMWare Solution service
- Support for application resources in the Data Integration service
- Support for multi-AVM on Exadata Cloud at Customer infrastructure in the Database service
- Support for heterogeneous (VM and AVM) clusters on Exadata Cloud at Customer infrastructure in the Database service
- Support for custom maintenance schedules for AVM clusters on Exadata Cloud at Customer infrastructure in the Database service
- Support for listing vulnerabilities, vulnerability-impacted containers, and vulnerability-impacted hosts in the Vulnerability Scanning service
- Support for specifying an image count when creating or updating container scan recipes in the Vulnerability Scanning service

### Changed
- Improved error message for service error, auth provider error, upload manager and other miscellaneous errors

### Breaking Changes
- LifecycleState type in `workspace_summary` model in dataintegration service changed from `WorkspaceLifecycleStateEnum` to `WorkspaceSummaryLifecycleStateEnum`

## 61.0.0 - 2022-03-08
### Added
- Support for the Sales Accelerator license option in the Content Management service
- Support for VCN hostname cluster endpoints in the Container Engine for Kubernetes service
- Support for optionally specifying an admin username and password when creating a database system during a restore operation in the MySQL Database service
- Support for automatic tablespace creation on non-autonomous and autonomous database dedicated targets in the Database Migration service
- Support for reporting excluded objects based on static exclusion rules and dynamic exclusion settings in the Database Migration service
- Support for removing, listing, and adding database objects reported by the Cloud Premigration Advisor Tool (CPAT) in the Database Migration service
- Support for migrating Oracle databases from the AWS RDS service to OCI as autonomous databases, using the AWS S3 service and DBLINK for data transfer, in  the Database Migration service
- Support for querying additional fields of a resource using return clauses in the Search service
- Support for clusters and station clusters in the Roving Edge Infrastructure service
- Support for creating database systems and database homes using customer-managed keys in the Database service
 
### Breaking Changes
- Support for retries enabled by default on operations in the Container Engine for Kubernetes service
- Support for retries enabled by default on operations in the Resource Manager service
- Support for retries enabled by default on operations in the Search service

## 60.0.0 - 2022-03-01
### Added
- Support for DRG route distribution statements to be specified with a new match type 'MATCH_ALL' for matching criteria in the Networking service
- Support for VCN route types on DRG attachments for deciding whether to import VCN CIDRs or subnet CIDRs into route rules in the Networking service
- Support for CPS offline reports in the Database service
- Support for infrastructure patching v2 features in the Database service
- Support for auto-scaling the storage of an autonomous database, as well as shrinking an autonomous database, in the Database service
- Support for managed egress via a default networking option on jobs and notebooks in the Data Science service
- Support for more types of saved search enums in the Management Dashboard service
### Breaking Changes
- Support for retries enabled by default on some operations in the AI Vision service


## 59.0.0 - 2022-02-22
### Added
- Support for the Data Connectivity Management service
- Support for the AI Speech service
- Support for disabling crash recovery in the MySQL Database service
- Support for detector recipes of type 'threat', new detector rule of type 'rogue user', and sightings operations in the Cloud Guard service
- Support for more VM shape configurations when listing shapes in the Compute service
- Support for customer-managed encryption keys in the Analytics Cloud service
- Support for FastConnect device information in the Networking service
 
### Breaking Changes
- Update the property `riskLevel` from required to optional in `TargetDetectorDetails` in Cloud Guard service.
- Update the property `riskLevel` from required to optional in `DetectorDetails` in Cloud Guard service.
- Support for retries enabled by default on all operations in the Application Performance Monitoring control plane service

## 58.0.0 - 2022-02-15
### Added
- Support for the AI Vision service
- Support for the Threat Intelligence service
- Support for creation of NoSQL database tables with on-demand throughput capacity in the NoSQL Database Cloud service
- Support for tagging features in the Oracle Container Engine for Kubernetes (OKE) service
- Support for trace snapshots in the Application Performance Monitoring service
- Support for auditing and alerts in the Data Safe service
- Support for data discovery and data masking in the Data Safe service
- Support for customized subscriptions and delivery of announcements by email and SMS in the Announcements service
- Support for case insensitive validation for enum values
### Breaking Changes
- API `QueryOld` is removed from QueryClient in APM Traces service

## 57.0.0 - 2022-02-08
### Added
- Support for managing tablespaces in the Database Management service
- Support for upgrading and managing payment for subscriptions in the Account Management service
- Support for listing fast launch job configurations in the Data Science service

### Breaking Changes
- Support for retries enabled by default on all operations in the Application Performance Monitoring service
- Support for enum value validation when sending API requests
- The data type of the property BillToAddress was changed from `Address` to `BillToAddress` for the Invoice model of the Account Management service

## 56.1.0 - 2022-02-01
### Added
- Support for calling Oracle Cloud Infrastructure services in the ap-dcc-canberra-1 region
- Support for the Console Dashboard service
- Support for capacity reservation in the Container Engine for Kubernetes service
- Support for tagging in the Container Engine for Kubernetes service
- Support for fetching listings by image OCID in the Marketplace service
- Support for underscores and hyphens in project resource names in the DevOps service
- Support for cross-region cloning in the Database service

## 56.0.0 - 2022-01-25
### Added
- Support for OneSubscription services
- Support for specifying if a run or application is streaming or batch in the Data Flow service
- Support for updating the Instance Configuration of an Instance Pool within a Cluster Network in the Compute Management service
- Updated documentation for Cross Region ADG feature for Autonomous Database in the Database service

### Breaking Changes
- Support for retries enabled by default on all operations in the Object Storage service

## 55.1.0 - 2022-01-18
### Added
- Support for calling Oracle Cloud Infrastructure services in the me-dcc-muscat-1 region
- Support for the Visual Builder service
- Support for cross-region replication of volume groups in the Block Storage service
- Support for boot volume encryption in the Container Engine for Kubernetes service
- Support for adding metadata to records when creating and updating records in the Data Labeling service
- Support for global export formats in snapshot datasets in the Data Labeling service
- Support for adding labeling instructions to datasets in the Data Labeling service
- Support for updating autonomous dataguard associations for autonomous container databases in the Database service
- Support for setting up automatic failover when creating autonomous container databases in the Database service
- Support for setting the RECO storage size when updating a database system in the Database service
- Support for reconnecting refreshable clones to source for autonomous databases on shared infrastructure in the Database service
- Support for checking if an autonomous database on shared infrastructure can be reconnected to source, in the Database service

## 55.0.0 - 2022-01-11
### Added
- Support for calling Oracle Cloud Infrastructure services in the af-johannesburg-1 and eu-stockholm-1 regions
- Support for multiple protocols on the same listener in the Network Load Balancing service
- IPv6 support in the Network Load Balancing service
- Support for creating Enterprise Manager-based Solaris and SunOS host targets in the Operations Insights service
- Support for choosing Data Guard type (Active Data Guard or regular) on databases in the Database service

### Breaking Changes
- Support for retries enabled by default on all operations in the Java Management service'

## 54.0.0 - 2021-12-14
### Added
- Support for node replacement in the VMWare Solution service
- Support for ingestion of SQL stats metrics in the Operations Insights service
- Support for AWR hub integration in the Operations Insights service
- Support for automatically generating logical entities from filename patterns and relationships between business terms across glossaries in the Data Catalog service
- Support for automatic start/stop at scheduled times in the Database service
- Support for cloud VM cluster resources on autonomous dedicated databases in the Database service
- Support for external Hive metastores in the Big Data service
- Support for batch detection/inference in the AI Language service
- Support for dimensions on monitoring targets in the Service Connector Hub service
- Support for invoice operations in the Account Management service
- Support for custom CA trust stores in the API Gateway service
- Support for generating scoped database tokens in the Identity service
- Support for database passwords for users, for logging into database accounts, in the Identity service

### Breaking changes
- Support for retries enabled by default on some operations in the Data Catalog service
- Support for retries enabled by default on all operations in the Ocvp service

## 53.1.0 - 2021-12-07
### Added
- Support for the Application Management service
- Support for getting the inventory of JMS resources and listing Java runtime usage in a specified host in the Java Management service
- Support for categories, entity topology, and verifying scheduled tasks in the Logging Analytics service
- Support for RAC databases in the GoldenGate service
- Support for querying additional fields of a resource using return clauses in the Search service
- Support for key versions and key version OCIDs in the Key Management service

## 53.0.0 - 2021-11-30
### Added
- Support for SQL Tuning Advisor in the Database Management service
- Support for listing users and getting user details in the Database Management service
- Support for autonomous databases in the Database Management service
- Support for enabling and disabling Database Management features on autonomous databases in the Database service
- Support for the Solaris platform in the Management Agent Cloud service
- Support for cross-compartment operations in the Operations Insights service
- Support for listing deployment backups in the GoldenGate service
- Support for standard tags in the Identity service
- Support for viewing problems for deleted targets in the Cloud Guard service
- Support for choosing a platform version while creating a platform instance in the     Blockchain Platform service
- Support for custom IPSec connection tunnel internet key exchange phase 1 and phase 2 encryption algorithms in the Networking service
- Support for pagination when listing work requests corresponding to an APM domain in the Application Performance Monitoring service
- Support for the "deleted" lifecycle state on APM domains in the Application Performance Monitoring service
- Support for calling Oracle Cloud Infrastructure services in the eu-milan-1 and me-abudhabi-1 regions

### Breaking
- Support for retries enabled by default in all operations of the DevOps, Build, and Source Code Management services

## 52.0.0 - 2021-11-17
### Added
- Support for getting subnet topology in the Networking service
- Support for encrypted FastConnect resources in the Networking service
- Support for performance and high availability, as well as recommendation metrics, in the Optimizer service
- Support for optional TDE wallet passwords in the Database service
- Support for Object Storage service integration in the Big Data service

### Breaking
- Circuit breakers enabled by default in all services except Streaming and Compute
- Retries enabled by default in all operations of the Functions and Roving Edge services, and in some operations of the Streaming service.

## 51.0.0 - 2021-11-09
### Added
- Support for drill down metadata in the Management Dashboard service
- Support for operator access control on dedicated autonomous databases in the Operator Access Control service

### Breaking changes
- `OperatorControlName`, `ApproverGroupsList` and `IsFullyPreApproved` changed from optional to required in UpdateOperatorControlDetails
- `IsEnforcedAlways` changed from optional to required in UpdateOperatorControlAssignmentDetails
- `ResourceType` in OperatorControlAssignmentSummary changed return type from *string to ResourceTypesEnum
- `ApproverGroupsList`, `IsFullyPreApproved` and `ResourceType` in CreateOperatorControlDetails changed from optional to required
- `ResourceType` and `IsEnforcedAlways` in CreateOperatorControlAssignmentDetail changed from optional to required

## 50.1.0 - 2021-11-02
### Added
- Support for the Database Tools service
- Support for scan listener port TCP and TCP SSL on cloud VM clusters in the Database service
- Support for domains in the Identity service
- Support for redeemable users and support rewards in the Usage service
- Support for calling Oracle Cloud Infrastructure services in the ap-singapore-1 and eu-marseille-1 regions
- Endpoint for Identity service changed to include ".oci" subdomain
- Support for unknown region fallback to secondary level domain configured through environment variable `OCI_DEFAULT_REALM`

## 50.0.0 - 2021-10-26
### Added
- Support for the Source Code Management service
- Support for the Build service
- Support for the Certificates service
- Support to create child tenancies in an organization and manage subscriptions in the Organizations service
- Support for Certificates service integration in the Load Balancing service
- Support for creating hosts in specific availability domains in the VMWare Solution service
- Support for user-defined functions and libraries, as well as scheduling and orchestration, in the Data Integration service
- Support for EM-managed Exadatas and EM-managed hosts in the Operations Insights service

### Breaking changes
- Model `ComputeInstanceGroupBlueGreenDeployStageExecutionProgress`, `ComputeInstanceGroupBlueGreenTrafficShiftDeployStageExecutionProgress`, 
`ComputeInstanceGroupCanaryApprovalDeployStageExecutionProgress`, `ComputeInstanceGroupCanaryDeployStageExecutionProgress`,
  `ComputeInstanceGroupCanaryTrafficShiftDeployStageExecutionProgress`, `RunPipelineDeployStageExecutionProgress` and 
  `RunValidationTestOnComputeInstanceDeployStageExecutionProgress` were removed in the Build service

## 49.2.0 - 2021-10-19
### Added
- Support for creating database systems from backups with database software images in the Database service
- Support for optionally providing a SID prefix during Exadata database creation in the Database service
- Support for node subsetting on VM clusters in the Database service
- Support for non-CDB to PDB conversion in the Database service
- Support for default homepages, unprocessed data buckets, and parsing geostats in the Logging Analytics service
- Support for creating instance principal delegation token in a specific region
- Support for circuit breaker feature

## 49.1.0 - 2021-10-12
### Added
- Support for the Data Labeling Service
- Support for the Web Application Firewall service
- Support for querying and setting Application Performance Monitoring configurations in the Application Performance Monitoring service
- Support for the run-once monitor feature and network data collection in the Application Performance Monitoring service
- Support for Oracle Enterprise Manager bridges, source auto-association, source event types mapping, and partitioning and searching data by LogSet in the Logging Analytics service
- Support for Log events APIs used by plugins like fluentd, fluentbit, etc. to upload data in the Logging Analytics service
- Support for a new ActionType: FAILED in work requests in the VMware Provisioning service
- Support for calling Oracle Cloud Infrastructure services in the il-jerusalem-1 region

## 49.0.0 - 2021-10-05
### Added
- Support for configuring Binlog variables in the MySQL Database service.
- Support new response value "OPERATOR" for backup creationType in list and get MDS - backup API in the MySQL Database service.
- Support for SetAutoUpgradableConfig and GetAutoUpgradableConfig operations in - Management Agent Cloud service.
- Support for additional installType filter for List Management Agents, Images and Count - API operations in Management Agent Cloud service.
- Support for list and read DeploymentUpgrade, cancel and restore DeploymentBackup in - the Golden Gate service.
- Support for non-autonomous databases targets, executing Pre-Migration advisor, - uploading Datapump logs into Object Storage bucket, and filtering Database Objects in - the Database Migration service.
- Support for calling Oracle Cloud Infrastructure services in the ap-ibaraki-1 region.

### Breaking
- Property `Display` was removed from `ListWorkRequestErrorRequest`, `ListWorkRequestLogsRequest`, `ListWorkRequestsRequest`  models in Database Migration service
- Property `LifecycleState` was changed to `MigrationLifecycleStatesEnum` type from `Migration`, `MigrationSummary` models in Database Migration service
- Property `CompartmentId` was removed from `UpdateAgentDetails` model in Database Migration service
- Property `TimeStamp` was renamed to `Timestamp` from `WorkRequestError`, `WorkRequestLogEntry` models in Database Migration service
- Property `IsAgentAutoUpgradable` was removed from `UpdateManagementAgentDetails` model in Management Agent service

## 48.0.0 - 2021-09-28
### Added
- Support for autonomous databases and clones on shared infrastructure not requiring mTLS in the Database service
- Support for server-side encryption using object-specific KMS keys in the Object Storage service
- Support for Windows in the Java Management service
- Support for using network security groups in the API Gateway service
- Support for network security groups in the Functions service
- Support for signed container images in the Functions service
- Support for setting message format when creating and updating alarms in the Monitoring service
- Support for user and security assessment features in the Data Safe service

### Breaking changes
- Model `RequestSummarizedApplicationUsageDetails`, `RequestSummarizedInstallationUsageDetails`, `RequestSummarizedJreUsageDetails` 
  and `RequestSummarizedManagedInstanceUsageDetails` were removed in the Java Management service
- Operation `RequestSummarizedApplicationUsage`, `RequestSummarizedInstallationUsage`, `RequestSummarizedJreUsage` and 
`RequestSummarizedManagedInstanceUsage` were removed in the Java Management service
  

## 47.1.0 - 2021-09-14
### Added
- Support for serviceHostKeyFingerprint property for InstanceConsoleConnection in Core service
- Support for Shielded Instances in Core service
- Support for ML Jobs in the Data Science service

## 47.0.0 - 2021-09-07
### Added
- Support for terraform advanced options (detailed log level, refresh, and parallelism) on jobs in the Resource Manager service
- Support for forced cancellation when cancelling jobs in the Resource Manager service
- Support for getting the detailed log content of a job in the Resource Manager service
- Support for provider information in the responses of list operations in the Management Dashboard service
- Support for scheduled jobs in Database Management service
- Support for monitoring and management of OCI virtual machine, bare metal, and ExaCS databases in the Database Management service
- Support for a unified way of managing both external and cloud databases in the Database Management service
- Support for metrics and Performance Hub on virtual machine, bare metal, and ExaCS databases in the Database Management service

### Breaking changes:
- Parameter `OciSplatGeneratedOcids` was removed from operation `CreateTemplate` in the Resource Manager service

## 46.2.0 - 2021-08-31
### Added
- Support for Oracle Analytics Cloud and OCI Vault integration on connections in the Data Catalog service
- Support for critical event monitoring in the OS Management service
- Added eventual consistency for retry strategies

## 46.1.0 - 2021-08-24
### Added
- Support for generating recommended VM cluster networks in the Database service
- Support for creating VM cluster networks with a specified listener port in the Database service

## 46.0.0 - 2021-08-17
### Added
- Support for getting management agent hosts which are eligible to create Operations Insights host resources on, in the Operations Insights service
- Support for getting summarized agent counts and summarized plugin counts in the Management Agent Cloud service

### Breaking
- The type for property `PluginName` was changed from `*string` to `[]string` for `ListManagementAgentsRequest` model under the ManagementAgent service
- The type for property `Version` was changed from `*string` to `[]string` for `ListManagementAgentsRequest` model under the ManagementAgent service
- The type for property `PlatformType` was changed from `ListManagementAgentsPlatformTypeEnum` to `[]PlatformTypesEnum` for  `ListManagementAgentsRequest` model under the ManagementAgent service

## 45.2.0 - 2021-08-03
### Added
- Support for manually copying volume group backups across regions in the Block Volume service
- Support for work requests for the copy volume backup and copy boot volume backup operations in the Block Volume service
- Support for specifying external Hive metastores during application creation in the Data Flow service
- Support for changing the compartment of a backup in the MySQL Database service
- Support for model catalog features including provenance, metadata, schemas, and artifact introspection in the Data Science service
- Support for Exadata system network bonding in the Database service
- Support for creating autonomous databases with early patching enabled in the Database service

## 45.1.0 - 2021-07-27
### Added
- Support for filtering by tag on capacity planning and SQL warehouse list operations in the Operations Insights service
- Support for creating cross-region autonomous data guards in the Database service
- Support for the customer contacts feature on cloud exadata infrastructure in the Database service
- Support for cost analysis custom tables in the Usage service
- Updated THIRD_PARTY_LICENSES and added THIRD_PARTY_LICENSES_DEV file

## 45.0.0 - 2021-07-20
### Added
- Support for schedules, schedule tasks, REST tasks, operators, S3, and Fusion Apps in the Data Integration service
- Support for getting available updates and update histories for VM clusters in the Database service
- Support for downloading network validation reports for Exadata network resources in the Database service
- Support for patch and upgrade of Grid Infrastructure (GI), and update of DomU OS software for VM clusters in the Database service
- Support for updating data guard associations in the Database service

### Breaking
- The property `ModelType` was removed and property `BucketName` was replaced by `BucketSchema` in the models OracleAdwcWriteAttributes and OracleAtpWriteAttributes under the Data Integration service
- The type for property `Type` was changed from `BaseType` to `*interface{}` for Parameter model under the Data Integration service
- The type for property `Type` was changed from `*string` to `*interface{}` for ShapeField and NativeShapeField models under the Data Integration service
- Added extraHeaders parameter to `HTTPRequest` method in OCIRquest interface

## 44.0.0 - 2021-07-13
### Added
- Support for the AI Anomaly Detection service
- Support for retrieving a DNS zone as a zone file in the DNS service
- Support for querying manual adjustments in the Usage service
- Support for searching Marketplace listings in the Marketplace service
- Support for new cluster type 'ODH' in the Big Data service
- Support for availability domain as an optional parameter when creating VLANs in the Networking service
- Support for search domain type on DHCP options, to support multi-level domain search in the Networking service

### Breaking
- Parameter `Tsig` in model `external_master` was removed in the DNS service
- model `create_custom_table_details`, `create_schedule_report_details`, `custom_table`, `custom_table_collection`, `custom_table_summary`, `saved_schedule_report`, `schedule_report`, `schedule_report_collection`, `schedule_report_summary`, `update_custom_table_details`, `update_schedule_report_details` were removed in the Usage service

## 43.1.0 - 2021-07-06
### Added
- Support for order activation in the Organizations service
- Support for resource principal authorization on Enterprise Manager bridge resources in the Operations Insights service
- Support for the starter edition license type in the Content and Experience service
- Support for the Generic Artifacts service's new domain name

## 43.0.0 - 2021-06-29
### Added
- Support for the DevOps service
- Support for configuring network security groups for node pools in the Container Engine for Kubernetes service
- Support for optionally specifying CPU core count and data storage size when creating autonomous databases in the Database service
- Support for metastore and initial data asset import/export in the Data Catalog service
- Support for associating domain names to emails and managing email domain names / DKIM in the Email Delivery service
- Support for email domain names on senders and suppressions in the Email Delivery service
- Add multipart download example

### Breaking changes
- Property `LifecycleState` in model `SenderSummary`'s type was changed from `SenderSummaryLifecycleStateEnum` to `SenderLifecycleStateEnum` 
  in the Email Delivery service
- Parameter `SoryBy` in the operation `ListJobExecutions`'s type `ListJobExecutionsSortByEnum`, item `ListJobExecutionsSortByDisplayname`
  was removed in the Data Catalog service

    

## 42.1.0 - 2021-06-22
### Added
- Support for virtual machine and bare metal pluggable databases in the Database service

## 42.0.0 - 2021-06-15
### Added
- Support for elastic storage on Exadata Infrastructure resources for Cloud at Customer in the Database service
- Support for registration and management of target databases in the Data Safe service
- Support for config on metadata in the Management Dashboard service
- Support for a new work request operation type for node pool reconciliation events in the Container Engine for Kubernetes service
- Support for migrating clusters with a public Kubernetes API endpoint which are not integrated with a customer's VCN to a VCN-native cluster in the Container Engine for Kubernetes service
- Support for getting the spark version of applications, and filtering applications by spark version, in the Data Flow service

### Breaking
- Propertry `FreeformTags` and `DefinedTags` were removed from the management_dashboard_export_details model in the Management Dashboard service

## 41.2.0 - 2021-06-08
### Added
- Support for Java Management service
- Support for resource principals for the Enterprise Manager bridge resource in Operations Insights service
- Support for encryptionInTransitType in BootVolumeAttachment and IScsiVolumeAttachment in Core service
- Support for updating iscsiLoginState for VolumeAttachment in Core service
- Support for a new type of Source called Import for use with the Export tool in Application Migration service
- Support for Expect/100-continue HTTP header. Expect headers are added by default for all PUT/POST operations

## 41.1.0 - 2021-06-01
### Added
- Support for configuration of autonomous database KMS keys in the Database service
- Support for creating database software images with any supported RUs in the Database service
- Support for creating database software images from an existing database home in the Database service
- Support for listing all NSGs associated with a given VLAN in the Networking service
- Support for a duration windows, task failure reasons, and next execution times on scheduled tasks in the Logging Analytics service 
- Support for calling Oracle Cloud Infrastructure services in the sa-vinhedo-1 region

## 41.0.0 - 2021-05-25
### Added
- Support for the Generic Artifacts service
- Support for the Bastion service
- Support for reading secrets by name in the Vault service
- Support for the isDynamic field when listing definitions in the Limits service
- Support for getting billable image sizes in the Compute service
- Support for getting Automatic Workload Repository (AWR) data on external databases in the Database Management service
- Support for the VM.Standard.E3.Flex flexible compute shape with customizable OCPUs and memory on notebooks in the Data Science service
- Support for container images and generic artifacts billing in the Registry service
- Support for the HCX Enterprise add-on in the VMware Solution service

### Breaking changes
- Property `Name` of Model `SupportedSkuSummary` type changed from `SupportedSkuSummaryNameEnum` to `SkuEnum` in the VMware Solution service

## 40.4.0 - 2021-05-18
### Added
- Support for spark-submit compatible options in the Data Flow service
- Support for Object Storage as a configuration source in the Resource Manager service
  
### Fixed
- Fixed UploadManager creates too many small parts issue

## 40.3.0 - 2021-05-11
### Added
- Support for creating notebook sessions with larger block volumes in the Data Science service
- Support for database maintenance run patch modes in the Database service

## 40.2.0 - 2021-05-04
### Added
- Support for the Operator Access Control service
- Support for the Service Catalog service
- Support for the AI Language service
- Support for autonomous database on Exadata Cloud at Customer infrastructure patching in the Database service
- Added default retry policy, which retries on 409(IncorrectState), 429(TooManyRequests) and any 5XX errors except 
  501(MethodNotImplemented), and uses exponential backoff




## 40.1.0 - 2021-04-27
### Added
- VCN id parameters were moved from being required to being optional on all list operations in the Networking service
- Support for RACs (real application clusters) for external container, non-container, and pluggable databases in the Database service
- Support for data masking in the Cloud Guard service
- Support for opting out of DNS records during instance launch, as well as attaching secondary VNICs, in the Compute service
- Support for mutable sizes on cluster networks in the Autoscaling service
- Support for auto-tiering on buckets in the Object Storage service




## 40.0.0 - 2021-04-20
### Added
- Support for opting in/out of live migration on instances in the Compute service
- Support for enabling/disabling Operations Insights on external non-container and external pluggable databases in the Database service
- Support for a GraphStudio URL as a connection URL on databases in the Database service
- Support for adding customer contacts on autonomous databases in the Database service
- Support for name annotations on harvested objects in the Data Catalog service
- Fixed retry doesn't work once the request is with binary request body issue, for detail, can refer https://github.com/oracle/oci-go-sdk/blob/master/oci.go#L271

### Breaking changes
- Added a method `BinaryRequestBody()` to interface `OCIRetryableRequest`, any data type inherit the interface has to implement the method

## 39.0.0 - 2021-04-13
### Added
- Support for the Database Migration service
- Support for the Networking Topology service
- Support for getting the id of peered VCNs on local peering gateways in the Networking service
- Support for burstable instances in the Compute service
- Support for preemptible instances in the Compute service
- Support for fractional resource usage and availability in the Limits service
- Support for streaming analytics in the Service Connector Hub service
- Support for flexible routing inside DRGs to enable packet flow between any two attachments in the Networking service 
- Support for routing policy to customize dynamic import/export of routes in the Networking service
- Support for IPv6, including on FastConnect and IPsec resources, in the Networking service
- Support for request validation policies in the API Gateway service
- Support for RESP-compliant (e.g. REDIS) response caches, and for configuring response caching per-route in the API Gateway service
- Support for flexible billing in the VMWare Solution service
- Support for new DNS format for the Web Application Acceleration and Security service
- Support for configuring APM tracing on applications and functions in the Functions service
- Support for Enterprise Manager external databases and Management Agent Service managed external databases and hosts in the Operations Insights service
- Support for getting cluster cache metrics for RAC CDB managed databases in the Database Management service

### Breaking Changes
- Property `IsInternetAccessAllowed` in model `CreateIpv6Details` was removed in the Networking service
- Property `Ipv6CidrBlock` in model `CreateVcnDetails` was removed in the Networking service
- Property `IsInternetAccessAllowed` and `PublicIpAddress` in model `Ipv6` were removed in the Networking service
- Property `Ipv6PublicCidrBlock` in model `Subnet` was removed in the Networking service
- Property `IsInternetAccessAllowed` in model `UpdateIpv6Details` was removed in the Networking service
- Property `Ipv6CidrBlock` and `Ipv6PublicCidrBlock` in model `Vcn` were removed in the Networking service
- Property `CurrentSku` in model `CreateEsxiHostDetails` was added in the VMWare Solution service
- Property `InitialSku` in model `CreateSddcDetails` was added in the VMWare Solution service
- Model `DatabaseInsightSummary` type was changed from struct to interface in the Operations Insights service

## 38.1.0 - 2021-04-06
### Added
- Support for scheduling the suspension and resumption of compute instance pools based on predefined schedules in the Autoscaling service
- Support for database software images for Cloud@Customer in the Database service
- Support for OCIC IDCS authorization details in the Application Migration service
- Support for cross-region asynchronous volume replication in the Block Storage service
- Support for SDK generation in the API Gateway service
- Support for container image signing in the Registry service
- Support for cluster features as a part of the Container Engine for Kubernetes service
- Support for filtering dedicated virtual machine hosts by remaining memory and OCPUs in the Compute service
- Support for read/write-any object from buckets using pre-authenticated requests in the Object Storage service
- Support for restricting pre-authenticated requests by prefix in the Object Storage service
- Support for route filtering on public virtual circuits in the Virtual Networking service

## 38.0.0 - 2021-03-30
### Added
- Support for the Vulnerability Scanning service
- Support for vSphere 7.0 in the VMware Solution service
- Support for forecasting in the Usage service
- Support for viewing, searching, and modifying parameters for on-premise Oracle databases in the Database Management service
- Support for listing tablespaces of managed databases in the Database Management service
- Support for cross-regional replication of keys in the Key Management service
- Support for highly-available database systems in the MySQL Database service
- Support for Oracle Enterprise Manager bridges, source auto-association, source event type mappings, and plugins to upload data in the Logging Analytics service

### Breaking changes
- Model `Forecast`'s Enum value was changed from `ForecastForcastTypeEnum` to `ForecastForecastTypeEnum` in the Usage service
- Operation `ListLookups`'s Enum value was changed from `ListLookupsStatusSuccesful` to `ListLookupsStatusSuccessful` in the
Logging Analytics service

## 37.0.0 - 2021-03-23
### Added
- Support for the Network Load Balancing service
- Support for maintenance runs on autonomous databases in the Database service
- Support for announcement preferences in the Announcements service
- Support for domain claiming in the Organizations service
- Support for saved reports in the Usage service
- Support for the HeatWave in-memory analytics accelerator in the MySQL Database service
- Support for community applications in the Marketplace service
- Support for capacity reservations in the Compute service

### Breaking changes
- Operation `ListWorkRequests`'s param `Status`'s type was changed from `[]ListWorkRequestsStatusEnum` to 
`[]WorkRequestStatusEnum` in the Analytics service
- Operation `RequestSummarizedProblems`'s parameter `ListDimensions`'s type was changed from 
`[]RequestSummarizedProblemsListDimensionsEnum` to `[]ProblemDimensionEnum` in the Cloudguard service
- Operation `RequestSummarizedResponderExecutions`'s parameter `ResponderExecutionsDimensions`'s type was changed from 
`[]RequestSummarizedResponderExecutionsResponderExecutionsDimensionsEnum` to `[]ResponderDimensionEnum` in the Cloudguard service
- Operation `ListClusters`'s parameter `LifecycleState`'s type was changed from `[]ListClustersLifecycleStateEnum` to 
`[]ClusterLifecycleStateEnum` in the ContainerEngine service
- Model `Attribute`'s property `AssociatedRuleTypes`'s type was changed from `[]AttributeAssociatedRuleTypesEnum` to 
`[]RuleTypeEnum` in the Datacatalog service
- Model `AttributeSummary`'s property `AssociatedRuleTypes`'s type was changed from `[]AttributeSummaryAssociatedRuleTypesEnum`
to `[]RuleTypeEnum` in the Datacatalog service
- Operation `ListCustomProperties`'s parameter `DataTypes`'s type was changed from `[]ListCustomPropertiesDataTypesEnum` to 
`[]CustomPropertyDataTypeEnum` in the Datacatalog service
- Operation `Recommendations`'s parameter `RecommendationType`'s type was changed from `[]RecommendationsRecommendationTypeEnum`
to `[]RecommendationTypeEnum` in the Datacatalog service
- Operation `ListListings`'s parameter `Pricing`'s type was changed from `[]ListListingsPricingEnum`  to `[]PricingTypeEnumEnum`
in the Marketplace service
- Operation `ListListings`'s parameter `ListingTypes`'s type was changed from `[]ListListingsListingTypesEnum`  to `[]ListingTypeEnum`
in the Marketplace service
- Operation `ListAddressLists`'s parameter `LifecycleState`'s type was changed from `[]ListAddressListsLifecycleStateEnum` to
`LifecycleStatesEnum` in the Waas service
- Operation `ListCertificates`'s parameter `LifecycleState`'s type was changed from `[]ListCertificatesLifecycleStateEnum` to 
`[]LifecycleStatesEnum` in the Waas service
- Operation `ListCustomProtectionRules`'s parameter `LifecycleState`'s type was changed from `[]ListCustomProtectionRulesLifecycleStateEnum`
to `[]LifecycleStatesEnum` in the Waas service
- Operation `ListHttpRedirects`'s parameter `LifecycleState`'s type was changed from `[]ListHttpRedirectsLifecycleStateEnum`
to `[]LifecycleStatesEnum` in the Waas service
- Operation `ListWaasPolicies`'s parameter `LifecycleState`'s type was changed from `[]ListWaasPoliciesLifecycleStateEnum` to
`LifecycleStatesEnum` in the Waas service
- Operation `ListWorkRequestErrors`'s parameter `CompartmentId` was removed in the Tenantmanagercontrolplane Service
- Model `Ipv6`'s property `VnicId` was tagged as mandatory in the VCN service
- Model `CreateIpv6Details`'s property `VnicId` was tagged as mandatory in the VCN service

## 36.2.0 - 2021-03-16
### Added
- Support for routing policies and HTTP2 listener protocols in the Load Balancing service
- Support for model deployments in the Data Science service
- Support for private clusters in the Container Engine for Kubernetes service
- Support for updating an instance's usage type in the Content and Experience service

## 36.1.0 - 2021-03-09
### Added
- Support for the Application Performance Monitoring service
- Support for the Golden Gate service
- Support for SMS subscriptions in the Notifications service
- Support for friendly-formatted messages in the Service Connector Hub service
- Support for attaching and detaching instances to instance pools in the Autoscaling service

## 36.0.0 - 2021-03-02
### Added
- Support for pipelines, pipeline tasks, and favorites in the Data Integration service
- Support for publishing tasks to OCI Data Flow in the Data Integration service
- Support for clones in the File Storage service

### Breaking changes
- Changed model `UniqueKey` type from struct to interface in the Data Integration service
- Removed property `ModelType` from Model `PrimaryKey` in the Data Integration service
- Changed model `ForeignKey` property `ReferenceUniqueKey` type from `*UniqueKey` to `UniqueKey` in the Data Integration service
- Removed KeyModelTypeEnum enum type `PRIMARY_KEY` and `UNIQUE_KEY` from model `key` in the Data Integration service

## 35.3.0 - 2021-02-23
### Added
- Support for the OCI Registry service
- Support for exporting an existing running VM, or a copy of VM, into a VMDK, QCOW2, VDI, VHD, or OCI formatted image in the Compute service
- Support for platform configurations on instances in the Compute service
- Support for providing target tags and target compartments on profiles in the Optimizer service
- Support for the 'Fix it' feature in the Optimizer service

## 35.2.0 - 2021-02-16
### Added
- Support for scan DNS names and zone ids on database system, cloud VM cluster, and autonomous Exadata infrastructure responses in the Database service
- Support for specifying ACL rules to limit ingress into public load balancers in the Integration service
- Support for Cloud at Customer as a source type in the Application Migration service
- Support for selective migration of specific resources in the Application Migration service

## 35.1.0 - 2021-02-09
### Added
- Support for the Database Management service
- Support for setting an offset for budget processing in the Budgets service
- Support for enabling and disabling Oracle Cloud Agent plugins in the Compute service
- Support for listing available plugins and for getting the status of plugins in the Oracle Cloud Agent service
- Support for one-off patching in autonomous transaction processing - dedicated databases in the Database service
- Support for additional database upgrade options in the Database service
- Support for glossary term recommendations in the Data Catalog service
- Support for listing errata in the OS Management service

## 35.0.0 - 2021-02-02
### Added
- Support for checking if a contact for Exadata infrastructure is valid in My Oracle Support in the Database service
- Support for checking if Exadata infrastructure is in a degraded state in the Database service
- Support for updating the operating system on a VM cluster in the Database service
- Support for external databases in the Database service
- Support for uploading objects to the infrequent access storage tier in the Object Storage service
- Support for changing the storage tier of existing objects in the Object Storage service
- Support for private templates in the Resource Manager service
- Support for multiple encryption domains on IPSec tunnels in the Networking service

### Breaking changes
- Header Parameter `Etag` in Operation `ListAppCatalogListingResourceVersions` response was removed from the Core service
- Property `VnicId` in model `Ipv6` was removed from from the Core service
- Const `GetObjectArchivalStateAvailable` was removed from operation `GetObject` response from the Object Storage service

## 34.0.0 - 2021-01-26
### Added
- Support for creating, managing, and using asymmetric keys in the Key Management service
- Support for peer ACD unique names in Exadata Cloud at Customer in the Database service
- Support for ACLs on autonomous databases in Exadata Cloud at Customer Data Guard in the Database service
- Support for drift detection on individual resources of a stack in the Resource Manager service
- Support for private access channels and vanity URLs in the Analytics Cloud service
- Support for updating load balancer shapes in the Blockchain Platform service
- Support for assigning volume backup policies to volume groups in the Block Volume service

### Breaking changes
- Property `IdcsAccessToken` in model `CreateBlockchainPlatformDetails` changed from `optional` to `required` in the Blockchain Platform service
- Const `WrappedImportKeyWrappingAlgorithmRsaOaepSha256` was removed from model `WrappedImportKey` in the Key Management service

## 33.0.0 - 2021-01-19
### Added
- Support for Logging Analytics as a target in the Service Connector Hub service
- Support for lookups, agent collection warnings, task commands, and data archive/recall in the Logging Analytics service

### Fixed
- Fixed a bug in the endpoint used for the Management Dashboard service

### Breaking changes
- Parameter `SortBy` type in request `ListMetaSourceTypesRequest` changed from `*string` to `ListMetaSourceTypesSortByEnum` in the Logging Analytics service
- Parameter `SortBy` type in request `ListParserFunctionsRequest` changed from `*string` to `ListParserFunctionsSortByEnum` in the Logging Analytics service
- Parameter `SortBy` type in request `ListParserMetaPluginsRequest` changed from `*string` to `ListParserMetaPluginsSortByEnum` in the Logging Analytics service
- Parameter `SortBy` type in request `ListSourceLabelOperatorsRequest` changed from `*string` to `ListSourceLabelOperatorsSortByEnum` in the Logging Analytics service
- Parameter `SortBy` type in request `ListSourceMetaFunctionsRequest` changed from `*string` to `ListSourceMetaFunctionsSortByEnum` in the Logging Analytics service
- Model `UpdateScheduledTaskDetails` type changed from struct to interface
- Model `ScheduledTask` type changed from struct to interface in the Logging Analytics service

## 32.0.0 - 2021-01-12
### Added
- Support for auto-scaling in the Big Data service
- Documentation fixes for the Logging Search service

### Breaking changes
- Removed `NodeLifecycleStateStarting` and `NodeLifecycleStateStopping` from the model of `NodeLifecycleStateEnum` in the Big Data service
- Removed `BdsInstanceLifecycleStateUpdatingInfra` from `BdsInstanceLifecycleStateEnum` from the model of `BdsInstance` in the Big Data service

## 31.0.0 - 2020-12-15
### Added
- Support for filtering listKeys based on KeyShape in KeyManagement service
- Support for the Oracle Roving Edge Infrastructure service
- Support for flexible ShapeDetails in Load Balancer service
- Support for listing of harvested Rules, additional filtering for Logical Entity list calls in Data Catalog service
- Support second level domain for audit SDK
- Support for listing flex components in Database service
- Support for APEX service for ADBS on OCI console for Database service
- Support for Customer-Managed Key features as a part of the Database service
- Support for Github configuration source provider as part of the Resource Manager service

### Breaking changes
- Removed deprecated API `CreateAutonomousDataWarehouse` from Database service
- Removed deprecated API `CreateAutonomousDataWarehouseBackup` from Database service
- Removed deprecated API `DeleteAutonomousDataWarehouse` from Database service
- Removed deprecated API `GenerateAutonomousDataWarehouseWallet` from Database service
- Removed deprecated API `GetAutonomousDataWarehouse` from Database service
- Removed deprecated API `GetAutonomousDataWarehouseBackup` from Database service
- Removed deprecated API `ListAutonomousDataWarehouseBackups` from Database service
- Removed deprecated API `ListAutonomousDataWarehouses` from Database service
- Removed deprecated API `RestoreAutonomousDataWarehouse` from Database service
- Removed deprecated API `StartAutonomousDataWarehouse` from Database service
- Removed deprecated API `StopAutonomousDataWarehouse` from Database service
- Removed deprecated API `UpdateAutonomousDataWarehouse` from Database service
- Method GetLifecycleState()'s return type in the interface `ConfigurationSourceProviderSummary` was changed from 
`ConfigurationSourceProviderSummary` to `ConfigurationSourceProviderLifecycleStateEnum`

## 30.1.0 - 2020-12-08
### Added
- Support for Integration Service custom endpoint feature
- Support for metadata field in IdentityProvider Get and List response
- Support for fine-grained data analysis and improved SQL insights
- Support for ADB Dedicated - ORDS and SSL cert rotation at AEI
- Support for Maintenance Schedule feature for Exadata Infrastructure resources for ExaCC

## 30.0.0 - 2020-12-01
### Added
- Support for calling Oracle Cloud Infrastructure services in the sa-santiago-1 region
- Support for peer and OSN resources, as well as retry tokens, in the - Blockchain Platform service
- Support for getting the availability status of management agents in the Management Agent service
- Support for the on-prem-connector resource type in the Data Safe service
- Support for service channels in the MySQL Database service
- Support for getting the creation type of backups, and for filtering backups by creation type in the MySQL Database service

### Breaking changes
- Update property `IsEnabled` in model `EnableDataSafeConfigurationDetails` of service `datasafe`

## 29.0.0 - 2020-11-17
### Added
- Support for specifying memory for AMD E3 shapes during node pool creation and update in the Container Engine for Kubernetes service
- Support for upgrading a database on a VM database system in the Database service
- Support for listing autonomous database clones in the Database service
- Support for Data Guard with autonomous container databases on Exadata Cloud at Customer in the Database service
- Support for getting the last login time of a user in the Identity service
- Support to bulk editing tags on resources in the Identity service

### Breaking changes
- Models `AgentUpload`, `Attribute`, `CreateNamespaceDetails`, `FieldMap`, `GenerateAgentObjectNameDetails`, 
`LogAnalytics`, `LogAnalyticsCollectionWarning`, `LogAnalyticsSummary`, `OutOfBoxEntityTypeDetails`, `Query`, 
`QueryWorkRequestResource`, `RegisterEntityTypesDetails`, `ServiceTenancy`, `StringListDetails` and property 
`SortOrdersEnum` are removed in loganalytics service
- Property 'Id' type changed from *interface{} to *string of model `LogAnalyticsParserFilter` in loganalytics service
- Property `mappingAbstractCommandDescriptorName` key `CUSLTER_SPLIT` and value `AbstractCommandDescriptorNameCuslterSplit`
were changed to `CLUSTER_SPLIT` and `AbstractCommandDescriptorNameClusterSplit` in loganalytics service

## 28.0.0 - 2020-11-10
### Added
- Support for the 21C autonomous database version in the Database service
- Support for creating a Data Guard association with a standby database from a database software image in the Database service
- Support for specifying a TDE wallet password when creating a database or database system in the Database service
- Support for enabling access control lists for autonomous databases on Exadata Cloud At Customer in the Database service
- Support for private DNS resolvers, resolver endpoints, and views in the DNS service
- Support for getting a VCN and resolver association in the Networking service
- Support for additional parameters when updating subnets and VLANs in the Networking service
- Support for analytics clusters (database accelerators) in the MySQL Database service
- Support for migrations to Java Cloud Service and Oracle Weblogic Server instances that use existing databases in the Application Migration service
- Support for specifying reserved IPs when creating load balancers in the Load Balancing service
- Fix once request header content-length is 0, request body is not nil, then send chunked-encoding header issue

### Breaking changes
- Parameter `LifecycleState` type in `listMigrations` API from applicationmigration service is changed to `ListMigrationsLifecycleStateEnum`
- Parameter `LifecycleState` type in `listSources` API from applicationmigration service is changed to `ListSourcesLifecycleStateEnum`

## 27.3.0 - 2020-11-03
### Added
- Support for calling Oracle Cloud Infrastructure services in the uk-cardiff-1 region
- Support for the Organizations service
- Support for the Optimizer service
- Support for tenancy ID and name on responses in the Usage service
- Support for object versioning in object lifecycle management in the Object Storage service
- Support for specifying a syslog URL for applications in the Functions service
- Support for creation of always-free NoSQL database tables in the NoSQL Database service

## 27.2.0 - 2020-10-27
### Added
- Support for the Compute Instance Agent service
- Support for key store resources and operations in the Database service
- Support for specifying a key store when creating autonomous container databases in the Database service

## 27.1.0 - 2020-10-20
### Added
- Support for the Operations Insights service
- Support for updating autonomous databases to enable/disable Operations Insights service integration, in the Database service
- Support for the NEEDS_ATTENTION lifecycle state on database systems in the Database service
- Support for HCX in the VMware Solutions service

## 27.0.0 - 2020-10-13
### Added
- Support for API definitions in the API Gateway service
- Support for pattern-based logical entities, namespace-bound custom properties, and faceted search in the Data Catalog service
- Support for autonomous Data Guard on autonomous infrastructure in the Database service
- Support for creating a Data Guard association on an existing standby database home in the Database service
- Support for upgrading cloud VM cluster grid infrastructure in the Database service
- Support for applying same retry policy across multiple requests in the same service
- Support for resource principal v1.1 authentication.

### Breaking changes
- Removed property `IsQuickStart` from model `LogSavedSearch`, `LogSavedSearchSummary`, `UpdateLogSavedSearchDetails` and 
`CreateLogSavedSearchDetails` in logging service.
- Removed LogSavedSearchLifecycleStateEnum `DELETED` in logging service

## 26.0.0 - 2020-10-06
### Added
- Support for calling Oracle Cloud Infrastructure services in the me-dubai-1 region
- Support for rotating keys on autonomous container databases and autonomous databases in the Database service
- Support for cloud Exadata infrastructure and cloud VM clusters in the Database service
- Support for controlling the display of tax banners in the Marketplace service
- Support for application references, patch changes, generic JDBC and MySQL data asset types, and publishing tasks to OCI Dataflow in the Data Integration service
- Support for disabling the legacy Instance Metadata endpoints v1 in the Compute service
- Support for instance configurations specifying instance options in the Compute Management service

### Breaking changes
- Removed model `TypedNamePatternRule` method `UnmarshalJSON` in dataintegration service.

## 25.2.0 - 2020-09-29
### Added
- Support for specifying custom content dispositions when downloading objects in the Object Storage service
- Support for the “bring your own IP address” feature in the Virtual Networking service
- Support for updating the tags of instance console connections in the Compute service
- Support for custom SSL certificates on gateways in the API Gateway service

## 25.1.0 - 2020-09-22
### Added
- Support for software keys in the Key Management service
- Support for customer contacts on Exadata Cloud at Customer in the Database service
- Support for updating open modes and permission levels of autonomous databases in the Database service
- Support for flexible memory on VM instances in the Compute and Compute Management services

## 25.0.0 - 2020-09-15
### Added
- Support for the Cloud Guard service
- Support for specifying desired consumption models when creating instances in the Integration service
- Support for dynamic shapes in the Load Balancing service
- Support for allowing clients to return their currently configured endpoint
- Support for running existing code/samples which call the SDK in Cloud Shell without any changes
- Support for dumping request/response body in SDK logging error for 4XX/5XX errors
- Support for Go Modules

### Breaking changes
- All oci-go-sdk imports need to add specific major version for go module users, for more information please refer README.md
- Method `AuthType()` was added to the interface `ConfigurationProvider`, any interface/struct that inherits the interface is expected to implement AuthType()
- Client end logging level was updated from debug to info level for errors in `client.HTTPClient.Do(request)`

## 24.3.0 - 2020-09-08
### Added
- Support for Logging Service
- Support for Logging Analytics Service
- Support for Logging Search Service
- Support for Logging Ingestion Service
- Support for Management Agent Cloud Service
- Support for Management Dashboard Service
- Support for Service Connector Hub service
- Support for Policy based Request/Response transformation in the API Gateway Service
- Support for sending diagnostic interrupt to a VM instance in the Compute Service
- Support for custom Database Software Images in the Database Service
- Support for getting and listing container database patches for Autonomous Container Database resources in the Database Service
- Support for updating patch id on maintenance run for Autonomous Container Database resources in the Database Service
- Support for searching Oracle Cloud resources across tenancies in the Search Service
- Documentation update for Logging Policies in the API Gateway service

## 24.2.0 - 2020-09-01
### Added
- Support for calling Oracle Cloud Infrastructure services in the ap-chiyoda-1 region
- Support for VM database cloning in the Database service
- Support for the MAINTENANCE_IN_PROGRESS lifecycle state on database systems, VM clusters, and Cloud Exadata in the Database service
- Support for provisioning refreshable clones in the Database service
- Support for new options on listeners and backend sets for specifying SSL protocols, SSL cipher suites, and server ordering preferences in the Load Balancing service
- Support for AMD flexible shapes with configurable CPU in the Container Engine for Kubernetes service
- Support for network sources in authentication policies in the Identity service

## 24.1.0 - 2020-08-18
### Added
- Support for custom boot volume size and other node pool updates in the Container Engine for Kubernetes service
- Support for Data Guard on Exadata Cloud at Customer VM clusters in the Database service
- Support for stopping VM instances after scheduled maintenance or hypervisor reboots in the Compute service
- Support for creating and managing private endpoints in the Data Flow service
- Fix upload manager upload extreme large file out of memory issue
- Temporarily remove go mod feature

## 24.0.0 - 2020-08-11
### Added
- Support for autonomous json databases in the Database service
- Support for cleaning up uncommitted multipart uploads in the Object Storage service
- Support for additional list API filters in the Data Catalog service
- Support for Go SDK logging to file 
- Support for Go Modules

### Breaking changes
- Some unusable region enums were removed from the Support Management service
- `CreateIncidentRequest` parameter `OpcRetryToken` was removed from the Support Management service

## 23.0.0 - 2020-08-04
### Added
- Support for calling Oracle Cloud Infrastructure services in the uk-gov-cardiff-1 region
- Support for creating and managing private endpoints in the Data Flow service
- Support for changing instance shapes and restarting nodes in the Big Data service
- Support for additional versions (for example CSQL) in the Big Data service
- Support for creating stacks from compartments in the Resource Manager service

### Breaking changes
- Updated the property of `LifeCycleDetails` to `LifecycleDetails` from the model of `BlockchainPlatformSummary` and `BlockchainPlatformByHostname` in the blockchain service

## 22.0.0 - 2020-07-28
### Added
- Support for calling Oracle Cloud Infrastructure services in the us-sanjose-1 region
- Support for PKCS#8 format API Keys
- Support for updating the fault domain and launch options of VM instances in the Compute service
- Support for image capability schemas and schema versions in the Compute service
- Support for 'Patch Now' maintenance runs for autonomous Exadata infrastructure and autonomous container database resources in the Database service
- Support for automatic performance and cost tuning on volumes in the Block Storage service

### Breaking changes
- Removed the accessToken field from the GitlabAccessTokenConfigurationSourceProvider model in the Resource Manager service

## 21.4.0 - 2020-07-21
### Added
- Support for license types on instances in the Content and Experience service

## 21.3.0 - 2020-07-14
### Added
- Support for the Blockchain service
- Support for failing over an autonomous database that has Data Guard enabled in the Database service
- Support for switching over an autonomous database that has Data Guard enabled in the Database service
- Support for git configuration sources in the Resource Manager service
- Support for optionally specifying a VCN id on list operations of DHCP options, subnets, security lists, route tables, internet gateways, and local peering gateways in the Networking service

## 21.2.0 - 2020-07-07
### Added
- Support for registering and deregistering autonomous dedicated databases with Data Safe in the Database service
- Support for switching between non-private-endpoints and private endpoints on autonomous databases in the Database service
- Support for returning group names when listing identity provider groups in the Identity service
- Support for server-side object re-encryption in the Object Storage service
- Support for private endpoint (ingress) and public endpoint whitelisting in the Analytics Cloud service

### Fixed
- Fixed a bug where setting a Content-Type header twice in the same request

## 21.1.0 - 2020-06-30
### Added
- Support for the Usage service
- Support for the VMware Provisioning service
- Support for applying one-off patches to databases in the Database service
- Support for layer-2 virtualization features on vlans in the Networking service
- Support for all AttachVolumeDetails and ParavirtualizedAttachVolumeDetails properties on instance configurations in the Compute Management service
- Support for setting HTTP header size and allowing invalid characters in HTTP request headers in the Load Balancing service

## 21.0.0 - 2020-06-23
### Added
- Support for the Data Integration service
- Support for updating database home IDs on databases in the Database service
- Support for backing up autonomous databases on Cloud at Customer in the Database service
- Support for managing autonomous VM clusters on Cloud at Customer in the Database service
- Support for accessing data assets via private endpoints in the Data Catalog service
- Support for dependency archive zip files to be specified for use by applications in the Data Flow service

### Breaking changes
- Property `LifecycleState` type changed from `JobLifecycleStateEnum` to `ListJobsLifecycleStateEnum` in the Data Catalog service
- Property `JobType` type changed from `JobTypeEnum` to `ListJobsJobTypeEnum` in the Data Catalog service

## 20.1.0 - 2020-06-16
### Added
- Support for creating a new database from an existing database based on a given timestamp in the Database service
- Support for enabling archive log backups of databases in the Database service
- Support for returning the database version on autonomous container databases in the Database service
- Support for the new DNS format of the Data Transfer service
- Support for scheduled autoscaling, which allows for scaling actions triggered at particular times based on CRON expressions, in the Compute Autoscaling service
- Support for filtering of list APIs for groups, identity providers, identity provider groups, compartments, dynamic groups, network sources, policies, and users by name or lifecycle state in the Identity Service

## 20.0.0 - 2020-06-09
### Added
- Support for returning the database version of backups in the Database service
- Support for patching on Exadata Cloud at Customer resources in the Database service
- Support for new lifecycle substates on instances in the Digital Assistant service
- Support for file servers in the Integration service
- Support for deleting non-empty tag namespaces and bulk deleting tags in the Identity service
- Support for bulk move and bulk delete of resources by compartment in the Identity service
- Support for allowing config file location to be set via env var

### Breaking changes
- Updated property DataStorageSizeInTBs type from *int to *float64 in the database service
- Removed state 'OFFLINE' and added 'DISCONNECTED' for property ExadataInfrastructureLifecycleStateEnum in database service


## 19.4.0 - 2020-06-02
### Added
- Support for optionally supplying a signature when deleting an agreement in the Marketplace service
- Support for launching paid listings in non-US regions in the Marketplace service
- Support for returning the image id of packages in the Marketplace service
- Support for calling Oracle Cloud Infrastructure services in the ap-chuncheon-1 region
- Support security-token based authentication for all services


## 19.3.0 - 2020-05-18
### Added
- Support for returning the private IP of a private endpoint database in the Database service
- Support for native JWT validation in the API Gateway service


## 19.2.0 - 2020-05-12
### Added
- Support for drift detection in the Resource Manager service


## 19.1.0 - 2020-05-05
### Added
- Support for updating the license type of database systems in the Database service
- Support for updating the version of 19c autonomous databases in the Database service
- Support for backup and restore functionality in the Key Management service
- Support for reports in the Marketplace service
- Support for calling Oracle Cloud Infrastructure services in the ap-hyderabad-1 region

## 19.0.0 - 2020-04-28
### Added
- Support for the MySQL Database service
- Support for updating the database home of a database in the Database service
- Support for government regions in the Marketplace service
- Support for starting and stopping instances in the Integration service
- Support for installing Windows updates in the OS Management service

### Breaking changes
- Removed the models of `UpdatablePackageSummary`, `ManagedInstanceUpdateDetails` and the header parameter `etag` from `WorkRequestErrorsResponse` and `WorkRequestLogsResponse` in the osmanagement service

## 18.0.0 - 2020-04-21
### Added
- Support for the Data Safe service
- Support for the Incident Management service
- Support for showing which database versions support always-free in the Database service
- Support in instance configurations for flex shapes, dedicated VM hosts, encryption in transit, and KMS keys in the Compute Autoscaling service
- Support for server-side object encryption using a customer-provided encryption key in the Object Storage service
- Support for specifying maintenance preferences while launching and updating Exadata Database systems in the Database service
- Support for flexible-shaped VM instances in the Compute service
- Support for scheduled cross-region backups in the Block Volume service
- Support for object versioning in the Object Storage service

### Breaking changes
- Removed the models of `Archiver`, `CreateArchiverDetails` and `UpdateArchiverDetails`, operations of `CreateArchiver`, `GetArchiver`, `StartArchiver`, `StopArchiver` and `UpdateArchiver` in the streaming service

## 17.4.0 - 2020-04-14
### Added
- Support for access types on instances in the Content and Experience service
- Support for identity contexts in the Search service

## 17.3.0 - 2020-04-07
### Added
- Support for changing compartments of runs and applications in the Data Flow service
- Support for getting usage information in the Key Management Vault service
- Support for custom Key Management service endpoints and private endpoints on stream pools in the Streaming service

## 17.2.0 - 2020-03-31
### Added
- Support for the Secrets Management service 
- Support for the Big Data service
- Support for updating class name, file URI, language, and spark version of applications in the Data Flow service
- Support for cross-region replication in the Object Storage service
- Support for retention rules in the Object Storage service
- Support for enabling and disabling pod security policy admission controllers in the Container Engine for Kubernetes service

## 17.1.0 - 2020-03-24
### Added
- Support for Web Application Acceleration and Security configurations on instances in the Content and Experience service
- Support for shared database homes on Exadata Cloud at Customer resources in the Database service
- Support for Exadata database creation from backup in the Database service
- Support for conditions on JavaScript challenges, new action types on access rules, new policy configuration settings, exclusions on custom protection rules, and IP address lists on IP whitelists in the Web Application Acceleration and Security service

## 17.0.0 - 2020-03-17
### Added
- Support for serial console connections in the Database service
- Support for preview database versions in the Database service
- Support for node reboot migration maintenance status and maintenance windows in the Database service
- Support for using instance metadata API v2 for instance principals authentication


### Breaking changes
- Removed the model of `AutonomousExadataInfrastructureMaintenanceWindow` from Database service

## 16.0.0 - 2020-03-10
### Added
- Support for Events service integration with alerts in the Budgets service
- Support delegation-token auth for all services

### Breaking changes
- The parameters sort_by and lifecycle_state type from Budget service are changed from str to enum

## 15.8.0 - 2020-03-03
### Added
- Support for updating the shape of a Database System in the Database service
- Support for generating CPE configurations for download in the Networking service
- Support for private IPs and fault domains of cluster nodes in the Container Engine for Kubernetes service
- Support for calling Oracle Cloud Infrastructure services in the ca-montreal-1 region
- Support for exposing error after retrying failed in all services.

## 15.7.0 - 2020-02-25
### Added
- Support for restarting autonomous databases in the Database service
- Support for private endpoints on autonomous databases in the Database service
- Support for IP-based policies in the Identity service
- Support for management of OAuth 2.0 client credentials in the Identity service
- Support for OCI Functions as a subscription protocol in the Notifications service

## 15.6.0 - 2020-02-18
### Added
- Support for the NoSQL Database service
- Support for filtering database versions by storage management type in the Database service
- Support for specifying paid listing types within pricing models in the Marketplace service
- Support for primary and non-primary instance types in the Content and Experience service

## 15.5.0 - 2020-02-11
### Added
- Support for listing supported database versions for Autonomous Database Serverless, and selecting a version at provisioning time in the Database service
- Support for TCP proxy protocol versions on listener connection configurations in the Load Balancer service
- Support for calling the Notifications service in alternate realms
- Support for calling Oracle Cloud Infrastructure services in the eu-amsterdam-1 and me-jeddah-1 regions
- Support for non-default profiles for credentials

## 15.4.0 - 2020-02-04
## Added
- Support for the Data Science service
- Support for calling Oracle Cloud Infrastructure services in the ap-osaka-1 and ap-melbourne-1 regions

## 15.3.0 - 2020-01-28
## Added
- Support for the Application Migration service
- Support for the Data Flow service
- Support for the Data Catalog service
- Support for cross-shape Data Guard in the Database service
- Support for offline data export in the Data Transfer service

## 15.2.0 - 2020-01-21
## Added
- Support for getting DRG redundancy status in the Networking service
- Support for cloning autonomous databases from backups in the Database service

## 15.1.0 - 2020-01-14
### Added
- Support for a description field on route rules and security rules in the Networking service
- Support for starting and stopping Digital Assistant instances in the Digital Assistant service
- Support for shared database homes on Exadata, bare metal, and virtual machine instances in the Database service
- Support for tracking a number of Database service operations through the Work Requests service

## 15.0.0 - 2020-01-07
### Added
- Support for optionally specifying the corporate proxy field when creating Exadata infrastructure in the Database service
- Support for maintenance windows, and rescheduling maintenance runs, on autonomous container databases in the Database service

### Breaking changes
- Field `hostname` in `NodeDetails` from Database service is changed to mandatory

## 14.0.0 - 2019-12-17
### Added
- Support for the API Gateway service
- Support for the OS Management service
- Support for the Marketplace service
- Support for "default"-type vaults in the Key Management service
- Support for bringing your own keys in the Key Management service 
- Support for cross-region backups of boot volumes in the Block Storage service
- Support for top-level TSIG keys in the DNS service
- Support for resizing virtual machine instances to different shapes in the Compute service
- Support for management configuration of cloud agents in the Compute service
- Support for launching node pools using image IDs in the Container Engine for Kubernetes service

### Breaking changes
- Removed support for v1 auth tokens in kubeconfig files in the `CreateClusterKubeconfigContentDetails` class of the Container Engine for Kubernetes service
- Removed the IDCS access token requirement on the delete deleteOceInstance operation in the Content and Experience service, which is why the `DeleteOceInstanceDetails` class was removed
- Parameter `compartment_id` in `list_stream_pools` API from Streaming service is changed to required parameter

## 13.1.0 - 2019-12-10
### Added
- Support for etags on results of the List Objects API in the Object Storage service
- Support for OCIDs on buckets in the Object Storage service
- Support for content-disposition and cache-control headers on objects in the Object Storage service
- Support for recovering deleted compartments in the Identity service
- Support for sharing volumes across multiple instances in the Block Storage service
- Support for connect harnesses and stream pools in the Streaming service
- Support for associating file storage mount targets with network security groups in the File Storage service 
- Support for calling Oracle Cloud Infrastructure services in the uk-gov-london-1 region

## 13.0.0 - 2019-11-26
### Added
- Support for maintenance windows on autonomous databases in the Database service
- Support for getting the compute units (OCPUs) of an Exadata autonomous transaction processing - dedicated resource in the Database service

### Breaking changes
- Create database home from VM_CLUSTER_BACKUP is removed from Database Service
- Response type is changed for following two APIs in Virtual Network Service 
    - Before

    ```golang
    BulkAddVirtualCircuitPublicPrefixes (err error)

    BulkDeleteVirtualCircuitPublicPrefixes (err error)
    ```

    - After

    ```golang
    BulkAddVirtualCircuitPublicPrefixes (response BulkAddVirtualCircuitPublicPrefixesResponse, err error)

    BulkDeleteVirtualCircuitPublicPrefixes (response BulkDeleteVirtualCircuitPublicPrefixesResponse, err error)
    ```

## 12.5.0 - 2019-11-19
### Added
- Support for four-byte autonomous system numbers (ASNs) on FastConnect resources in the Networking service
- Support for choosing fault domains when creating instance pools in the Compute service
- Support for allowing connections from only specific VCNs to autonomous data warehouse and autonomous transaction processing instances in the Database service
- Support for Streaming Client Non-Regional

## 12.4.0 - 2019-11-12
### Added
- Support for access to APEX and SQL Dev features on autonomous transaction processing and autonomous data warehouse resources in the Database service
- Support for registering / deregistering autonomous transaction processing and autonomous data warehouse resources with Data Safe in the Database service
- Support for redirecting HTTP / HTTPS request URIs to different URIs in the Load Balancing service
- Support for specifying compartments on options APIs in the Container Engine for Kubernetes service
- Support for volume performance units on block volumes in the Block Storage service
- Support for opc-multipart-md5 verification for UploadManager. Example can be found on [Github](https://github.com/oracle/oci-go-sdk/blob/v8.0.0/example/example_objectstorage_test.go#L57)

## 12.3.0 - 2019-11-05
### Added
- Support for the Analytics Cloud service
- Support for the Integration Cloud service
- Support for IKE versions in IPSec connections in the Virtual Networking service
- Support for getting a stack's Terraform state in the Resource Manager service

## 12.2.0 - 2019-10-29
### Added
- Support for wallet rotation operations on Autonomous Databases in the Database service
- Support for adding and removing image shape compatibility entries in the Compute service
- Support for managing redirects in the Web Application Acceleration and Security service
- Support for migrating zones from the Dyn HTTP Redirect Service to Oracle Cloud Infrastructure in the DNS service

## 12.1.0 - 2019-10-15
### Added
- Support for the Digital Assistant service
- Support for work requests on Instance Pool operations in the Compute service

## 12.0.0 - 2019-10-08
### Added
- Support for the new schema for events in the Audit service
- Support for entitlements in the Data Transfer service
- Support for custom scheduled backup policies on volumes in the Block Storage service
- Support for specifying the network type when launching virtual machine instances in the Compute service
- Support for Monitoring service integration in the Health Checks service

### Fixed
- OCI Golang SDK hook/callback to display progress bar for uploads [Github issue 187](https://github.com/oracle/oci-go-sdk/issues/187)

### Breaking changes
* The TenantId parameter is now Id (Id of the Transfer Application Entitlement) for GetTransferApplianceEntitlementRequest in TransferApplianceEntitlementClient
* The Audit service version was bumped to 20190901, use older version of Go SDK for Audit service version 20160918 

## 11.0.0 - 2019-10-01
### Added
- Support for required tags in the Identity service
- Support for work requests on tagging operations in the Identity service
- Support for enumerated tag values in the Identity service
- Support for moving dynamic routing gateway resources across compartments in the Networking service
- Support for migrating zones from Dyn managed DNS to OCI in the DNS service
- Support for fast provisioning for virtual machine databases in the Database service

### Breaking changes
- The field``CreateZoneDetails`` is no longer an anonymous field and the type changed from struct to interface in struct ``CreateZoneRequest``. Here is sample code that shows how to update your code to incorporate this change. 


    - Before

    ```golang
    // There were two ways to initialize the CreateZoneRequest struct.
    // This breaking change only impact option #2
    request := dns.CreateZoneRequest{}

    // #1. Instantiate CreateZoneDetails directly: no impact
    details := dns.CreateZoneDetails{}
    details.Name = common.String('some name')
    // ... other properties
    // Set it to the request class
    request.CreateZoneDetails = details

    // #2. Instantiate CreateZoneDetails through anonymous fields: will break
    request.Name = common.String('some name')
    // ... other properties
    ```

    - After

    ```golang
    // #2 no longer supported. Create CreateZoneDetails directly
    details := dns.CreateZoneDetails{}
    details.Name = common.String('some name')
    // ... other properties

    request := dns.CreateZoneRequest{
        CreateZoneDetails: details
    }
    // ...
    ```

## 10.1.0 - 2019-09-24
### Added
- Support for selecting the Terraform version to use in the Resource Manager service
- Support for bucket re-encryption in the Object Storage service
- Support for enabling / disabling bucket-level events in the Object Storage service

## 10.0.0 - 2019-09-17
### Added
- Support for importing state files in the Resource Manager service
- Support for Exadata Cloud at Customer in the Database service
- Support for free tier resources and system tags in the Load Balancing service
- Support for free tier resources and system tags in the Compute service
- Support for free tier resources and system tags in the Block Storage service
- Support for free tier and system tags on autonomous databases in the Database service

### Breaking
- Interface CreateDbHomeWithDbSystemIdBase is renamed to CreateDbHomeBase and dbSystemId property is removed from it
- CreateDbHomeWithDbSystemIdBase in CreateDbHomeRequest is replaced with CreateDbHomeWithDbSystemIdDetails

## 9.0.0 - 2019-09-10
### Added
- Support for specifying the autoBackupWindow field for scheduling backups in the Database service
- Support for network security groups on autonomous Exadata infrastructure in the Database service
- Support for Kubernetes secrets encryption in customer clusters, regional subnets, and cluster authentication for instance principals in the Container Engine for Kubernetes service
- Support for the Oracle Content and Experience service

### Breaking
- The etag field has been removed from the ChangeSubscriptionCompartmentResponse and ChangeTopicCompartmentResponse structs of the Notifications service

## 8.1.0 - 2019-09-03
### Added
- Support for the Sydney (SYD) region
- Support for managing cluster networks in the Compute Autoscaling service
- Support for tracking asynchronous operations via work requests in the Database service

## 8.0.0 - 2019-08-27
### Added
- Support for the Sao Paulo (GRU) region
- Support for dedicated virtual machine hosts in the Compute service
- Support for resource groups in metrics and alarms in the Monitoring service
- Support for resource principle auth. Example can be found on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_resource_principal_function/README.md)

### Breaking changes
- Breaking changes were made for following enum values
    - Before
    ```golang
    autoscaling.ActionTypeEnum.ActionTypeBy
    keymanagement.CreateVaultDetailsVaultTypeEnum.CreateVaultDetailsVaultTypePrivate
    keymanagement.VaultSummaryVaultTypeEnum.VaultSummaryVaultTypePrivate
    keymanagement.VaultVaultTypeEnum.VaultVaultTypePrivate
    objectstorage.WorkRequestSummaryOperationTypeEnum.WorkRequestSummaryOperationTypeObject
    objectstorage.WorkRequestOperationTypeEnum.WorkRequestOperationTypeObject
    resourcemanager.LogEntryTypeEnum.LogEntryTypeConsole
    resourcemanager.WorkRequestOperationTypeEnum.WorkRequestOperationTypeCompartment
    ```

    - After
    ```golang
    autoscaling.ActionTypeEnum.ActionTypeChangeCountBy
    keymanagement.CreateVaultDetailsVaultTypeEnum.CreateVaultDetailsVaultTypeVirtualPrivate
    keymanagement.VaultSummaryVaultTypeEnum.VaultSummaryVaultTypeVirtualPrivate
    keymanagement.VaultVaultTypeEnum.VaultVaultTypeVirtualPrivate
    objectstorage.WorkRequestSummaryOperationTypeEnum.WorkRequestSummaryOperationTypeCopyObject
    objectstorage.WorkRequestOperationTypeEnum.WorkRequestOperationTypeCopyObject
    resourcemanager.LogEntryTypeEnum.LogEntryTypeTerraformConsole
    resourcemanager.WorkRequestOperationTypeEnum.WorkRequestOperationTypeChangeStackCompartment
    ```

## 7.1.0 - 2019-08-20
### Added
- Support for the Limits service
- Support for archiving to Object Storage in the Streaming service
- Support for etags on resources in the Streaming service
- Support for Key Management service (KMS) encryption of file systems in the File Storage service
- Support for moving public IP, DHCP, local peering gateway, internet gateway, network security group, and DRG attachment resources across compartments in the Networking service
- Support for multi-origin, basic cache, certificate mapping, and OCI Monitoring service integration in the Web Application Acceleration and Security service

## 7.0.0 - 2019-08-13
### Added
- Support for the Data Transfer service
- Support for the Zurich (ZRH) region

### Breaking changes
- Breaking changes were made in the Web Application Acceleration and Security (WAAS) service
  - `WorkRequestSummaryOperationTypePurgeWaasPolicy` const removed from `waas/work_request_summary.go`
  - `WorkRequestOperationTypesPurgeWaasPolicy` const removed from `waas/work_request_operation_types.go`
  - `WorkRequestOperationTypesPurgeWaasPolicy` const removed from `waas/work_request.go`
  - `IssuerName` in `Certificate` struct changed type from `*CertificateSubjectName` to `*CertificateIssuerName`
  - `LifecycleState` changed from array of string to array of `ListCertificateLifeCycleStateEnum` in `waas/list_certificates_request_response.go` and `waas/list_waas_policies_request_response.go`
  - `Etag` was removed from the following structs:
     - `AcceptRecommendationsResponse`
     - `DeleteWaasPolicyResponse`
     - `UpdateAccessRulesResponse`
     - `UpdateCaptchasResponse`
     - `UpdateDeviceFingerprintChallengeResponse`
     - `UpdateGoodBotsResponse`
     - `UpdateHumanInteractionChallengeResponse`
     - `UpdateJsChallengeResponse`
     - `UpdatePolicyConfigResponse`
     - `UpdateProtectionRulesResponse`
     - `UpdateProtectionSettingsResponse`
     - `UpdateThreatFeedsResponse`
     - `UpdateWaasPolicyResponse`
     - `UpdateWafAddressRateLimitingResponse`
     - `UpdateWafConfigResponse`
     - `UpdateWhitelistsResponse`

## 6.2.0 - 2019-08-06
### Added
- Support for IPv6 load balancers in the Load Balancing service
- Support for IPv6 on VCN and FastConnect resources in the Networking service

## 6.1.0 - 2019-07-30
### Added
- Support for the Mumbai (BOM) region
- Support for the Events service
- Support for moving streams across compartments in the Streaming service
- Support for moving FastConnect resources across compartments in the Networking service
- Support for moving policies across compartments in the Web Application Acceleration and Security service
- Support for tagging FastConnect resources in the Networking service

## 6.0.0 - 2019-07-23
### Added
- Support for moving resources across compartments in the Database service
- Support for moving resources across compartments in the Health Checks service
- Support for moving alarms across compartments in the Monitoring service
- Support for creating instance configurations from running instances in the Compute service
- Support for setting up budget alerts for cost tracking tags in the Budgets service

## 5.15.0 - 2019-07-16
### Added
- Support for the Functions service
- Support for the Quotas service
- Support for moving resources across compartments in the DNS service
- Support for moving instances across compartments in the Compute service
- Support for moving keys and vaults across compartments in the Key Management service
- Support for moving topics and subscriptions across compartments in the Notifications service
- Support for moving load balancers across compartments in the Load Balancing service
- Support for specifying permitted REST methods in load balancer rule sets in the Load Balancing service
- Support for configuring cookie session persistence in backend sets in the Load Balancing service
- Support for ACL rules in rule sets in the Load Balancing service
- Support for move compartment tree in the Identity service
- Support for specifying and returning a KMS key in backup operations in the Block Storage service
- Support for transit routing in the Networking service

## 5.14.0 - 2019-07-09
### Added
- Support for network security groups in the Load Balancing service
- Support for network security groups in Core Services
- Support for network security groups on database systems in the Database service
- Support for creating autonomous transaction processing and autonomous data warehouse previews in the Database service
- Support for getting the load balancer attachments of instance pools in the Compute service
- Support for moving resources across compartments in the Resource Manager service
- Support for moving VCN resources across compartments in the Networking service

## 5.13.0 - 2019-07-02
### Added
- Support for moving images, instance configurations, and instance pools across compartments in Core Services
- Support for moving autoscaling configurations across compartments in the Compute Autoscaling service

### Fixed
- Fixed a bug where the Streaming service's endpoints in Tokyo, Seoul, and future regions were not reachable from the SDK

## 5.12.0 - 2019-06-25
### Added
- Support for moving senders across compartments in the Email service
- Support for moving NAT gateway resources across compartments in Core Services

## 5.11.0 - 2019-06-18
### Added
- Support for moving service gateway resources across compartments in Core Services
- Support for moving block storage resources across compartments in Core Services
- Support for key deletion in the Key Management service

## 5.10.0 - 2019-06-11
### Added
- Support for specifying custom boot volume sizes on instance configurations in the Compute Autoscaling service
- Support for 'Autonomous Transaction Processing - Dedicated' features, as well as maintenance run and backup operations on autonomous databases, autonomous container databases, and autonomous Exadata infrastructure in the Database service

## 5.9.0 - 2019-06-04
### Added
- Support for autoscaling autonomous databases and autonomous data warehouses in the Database service
- Support for specifying fault domains as part of instance configurations in the Compute Autoscaling service
- Support for deleting tag definitions and tag namespaces in the Identity service

### Fixed
- Support for regions in realms other than oraclecloud.com in the Load Balancing service

## 5.8.0 - 2019-05-28
### Added
- Support for the Work Requests service, and tracking of a number of Core Services operations through work requests
- Support for emulated volume attachments in Core Services
- Support for changing the compartment of resources in the File Storage service
- Support for tags in list operations in the File Storage service
- Support for returning UI password creation dates in the Identity service

## 5.7.0 - 2019-05-21
### Added
- Support for returning tags when listing instance configurations, instance pools, or autoscaling configurations in the Compute Autoscaling service
- Support for getting the namespace of another tenancy than the caller's tenancy in the Object Storage service
- Support for BGP dynamic routing and providing pre-shared secrets (PSKs) when establishing tunnels in the Networking service

## 5.6.0 - 2019-05-14
### Added
- Support for the Seoul (ICN) region
- Support for logging context fields on data-plane APIs of the Key Management Service
- Support for reverse pagination on list operations of the Email service
- Support for configuring backup retention windows on database backups in the Database service

## 5.5.0 - 2019-05-07
### Added
- Support for the Tokyo (NRT) region

- Support UploadManager for uploading large objects. Sample is available on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_objectstorage_test.go)

## 5.4.0 - 2019-04-16
### Added
- Support for tagging dynamic groups in the Identity service
- Support for updating network ACLs and license types for autonomous databases and autonomous data warehouses in the Database service
- Support for editing static routes and IPSec remote IDs in the Virtual Networking service

## 5.3.0 - 2019-04-09
### Added
- Support for etag and if-match headers (for optimistic concurrency control) in the Email service

## 5.2.0 - 2019-04-02
### Added
- Support for provider service key names on virtual circuits in the FastConnect service
- Support for customer reference names on cross connects and cross connect groups in the FastConnect service

## 5.1.0 - 2019-03-26
### Added
- Support for glob patterns and exclusions for object lifecycle management in the Object Storage service
- Documentation enhancements and corrections for traffic management in the DNS service

### Fixed
- The 'tag' info is always ignored in the returned string of Version() function [Github issue 157](https://github.com/oracle/oci-go-sdk/issues/157)

## 5.0.0 - 2019-03-19
### Added

- Support for specifying metadata on node pools in the Container Engine for Kubernetes service
- Support for provisioning a new autonomous database or autonomous data warehouse as a clone of another in the Database service
### Breaking changes
- The field``CreateAutonomousDatabaseDetails`` is no longer an anonymous field and the type changed from struct to interface in struct ``CreateAutonomousDatabaseRequest``. Here is sample code that shows how to update your code to incorporate this change. 

    - Before

    ```golang
    // create a CreateAutonomousDatabaseRequest
    // There were two ways to initialize the CreateAutonomousDatabaseRequest struct.
    // This breaking change only impact option #2
    request := database.CreateAutonomousDatabaseRequest{}

    // #1. Instantiate CreateAutonomousDatabaseDetails directly: no impact
    details := database.CreateAutonomousDatabaseDetails{}
    details.CompartmentId = common.String(getCompartmentID())
    // ... other properties

    // Set it to the request class
    request.CreateAutonomousDatabaseDetails = details

    // #2. Instantiate CreateAutnomousDatabaseDetails through  anonymous fields: will break
    request.CompartmentId = common.String(getCompartmentID())
    // ... other properties
    ```

    - After

    ```golang
    // #2 no longer supported. Create CreateAutonomousDatabaseDetails directly
    details := database.CreateAutonomousDatabaseDetails{}
    details.CompartmentId = common.String(getCompartmentID())
    // ... other properties

    // and set the details to CreateAutonomousDatabaseBase
    request := database.CreateAutonomousDatabaseRequest{}
    request.CreateAutonomousDatabaseDetails = details
    // ...
    ```


## 4.2.0 - 2019-03-12
### Added
- Support for the Budgets service
- Support for managing multifactor authentication in the Identity service
- Support for managing default tags in the Identity service
- Support for account recovery in the Identity service
- Support for authentication policies in the Identity service
- Support for specifying the workload type when creating autonomous databases in the Database service
- Support for I/O resource management for Exadata database systems in the Database service
- Support for customer-specified timezones on database systems in the Database service

## 4.1.0 - 2019-02-28
### Added
- Support for the Monitoring service
- Support for the Notification service
- Support for the Resource Manager service
- Support for the Compute Autoscaling service
- Support for changing the compartment of a tag namespace in the Identity service
- Support for specifying fault domains in the Database service
- Support for managing instance monitoring in the Compute service
- Support for attaching/detaching load balancers to instance pools in the Compute service

## 4.0.0 - 2019-02-21
### Added
- Support for government-realm regions
- Support for the Streaming service
- Support for tags in the Key Management service
- Support for regional subnets in the Virtual Networking service

### Fixed
- Removed unused Announcements service 'NotificationFollowupDetails' struct and 'GetFollowups' operation
- InstancePrincipals now invalidates a token shortly before its expiration time to avoid making  a service call with an expired token
- Requests with binary bodies that require its body to be included in the signature are now being signed correctly

## 3.7.0 - 2019-02-07
### Added
- Support for the Web Application Acceleration and Security (WAAS) service
- Support for the Health Checks service
- Support for connection strings on Database resources in the Database service
- Support for traffic management in the DNS service
- Support for tagging in the Email service
### Fixed
- Retry context in now cancelable during wait for new retry

## 3.6.0 - 2019-01-31
### Added
- Support for the Announcements service

## 3.5.0 - 2019-01-24
### Added

- Support for renaming databases during restore-from-backup operations in the Database service
- Built-in logging now supports log levels. More information about the changes can be found in the [go-docs page](https://godoc.org/github.com/oracle/oci-go-sdk#hdr-Logging_and_Debugging)
- Support for calling Oracle Cloud Infrastructure services in the ca-toronto-1 region

## 3.4.0 - 2019-01-10
### Added 
- Support for device attributes on volume attachments in the Compute service
- Support for custom header rulesets in the Load Balancing service


## 3.3.0 - 2018-12-13
### Added 
- Support for Data Guard for VM shapes in the Database service
- Support for sparse disk groups for Exadata shapes in the Database service
- Support for a new field, isLatestForMajorVersion, when listing DB versions in the Database service
- Support for in-transit encryption for paravirtualized boot volume and data volume attachments in the Block Storage service
- Support for tagging DNS Zones in the DNS service
- Support for resetting credentials for SCIM clients associated with an Identity provider and updating user capabilities in the Identity service

## 3.2.0 - 2018-11-29
### Added 
- Support for getting bucket statistics in the Object Storage service

### Fixed
- Block Storage service for copying volume backups across regions is now enabled 
- Object storage `PutObject` and `UploadPart` operations now do not override the client's signer

## 3.1.0 - 2018-11-15
### Added
- Support for VCN transit routing in the Networking service 

## 3.0.0 - 2018-11-01
### Added
- Support for modifying the route table, DHCP options and security lists associated with a subnet in the Networking service.
- Support for tagging of File Systems, Mount Targets and Snapshots in the File Storage service.
- Support for nested compartments in the Identity service

### Notes
- The version is bumped due to breaking changes in previous release.

## 2.7.0 - 2018-10-18
### Added
- Support for cost tracking tags in the Identity service
- Support for generating and downloading wallets in the Database service
- Support for creating a standalone backup from an on-premises database in the Database service
- Support for db version and additional connection strings in the Autonomous Transaction Processing and Autonomous Data Warehouse resources of the Database service
- Support for copying volume backups across regions in the Block Storage service
- Support for deleting compartments in the Identity service
- Support for reboot migration for virtual machines in the Compute service
- Support for Instance Pools and Instance Configurations in the Compute service

### Fixed
- The signing algorithm does not lower case the header fields [Github issue 132](https://github.com/oracle/oci-go-sdk/issues/132)
- Raw configuration provider does not check for empty strings [Github issue 134](https://github.com/oracle/oci-go-sdk/issues/134)

### Breaking change
- DbDataSizeInMBs field in Backup and BackupSummary struct was renamed to DatabaseSizeInGBs and type changed from *int to *float64 
    - Before
    ```golang
    // Size of the database in megabytes (MB) at the time the backup was taken.
    DbDataSizeInMBs *int `mandatory:"false" json:"dbDataSizeInMBs"`
    ```

    - After

    ```golang
    // The size of the database in gigabytes at the time the backup was taken.
    DatabaseSizeInGBs *float64 `mandatory:"false" json:"databaseSizeInGBs"`
    ```
- Data type for DatabaseEdition in Backup and BackupSummary struct was changed from *string to BackupDatabaseEditionEnum
    - Before

    ```golang
    // The Oracle Database edition of the DB system from which the database backup was taken.
    DatabaseEdition *string `mandatory:"false" json:"databaseEdition"`
    ```

    - After

    ```golang
     // The Oracle Database edition of the DB system from which the database backup was taken.
     DatabaseEdition BackupDatabaseEditionEnum `mandatory:"false" json:"databaseEdition,omitempty"`
    ```

## 2.6.0 - 2018-10-04
### Added
- Support for trusted partner images through application listings and subscriptions in the Compute service
- Support for object lifecycle policies in the Object Storage service
- Support for copying objects across regions in the Object Storage service
- Support for network address translation (NAT) gateways in the Networking service

## 2.5.0 - 2018-09-27
### Added
- Support for paravirtualized launch mode when importing images in the Compute service
- Support for Key Management service
- Support for encrypting the contents of an Object Storage bucket using a Key Management service key
- Support for specifying a Key Management service key when launching a compute instance in the Compute service
- Support for specifying a Key Management service key when backing up or restoring a block storage volume in the Block Volume service

## 2.4.0 - 2018-09-06
### Added
- Added support for updating metadata fields on an instance in the Compute service

## 2.3.0 - 2018-08-23
### Added
- Support for fault domain in the Identity Service
- Support for Autonomous Data Warehouse and Autonomous Transaction Processing in the Database service
- Support for resizing an offline volume in the Block Storage service
- Nil interface when polymorphic json response object is null

## 2.2.0 - 2018-08-09
### Added
- Support for fault domains in the Compute service
- A sample showing how to use Search service from the SDK is available on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_resourcesearch_test.go)

## 2.1.0 - 2018-07-26
### Added
- Support for the Search service
- Support for specifying a backup policy when creating a boot volume in the Block Storage service

### Fixed
- OCI error is missing opc-request-id value [Github Issue 120](https://github.com/oracle/oci-go-sdk/issues/120)
- Include raw http response when service error occurred

## 2.0.0 - 2018-07-12
### Added
- Support for tagging Load Balancers in the Load Balancing service
- Support for export options in the File Storage service
- Support for retrieving compartment name and user name as part of events in the Audit service

### Fixed
- CreateKubeconfig function should not close http reponse body [Github Issue 116](https://github.com/oracle/oci-go-sdk/issues/116)

### Breaking changes
- Datatype changed from *int to *int64 for several request/response structs. Here is sample code that shows how to update your code to incorporate this change. 

    - Before

    ```golang
    // Update the impacted properties from common.Int to common.Int64.
    // Here is the updates for CreateBootVolumeDetails
    details := core.CreateBootVolumeDetails{
        SizeInGBs: common.Int(10),
    }
    ```

    - After

    ```golang
    details := core.CreateBootVolumeDetails{
        SizeInGBs: common.Int64(10),
    }
    ```

- Impacted packages and structs
    - core
        - BootVolume.(SizeInGBs, SizeInMBs)
        - BootVolumeBackup.(SizeInGBs, UniqueSizeInGBs)
        - CreateBootVolumeDetails.SizeInGBs
        - CreateVolumeDetails.(SizeInGBs, SizeInMBs)
        - Image.SizeInMBs
        - InstanceSourceViaImageDetails.BootVolumeSizeInGBs
        - Volume.(SizeInGBs, SizeInMBs)
        - VolumeBackup.(SizeInGBs, SizeInMBs, UniqueSizeInGBs, UniqueSizeInMbs)
        - VolumeGroup.(SizeInMBs, SizeInGBs)
        - VolumeGroupBackup.(SizeInMBs, SizeInGBs, UniqueSizeInMbs, UniqueSizeInGbs)
    - dns
        - GetDomainRecordsRequest.Limit
        - GetRRSetRequest.Limit
        - GetZoneRecordsRequest.Limit
        - ListZonesRequest.Limit
        - Zone.Serial
        - ZoneSummary.Serial
    - filestorage
        - ExportSet.(MaxFsStatBytes, MaxFsStatFiles)
        - FileSystem.MeteredBytes
        - FileSystemSummary.MeteredBytes
        - UpdateExportSetDetails.(MaxFsStatBytes, MaxFsStatFiles)
    - identity
        - ApiKey.InactiveStatus
        - AuthToken.InactiveStatus
        - Compartment.InactiveStatus
        - CustomerSecretKey.InactiveStatus
        - CustomerSecretKeySummary.InactiveStatus
        - DynamicGroup.InactiveStatus
        - Group.InactiveStatus
        - IdentityProvider.InactiveStatus
        - IdpGroupMapping.InactiveStatus
        - Policy.InactiveStatus
        - Saml2IdentityProvider.InactiveStatus
        - SmtpCredential.InactiveStatus
        - SmtpCredentialSummary.InactiveStatus
        - SwiftPassword.InactiveStatus
        - UiPassword.InactiveStatus
        - User.InactiveStatus
        - UserGroupMembership.InactiveStatus
    - loadbalancer
        - ConnectionConfiguration.IdleTimeout
        - ListLoadBalancerHealthsRequest.Limit
        - ListLoadBalancersRequest.Limit
        - ListPoliciesRequest 
        - ListProtocolsRequest.Limit
        - ListShapesRequest.Limit
        - ListWorkRequestsRequest.Limit
    - objectstorage
        - GetObjectResponse.ContentLength
        - HeadObjectResponse.ContentLength
        - MultipartUploadPartSummary.Size
        - ObjectSummary.Size
        - PutObjectRequest.ContentLength
        - UploadPartRequest.ContentLength

## 1.8.0 - 2018-06-28
### Added
- Support for service gateway management in the Networking service
- Support for backup and clone of boot volumes in the Block Storage service

## 1.7.0 - 2018-06-14
### Added
- Support for the Container Engine service. A sample showing how to use this service from the SDK is available [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_containerengine_test.go)

### Fixed
- Empty string was send to backend service for optional enum if it's not set

## 1.6.0 - 2018-05-31
### Added
- Support for the "soft shutdown" instance action in the Compute service
- Support for Auth Token management in the Identity service
- Support for backup or clone of multiple volumes at once using volume groups in the Block Storage service
- Support for launching a database system from a backup in the Database service

### Breaking changes
- ``LaunchDbSystemDetails`` is renamed to ``LaunchDbSystemBase`` and the type changed from struct to interface in ``LaunchDbSystemRequest``. Here is sample code that shows how to update your code to incorporate this change. 

    - Before

    ```golang
    // create a LaunchDbSystemRequest
    // There were two ways to initialize the LaunchDbSystemRequest struct.
    // This breaking change only impact option #2
    request := database.LaunchDbSystemRequest{}

    // #1. explicity create LaunchDbSystemDetails struct (No impact)
    details := database.LaunchDbSystemDetails{}
    details.AvailabilityDomain = common.String(validAD())
    details.CompartmentId = common.String(getCompartmentID())
    // ... other properties
    request.LaunchDbSystemDetails = details

    // #2. use anonymous fields (Will break)
    request.AvailabilityDomain = common.String(validAD())
    request.CompartmentId = common.String(getCompartmentID())
    // ...
    ```

    - After

    ```golang
    // create a LaunchDbSystemRequest
    request := database.LaunchDbSystemRequest{}
    details := database.LaunchDbSystemDetails{}
    details.AvailabilityDomain = common.String(validAD())
    details.CompartmentId = common.String(getCompartmentID())
    // ... other properties

    // set the details to LaunchDbSystemBase
    request.LaunchDbSystemBase = details
    // ...
    ```

## 1.5.0 - 2018-05-17
### Added
- ~~Support for backup or clone of multiple volumes at once using volume groups in the Block Storage service~~
- Support for the ability to optionally specify a compartment filter when listing exports in the File Storage service
- Support for tagging virtual cloud network resources in the Networking service
- Support for specifying the PARAVIRTUALIZED remote volume type when creating a virtual image or launching a new instance in the Compute service
- Support for tilde in private key path in configuration files

## 1.4.0 - 2018-05-03
### Added
- Support for ``event_name`` in Audit Service
- Support for multiple ``hostnames`` for loadbalancer listener in LoadBalance service
- Support for auto-generating opc-request-id for all operations
- Add opc-request-id property for all requests except for Object Storage which use opc-client-request-id

## 1.3.0 - 2018-04-19
### Added
- Support for retry on Oracle Cloud Infrastructure service APIs. Example can be found on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_retry_test.go)
- Support for tagging DbSystem and Database resources in the Database Service
- Support for filtering by DbSystemId in ListDbVersions operation in Database Service

### Fixed
- Fixed a request signing bug for PatchZoneRecords API
- Fixed a bug in DebugLn

## 1.2.0 - 2018-04-05
### Added
- Support for Email Delivery Service. Example can be found on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_email_test.go)
- Support for paravirtualized volume attachments in Core Services
- Support for remote VCN peering across regions
- Support for variable size boot volumes in Core Services
- Support for SMTP credentials in the Identity Service
- Support for tagging Bucket resources in the Object Storage Service

## 1.1.0 - 2018-03-27
### Added
- Support for DNS service
- Support for File Storage service
- Support for PathRouteSets and Listeners in Load Balancing service
- Support for Public IPs in Core Services
- Support for Dynamic Groups in Identity service
- Support for tagging in Core Services and Identity service. Example can be found on [Github](https://github.com/oracle/oci-go-sdk/tree/master/example/example_tagging_test.go)
- Fix ComposingConfigurationProvider to not accept a nil ConfigurationProvider
- Support for passphrase configuration to FileConfiguration provider

## 1.0.0 - 2018-02-28 Initial Release
### Added
- Support for Audit service
- Support for Core Services (Networking, Compute, Block Volume)
- Support for Database service
- Support for IAM service
- Support for Load Balancing service
- Support for Object Storage service
