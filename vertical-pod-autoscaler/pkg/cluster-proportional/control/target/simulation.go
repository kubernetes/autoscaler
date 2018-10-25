package target

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type SimulationTarget struct {
	Current *v1.PodSpec

	ClusterState *ClusterStats

	UpdateCount int
}

var _ Interface = &SimulationTarget{}

func NewSimulationTarget() *SimulationTarget {
	return &SimulationTarget{}
}

func (s *SimulationTarget) Read(kind, namespace, name string) (*v1.PodSpec, error) {
	if s.Current == nil {
		return nil, fmt.Errorf("simulated value not set")
	}
	return s.Current.DeepCopy(), nil
}

func (s *SimulationTarget) UpdateResources(kind, namespace, name string, updates *v1.PodSpec, dryrun bool) error {
	for _, c := range updates.Containers {
		currentContainer := findContainerByName(s.Current.Containers, c.Name)
		if currentContainer == nil {
			glog.Warningf("cannot find container %q", c.Name)
			continue
		}

		for k, r := range c.Resources.Limits {
			if currentContainer.Resources.Limits == nil {
				currentContainer.Resources.Limits = make(v1.ResourceList)
			}
			currentContainer.Resources.Limits[k] = r
		}

		for k, r := range c.Resources.Requests {
			if currentContainer.Resources.Requests == nil {
				currentContainer.Resources.Requests = make(v1.ResourceList)
			}
			currentContainer.Resources.Requests[k] = r
		}
	}
	s.UpdateCount++
	return nil
}

func (s *SimulationTarget) ReadClusterState() (*ClusterStats, error) {
	if s.ClusterState == nil {
		return nil, fmt.Errorf("simulated cluster state not set")
	}
	return s.ClusterState, nil
}

// TODO: Duplicated - move to a util package?
func findContainerByName(containers []v1.Container, name string) *v1.Container {
	for i := range containers {
		c := &containers[i]
		if c.Name == name {
			return c
		}
	}
	return nil
}
