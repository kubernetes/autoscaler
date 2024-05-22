package hetzner

import (
	"k8s.io/klog/v2"
)

// DebugWriter is a writer that logs to klog at level 5.
type DebugWriter struct{}

func (d DebugWriter) Write(p []byte) (n int, err error) {
	klog.V(5).Info(string(p))
	return len(p), nil
}
