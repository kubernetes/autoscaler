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

	"k8s.io/utils/exec"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// OnScaleUpFunc is a function called on node group increase in AzToolsCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
func OnScaleUp(id string, delta int) error {
	// Modify worker number in config.yaml
	modifyConfigYaml(delta)

	// Create new vm
	_, err := execRun("./az_tools.py", "scaleup")
	if err != nil {
		return err
	}

	// Generate new cluster.yaml
	_, err = execRun("./az_tools.py", "genconfig")
	if err != nil {
		// TODO(harry): delete the new scaled node. Restore config.yaml
		return err
	}

	// Run scripts in new workers
	// TODO(harry): do we need to wait for this command finish?
	_, err = execRun("./deploy.py", "scriptblocks", "add_scaled_worker")
	if err != nil {
		// TODO(harry): delete the new scaled node. Restore config.yaml cluster.yaml
		return err
	}
	// TODO(harry): should we handle labels separately for `kubernetes labels`
	glog.Infof("Scale up successes with %v nodes added", delta)
	return nil
}

// modifyConfigYaml modifies config.yaml
func modifyConfigYaml(delta int) error {
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	if clusterMap, ok := m["azure_cluster"]; ok {
		for name, cluster := range clusterMap.(map[interface{}]interface{}) {
			if num, ok := cluster.(map[interface{}]interface{})["worker_node_num"]; ok {
				// Changes the nodes number to:
				//   worker_node_num: curr + delta
				cluster.(map[interface{}]interface{})["worker_node_num"] = num.(int) + delta
				// Add delta in the file:
				//   last_scale_up_node_num: delta
				cluster.(map[interface{}]interface{})["last_scale_up_node_num"] = delta
				break
			} else {
				return fmt.Errorf("cluster %v has no worker_node_num defined, autoscaling is not supported for it.", name)
			}
		}
	} else {
		return fmt.Errorf("no azure_cluster defined in config.yaml")
	}

	d, err := yaml.Marshal(&m)
	if err != nil {
		return err
	}

	// Write back
	err = ioutil.WriteFile("config.yaml", d, 0644)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Updated config.yaml with new worker node number.")

	return nil
}

// OnScaleDownFunc is a function called on cluster scale down
func OnScaleDown(id string, node string) error {
	// TODO(harry): this will not be implemented for now, we may want to schedule a cronjob or send some alert instead.
	return fmt.Errorf("Not implemented")
}

// OnNodeGroupCreateFunc is a fuction called when a new node group is created.
func OnNodeGroupCreate(id string) error {
	return fmt.Errorf("Not implemented")
}

// OnNodeGroupDeleteFunc is a function called when a node group is deleted.
func OnNodeGroupDelete(id string) error {
	return fmt.Errorf("Not implemented")
}

// execRun execute command and return outputs.
func execRun(cmd string, args ...string) ([]byte, error) {
	exe := exec.New()
	glog.V(4).Infof("Executing: %v, %v", cmd, args)
	return exe.Command(cmd, args...).CombinedOutput()
}
