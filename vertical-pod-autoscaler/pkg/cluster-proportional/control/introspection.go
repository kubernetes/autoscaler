package control

import (
	api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
)

type PolicyInfo struct {
	Policy *api.ClusterProportionalScaler `json:"policy"`
}

func (s *PolicyState) Query() *PolicyInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	info := &PolicyInfo{
		Policy: s.policy,
		//State:  s.evaluator.Query(),
	}
	return info
}

type StateInfo struct {
	Policies map[string]*PolicyInfo `json:"policies"`
}

// Query returns the current state, for reporting e.g. via the /statz endpoint
func (c *State) Query() interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	info := &StateInfo{
		Policies: make(map[string]*PolicyInfo),
	}
	for k, v := range c.policies {
		info.Policies[k.String()] = v.Query()
	}
	return info
}
