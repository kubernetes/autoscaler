package scaling

import (
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/clock"
	scalingpolicy "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
)

// resourceScalingRuleEvaluator holds the state for evaluation of a ResourceScalingRule
type resourceScalingRuleEvaluator struct {
	mutex sync.Mutex
	clock clock.Clock

	// rule holds a copy of the current rule
	policy *scalingpolicy.ResourceScalingRule

	// target holds the window of "raw" target values
	target windowValues

	// scaleDownThresholds holds the scale-down threshold values (computed by adding some padding to the input resource value)
	scaleDownThresholds windowValues

	// lastScaleDown is the time of the last scale-down, to prevent rapid repeated scale-down
	lastScaleDown time.Time
}

// updatePolicy updates for a change in the resource scaling policy API
func (e *resourceScalingRuleEvaluator) updatePolicy(policy *scalingpolicy.ResourceScalingRule) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if reflect.DeepEqual(policy, e.policy) {
		return
	}

	scaleDownAtWindowRetention := time.Duration(0)
	//if policy.Function.DelayScaleDown != nil {
	//	scaleDownAtWindowRetention = time.Duration(policy.Function.DelayScaleDown.DelaySeconds) * time.Second
	//}

	maxRetention := time.Duration(0)
	if maxRetention < scaleDownAtWindowRetention {
		maxRetention = scaleDownAtWindowRetention
	}

	// We retain the values for the biggest retention, because e.g. the scale-down-after-delay
	// uses the delay from the scale-down policy but the max value from the unshifted target values
	e.target.Reset(e.clock, maxRetention)
	e.scaleDownThresholds.Reset(e.clock, maxRetention)

	e.policy = policy.DeepCopy()
}

// computeResources computes the new resource value we should use, or returns nil if no change is needed
func (e *resourceScalingRuleEvaluator) computeResources(parentPath string) (*resource.Quantity, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	now := e.clock.Now()

	latestStats := e.target.stats(now, time.Duration(0))
	if !latestStats.HasLatest {
		glog.Infof("No data points, won't consider scaling %s", parentPath)
		return nil, nil
	}

	return e.toResourceQuantity(latestStats.LatestValue), nil

	/*
		currentV := float64(current.ScaledValue(internalScale))

		// TODO: We should have the same delay options for scale up

		// We scale up to the current value, with no delay - i.e. whenever the current value is LTE our value
		if currentV < latestStats.LatestValue {
			glog.Infof("Will scale up to target value for %s", parentPath)
			return e.toResourceQuantity(latestStats.LatestValue), nil
		}

		// We scale down immediately to the current value when we exceed the shiftedValue
		// i.e. we are "too far away"

			if e.policy.Function.DelayScaleDown != nil && e.policy.Function.DelayScaleDown.Max != 0 {
				latestScaleDownStats := e.scaleDownThresholds.stats(now, time.Duration(0))
				if latestScaleDownStats.HasLatest && currentV > latestScaleDownStats.LatestValue {
					glog.Infof("Current value broke threshold for scale-down; scaling down %s", parentPath)

					// Note that we scale down to the unshifted target value - we're just delaying the scale-down here
					// until it is of a big enough magnitude to be worth the cost
					return e.toResourceQuantity(latestStats.LatestValue), nil
				} else {
					glog.Infof("current value is not over scale-down threshold; won't scale down %s", parentPath)
				}
			}


		// We also scale down after e.g. 10 minutes, so we aren't permanently above-target.
		// The challenge is how to do so without introducing flapping when the target
		// value is gradually decreasing.
		//
		// We keep track of the last scaling event, and scale down to the maximum target value
		// we've observed in that window.  We also won't even consider doing so more often than
		// every e.g. 10 minutes.

			if e.policy.Function.DelayScaleDown != nil && e.policy.Function.DelayScaleDown.DelaySeconds != 0 {
				delay := time.Second * time.Duration(e.policy.Function.DelayScaleDown.DelaySeconds)
				// Don't scale more often than delay
				if now.Sub(e.lastScaleDown) > delay {
					// Ensure we have enough history
					if now.Sub(e.target.Start) > delay {
						windowStats := e.target.stats(now, delay)
						if windowStats.N != 0 && currentV > windowStats.Max {
							glog.Infof("Scale-down time-window exceeded, scaling down %s", parentPath)
							return e.toResourceQuantity(windowStats.Max), nil
						}
					}
				}
			}


		// If there is no DelayScaleDown, we scale down immediately
		 if e.policy.Function.DelayScaleDown == nil ||
		// TODO: IsZero()
		(e.policy.Function.DelayScaleDown.DelaySeconds == 0 && e.policy.Function.DelayScaleDown.Max == 0) {
			if currentV > latestStats.LatestValue {
				glog.Infof("Will scale down to target value for %s", parentPath)
				return e.toResourceQuantity(latestStats.LatestValue), nil
			}
		}
	*/

	return nil, nil
}

// addObservation is called whenever we observe input values
func (e *resourceScalingRuleEvaluator) addObservation(inputs factors.Snapshot) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	{
		v, err := computeValue(&e.policy.Function, inputs, 0)
		if err != nil {
			glog.Warningf("error computing shifted value: %v", err)
		} else {
			e.target.addObservation(inputs.Timestamp(), v)
		}
	}

	/*
		if e.policy.Function.DelayScaleDown != nil {
			if e.policy.Function.DelayScaleDown.Max != 0 {
				v, err := computeValue(&e.policy.Function, inputs, e.policy.Function.DelayScaleDown.Max)
				if err != nil {
					glog.Warningf("error computing scale-down threshold value: %v", err)
				} else {
					e.scaleDownThresholds.addObservation(inputs.Timestamp(), v)
				}
			}
		}
	*/

}

func (e *resourceScalingRuleEvaluator) toResourceQuantity(v float64) *resource.Quantity {
	q := resource.NewScaledQuantity(int64(v), internalScale)
	q.Format = e.policy.Function.Base.Format
	if q.Format == "" {
		q.Format = e.policy.Function.Slope.Format
	}
	return q
}
