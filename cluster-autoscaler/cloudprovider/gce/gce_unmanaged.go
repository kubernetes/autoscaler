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

package gce

import (
	"fmt"
	"sort"
	"time"

	"github.com/golang/glog"
	gce "google.golang.org/api/compute/v1"
)

const (
	// maxPages prevents bugs from causing infinitely looping
	maxPages = 25

	startOperationTimeout  = 1 * time.Minute
	startOperationInterval = 1 * time.Second
	stopOperationTimeout   = 1 * time.Minute
	stopOperationInterval  = 1 * time.Second
)

func (m *GceManager) ListInstances(mig *Mig) (all, running, stopping []string, err error) {
	names := make(map[string]struct{})
	pageToken := ""
	page := 0
	for ; page == 0 || (pageToken != "" && page < maxPages); page++ {
		opts := &gce.InstanceGroupsListInstancesRequest{}
		listCall := m.gceService.InstanceGroups.ListInstances(mig.Project, mig.Zone, mig.Name, opts)
		if pageToken != "" {
			listCall.PageToken(pageToken)
		}
		res, err := listCall.Do()
		if err != nil {
			return nil, nil, nil, err
		}
		pageToken = res.NextPageToken
		for _, i := range res.Items {
			name := i.Instance
			if _, ok := names[name]; ok {
				continue
			}
			all = append(all, name)
			names[name] = struct{}{}

			switch i.Status {
			case "STOPPING":
				stopping = append(stopping, name)
			case "RUNNING":
				running = append(running, name)
			}
		}
		if page >= maxPages {
			sort.Strings(all)
			sort.Strings(stopping)
			sort.Strings(running)
			return all, running, stopping, fmt.Errorf("received too many pages of results for managed instance group %v", mig)
		}
	}
	sort.Strings(all)
	sort.Strings(stopping)
	sort.Strings(running)
	return all, running, stopping, nil
}

func (m *GceManager) GetMigInstanceSize(mig *Mig) (int64, error) {
	_, running, stopping, err := m.ListInstances(mig)
	if err != nil {
		return -1, err
	}
	return int64(len(running) - len(stopping)), nil
}

type zonedOp struct {
	op   *gce.Operation
	zone string
}

func (m *GceManager) SetMigInstanceSize(mig *Mig, size int64) error {
	all, running, _, err := m.ListInstances(mig)
	if err != nil {
		return err
	}
	needed := size - int64(len(running))
	free := make(map[string]struct{})
	for _, instance := range all {
		free[instance] = struct{}{}
	}
	for _, instance := range running {
		delete(free, instance)
	}

	var ops []zonedOp
	var errs []error

	for instance := range free {
		if needed <= 0 {
			break
		}
		needed--
		project, zone, name, err := ParseInstanceUrl(instance)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Starting unmanaged instance %s in project %s and zone %s to match scale goals", name, project, zone)
		op, err := m.gceService.Instances.Start(project, zone, name).Do()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ops = append(ops, zonedOp{op: op, zone: zone})
	}

	for _, op := range ops {
		if err := m.waitForOpTimeout(op.op, mig.Project, op.zone, startOperationTimeout, startOperationInterval); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	if needed > 0 {
		return fmt.Errorf("not enough instances available to reach %d, %d/%d are running", size, len(running), len(all))
	}
	return nil
}

// StopInstances stops the given instances. All instances must be controlled by the same instance group.
func (m *GceManager) StopInstances(instances []*GceRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonMig, err := m.GetMigForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		mig, err := m.GetMigForInstance(instance)
		if err != nil {
			return err
		}
		if mig != commonMig {
			return fmt.Errorf("Cannot stop instances which don't belong to the same MIG.")
		}
	}

	var ops []*gce.Operation
	var errs []error
	for _, instance := range instances {
		op, err := m.gceService.Instances.Stop(instance.Project, instance.Zone, instance.Name).Do()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ops = append(ops, op)
	}

	for _, op := range ops {
		if err := m.waitForOpTimeout(op, commonMig.Project, commonMig.Zone, stopOperationTimeout, stopOperationInterval); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// GCE
func (m *GceManager) waitForOpTimeout(operation *gce.Operation, project string, zone string, timeout, interval time.Duration) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(interval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		if op, err := m.gceService.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s %s status: %s", project, zone, operation.Name, op.Status)
			if op.Status == "DONE" {
				return nil
			}
		} else {
			glog.Warningf("Error while getting operation %s on %s: %v", operation.Name, operation.TargetLink, err)
		}
	}
	return fmt.Errorf("Timeout while waiting for operation %s on %s to complete.", operation.Name, operation.TargetLink)
}
