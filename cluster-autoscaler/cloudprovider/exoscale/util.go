/*
Copyright 2021 The Kubernetes Authors.

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

package exoscale

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
)

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", exoscaleProviderIDPrefix, nodeID)
}

// toNodeID returns a node or Compute instance ID from the given provider ID.
func toNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, exoscaleProviderIDPrefix)
}

// toInstance converts the given egoscale.VirtualMachine to a cloudprovider.Instance.
func toInstance(instance *egoscale.Instance) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(*instance.ID),
		Status: toInstanceStatus(*instance.State),
	}
}

// toInstanceStatus converts the given Exoscale API Compute instance status to a cloudprovider.InstanceStatus.
func toInstanceStatus(state string) *cloudprovider.InstanceStatus {
	if state == "" {
		return nil
	}

	switch state {
	case "starting":
		return &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}

	case "running":
		return &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}

	case "stopping":
		return &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}

	default:
		return &cloudprovider.InstanceStatus{ErrorInfo: &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-exoscale",
			ErrorMessage: "error",
		}}
	}
}

// pollCmd executes the specified callback function until either it returns true or a non-nil error,
// or the if context times out.
func pollCmd(ctx context.Context, callback func() (bool, error)) error {
	timeout := time.Minute * 10
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for t := time.Tick(time.Second * 10); ; { // nolint: staticcheck
		ok, err := callback()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		select {
		case <-c.Done():
			return fmt.Errorf("context timeout after: %v", timeout)
		case <-t:
			continue
		}
	}
}
