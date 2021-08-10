// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	clusterPath = "/_"
	kubeConfig  = "kubeconfig"
)

var _ KubernetesEngineService = (*kubernetesEngineService)(nil)

type kubernetesEngineService struct {
	client *Client
}

type KubernetesEngineService interface {
	List(ctx context.Context, opts *ListOptions) ([]*Cluster, error)
	Create(ctx context.Context, req *ClusterCreateRequest) (*ExtendedCluster, error)
	Get(ctx context.Context, id string) (*FullCluster, error)
	Delete(ctx context.Context, id string) error
	AddWorkerPools(ctx context.Context, id string, awp *AddWorkerPoolsRequest) ([]*ExtendedWorkerPool, error)
	RecycleNode(ctx context.Context, clusterUID string, PoolID string, NodePhysicalID string) error
	DeleteClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) error
	GetClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) (*WorkerPoolWithNodes, error)
	UpdateClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string, uwp *UpdateWorkerPoolRequest) error
	DeleteClusterWorkerPoolNode(ctx context.Context, clusterUID string, PoolID string, NodeID string) error
	GetKubeConfig(ctx context.Context, clusterUID string) (string, error)
}

type WorkerPool struct {
	Name              string   `json:"name" yaml:"name"`
	Version           string   `json:"version,omitempty" yaml:"version,omitempty"`
	Flavor            string   `json:"flavor" yaml:"flavor"`
	ProfileType       string   `json:"profile_type" yaml:"profile_type"`
	VolumeType        string   `json:"volume_type" yaml:"volume_type"`
	VolumeSize        int      `json:"volume_size" yaml:"volume_size"`
	AvailabilityZone  string   `json:"availability_zone" yaml:"availability_zone"`
	DesiredSize       int      `json:"desired_size" yaml:"desired_size"`
	EnableAutoScaling bool     `json:"enable_autoscaling,omitempty" yaml:"enable_autoscaling,omitempty"`
	MinSize           int      `json:"min_size,omitempty" yaml:"min_size,omitempty"`
	MaxSize           int      `json:"max_size,omitempty" yaml:"max_size,omitempty"`
	Tags              []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type ControllerVersion struct {
	ID          string `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description" yaml:"description"`
	K8SVersion  string `json:"kubernetes_version" yaml:"kubernetes_version"`
}

type Clusters struct {
	Clusters_ []Cluster `json:"clusters" yaml:"clusters"`
}

type Cluster struct {
	UID              string            `json:"uid" yaml:"uid"`
	Name             string            `json:"name" yaml:"name"`
	Version          ControllerVersion `json:"version" yaml:"version"`
	VPCNetworkID     string            `json:"private_network_id" yaml:"private_network_id"`
	AutoUpgrade      bool              `json:"auto_upgrade" yaml:"auto_upgrade"`
	Tags             []string          `json:"tags" yaml:"tags"`
	ProvisionStatus  string            `json:"provision_status" yaml:"provision_status"`
	ClusterStatus    string            `json:"cluster_status" yaml:"cluster_status"`
	CreatedAt        string            `json:"created_at" yaml:"created_at"`
	CreatedBy        string            `json:"created_by" yaml:"created_by"`
	WorkerPoolsCount int               `json:"worker_pools_count" yaml:"worker_pools_count"`
}

type ExtendedCluster struct {
	Cluster
	WorkerPools []ExtendedWorkerPool `json:"worker_pools" yaml:"worker_pools"`
}

type ClusterStat struct {
	WorkerPoolCount int `json:"worker_pools" yaml:"worker_pools"`
	TotalCPU        int `json:"total_cpu" yaml:"total_cpu"`
	TotalMemory     int `json:"total_memory" yaml:"total_memory"`
}

type FullCluster struct {
	ExtendedCluster
	Stat ClusterStat `json:"stat" yaml:"stat"`
}

type ExtendedWorkerPool struct {
	WorkerPool
	UID                string `json:"id" yaml:"id"`
	ProvisionStatus    string `json:"provision_status" yaml:"provision_status"`
	LaunchConfigID     string `json:"launch_config_id" yaml:"launch_config_id"`
	AutoScalingGroupID string `json:"autoscaling_group_id" yaml:"autoscaling_group_id"`
	CreatedAt          string `json:"created_at" yaml:"created_at"`
}

type ExtendedWorkerPools struct {
	WorkerPools []ExtendedWorkerPool `json:"worker_pools" yaml:"worker_pools"`
}

type ClusterCreateRequest struct {
	Name         string       `json:"name" yaml:"name"`
	Version      string       `json:"version" yaml:"version"`
	AutoUpgrade  bool         `json:"auto_upgrade,omitempty" yaml:"auto_upgrade,omitempty"`
	VPCNetworkID string       `json:"private_network_id" yaml:"private_network_id"`
	EnableCloud  bool         `json:"enable_cloud,omitempty" yaml:"enable_cloud,omitempty"`
	Tags         []string     `json:"tags,omitempty" yaml:"tags,omitempty"`
	WorkerPools  []WorkerPool `json:"worker_pools" yaml:"worker_pools"`
}

type AddWorkerPoolsRequest struct {
	WorkerPools []WorkerPool `json:"worker_pools" yaml:"worker_pools"`
}

type PoolNode struct {
	ID           string   `json:"id" yaml:"id"`
	Name         string   `json:"name" yaml:"name"`
	PhysicalID   string   `json:"physical_id" yaml:"physical_id"`
	IPAddresses  []string `json:"ip_addresses" yaml:"ip_addresses"`
	Status       string   `json:"status" yaml:"status"`
	StatusReason string   `json:"status_reason" yaml:"status_reason"`
}

type WorkerPoolWithNodes struct {
	ExtendedWorkerPool
	Nodes []PoolNode `json:"nodes" yaml:"nodes"`
}

type UpdateWorkerPoolRequest struct {
	DesiredSize       int  `json:"desired_size,omitempty" yaml:"desired_size,omitempty"`
	EnableAutoScaling bool `json:"enable_autoscaling,omitempty" yaml:"enable_autoscaling,omitempty"`
	MinSize           int  `json:"min_size,omitempty" yaml:"min_size,omitempty"`
	MaxSize           int  `json:"max_size,omitempty" yaml:"max_size,omitempty"`
}

func (c *kubernetesEngineService) resourcePath() string {
	return clusterPath + "/"
}

func (c *kubernetesEngineService) itemPath(id string) string {
	return strings.Join([]string{clusterPath, id}, "/")
}

func (c *kubernetesEngineService) List(ctx context.Context, opts *ListOptions) ([]*Cluster, error) {
	req, err := c.client.NewRequest(ctx, http.MethodGet, kubernetesServiceName, c.resourcePath(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Clusters []*Cluster `json:"clusters" yaml:"clusters"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Clusters, nil
}

func (c *kubernetesEngineService) Create(ctx context.Context, clcr *ClusterCreateRequest) (*ExtendedCluster, error) {
	var data struct {
		Cluster *ExtendedCluster `json:"cluster" yaml:"cluster"`
	}
	req, err := c.client.NewRequest(ctx, http.MethodPost, kubernetesServiceName, c.resourcePath(), &clcr)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Cluster, nil
}

func (c *kubernetesEngineService) Get(ctx context.Context, id string) (*FullCluster, error) {
	var cluster *FullCluster
	req, err := c.client.NewRequest(ctx, http.MethodGet, kubernetesServiceName, c.itemPath(id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *kubernetesEngineService) Delete(ctx context.Context, id string) error {
	req, err := c.client.NewRequest(ctx, http.MethodDelete, kubernetesServiceName, c.itemPath(id), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)

	if err != nil {
		fmt.Println("error send req")
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *kubernetesEngineService) AddWorkerPools(ctx context.Context, id string, awp *AddWorkerPoolsRequest) ([]*ExtendedWorkerPool, error) {
	req, err := c.client.NewRequest(ctx, http.MethodPost, kubernetesServiceName, c.itemPath(id), &awp)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var respData struct {
		Pools []*ExtendedWorkerPool `json:"worker_pools" yaml:"worker_pools"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}
	return respData.Pools, nil
}

func (c *kubernetesEngineService) RecycleNode(ctx context.Context, clusterUID string, poolID string, nodePhysicalID string) error {
	req, err := c.client.NewRequest(ctx, http.MethodPut, kubernetesServiceName, strings.Join([]string{clusterPath, clusterUID, poolID, nodePhysicalID}, "/"), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *kubernetesEngineService) DeleteClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) error {
	req, err := c.client.NewRequest(ctx, http.MethodDelete, kubernetesServiceName, strings.Join([]string{clusterPath, clusterUID, PoolID}, "/"), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *kubernetesEngineService) GetClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) (*WorkerPoolWithNodes, error) {
	var pool *WorkerPoolWithNodes
	req, err := c.client.NewRequest(ctx, http.MethodGet, kubernetesServiceName, strings.Join([]string{clusterPath, clusterUID, PoolID}, "/"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func (c *kubernetesEngineService) UpdateClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string, uwp *UpdateWorkerPoolRequest) error {
	req, err := c.client.NewRequest(ctx, http.MethodPatch, kubernetesServiceName, strings.Join([]string{clusterPath, clusterUID, PoolID}, "/"), &uwp)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *kubernetesEngineService) DeleteClusterWorkerPoolNode(ctx context.Context, clusterUID string, PoolID string, NodeID string) error {
	req, err := c.client.NewRequest(ctx, http.MethodDelete, kubernetesServiceName, strings.Join([]string{clusterPath, clusterUID, PoolID, NodeID}, "/"), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *kubernetesEngineService) GetKubeConfig(ctx context.Context, clusterUID string) (string, error) {
	req, err := c.client.NewRequest(ctx, http.MethodGet, kubernetesServiceName, strings.Join([]string{c.itemPath(clusterUID), kubeConfig}, "/"), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}
