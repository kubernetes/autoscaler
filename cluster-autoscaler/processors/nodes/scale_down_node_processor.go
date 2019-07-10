package nodes

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type ScaleDownNodeProcessor interface {
	Process(*context.AutoscalingContext, []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError)
	CleanUp()
}

type NoOpScaleDownNodeProcessor struct {
}

func (n *NoOpScaleDownNodeProcessor) Process(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

func (n *NoOpScaleDownNodeProcessor) CleanUp() {
}

func NewDefaultScaleDownNodeProcessor() ScaleDownNodeProcessor {
	return &NoOpScaleDownNodeProcessor{}
}
