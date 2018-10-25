package static

import (
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
	"k8s.io/apimachinery/pkg/util/clock"
)

type staticFactors struct {
	clock  clock.Clock
	values map[string]float64
}

var _ factors.Interface = &staticFactors{}

type staticFactorsSnapshot struct {
	timestamp time.Time
	values    map[string]float64
}

var _ factors.Snapshot = &staticFactorsSnapshot{}

func NewStaticFactors(clock clock.Clock, values map[string]float64) factors.Interface {
	p := &staticFactors{
		values: values,
		clock:  clock,
	}
	return p
}

func (k *staticFactors) Snapshot() (factors.Snapshot, error) {
	return &staticFactorsSnapshot{
		values:    k.values,
		timestamp: k.clock.Now(),
	}, nil
}

func (s *staticFactorsSnapshot) Get(key string) (float64, bool, error) {
	v, found := s.values[key]
	return v, found, nil
}

func (s *staticFactorsSnapshot) Timestamp() time.Time {
	return s.timestamp
}
