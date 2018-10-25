package scaling

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
	scalingpolicy "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
)

const internalScale = resource.Milli

func computeValue(fn *scalingpolicy.ResourceScalingFunction, inputs factors.Snapshot, shift float64) (float64, error) {
	var v float64
	if !fn.Base.IsZero() {
		v = float64(fn.Base.ScaledValue(internalScale))
	}

	if fn.Input != "" {
		input, found, err := inputs.Get(fn.Input)
		if err != nil {
			return 0, fmt.Errorf("error reading %q: %v", fn.Input, err)
		}

		if !found {
			glog.Warningf("value %q not found", fn.Input)
			// We still continue, we just apply the base value
		} else if !fn.Slope.IsZero() {
			input += shift

			increment := float64(fn.Slope.ScaledValue(internalScale)) * input
			if fn.Per > 1 {
				increment /= float64(fn.Per)
			}
			v += increment
		}
	}

	return v, nil
}
