/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utho

import (
	"context"
	"errors"
	"fmt"
)

// KubernetesService provides methods for interacting with Kubernetes resources via the Utho API.
type KubernetesService service

// KubernetesList represents a list of Kubernetes clusters returned by the API.
type KubernetesList struct {
	K8s     []K8s  `json:"k8s,omitempty"`
	Rcode   string `json:"rcode,omitempty"`
	Status  string `json:"status" faker:"oneof: success, failure"`
	Message string `json:"message" faker:"sentence"`
}
// K8s represents a single Kubernetes cluster.
type K8s struct {
	ID             int                 `json:"id,string"`
	CreatedAt      string              `json:"created_at" faker:"timestamp"`
	Dcslug         string              `json:"dcslug"`
	RefID          string              `json:"ref_id" faker:"uuid_digit"`
	Nodepool       string              `json:"nodepool"`
	Hostname       string              `json:"hostname"`
	RAM            int                 `json:"ram,string"`
	CPU            int                 `json:"cpu,string"`
	Disksize       int                 `json:"disksize,string"`
	Bandwidth      string              `json:"bandwidth,omitempty"`
	AppStatus      string              `json:"app_status" faker:"oneof: Active, Inactive"`
	IP             string              `json:"ip" faker:"ipv4"`
	Cost           int                 `json:"cost,omitempty"`
	Cloudid        int                 `json:"cloudid,string"`
	Powerstatus    string              `json:"powerstatus" faker:"oneof: On, Off"`
	Dclocation     K8sDclocation       `json:"dclocation"`
	Status         string              `json:"status" faker:"oneof: Running, Stopped"`
	WorkerCount    string              `json:"worker_count"`
	LoadBalancers  []K8sLoadbalancers  `json:"load_balancers"`
	TargetGroups   []K8sTargetGroups   `json:"target_groups"`
	SecurityGroups []K8sSecurityGroups `json:"security_groups"`
}

// Read
type KubernetesRead struct {
	Info           KubernetesClusterInfo      `json:"info"`
	Vpc            []VpcDetails               `json:"vpc"`
	Nodepools      map[string]NodepoolDetails `json:"nodepools"`
	LoadBalancers  []K8sLoadbalancers         `json:"load_balancers"`
	TargetGroups   []K8sTargetGroups          `json:"target_groups"`
	SecurityGroups []K8sSecurityGroups        `json:"security_groups"`
	Rcode          string                     `json:"rcode,omitempty"`
	Status         string                     `json:"status,omitempty"`
	Message        string                     `json:"message,omitempty"`
}
type KubernetesClusterInfo struct {
	Cluster KubernetesClusterMetadata `json:"cluster"`
	Master  MasterNodeDetails         `json:"master"`
}
type KubernetesClusterMetadata struct {
	ID              int           `json:"id,string"`
	Version         string        `json:"version,omitempty"`
	Label           string        `json:"label,omitempty"`
	Endpoint        string        `json:"endpoint,omitempty"`
	Dcslug          string        `json:"dcslug,omitempty"`
	AutoUpgrade     string        `json:"auto_upgrade,omitempty"`
	SurgeUpgrade    string        `json:"surge_upgrade,omitempty"`
	Ipv4            string        `json:"ipv4,omitempty"`
	ClusterSubnet   string        `json:"cluster_subnet,omitempty"`
	ServiceSubnet   string        `json:"service_subnet,omitempty"`
	Tags            string        `json:"tags,omitempty"`
	CreatedAt       string        `json:"created_at,omitempty"`
	UpdatedAt       string        `json:"updated_at,omitempty"`
	DeletedAt       string        `json:"deleted_at,omitempty"`
	Status          string        `json:"status,omitempty"`
	Nodepools       string        `json:"nodepools,omitempty"`
	Vpc             string        `json:"vpc,omitempty"`
	PublicIpEnabled string        `json:"public_ip_enabled,omitempty"`
	LoadBalancers   string        `json:"load_balancers,omitempty"`
	SecurityGroups  string        `json:"security_groups,omitempty"`
	TargetGroups    string        `json:"target_groups,omitempty"`
	Userid          string        `json:"userid,omitempty"`
	Powerstatus     string        `json:"powerstatus,omitempty"`
	Dns             string        `json:"dns,omitempty"`
	Dclocation      K8sDclocation `json:"dclocation"`
}
type MasterNodeDetails struct {
	Cloudid        int            `json:"cloudid,string"`
	Hostname       string         `json:"hostname"`
	Ram            int            `json:"ram,string"`
	Cpu            int            `json:"cpu,string"`
	Cost           string         `json:"cost,omitempty"`
	Disksize       string         `json:"disksize,omitempty"`
	Bandwidth      string         `json:"bandwidth,omitempty"`
	AppStatus      string         `json:"app_status,omitempty"`
	Dcslug         string         `json:"dcslug,omitempty"`
	Planid         int            `json:"planid,string,omitempty"`
	Ip             string         `json:"ip,omitempty"`
	PrivateNetwork PrivateNetwork `json:"private_network,omitempty"`
}

