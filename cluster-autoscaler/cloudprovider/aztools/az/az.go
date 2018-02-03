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

	"k8s.io/utils/exec"

	"github.com/golang/glog"
)

// OnScaleUpFunc is a function called on node group increase in AzToolsCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
func OnScaleUp(id string, delta int) error {
	args := []string{"cluster scaled up"}
	dataOut, err := execRun("echo", args...)
	if err != nil {
		return err
	}
	glog.Warningf("%v: groupID: %v, delta: %v", string(dataOut), id, delta)
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
	return exe.Command(cmd, args...).CombinedOutput()
}
