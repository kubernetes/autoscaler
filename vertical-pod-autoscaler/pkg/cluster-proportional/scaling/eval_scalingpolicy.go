package scaling

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/clock"
	api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
)

type ScalingPolicyEvaluator struct {
	mutex sync.Mutex
	clock clock.Clock
	rule  *api.ClusterProportionalScaler

	containers map[string]*containerScalingRuleEvaluator
}

func NewScalingPolicyEvaluator(clock clock.Clock, rule *api.ClusterProportionalScaler) *ScalingPolicyEvaluator {
	e := &ScalingPolicyEvaluator{
		rule:       rule,
		clock:      clock,
		containers: make(map[string]*containerScalingRuleEvaluator),
	}

	e.UpdatePolicy(rule)

	return e
}

func (e *ScalingPolicyEvaluator) UpdatePolicy(rule *api.ClusterProportionalScaler) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var containers []api.CPContainerResourcePolicy
	if rule.Spec.ResourcePolicy != nil {
		containers = rule.Spec.ResourcePolicy.ContainerPolicies
	}

	marked := make(map[string]bool)
	for i := range containers {
		r := &containers[i]
		ce := e.containers[r.ContainerName]
		if ce == nil {
			ce = newContainerScalingRuleEvaluator(r, e.clock)
			e.containers[r.ContainerName] = ce
		} else {
			ce.updatePolicy(r)
		}
		marked[r.ContainerName] = true
	}
	for k := range e.containers {
		if !marked[k] {
			delete(e.containers, k)
		}
	}
}

// ComputeResources computes a list of resource quantities based on the input state and the specified policy
// It returns a partial PodSpec with the resources we should apply
func (e *ScalingPolicyEvaluator) ComputeResources(parentPath string) (*api.RecommendedPodResources, error) {
	recommendations := &api.RecommendedPodResources{}

	for k, ce := range e.containers {
		c, err := ce.computeResources(parentPath + "[" + k + "]")
		if err != nil {
			return nil, err
		}
		if c != nil {
			recommendations.ContainerRecommendations = append(recommendations.ContainerRecommendations, *c)
		}
	}

	return recommendations, nil
}

// AddObservation is called whenever we observe input values
func (e *ScalingPolicyEvaluator) AddObservation(inputs factors.Snapshot) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, ce := range e.containers {
		ce.addObservation(inputs)
	}
}
