package paperspace

import (
	"fmt"
)

type ClusterPlatformType string

const (
	ClusterPlatformAWS        ClusterPlatformType = "aws"
	ClusterPlatformDGX        ClusterPlatformType = "nvidia-dgx"
	ClusterPlatformMetal      ClusterPlatformType = "metal"
	ClusterPlatformPaperspace ClusterPlatformType = "paperspace-cloud"
)

var ClusterAWSRegions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-2",
	"ca-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-central-1",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-southeast-1",
	"ap-southeast-2",
}

var ClusterPlatforms = []ClusterPlatformType{
	ClusterPlatformAWS,
	ClusterPlatformDGX,
	ClusterPlatformMetal,
	ClusterPlatformPaperspace,
}
var DefaultClusterType = 3

type Cluster struct {
	APIToken     APIToken            `json:"apiToken"`
	Domain       string              `json:"fqdn"`
	Platform     ClusterPlatformType `json:"cloud"`
	Name         string              `json:"name"`
	ID           string              `json:"id"`
	Region       string              `json:"region,omitempty"`
	S3Credential S3Credential        `json:"s3Credential"`
	TeamID       string              `json:"teamId"`
	Type         string              `json:"type,omitempty"`
}

type ClusterCreateParams struct {
	RequestParams

	ArtifactsAccessKeyID     string `json:"accessKey,omitempty" yaml:"artifactsAccessKeyId,omitempty"`
	ArtifactsBucketPath      string `json:"bucketPath,omitempty" yaml:"artifactsBucketPath,omitempty"`
	ArtifactsSecretAccessKey string `json:"secretKey,omitempty" yaml:"artifactsSecretAccessKey,omitempty"`
	Domain                   string `json:"fqdn" yaml:"domain"`
	IsDefault                bool   `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
	Name                     string `json:"name" yaml:"name"`
	Platform                 string `json:"cloud,omitempty" yaml:"platform,omitempty"`
	Region                   string `json:"region,omitempty" yaml:"region,omitempty"`
	Type                     int    `json:"type,omitempty" yaml:"type,omitempty"`
}

type ClusterGetParams struct {
	RequestParams
}
type ClusterListParams struct {
	RequestParams

	Filter Filter `json:"filter,omitempty"`
}

type ClusterUpdateAttributeParams struct {
	RequestParams

	Domain string `json:"fqdn,omitempty" yaml:"domain"`
	Name   string `json:"name,omitempty" yaml:"name"`
}

type ClusterUpdateParams struct {
	RequestParams

	Attributes         ClusterUpdateAttributeParams `json:"attributes,omitempty"`
	CreateNewToken     bool                         `json:"createNewToken,omitempty"`
	RegistryAttributes ClusterUpdateRegistryParams  `json:"registryAttributes,omitempty"`
	ID                 string                       `json:"id"`
	RetryWorkflow      bool                         `json:"retryWorkflow,omitempty"`
	S3Attributes       ClusterUpdateS3Params        `json:"s3Attributes,omitempty"`
}

type ClusterUpdateRegistryParams struct {
	RequestParams

	URL        string `json:"url,omitempty"`
	Password   string `json:"password,omitempty"`
	Repository string `json:"repository,omitempty"`
	Username   string `json:"username,omitempty"`
}

type ClusterUpdateS3Params struct {
	RequestParams

	AccessKey string `json:"accessKey,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

func NewClusterListParams() ClusterListParams {
	return ClusterListParams{}
}

func (c Client) CreateCluster(params ClusterCreateParams) (Cluster, error) {
	cluster := Cluster{}
	params.Type = DefaultClusterType

	url := fmt.Sprintf("/clusters/createCluster")
	_, err := c.Request("POST", url, params, &cluster, params.RequestParams)

	return cluster, err
}

func (c Client) GetCluster(id string, params ClusterGetParams) (Cluster, error) {
	cluster := Cluster{}

	url := fmt.Sprintf("/clusters/getCluster?id=%s", id)
	_, err := c.Request("GET", url, nil, &cluster, params.RequestParams)

	return cluster, err
}

func (c Client) GetClusters(params ClusterListParams) ([]Cluster, error) {
	clusters := []Cluster{}

	url := fmt.Sprintf("/clusters/getClusters")
	_, err := c.Request("GET", url, params, &clusters, params.RequestParams)

	return clusters, err
}

func (c Client) UpdateCluster(id string, params ClusterUpdateParams) (Cluster, error) {
	cluster := Cluster{}

	url := fmt.Sprintf("/clusters/updateCluster")
	_, err := c.Request("POST", url, params, &cluster, params.RequestParams)

	return cluster, err
}
