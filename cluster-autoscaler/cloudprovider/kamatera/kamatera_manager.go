/*
Copyright 2016 The Kubernetes Authors.

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

package kamatera

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/kubernetes"

	"k8s.io/klog/v2"
)

const (
	clusterServerTagPrefix string = "k8sca-"
	nodeGroupTagPrefix     string = "k8scang-"
)

// manager handles Kamatera communication and holds information about
// the node groups
type manager struct {
	client       kamateraAPIClient
	config       *kamateraConfig
	nodeGroupsMu sync.RWMutex
	nodeGroups   map[string]*NodeGroup // key: NodeGroup.id
	instancesMu  sync.RWMutex
	instances    map[string]*Instance // key: Instance.id (which is the cloud provider ID)
	kubeClient   kubernetes.Interface
}

func newManager(config io.Reader, kubeClient kubernetes.Interface) (*manager, error) {
	cfg, err := buildCloudConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}
	client := buildKamateraAPIClient(cfg.apiClientId, cfg.apiSecret, cfg.apiUrl)
	m := &manager{
		client:     client,
		config:     cfg,
		nodeGroups: make(map[string]*NodeGroup),
		instances:  make(map[string]*Instance),
		kubeClient: kubeClient,
	}
	return m, nil
}

func (m *manager) refresh() error {
	instancesSnapshot := m.snapshotInstances()
	servers, err := m.client.ListServers(
		context.Background(),
		instancesSnapshot,
		m.config.filterNamePrefix,
		m.config.providerIDPrefix,
	)
	if err != nil {
		return fmt.Errorf("failed to get list of Kamatera servers from Kamatera API: %v", err)
	}
	nodeGroups := make(map[string]*NodeGroup)
	for nodeGroupName, nodeGroupCfg := range m.config.nodeGroupCfg {
		nodeGroup, err := m.buildNodeGroup(nodeGroupName, nodeGroupCfg, servers)
		if err != nil {
			return fmt.Errorf("failed to build node group %s: %v", nodeGroupName, err)
		}
		nodeGroups[nodeGroupName] = nodeGroup
	}

	// show some debug info
	klog.V(4).Infof("Kamatera node groups after refresh:")
	for _, ng := range nodeGroups {
		klog.V(4).Infof("%s", ng.extendedDebug())
	}

	m.nodeGroupsMu.Lock()
	m.nodeGroups = nodeGroups
	m.nodeGroupsMu.Unlock()
	return nil
}

func (m *manager) buildNodeGroup(name string, cfg *nodeGroupConfig, servers []Server) (*NodeGroup, error) {
	// TODO: do validation of server args with Kamatera api
	instances, err := m.getNodeGroupInstances(name, servers)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for node group %s: %v", name, err)
	}
	password := cfg.Password
	if len(password) == 0 {
		password = "__generate__"
	}
	scriptBytes, err := base64.StdEncoding.DecodeString(cfg.ScriptBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode script for node group %s: %v", name, err)
	}
	script := string(scriptBytes)
	if len(script) < 1 {
		return nil, fmt.Errorf("script for node group %s is empty", name)
	}
	if len(cfg.Datacenter) < 1 {
		return nil, fmt.Errorf("datacenter for node group %s is empty", name)
	}
	if len(cfg.Image) < 1 {
		return nil, fmt.Errorf("image for node group %s is empty", name)
	}
	if len(cfg.Cpu) < 1 {
		return nil, fmt.Errorf("cpu for node group %s is empty", name)
	}
	if len(cfg.Ram) < 1 {
		return nil, fmt.Errorf("ram for node group %s is empty", name)
	}
	if len(cfg.Disks) < 1 {
		return nil, fmt.Errorf("no disks for node group %s", name)
	}
	if len(cfg.Networks) < 1 {
		return nil, fmt.Errorf("no networks for node group %s", name)
	}
	billingCycle := cfg.BillingCycle
	if billingCycle == "" {
		billingCycle = "hourly"
	} else if billingCycle != "hourly" && billingCycle != "monthly" {
		return nil, fmt.Errorf("billing cycle for node group %s is invalid", name)
	}
	tags := []string{
		fmt.Sprintf("%s%s", clusterServerTagPrefix, m.config.clusterName),
		fmt.Sprintf("%s%s", nodeGroupTagPrefix, name),
	}
	for _, tag := range tags {
		if len(tag) < 3 {
			return nil, fmt.Errorf("tag %s is too short, must be at least 3 characters", tag)
		}
		if len(tag) > 24 {
			return nil, fmt.Errorf("tag %s is too long, must be at most 24 characters", tag)
		}
		tagRegexp := `^[a-zA-Z0-9\-_\s\.]{3,24}$`
		matched, err := regexp.MatchString(tagRegexp, tag)
		if err != nil {
			return nil, fmt.Errorf("failed to validate tag %s: %v", tag, err)
		}
		if !matched {
			return nil, fmt.Errorf("tag %s is invalid, must contain only English letters, numbers, dash, underscore, space, dot", tag)
		}
	}
	serverConfig := ServerConfig{
		NamePrefix:     cfg.NamePrefix,
		Password:       password,
		SshKey:         cfg.SshKey,
		Datacenter:     cfg.Datacenter,
		Image:          cfg.Image,
		Cpu:            cfg.Cpu,
		Ram:            cfg.Ram,
		Disks:          cfg.Disks,
		Dailybackup:    cfg.Dailybackup,
		Managed:        cfg.Managed,
		Networks:       cfg.Networks,
		BillingCycle:   billingCycle,
		MonthlyPackage: cfg.MonthlyPackage,
		ScriptFile:     script,
		UserdataFile:   "",
		Tags:           tags,
	}
	ng, exists := m.nodeGroups[name]
	if exists {
		ng.minSize = cfg.minSize
		ng.maxSize = cfg.maxSize
		ng.instances = instances
		ng.serverConfig = serverConfig
		ng.templateLabels = cfg.TemplateLabels
	} else {
		ng = &NodeGroup{
			id:             name,
			manager:        m,
			minSize:        cfg.minSize,
			maxSize:        cfg.maxSize,
			instances:      instances,
			serverConfig:   serverConfig,
			templateLabels: cfg.TemplateLabels,
		}
	}
	return ng, nil
}

func (m *manager) getNodeGroupInstances(name string, servers []Server) (map[string]*Instance, error) {
	clusterTag := fmt.Sprintf("%s%s", clusterServerTagPrefix, m.config.clusterName)
	nodeGroupTag := fmt.Sprintf("%s%s", nodeGroupTagPrefix, name)
	instances := make(map[string]*Instance)
	m.instancesMu.Lock()
	defer m.instancesMu.Unlock()
	if m.nodeGroups[name] != nil {
		for _, instance := range m.nodeGroups[name].instances {
			instances[instance.Id] = instance
		}
	}
	var refreshedInstanceProviderIDs []string
	for _, server := range servers {
		hasClusterTag := false
		hasNodeGroupTag := false
		for _, tag := range server.Tags {
			if tag == nodeGroupTag {
				hasNodeGroupTag = true
			} else if tag == clusterTag {
				hasClusterTag = true
			}
		}
		cloudProviderID := formatKamateraProviderID(m.config.providerIDPrefix, server.Name)
		if m.instances[cloudProviderID] == nil {
			// create a new instance object
			instance := &Instance{
				Id:      cloudProviderID,
				PowerOn: server.PowerOn,
				Tags:    server.Tags,
			}
			refreshedInstanceProviderIDs = append(refreshedInstanceProviderIDs, cloudProviderID)
			if !instance.refresh(m.client, m.config.providerIDPrefix, m.config.PoweroffOnScaleDown, m.kubeClient, true) {
				m.instances[cloudProviderID] = instance
				if hasClusterTag && hasNodeGroupTag {
					instances[cloudProviderID] = instance
				}
			}
		} else {
			// update an existing instance
			instance := m.instances[cloudProviderID]
			instance.PowerOn = server.PowerOn
			instance.Tags = server.Tags
			refreshedInstanceProviderIDs = append(refreshedInstanceProviderIDs, cloudProviderID)
			if instance.refresh(m.client, m.config.providerIDPrefix, m.config.PoweroffOnScaleDown, m.kubeClient, true) {
				delete(m.instances, cloudProviderID)
			} else if hasClusterTag && hasNodeGroupTag {
				instances[cloudProviderID] = instance
			}
		}
	}
	for _, instance := range m.instances {
		wasRefreshed := false
		for _, refreshedID := range refreshedInstanceProviderIDs {
			if instance.Id == refreshedID {
				wasRefreshed = true
				break
			}
		}
		if !wasRefreshed {
			if m.instances[instance.Id].refresh(m.client, m.config.providerIDPrefix, m.config.PoweroffOnScaleDown, m.kubeClient, false) {
				delete(m.instances, instance.Id)
				delete(instances, instance.Id)
			}
		}
	}
	return instances, nil
}

func (m *manager) snapshotInstances() map[string]*Instance {
	m.instancesMu.RLock()
	defer m.instancesMu.RUnlock()
	instances := make(map[string]*Instance, len(m.instances))
	for key, value := range m.instances {
		instances[key] = value
	}
	return instances
}

func (m *manager) addCreatingInstance(serverName string, commandId string, tags []string) *Instance {
	cloudProviderID := formatKamateraProviderID(m.config.providerIDPrefix, serverName)
	klog.V(4).Infof("Adding creating instance %s with command ID %s", cloudProviderID, commandId)
	instance := &Instance{
		Id:                cloudProviderID,
		Status:            &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
		StatusCommandId:   commandId,
		StatusCommandCode: InstanceCommandCreating,
		Tags:              tags,
	}
	m.instancesMu.Lock()
	defer m.instancesMu.Unlock()
	m.instances[cloudProviderID] = instance
	return instance
}