type VpcDetails struct {
	ID         string `json:"id"`
	IsVpc      string `json:"is_vpc,omitempty"`
	VpcNetwork string `json:"vpc_network"`
}

type NodepoolDetails struct {
	ID        string       `json:"id"`
	Size      string       `json:"size"`
	Cost      float64      `json:"cost"`
	Planid    int          `json:"planid,string"`
	Ip        string       `json:"ip"`
	Count     int          `json:"count,string"`
	AutoScale bool         `json:"auto_scale,omitempty"`
	MinNodes  int          `json:"min_size,omitempty"`
	MaxNodes  int          `json:"max_size,omitempty"`
	Policies  []any        `json:"policies" faker:"-"`
	Workers   []WorkerNode `json:"workers"`
}
type WorkerNode struct {
	ID             int            `json:"cloudid,string"`
	Nodepool       string         `json:"nodepool"`
	Hostname       string         `json:"hostname"`
	Ram            int            `json:"ram,string"`
	Cost           string         `json:"cost"`
	Cpu            int            `json:"cpu,string"`
	Disksize       int            `json:"disksize,string"`
	Bandwidth      string         `json:"bandwidth"`
	AppStatus      string         `json:"app_status"`
	Ip             string         `json:"ip"`
	Planid         int            `json:"planid,string"`
	Status         string         `json:"status"`
	PrivateNetwork PrivateNetwork `json:"private_network"`
}

// Generic
type PrivateNetwork struct {
	Ip         string `json:"ip"`
	Vpc        string `json:"vpc"`
	VpcNetwork string `json:"vpc_network"`
}

type K8sDclocation struct {
	Location string `json:"location"`
	Country  string `json:"country"`
	Dc       string `json:"dc"`
	Dccc     string `json:"dccc"`
}

type K8sLoadbalancers struct {
	ID   int    `json:"lbid,string"`
	Name string `json:"name"`
	IP   string `json:"ip" faker:"ipv4"`
}

type K8sTargetGroups struct {
	ID       int    `json:"id,string"`
	Name     string `json:"name"`
	Protocol any    `json:"protocol" faker:"-"`
	Port     string `json:"port"`
}

type K8sSecurityGroups struct {
	ID   int    `json:"id,string"`
	Name string `json:"name"`
}

