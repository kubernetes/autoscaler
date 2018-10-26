package scaling

import (
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
)

type containerScalingRuleEvaluator struct {
	mutex sync.Mutex
	clock clock.Clock
	rule  *api.CPContainerResourcePolicy

	limits   map[v1.ResourceName]*resourceScalingRuleEvaluator
	requests map[v1.ResourceName]*resourceScalingRuleEvaluator
}

func newContainerScalingRuleEvaluator(rule *api.CPContainerResourcePolicy, clock clock.Clock) *containerScalingRuleEvaluator {
	e := &containerScalingRuleEvaluator{
		rule:     rule,
		clock:    clock,
		limits:   make(map[v1.ResourceName]*resourceScalingRuleEvaluator),
		requests: make(map[v1.ResourceName]*resourceScalingRuleEvaluator),
	}

	e.updatePolicy(rule)

	return e
}

func (e *containerScalingRuleEvaluator) updatePolicy(rule *api.CPContainerResourcePolicy) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.updateResourceMap(rule.Limits, e.limits)
	e.updateResourceMap(rule.Requests, e.requests)
}

func (e *containerScalingRuleEvaluator) updateResourceMap(rules []api.ResourceScalingRule, evaluators map[v1.ResourceName]*resourceScalingRuleEvaluator) {
	marked := make(map[v1.ResourceName]bool)
	for i := range rules {
		r := &rules[i]
		re := evaluators[r.Resource]
		if re == nil {
			re = &resourceScalingRuleEvaluator{clock: e.clock}
			evaluators[r.Resource] = re
		}
		re.updatePolicy(r)
		marked[r.Resource] = true
	}
	for k := range evaluators {
		if !marked[k] {
			delete(evaluators, k)
		}
	}
}

// ComputeResources computes a list of resource quantities based on the input state and the specified policy
// It returns a partial PodSpec with the resources we should apply
func (e *containerScalingRuleEvaluator) computeResources(parentPath string) (*api.RecommendedContainerResources, error) {
	container := &api.RecommendedContainerResources{
		ContainerName: e.rule.ContainerName,
	}

	/*
		for k, re := range e.limits {
			current := currentParent.Resources.Limits[k]
			r, err := re.computeResources(parentPath+".limits."+string(k), current)
			if err != nil {
				return nil, err
			}
			if r == nil {
				continue
			}
			if container.Resources.Limits == nil {
				container.Resources.Limits = make(v1.ResourceList)
			}
			container.Resources.Limits[k] = *r
		}
	*/

	for k, re := range e.requests {
		r, err := re.computeResources(parentPath + ".requests." + string(k))
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}
		/*		if container.Resources.Requests == nil {
					container.Resources.Requests = make(v1.ResourceList)
				}
				container.Resources.Requests[k] = *r
		*/

		if container.Target == nil {
			container.Target = make(v1.ResourceList)
		}
		container.Target[k] = *r

		// We set LowerBound to == target, so that we always scale up right away
		if container.LowerBound == nil {
			container.LowerBound = make(v1.ResourceList)
		}
		container.LowerBound[k] = *r

	}

	/*
		if len(container.Resources.Requests) == 0 && len(container.Resources.Limits) == 0 {
			return nil, nil
		}
	*/
	if len(container.Target) == 0 && len(container.UpperBound) == 0 && len(container.LowerBound) == 0 {
		return nil, nil
	}

	return container, nil
}

// AddObservation is called whenever we observe input values
func (e *containerScalingRuleEvaluator) addObservation(inputs factors.Snapshot) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, re := range e.limits {
		re.addObservation(inputs)
	}
	for _, re := range e.requests {
		re.addObservation(inputs)
	}
}
