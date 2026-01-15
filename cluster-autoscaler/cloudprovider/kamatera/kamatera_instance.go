/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Instance implements cloudprovider.Instance interface. Instance contains
// configuration info and functions to control a single Kamatera server instance.
type Instance struct {
	// Id is the cloud provider id.
	Id string
	// Status represents status of the node. (Optional)
	Status *cloudprovider.InstanceStatus
	// Kamatera specific fields
	PowerOn bool
	Tags    []string
}

func (i *Instance) delete(client kamateraAPIClient, providerIDPrefix string) error {
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}
	return client.DeleteServer(context.Background(), parseKamateraProviderID(providerIDPrefix, i.Id))
}

func (i *Instance) poweroff(client kamateraAPIClient, providerIDPrefix string) error {
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}
	i.PowerOn = false
	return client.PoweroffServer(context.Background(), parseKamateraProviderID(providerIDPrefix, i.Id))
}

func (i *Instance) poweron(client kamateraAPIClient, providerIDPrefix string) error {
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}
	err := client.PoweronServer(context.Background(), parseKamateraProviderID(providerIDPrefix, i.Id))
	if err != nil {
		return err
	}
	i.PowerOn = true
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}
	return nil
}

func (i *Instance) extendedDebug() string {
	state := ""
	if i.Status == nil {
		state = "Unknown"
	} else if i.Status.State == cloudprovider.InstanceRunning {
		state = "Running"
	} else if i.Status.State == cloudprovider.InstanceCreating {
		state = "Creating"
	} else if i.Status.State == cloudprovider.InstanceDeleting {
		state = "Deleting"
	}
	return fmt.Sprintf("instance ID: %s state: %s powerOn: %v", i.Id, state, i.PowerOn)
}
