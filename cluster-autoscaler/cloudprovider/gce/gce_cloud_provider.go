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
	"strconv"
	"strings"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// GceCloudProvider implements CloudProvider interface.
type GceCloudProvider struct {
	gceManager *GceManager
	migs       []*Mig
}

// BuildGceCloudProvider builds CloudProvider implementation for GCE.
func BuildGceCloudProvider(gceManager *GceManager, specs []string) (*GceCloudProvider, error) {
	gce := &GceCloudProvider{
		gceManager: gceManager,
		migs:       make([]*Mig, 0),
	}
	for _, spec := range specs {
		if err := gce.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}
	return gce, nil
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:migUrl
func (gce *GceCloudProvider) addNodeGroup(spec string) error {
	mig, err := buildMig(spec, gce.gceManager)
	if err != nil {
		return err
	}
	gce.migs = append(gce.migs, mig)
	gce.gceManager.RegisterMig(mig)
	return nil
}

// Name returns name of the cloud provider.
func (gce *GceCloudProvider) Name() string {
	return "gce"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (gce *GceCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(gce.migs))
	for _, mig := range gce.migs {
		result = append(result, mig)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (gce *GceCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	mig, err := gce.gceManager.GetMigForInstance(ref)
	return mig, err
}

// GceRef contains s reference to some entity in GCE/GKE world.
type GceRef struct {
	Project string
	Zone    string
	Name    string
}

// GceRefFromProviderId creates InstanceConfig object
// from provider id which must be in format:
// gce://<project-id>/<zone>/<name>
// TODO(piosz): add better check whether the id is correct
func GceRefFromProviderId(id string) (*GceRef, error) {
	splitted := strings.Split(id[6:], "/")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("Wrong id: expected format gce://<project-id>/<zone>/<name>, got %v", id)
	}
	return &GceRef{
		Project: splitted[0],
		Zone:    splitted[1],
		Name:    splitted[2],
	}, nil
}

// Mig implements NodeGroup interfrace.
type Mig struct {
	GceRef

	gceManager *GceManager

	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (mig *Mig) MaxSize() int {
	return mig.maxSize
}

// MinSize returns minimum size of the node group.
func (mig *Mig) MinSize() int {
	return mig.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kuberentes.
func (mig *Mig) TargetSize() (int, error) {
	size, err := mig.gceManager.GetMigSize(mig)
	return int(size), err
}

// IncreaseSize increases Mig size
func (mig *Mig) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size)+delta > mig.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, mig.MaxSize())
	}
	return mig.gceManager.SetMigSize(mig, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (mig *Mig) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be netative")
	}
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	nodes, err := mig.gceManager.GetMigNodes(mig)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return mig.gceManager.SetMigSize(mig, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (mig *Mig) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetMig, err := mig.gceManager.GetMigForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetMig == nil {
		return false, fmt.Errorf("%s doesn't belong to a known mig", node.Name)
	}
	if targetMig.Id() != mig.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteNodes deletes the nodes from the group.
func (mig *Mig) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size) <= mig.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*GceRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := mig.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different mig than %s", node.Name, mig.Id())
		}
		gceref, err := GceRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, gceref)
	}
	return mig.gceManager.DeleteInstances(refs)
}

// Id returns mig url.
func (mig *Mig) Id() string {
	return GenerateMigUrl(mig.Project, mig.Zone, mig.Name)
}

// Debug returns a debug string for the Mig.
func (mig *Mig) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", mig.Id(), mig.MinSize(), mig.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (mig *Mig) Nodes() ([]string, error) {
	return mig.gceManager.GetMigNodes(mig)
}

func buildMig(value string, gceManager *GceManager) (*Mig, error) {
	tokens := strings.SplitN(value, ":", 3)
	if len(tokens) != 3 {
		return nil, fmt.Errorf("wrong nodes configuration: %s", value)
	}

	mig := Mig{
		gceManager: gceManager,
	}
	if size, err := strconv.Atoi(tokens[0]); err == nil {
		if size <= 0 {
			return nil, fmt.Errorf("min size must be >= 1")
		}
		mig.minSize = size
	} else {
		return nil, fmt.Errorf("failed to set min size: %s, expected integer", tokens[0])
	}

	if size, err := strconv.Atoi(tokens[1]); err == nil {
		if size < mig.minSize {
			return nil, fmt.Errorf("max size must be greater or equal to min size")
		}
		mig.maxSize = size
	} else {
		return nil, fmt.Errorf("failed to set max size: %s, expected integer", tokens[1])
	}

	var err error
	if mig.Project, mig.Zone, mig.Name, err = ParseMigUrl(tokens[2]); err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", tokens[2], err)
	}
	return &mig, nil
}
