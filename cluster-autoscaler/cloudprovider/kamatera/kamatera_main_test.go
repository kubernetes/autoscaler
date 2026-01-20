package kamatera

import (
	"flag"
	"os"
	"testing"

	"k8s.io/klog/v2"
)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	_ = flag.Set("v", "4")
	flag.Parse()
	os.Exit(m.Run())
}
