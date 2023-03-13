package wrapper

import (
	"fmt"
	"strings"
)

// SkippedReasons contains information why given node group was skipped.
type SkippedReasons struct {
	messages []string
}

// Reasons returns a slice of reasons why the node group was not considered for scale up.
func (sr *SkippedReasons) Reasons() []string {
	return sr.messages
}

var (
	// BackoffReason node group is in backoff.
	BackoffReason         = &SkippedReasons{[]string{"in backoff after failed scale-up"}}
	// MaxLimitReachedReason node group reached max size limit.
	MaxLimitReachedReason = &SkippedReasons{[]string{"max node group size reached"}}
	// NotReadyReason node group is not ready.
	NotReadyReason        = &SkippedReasons{[]string{"not ready for scale-up"}}
)

// MaxResourceLimitReached returns a reason describing which cluster wide resource limits were reached.
func MaxResourceLimitReached(resources []string) *SkippedReasons {
	return &SkippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached", strings.Join(resources, ", "))}}
}
