package control

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/control/target"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/options"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/scaling"
)

// PolicyState is the state around a single scaling policy
type PolicyState struct {
	client  versioned.Interface
	target  target.Interface
	options *options.AutoScalerConfig

	mutex  sync.Mutex
	parent *State
	policy *api.ClusterProportionalScaler

	evaluator *scaling.ScalingPolicyEvaluator
}

func NewPolicyState(parent *State, policy *api.ClusterProportionalScaler) *PolicyState {
	s := &PolicyState{
		client:  parent.client,
		target:  parent.target,
		options: parent.options,
		parent:  parent,
		policy:  policy,
	}

	s.evaluator = scaling.NewScalingPolicyEvaluator(parent.clock, policy)

	return s
}

func (s *PolicyState) updatePolicy(o *api.ClusterProportionalScaler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.policy = o
	s.evaluator.UpdatePolicy(o)
}

// addObservation is called whenever we observe a set of input values
func (s *PolicyState) addObservation(snapshot factors.Snapshot) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	policy := s.policy

	namespace := policy.Namespace
	name := policy.Name

	path := fmt.Sprintf("%s/%s", namespace, name)

	glog.V(4).Infof("adding observation for %s", path)

	s.evaluator.AddObservation(snapshot)
}

func (s *PolicyState) updateValues() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	policy := s.policy

	namespace := policy.Namespace
	name := policy.Name

	path := fmt.Sprintf("%s/%s", namespace, name)

	changes, err := s.evaluator.ComputeResources(path)
	if err != nil {
		return err
	}

	if changes != nil && !equalRecommendations(changes, policy.Status.Recommendation) {
		// TODO: Just do a status update??
		u := policy.DeepCopy()
		u.Status.Recommendation = changes
		if _, err := s.client.AutoscalingV1beta1().ClusterProportionalScalers(namespace).Update(u); err != nil {
			glog.Warningf("failed to update %q: %v", path, err)
		} else {
			glog.V(4).Infof("applied update to %s", path)
		}
	} else {
		glog.V(4).Infof("no change needed for %s", path)
	}

	return nil
}

func equalRecommendations(l, r *api.RecommendedPodResources) bool {
	if l == nil {
		return r == nil
	}
	if r == nil {
		return l == nil
	}

	if len(l.ContainerRecommendations) != len(r.ContainerRecommendations) {
		return false
	}

	// We enforce order
	for i := range l.ContainerRecommendations {
		if l.ContainerRecommendations[i].ContainerName != r.ContainerRecommendations[i].ContainerName {
			return false
		}

		if !resourceListEquals(l.ContainerRecommendations[i].Target, r.ContainerRecommendations[i].Target) {
			return false
		}

		if !resourceListEquals(l.ContainerRecommendations[i].UpperBound, r.ContainerRecommendations[i].UpperBound) {
			return false
		}
	}

	return true
}

func resourceListEquals(l, r v1.ResourceList) bool {
	if len(l) != len(r) {
		return false
	}

	for k, lv := range l {
		rv, found := r[k]
		if !found {
			return false
		}
		if lv.Cmp(rv) != 0 {
			return false
		}
	}
	return true
}