type CreateKubernetesParams struct {
	Dcslug         string                  `json:"dcslug"`
	ClusterLabel   string                  `json:"cluster_label"`
	ClusterVersion string                  `json:"cluster_version"`
	Nodepools      []CreateNodepoolsParams `json:"nodepools"`
	Auth           string                  `json:"auth"`
	Vpc            string                  `json:"vpc"`
	Subnet         string                  `json:"subnet"`
	NetworkType    string                  `json:"network_type"`
	Firewall       string                  `json:"firewall"`
	Cpumodel       string                  `json:"cpumodel"`
	SecurityGroups string                  `json:"security_groups"`
}
type CreateNodepoolsParams struct {
	Label    string                           `json:"label"`
	Size     string                           `json:"size"`
	PoolType string                           `json:"pool_type"`
	MaxCount string                           `json:"maxCount,omitempty"`
	Count    string                           `json:"count"`
	Ebs      []CreateNodePoolEbs              `json:"ebs"`
	Policies []CreateKubernetesPoliciesParams `json:"policies,omitempty"`
}
type CreateKubernetesPoliciesParams struct {
	Adjust   int    `json:"adjust"`
	Compare  string `json:"compare"`
	Cooldown int    `json:"cooldown"`
	Period   string `json:"period"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Product  string `json:"product"`
	Maxsize  string `json:"maxsize"`
	Minsize  string `json:"minsize"`
}

type CreateKubernetesNodePoolParams struct {
	ClusterId int
	Nodepools []CreateNodepoolsDetails `json:"nodepools"`
}
type CreateNodepoolsDetails struct {
	Label string              `json:"label"`
	Size  string              `json:"size"`
	Count string              `json:"count"`
	Ebs   []CreateNodePoolEbs `json:"ebs"`
}
type CreateNodePoolEbs struct {
	Disk string `json:"disk"`
	Type string `json:"type"`
}

// Create creates a new Kubernetes cluster using the provided parameters.
func (k *KubernetesService) Create(ctx context.Context, params CreateKubernetesParams) (*CreateResponse, error) {
	reqUrl := "kubernetes/deploy"
	req, _ := k.client.NewRequest("POST", reqUrl, &params)

	var kubernetes CreateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

// Read retrieves information about a specific Kubernetes cluster by its ID.
func (k *KubernetesService) Read(ctx context.Context, clusterId int) (*KubernetesRead, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetes KubernetesRead
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Info.Cluster.ID != clusterId {
		return nil, errors.New("sorry we unable to find this cluster or you dont have access")
	}

	return &kubernetes, nil
}

// List retrieves a list of all Kubernetes clusters.
func (k *KubernetesService) List(ctx context.Context) ([]K8s, error) {
	reqUrl := "kubernetes"
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetes KubernetesList
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return kubernetes.K8s, nil
}

type DeleteKubernetesParams struct {
	ClusterId int
	// confirm message"I am aware this action will delete data and cluster permanently"
	Confirm string `json:"confirm"`
}

// Delete deletes a Kubernetes cluster specified by the given parameters.
func (k *KubernetesService) Delete(ctx context.Context, params DeleteKubernetesParams) (*DeleteResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/destroy", params.ClusterId)
	req, _ := k.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := k.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

type CreateKubernetesLoadbalancerParams struct {
	ClusterId      int
	LoadbalancerId int
}

// Loadbalancer
// CreateLoadbalancer attaches a load balancer to a Kubernetes cluster.
func (k *KubernetesService) CreateLoadbalancer(ctx context.Context, params CreateKubernetesLoadbalancerParams) (*CreateResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/loadbalancer/%d", params.ClusterId, params.LoadbalancerId)
	req, _ := k.client.NewRequest("POST", reqUrl, nil)

	var kubernetes CreateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

// ReadLoadbalancer retrieves information about a specific load balancer in a Kubernetes cluster.
func (k *KubernetesService) ReadLoadbalancer(ctx context.Context, clusterId, loadbalancerId int) (*K8sLoadbalancers, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Info.Cluster.Status != "Active" && kubernetess.Info.Cluster.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}
	var loadbalancers K8sLoadbalancers
	for _, r := range kubernetess.LoadBalancers {
		if r.ID == loadbalancerId {
			loadbalancers = r
		}
	}
	if loadbalancers.ID == 0 {
		return nil, errors.New("kubernetess loadbalancer not found")
	}

	return &loadbalancers, nil
}

// ListLoadbalancers retrieves all load balancers for a given Kubernetes cluster.
func (k *KubernetesService) ListLoadbalancers(ctx context.Context, clusterId int) ([]K8sLoadbalancers, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	return kubernetess.LoadBalancers, nil
}

// DeleteLoadbalancer deletes a load balancer from a Kubernetes cluster.
func (k *KubernetesService) DeleteLoadbalancer(ctx context.Context, clusterId, kubernetesLoadbalancerId int) (*DeleteResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/loadbalancer/%d", clusterId, kubernetesLoadbalancerId)
	req, _ := k.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := k.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

type CreateKubernetesSecurityGroupParams struct {
	ClusterId                 int
	KubernetesSecurityGroupId int
}

// SecurityGroup
// CreateSecurityGroup attaches a security group to a Kubernetes cluster.
func (k *KubernetesService) CreateSecurityGroup(ctx context.Context, params CreateKubernetesSecurityGroupParams) (*CreateResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/securitygroup/%d", params.ClusterId, params.KubernetesSecurityGroupId)
	req, _ := k.client.NewRequest("POST", reqUrl, nil)

	var kubernetes CreateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

// ReadSecurityGroup retrieves information about a specific security group in a Kubernetes cluster.
func (k *KubernetesService) ReadSecurityGroup(ctx context.Context, clusterId, securitygroupId int) (*K8sSecurityGroups, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}
	var securitygroups K8sSecurityGroups
	for _, r := range kubernetess.SecurityGroups {
		if r.ID == securitygroupId {
			securitygroups = r
		}
	}
	if securitygroups.ID == 0 {
		return nil, errors.New("kubernetess securitygroup not found")
	}

	return &securitygroups, nil
}

// ListSecurityGroups retrieves all security groups for a given Kubernetes cluster.
func (k *KubernetesService) ListSecurityGroups(ctx context.Context, clusterId int) ([]K8sSecurityGroups, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	return kubernetess.SecurityGroups, nil
}

// DeleteSecurityGroup deletes a security group from a Kubernetes cluster.
func (k *KubernetesService) DeleteSecurityGroup(ctx context.Context, clusterId, kubernetesSecurityGroupId int) (*DeleteResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/securitygroup/%d", clusterId, kubernetesSecurityGroupId)
	req, _ := k.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := k.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

type CreateKubernetesTargetgroupParams struct {
	ClusterId               int
	KubernetesTargetgroupId int
}

// Targetgroup
// CreateTargetgroup attaches a target group to a Kubernetes cluster.
func (k *KubernetesService) CreateTargetgroup(ctx context.Context, params CreateKubernetesTargetgroupParams) (*CreateResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/targetgroup/%d", params.ClusterId, params.KubernetesTargetgroupId)
	req, _ := k.client.NewRequest("POST", reqUrl, nil)

	var kubernetes CreateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

// ReadTargetgroup retrieves information about a specific target group in a Kubernetes cluster.
func (k *KubernetesService) ReadTargetgroup(ctx context.Context, clusterId, targetgroupId int) (*K8sTargetGroups, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	if kubernetess.Info.Cluster.ID == 0 {
		return nil, errors.New("no Cluster Found")
	}
	var targetgroups K8sTargetGroups
	for _, tg := range kubernetess.TargetGroups {
		if tg.ID == targetgroupId {
			targetgroups = tg
		}
	}
	if targetgroups.ID == 0 {
		return nil, errors.New("kubernetess targetgroup not found")
	}

	return &targetgroups, nil
}

// ListTargetgroups retrieves all target groups for a given Kubernetes cluster.
func (k *KubernetesService) ListTargetgroups(ctx context.Context, clusterId int) ([]K8sTargetGroups, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	return kubernetess.TargetGroups, nil
}

// DeleteTargetgroup deletes a target group from a Kubernetes cluster.
func (k *KubernetesService) DeleteTargetgroup(ctx context.Context, clusterId, kubernetesTargetgroupId int) (*DeleteResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/targetgroup/%d", clusterId, kubernetesTargetgroupId)
	req, _ := k.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := k.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

// PowerOff powers off a Kubernetes cluster by its ID.
func (k *KubernetesService) PowerOff(ctx context.Context, clusterId int) (*BasicResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/stop", clusterId)
	req, _ := k.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := k.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// PowerOn powers on a Kubernetes cluster by its ID.
func (k *KubernetesService) PowerOn(ctx context.Context, clusterId int) (*BasicResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/start", clusterId)
	req, _ := k.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := k.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// NodePool
// CreateNodePool creates a new node pool in a Kubernetes cluster.
func (k *KubernetesService) CreateNodePool(ctx context.Context, params CreateKubernetesNodePoolParams) (*CreateResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/nodepool/add", params.ClusterId)
	req, _ := k.client.NewRequest("POST", reqUrl, &params)

	var kubernetes CreateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

// ReadNodePool retrieves information about a specific node pool in a Kubernetes cluster.
func (k *KubernetesService) ReadNodePool(ctx context.Context, clusterId int, nodePoolId string) (*NodepoolDetails, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	if kubernetess.Info.Cluster.ID == 0 {
		return nil, errors.New("no Cluster Found")
	}
	var nodepools NodepoolDetails
	for id, np := range kubernetess.Nodepools {
		if id == nodePoolId {
			np.ID = id
			nodepools = np
		}
	}
	if len(nodepools.ID) == 0 {
		return nil, errors.New("kubernetess NodePool not found")
	}

	return &nodepools, nil
}

// ListNodePools retrieves all node pools for a given Kubernetes cluster.
func (k *KubernetesService) ListNodePools(ctx context.Context, clusterId string) ([]NodepoolDetails, error) {
	reqUrl := fmt.Sprintf("kubernetes/%s", clusterId)
	req, _ := k.client.NewRequest("GET", reqUrl)

	var kubernetess KubernetesRead
	_, err := k.client.Do(req, &kubernetess)
	if err != nil {
		return nil, err
	}
	if kubernetess.Status != "success" && kubernetess.Status != "" {
		return nil, errors.New(kubernetess.Message)
	}

	nodepools := make([]NodepoolDetails, 0, len(kubernetess.Nodepools))
	for id, np := range kubernetess.Nodepools {
		np.ID = id
		nodepools = append(nodepools, np)
	}
	return nodepools, nil
}

type UpdateKubernetesAutoscaleNodepool struct {
	ClusterId  int
	NodePoolId string
	Count      string                           `json:"count"`
	Label      string                           `json:"label,omitempty"`
	PoolType   string                           `json:"pool_type,omitempty"`
	Size       string                           `json:"size,omitempty"`
	Policies   []CreateKubernetesPoliciesParams `json:"policies,omitempty"`
	MinNodes   int                              `json:"min_nodes,omitempty"`
	MaxNodes   int                              `json:"max_nodes,omitempty"`
}

type UpdateKubernetesAutoscaleNodepoolResponse struct {
	ID        string        `json:"id"`
	Size      string        `json:"size"`
	Cost      float64       `json:"cost"`
	Planid    string        `json:"planid"`
	Count     int           `json:"count"`
	AutoScale int           `json:"auto_scale,string"`
	MinNodes  int           `json:"min_nodes"`
	MaxNodes  int           `json:"max_nodes"`
	Policies  []interface{} `json:"policies"`
	Workers   []WorkerNode  `json:"workers"`

	Status  string `json:"status"`
	Message string `json:"message"`
}

// UpdateNodePool updates an autoscale node pool in a Kubernetes cluster.
func (k *KubernetesService) UpdateNodePool(ctx context.Context, params UpdateKubernetesAutoscaleNodepool) (*UpdateKubernetesAutoscaleNodepoolResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/nodepool/%s/update", params.ClusterId, params.NodePoolId)
	req, _ := k.client.NewRequest("POST", reqUrl, &params)

	var kubernetes UpdateKubernetesAutoscaleNodepoolResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

type UpdateKubernetesStaticNodepool struct {
	ClusterId  int
	NodePoolId string
	Count      string `json:"count"`
	Label      string `json:"label"`
	PoolType   string `json:"pool_type"`
	Size       string `json:"size"`
}

// UpdateStaticNodepool updates a static node pool in a Kubernetes cluster.
func (k *KubernetesService) UpdateStaticNodepool(ctx context.Context, params UpdateKubernetesStaticNodepool) (*UpdateResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/nodepool/%s/update", params.ClusterId, params.NodePoolId)
	req, _ := k.client.NewRequest("POST", reqUrl, &params)

	var kubernetes UpdateResponse
	_, err := k.client.Do(req, &kubernetes)
	if err != nil {
		return nil, err
	}
	if kubernetes.Status != "success" && kubernetes.Status != "" {
		return nil, errors.New(kubernetes.Message)
	}

	return &kubernetes, nil
}

type DeleteNodeParams struct {
	ClusterId int
	PoolId    string
	NodeId    string
}

// DeleteNode deletes a node from a node pool in a Kubernetes cluster.
func (k *KubernetesService) DeleteNode(ctx context.Context, params DeleteNodeParams) (*DeleteResponse, error) {
	reqUrl := fmt.Sprintf("kubernetes/%d/nodepool/%s/%s/delete", params.ClusterId, params.PoolId, params.NodeId)
	req, _ := k.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := k.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}
