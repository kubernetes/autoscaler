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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/kubernetes"
	"regexp"

	klog "k8s.io/klog/v2"
)

const (
	clusterServerTagPrefix string = "k8sca-"
	nodeGroupTagPrefix     string = "k8scang-"
)

// manager handles Kamatera communication and holds information about
// the node groups
type manager struct {
	client     kamateraAPIClient
	config     *kamateraConfig
	nodeGroups map[string]*NodeGroup // key: NodeGroup.id
	instances  map[string]*Instance  // key: Instance.id (which is also the Kamatera server name)
	kubeClient kubernetes.Interface
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
	servers, err := m.client.ListServers(
		context.Background(),
		m.instances,
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
	klog.V(2).Infof("Kamatera node groups after refresh:")
	for _, ng := range nodeGroups {
		klog.V(2).Infof("%s", ng.extendedDebug())
	}

	m.nodeGroups = nodeGroups
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
	ng := &NodeGroup{
		id:           name,
		manager:      m,
		minSize:      cfg.minSize,
		maxSize:      cfg.maxSize,
		instances:    instances,
		serverConfig: serverConfig,
	}
	return ng, nil
}

func (m *manager) getNodeGroupInstances(name string, servers []Server) (map[string]*Instance, error) {
	clusterTag := fmt.Sprintf("%s%s", clusterServerTagPrefix, m.config.clusterName)
	nodeGroupTag := fmt.Sprintf("%s%s", nodeGroupTagPrefix, name)
	instances := make(map[string]*Instance)
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
		if m.instances[server.Name] == nil {
			var state cloudprovider.InstanceState
			if server.PowerOn {
				state = cloudprovider.InstanceRunning
			} else {
				// for new servers that are stopped we assume they were deleted previously and deletion failed
				state = cloudprovider.InstanceDeleting
			}
			m.instances[server.Name] = &Instance{
				Id:      server.Name,
				Status:  &cloudprovider.InstanceStatus{State: state},
				PowerOn: server.PowerOn,
				Tags:    server.Tags,
			}
		} else {
			if server.PowerOn {
				m.instances[server.Name].Status.State = cloudprovider.InstanceRunning
			} else {
				// we can only make assumption about server state being powered on
				// for other conditions we can't know why server is powered off, so we can't update state
			}
			m.instances[server.Name] = m.instances[server.Name]
			m.instances[server.Name].PowerOn = server.PowerOn
			m.instances[server.Name].Tags = server.Tags
		}
		if hasClusterTag && hasNodeGroupTag {
			instances[server.Name] = m.instances[server.Name]
		}
	}
	return instances, nil
}

func (m *manager) addInstance(server Server, state cloudprovider.InstanceState) (*Instance, error) {
	if m.instances[server.Name] == nil {
		m.instances[server.Name] = &Instance{
			Id:      server.Name,
			Status:  &cloudprovider.InstanceStatus{State: state},
			PowerOn: server.PowerOn,
			Tags:    server.Tags,
		}
	} else {
		m.instances[server.Name].Status.State = state
		m.instances[server.Name].PowerOn = server.PowerOn
		m.instances[server.Name].Tags = server.Tags
	}
	return m.instances[server.Name], nil
}
