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

package spotinst

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"
)

type Group struct {
	manager *CloudManager
	group   *aws.Group
	groupID string
	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (grp *Group) MaxSize() int {
	return grp.maxSize
}

// MinSize returns minimum size of the node group.
func (grp *Group) MinSize() int {
	return grp.minSize
}

// TargetSize returns the current target size of the node group.
func (grp *Group) TargetSize() (int, error) {
	size, err := grp.manager.GetGroupSize(grp)
	return int(size), err
}

// IncreaseSize increases the size of the node group.
func (grp *Group) IncreaseSize(delta int) error {
	if delta <= 0 {
		return errors.New("size increase must be positive")
	}
	size, err := grp.manager.GetGroupSize(grp)
	if err != nil {
		return err
	}
	if int(size)+delta > grp.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, grp.MaxSize())
	}
	return grp.manager.SetGroupSize(grp, size+int64(delta))
}

// DeleteNodes deletes nodes from this node group.
func (grp *Group) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := grp.manager.GetGroupSize(grp)
	if err != nil {
		return fmt.Errorf("error when deleting nodes, retrieving size of group %s failed: %v", grp.Id(), err)
	}
	if int(size) <= grp.MinSize() {
		return errors.New("min size reached, nodes will not be deleted")
	}
	toBeDeleted := make([]string, 0)
	for _, node := range nodes {
		belongs, err := grp.Belongs(node)
		if err != nil {
			return fmt.Errorf("failed to check membership of node %s in group %s: %v", node.Name, grp.Id(), err)
		}
		if !belongs {
			return fmt.Errorf("%s belongs to a different group than %s", node.Name, grp.Id())
		}
		instanceID, err := extractInstanceId(node.Spec.ProviderID)
		if err != nil {
			return fmt.Errorf("node %s's cloud provider ID is malformed: %v", node.Name, err)
		}
		toBeDeleted = append(toBeDeleted, instanceID)
	}
	return grp.manager.DeleteInstances(toBeDeleted)
}

// DecreaseTargetSize decreases the target size of the node group.
func (grp *Group) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return errors.New("size decrease size must be negative")
	}
	size, err := grp.manager.GetGroupSize(grp)
	if err != nil {
		return err
	}
	nodes, err := grp.Nodes()
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("size decrease too large - desired:%d existing:%d", int(size)+delta, len(nodes))
	}
	return grp.manager.SetGroupSize(grp, size+int64(delta))
}

// Id returns an unique identifier of the node group.
func (grp *Group) Id() string {
	return grp.groupID
}

// Debug returns a string containing all information regarding this node group.
func (grp *Group) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", grp.Id(), grp.MinSize(), grp.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (grp *Group) Nodes() ([]string, error) {
	in := &aws.StatusGroupInput{
		GroupID: spotinst.String(grp.Id()),
	}
	status, err := grp.manager.groupService.CloudProviderAWS().Status(context.Background(), in)
	if err != nil {
		return []string{}, err
	}
	out := make([]string, 0)
	for _, instance := range status.Instances {
		if instance.ID != nil && instance.AvailabilityZone != nil {
			out = append(out, fmt.Sprintf("aws:///%s/%s",
				*instance.AvailabilityZone, *instance.ID))
		}
	}
	return out, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (grp *Group) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	glog.Infof("No working nodes in node group %s, trying to generate from template", grp.Id())

	template, err := grp.manager.buildGroupTemplate(grp.Id())
	if err != nil {
		return nil, err
	}

	node, err := grp.manager.buildNodeFromTemplate(grp, template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(grp.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (grp *Group) Belongs(node *apiv1.Node) (bool, error) {
	instanceID, err := extractInstanceId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	group, err := grp.manager.GetGroupForInstance(instanceID)
	if err != nil {
		return false, err
	}
	if group == nil {
		return false, fmt.Errorf("%s does not belong to a known group", node.Name)
	}
	return true, nil
}

// Exist checks if the node group really exists on the cloud provider side.
func (grp *Group) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (grp *Group) Create() error {
	return cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (grp *Group) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (grp *Group) Autoprovisioned() bool {
	return false
}

var (
	spotinstProviderRE = regexp.MustCompile(`^spotinst\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*$`)
	awsProviderRE      = regexp.MustCompile(`^aws\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*$`)
)

func extractInstanceId(providerID string) (string, error) {
	var prefix string

	if spotinstProviderRE.FindStringSubmatch(providerID) != nil {
		prefix = "spotinst:///"
	}

	if awsProviderRE.FindStringSubmatch(providerID) != nil {
		prefix = "aws:///"
	}

	if prefix == "" {
		return "", fmt.Errorf("expected node provider ID to be one of the "+
			"forms `spotinst:///<zone>/<instance-id>` or `aws:///<zone>/<instance-id>`, got `%s`", providerID)
	}

	parts := strings.Split(providerID[len(prefix):], "/")
	instanceID := parts[1]

	glog.Infof("Instance ID `%s` extracted from provider `%s`", instanceID, providerID)
	return instanceID, nil
}
