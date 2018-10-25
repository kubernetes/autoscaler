package factors

import "time"

type Interface interface {
	Snapshot() (Snapshot, error)
}

// Snapshot is a set of values, which enables batch querying
type Snapshot interface {
	Get(key string) (float64, bool, error)
	Timestamp() time.Time
}
