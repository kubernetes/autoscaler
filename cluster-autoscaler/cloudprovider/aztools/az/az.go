/*
Copyright 2018 The Kubernetes Authors.

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

package az

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"k8s.io/utils/exec"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// GetWorkerList get the worker list from given group.
func GetWorkerList(groupID string) ([]string, error) {
	data, err := ioutil.ReadFile("./cluster.yaml")
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	// If there's no worker in cluster.yaml (e.g. only master), the node group will be
	// initialized with 0 workers.
	// And automatic scale up later should not add workers in this node group, the node state
	// will be maintained in autoscaler with node lister. The target size is the only thing
	// been updated during scaling.
	machines := []string{}
	if machineMap, ok := m["machines"]; ok {
		for vm, roleMap := range machineMap.(map[interface{}]interface{}) {
			if roleMap.(map[interface{}]interface{})["role"] == "worker" &&
				roleMap.(map[interface{}]interface{})["node-group"] == groupID {
				machines = append(machines, vm.(string))
			}
		}
	}
	return machines, nil
}

// OnScaleUp is a function called on node group increase in AzToolsCloudProvider.
// First parameter is the NodeGroupInfo id, second is the increase delta.
func OnScaleUp(id string, delta int) error {
	// Backup config.yaml
	output, err := execRun("cp", "deploy/scaler.yaml", "deploy/.scaler.yaml.bak")
	if err != nil {
		return fmt.Errorf("%v, %s", err, output)
	}

	// 1. Scale up group id with delta nodes, it will modify scaler.yaml so we need to
	// back up the file first.
	output, err = execRun("./az_tools.py", "scaleup", id, strconv.Itoa(delta))
	if err != nil {
		restoreScalerConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// Backup cluster.yaml
	output, err = execRun("cp", "cluster.yaml", "deploy/.cluster.yaml.bak")
	if err != nil {
		restoreScalerConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// 2. Generate new cluster.yaml
	output, err = execRun("./az_tools.py", "genconfig")
	if err != nil {
		restoreScalerConfig()
		restoreClusterConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// 3. Run scripts in new workers
	output, err = execRun("./deploy.py", "scriptblocks", "add_scaled_worker")
	if err != nil {
		// TODO(harry): delete the new scaled node.
		restoreScalerConfig()
		restoreClusterConfig()
		return fmt.Errorf("%v, %s", err, output)
	}
	// TODO(harry): should we handle labels separately for `kubernetes labels`
	glog.Infof("Scale up successfully with %v nodes added", delta)
	return nil
}

type ScalerConfig struct {
	NodeGroupInfos map[string]NodeGroupInfo `yaml:"node_groups"`
}

type NodeGroupInfo struct {
	WorkerNodeNum     int      `yaml:"worker_node_num"`
	LastScaledUpNodes []string `yaml:"last_scaled_up_nodes"`
}

// InitScalerFromConfig is used to initialize new deploy/scaler.yaml from cluster.yaml
func InitScalerFromConfig(grouNames []string) error {
	data, err := ioutil.ReadFile("./cluster.yaml")
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})

	config := ScalerConfig{
		NodeGroupInfos: map[string]NodeGroupInfo{},
	}

	// Initialize, so we don't need to check if exists later.
	for _, grp := range grouNames {
		config.NodeGroupInfos[grp] = NodeGroupInfo{
			WorkerNodeNum:     0,
			LastScaledUpNodes: []string{},
		}
	}

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	if machineMap, ok := m["machines"]; ok {
		for _, machine := range machineMap.(map[interface{}]interface{}) {
			if machine.(map[interface{}]interface{})["role"].(string) == "worker" {
				nodeGroup := machine.(map[interface{}]interface{})["node-group"].(string)
				// Maybe use pointer to avoid copy back?
				ng := config.NodeGroupInfos[nodeGroup]
				ng.WorkerNodeNum += 1
				config.NodeGroupInfos[nodeGroup] = ng
			}
		}
	} else {
		return fmt.Errorf("no machines defined in cluster.yaml")
	}

	d, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Write back
	err = ioutil.WriteFile("deploy/scaler.yaml", d, 0644)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Initialized deploy/scaler.yaml from cluster.yaml, with data: %#v", config)

	return nil
}

func restoreScalerConfig() {
	// Restore scaler.yaml
	execRun("cp", "deploy/.scaler.yaml.bak", "deploy/scaler.yaml")
}

func restoreClusterConfig() {
	// Restore cluster.yaml
	execRun("cp", "deploy/.cluster.yaml.bak", "cluster.yaml")
}

// OnScaleDown is a function called on cluster scale down
func OnScaleDown(id string, nodeName string) error {
	// Backup config.yaml
	output, err := execRun("cp", "deploy/scaler.yaml", "deploy/.scaler.yaml.bak")
	if err != nil {
		return fmt.Errorf("%v, %s", err, output)
	}

	// 1. Delete vm by group ID and name
	output, err = execRun("./az_tools.py", "scaledown", id, nodeName)
	if err != nil {
		restoreScalerConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// Backup cluster.yaml
	output, err = execRun("cp", "cluster.yaml", "deploy/.cluster.yaml.bak")
	if err != nil {
		restoreScalerConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// 2. Generate new cluster.yaml
	output, err = execRun("./az_tools.py", "genconfig")
	if err != nil {
		restoreScalerConfig()
		restoreClusterConfig()
		return fmt.Errorf("%v, %s", err, output)
	}

	// 3. Delete node from kubernetes cluster
	output, err = execRun("./deploy.py", "kubectl", "delete", "node", nodeName)
	if err != nil {
		return fmt.Errorf("%v, %s", err, output)
	}

	glog.Infof("Scale down node: %v successfully", nodeName)
	return nil
}

// OnNodeGroupInfoCreate is a fuction called when a new node group is created.
func OnNodeGroupInfoCreate(id string) error {
	return fmt.Errorf("Not implemented")
}

// OnNodeGroupInfoDelete is a function called when a node group is deleted.
func OnNodeGroupInfoDelete(id string) error {
	return fmt.Errorf("Not implemented")
}

// execRun execute command and return outputs.
func execRun(cmd string, args ...string) ([]byte, error) {
	exe := exec.New()
	glog.V(4).Infof("Executing: %v, %v", cmd, args)
	return exe.Command(cmd, args...).CombinedOutput()
}
