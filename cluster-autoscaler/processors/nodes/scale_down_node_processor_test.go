package nodes

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestDefaultScaleDownNodeProcessor(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	n2 := BuildTestNode("n2", 100, 1000)
	ctx := &context.AutoscalingContext{}
	defaultProcessor := NewDefaultScaleDownNodeProcessor()
	expectedNodes := []*apiv1.Node{n1, n2}
	nodes := []*apiv1.Node{n1, n2}
	nodes, err := defaultProcessor.Process(ctx, nodes)
	if err != nil {
		t.Fatalf("Unexpected error; want: %v, got: %v", nil, err)
	}
	if diff := cmp.Diff(nodes, expectedNodes); diff != "" {
		t.Fatalf("Unexoected result; want: %v, got: %v", expectedNodes, nodes)
	}
}
